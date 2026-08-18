package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"records-manager/backend/internal/games"
	"records-manager/backend/internal/games/repository"
)

type GameService struct {
	repo       repository.GameRepository
	rawgAPIKey string
	uploadsDir string
}

func NewGameService(repo repository.GameRepository, rawgAPIKey string, uploadsDir string) *GameService {
	return &GameService{
		repo:       repo,
		rawgAPIKey: rawgAPIKey,
		uploadsDir: uploadsDir,
	}
}

// === Types RAWG ===

type RAWGSearchResponse struct {
	Results []RAWGGameResult `json:"results"`
}

type RAWGGenre struct {
	Name string `json:"name"`
}

type RAWGPlatformInfo struct {
	Name string `json:"name"`
}

type RAWGPlatformEntry struct {
	Platform RAWGPlatformInfo `json:"platform"`
}

type RAWGPublisher struct {
	Name string `json:"name"`
}

type RAWGGameResult struct {
	ID              int                 `json:"id"`
	Name            string              `json:"name"`
	Released        string              `json:"released"`
	BackgroundImage string              `json:"background_image"`
	Genres          []RAWGGenre         `json:"genres"`
	Platforms       []RAWGPlatformEntry `json:"platforms"`
}

type RAWGGameDetails struct {
	ID              int             `json:"id"`
	Name            string          `json:"name"`
	Released        string          `json:"released"`
	BackgroundImage string          `json:"background_image"`
	Genres          []RAWGGenre     `json:"genres"`
	Publishers      []RAWGPublisher `json:"publishers"`
}

// GamePreview reflète les informations récupérées via RAWG pour un jeu,
// avant que l'utilisateur ne les valide dans le formulaire — même rôle que
// CoverPreview côté disques/Discogs.
type GamePreview struct {
	CoverURL  string           `json:"cover_url"`
	Title     string           `json:"title"`
	Year      string           `json:"year"`
	Genres    []string         `json:"genres"`
	Platforms []string         `json:"platforms"`
	Publisher string           `json:"publisher"`
	Found     bool             `json:"found"`
	Results   []RAWGGameResult `json:"results,omitempty"`
	RAWGID    int64            `json:"rawg_id,omitempty"`
}

func releaseYearFromDate(dateStr string) *int {
	if dateStr == "" {
		return nil
	}
	parts := strings.SplitN(dateStr, "-", 2)
	year, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil
	}
	return &year
}

// searchRAWGWithResults recherche des jeux par titre sur RAWG et récupère
// les détails (dont l'éditeur, absent de la réponse de recherche) du
// premier résultat.
func (s *GameService) searchRAWGWithResults(title string) (*GamePreview, error) {
	if s.rawgAPIKey == "" {
		return nil, fmt.Errorf("clé RAWG_API_KEY non configurée sur le serveur")
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return &GamePreview{Found: false}, nil
	}

	url := fmt.Sprintf("https://api.rawg.io/api/games?key=%s&search=%s&page_size=20",
		s.rawgAPIKey, strings.ReplaceAll(title, " ", "%20"))

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erreur API RAWG: %d", resp.StatusCode)
	}

	var searchResp RAWGSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, err
	}

	if len(searchResp.Results) == 0 {
		return &GamePreview{Found: false}, nil
	}

	first := searchResp.Results[0]
	detailed, err := s.getRAWGGameDetails(int64(first.ID))
	if err == nil && detailed != nil {
		detailed.Results = searchResp.Results
		return detailed, nil
	}

	// Repli si la récupération des détails échoue : on construit quand même
	// une preview à partir des données de la recherche (sans éditeur).
	var genreNames []string
	for _, g := range first.Genres {
		genreNames = append(genreNames, g.Name)
	}
	var platformNames []string
	for _, p := range first.Platforms {
		platformNames = append(platformNames, p.Platform.Name)
	}
	return &GamePreview{
		CoverURL:  first.BackgroundImage,
		Title:     first.Name,
		Year:      first.Released,
		Genres:    genreNames,
		Platforms: platformNames,
		Found:     true,
		Results:   searchResp.Results,
		RAWGID:    int64(first.ID),
	}, nil
}

// getRAWGGameDetails récupère la fiche complète d'un jeu (dont l'éditeur).
func (s *GameService) getRAWGGameDetails(rawgID int64) (*GamePreview, error) {
	if s.rawgAPIKey == "" {
		return nil, fmt.Errorf("clé RAWG manquante")
	}
	url := fmt.Sprintf("https://api.rawg.io/api/games/%d?key=%s", rawgID, s.rawgAPIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erreur API RAWG (détails): %d", resp.StatusCode)
	}

	var details RAWGGameDetails
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return nil, err
	}

	var genreNames []string
	for _, g := range details.Genres {
		genreNames = append(genreNames, g.Name)
	}
	publisher := ""
	if len(details.Publishers) > 0 {
		publisher = details.Publishers[0].Name
	}

	return &GamePreview{
		CoverURL:  details.BackgroundImage,
		Title:     details.Name,
		Year:      details.Released,
		Genres:    genreNames,
		Publisher: publisher,
		Found:     true,
		RAWGID:    int64(details.ID),
	}, nil
}

// PreviewGame - recherche par titre exposée au handler (équivalent SearchRAWG)
func (s *GameService) PreviewGame(title string) (*GamePreview, error) {
	return s.searchRAWGWithResults(title)
}

// SelectRAWGResult - récupère les détails d'un résultat choisi par l'utilisateur
func (s *GameService) SelectRAWGResult(rawgID int64) (*GamePreview, error) {
	return s.getRAWGGameDetails(rawgID)
}

// === Téléchargement de jaquette ===

func sanitizeFilename(name string) string {
	cleaned := name
	replacements := map[string]string{
		"à": "a", "á": "a", "â": "a", "ã": "a", "ä": "a", "å": "a",
		"è": "e", "é": "e", "ê": "e", "ë": "e",
		"ì": "i", "í": "i", "î": "i", "ï": "i",
		"ò": "o", "ó": "o", "ô": "o", "õ": "o", "ö": "o", "ø": "o",
		"ù": "u", "ú": "u", "û": "u", "ü": "u",
		"ý": "y", "ÿ": "y",
		"ñ": "n", "ç": "c",
		"À": "A", "Á": "A", "Â": "A", "Ã": "A", "Ä": "A", "Å": "A",
		"È": "E", "É": "E", "Ê": "E", "Ë": "E",
		"Ì": "I", "Í": "I", "Î": "I", "Ï": "I",
		"Ò": "O", "Ó": "O", "Ô": "O", "Õ": "O", "Ö": "O", "Ø": "O",
		"Ù": "U", "Ú": "U", "Û": "U", "Ü": "U",
		"Ý": "Y", "Ÿ": "Y",
		"Ñ": "N", "Ç": "C",
		"æ": "ae", "œ": "oe",
		"Æ": "AE", "Œ": "OE",
	}
	for old, new := range replacements {
		cleaned = strings.ReplaceAll(cleaned, old, new)
	}
	cleaned = strings.ReplaceAll(cleaned, "/", "-")
	cleaned = strings.ReplaceAll(cleaned, "\\", "-")
	cleaned = strings.ReplaceAll(cleaned, ":", "-")
	cleaned = strings.ReplaceAll(cleaned, "*", "")
	cleaned = strings.ReplaceAll(cleaned, "?", "")
	cleaned = strings.ReplaceAll(cleaned, "\"", "")
	cleaned = strings.ReplaceAll(cleaned, "<", "")
	cleaned = strings.ReplaceAll(cleaned, ">", "")
	cleaned = strings.ReplaceAll(cleaned, "|", "")
	cleaned = strings.ReplaceAll(cleaned, "&", "and")
	cleaned = strings.ReplaceAll(cleaned, "'", "")
	cleaned = strings.ReplaceAll(cleaned, "`", "")
	cleaned = strings.Join(strings.Fields(cleaned), " ")
	cleaned = strings.ReplaceAll(cleaned, " ", "_")
	return strings.TrimSpace(cleaned)
}

func (s *GameService) downloadCoverWithName(imageURL, title string, gameID int) (string, error) {
	if imageURL == "" {
		return "", nil
	}
	coversDir := filepath.Join(s.uploadsDir, "covers")
	if err := os.MkdirAll(coversDir, 0755); err != nil {
		return "", fmt.Errorf("erreur création dossier covers: %w", err)
	}
	finalTitle := title
	if finalTitle == "" {
		finalTitle = "Unknown"
	}
	cleanTitle := sanitizeFilename(finalTitle)
	if len(cleanTitle) > 60 {
		cleanTitle = cleanTitle[:60]
	}
	ext := ".jpg"
	if strings.Contains(imageURL, ".png") {
		ext = ".png"
	} else if strings.Contains(imageURL, ".webp") {
		ext = ".webp"
	}
	filename := fmt.Sprintf("%s (game-%d)%s", cleanTitle, gameID, ext)
	filePath := filepath.Join(coversDir, filename)
	if _, err := os.Stat(filePath); err == nil {
		return "/uploads/covers/" + filename, nil
	}

	resp, err := http.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("erreur téléchargement image: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("erreur HTTP %d lors du téléchargement", resp.StatusCode)
	}

	out, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("erreur création fichier: %w", err)
	}
	defer out.Close()
	if _, err := io.Copy(out, resp.Body); err != nil {
		return "", fmt.Errorf("erreur écriture fichier: %w", err)
	}

	return "/uploads/covers/" + filename, nil
}

// === CRUD ===

func (s *GameService) CreateGame(
	ctx context.Context,
	title string,
	platformID int,
	genreID *int,
	publisherID *int,
	releaseYear *int,
	barcode *string,
	notes *string,
	price *float64,
	quantity *int,
	coverImage *string,
	rawgID *int64,
) (*games.GameWithDetails, error) {
	game := &games.Game{
		Title:       title,
		PlatformID:  platformID,
		GenreID:     genreID,
		PublisherID: publisherID,
		ReleaseYear: releaseYear,
		Barcode:     barcode,
		Notes:       notes,
		Price:       price,
		Quantity:    quantity,
		RAWGID:      rawgID,
		CoverURL:    coverImage,
	}

	if err := s.repo.Create(ctx, game); err != nil {
		return nil, err
	}

	created, err := s.repo.FindByID(ctx, *game.ID)
	if err != nil {
		return nil, fmt.Errorf("échec de la récupération après création : %w", err)
	}
	if created == nil {
		return nil, fmt.Errorf("jeu créé mais introuvable")
	}

	// Le téléchargement de la jaquette (réseau externe vers RAWG) part en
	// arrière-plan pour ne pas ralentir l'enregistrement — même optimisation
	// que pour les disques (voir disc_service.go, CreateDisc).
	if coverImage != nil && strings.HasPrefix(*coverImage, "http") {
		gameID := created.ID
		gameTitle := created.Title
		imageURL := *coverImage
		go func() {
			bgCtx := context.Background()
			coverPath, err := s.downloadCoverWithName(imageURL, gameTitle, gameID)
			if err != nil || coverPath == "" {
				fmt.Printf("❌ Erreur téléchargement jaquette jeu: %v\n", err)
				return
			}
			updateGame := &games.Game{
				ID:          &gameID,
				Title:       created.Title,
				PlatformID:  created.PlatformID,
				GenreID:     created.GenreID,
				PublisherID: created.PublisherID,
				ReleaseYear: created.ReleaseYear,
				Barcode:     created.Barcode,
				CoverURL:    &coverPath,
				Notes:       created.Notes,
				Price:       created.Price,
				Quantity:    created.Quantity,
			}
			if err := s.repo.Update(bgCtx, updateGame); err != nil {
				fmt.Printf("❌ Erreur mise à jour cover_url jeu: %v\n", err)
			}
		}()
	}

	return created, nil
}

func (s *GameService) GetAllGames(ctx context.Context) ([]games.GameWithDetails, error) {
	return s.repo.FindAll(ctx)
}

func (s *GameService) GetGameByID(ctx context.Context, id int) (*games.GameWithDetails, error) {
	return s.repo.FindByID(ctx, id)
}

// removeCoverFile supprime du disque le fichier pointé par coverURL, s'il
// s'agit bien d'un chemin local uploadé (jamais une URL RAWG externe).
func (s *GameService) removeCoverFile(coverURL *string) {
	if coverURL == nil || *coverURL == "" || !strings.HasPrefix(*coverURL, "/uploads/") {
		return
	}
	diskPath := filepath.Join(s.uploadsDir, strings.TrimPrefix(*coverURL, "/uploads/"))
	if err := os.Remove(diskPath); err != nil && !os.IsNotExist(err) {
		fmt.Printf("⚠️ Erreur suppression jaquette jeu: %v\n", err)
	}
}

func (s *GameService) UpdateGame(
	ctx context.Context,
	id int,
	title string,
	platformID int,
	genreID *int,
	publisherID *int,
	releaseYear *int,
	barcode *string,
	notes *string,
	price *float64,
	quantity *int,
	coverImage *string,
) (*games.GameWithDetails, error) {
	currentGame, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if currentGame == nil {
		return nil, fmt.Errorf("jeu introuvable")
	}

	// Même logique que UpdateDisc (disc_service.go) : nil = pas de
	// changement demandé, "" = suppression explicite, URL http = nouvelle
	// recherche web à télécharger, sinon = chemin local déjà uploadé via
	// /api/upload-cover. Dans tous les cas où la jaquette change, l'ancien
	// fichier local est nettoyé pour ne pas laisser d'orphelins sur disque.
	var finalCoverURL *string
	switch {
	case coverImage == nil:
		finalCoverURL = currentGame.CoverURL
	case *coverImage == "":
		s.removeCoverFile(currentGame.CoverURL)
		finalCoverURL = nil
	case strings.HasPrefix(*coverImage, "http"):
		coverPath, err := s.downloadCoverWithName(*coverImage, title, id)
		if err != nil || coverPath == "" {
			fmt.Printf("❌ Échec téléchargement jaquette jeu: %v\n", err)
			finalCoverURL = currentGame.CoverURL
		} else {
			s.removeCoverFile(currentGame.CoverURL)
			finalCoverURL = &coverPath
		}
	default:
		if currentGame.CoverURL == nil || *currentGame.CoverURL != *coverImage {
			s.removeCoverFile(currentGame.CoverURL)
		}
		finalCoverURL = coverImage
	}

	idCopy := id
	game := &games.Game{
		ID:          &idCopy,
		Title:       title,
		PlatformID:  platformID,
		GenreID:     genreID,
		PublisherID: publisherID,
		ReleaseYear: releaseYear,
		Barcode:     barcode,
		Notes:       notes,
		Price:       price,
		Quantity:    quantity,
		CoverURL:    finalCoverURL,
	}

	if err := s.repo.Update(ctx, game); err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, id)
}

func (s *GameService) DeleteGame(ctx context.Context, id int) error {
	game, err := s.repo.FindByID(ctx, id)
	if err == nil && game != nil {
		s.removeCoverFile(game.CoverURL)
	}
	return s.repo.Delete(ctx, id)
}

func (s *GameService) CheckBarcodeExists(ctx context.Context, barcode string, excludeID *int) (bool, *games.GameWithDetails, error) {
	if barcode == "" {
		return false, nil, nil
	}
	game, err := s.repo.FindByBarcode(ctx, barcode)
	if err != nil {
		return false, nil, err
	}
	if game == nil {
		return false, nil, nil
	}
	if excludeID != nil && game.ID == *excludeID {
		return false, nil, nil
	}
	return true, game, nil
}
