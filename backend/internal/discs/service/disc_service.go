package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"records-manager/backend/internal/discs"
	"records-manager/backend/internal/discs/repository"
	"records-manager/backend/internal/tracks"
	tracksrepo "records-manager/backend/internal/tracks/repository"
)

type DiscService struct {
	repo         repository.DiscRepository
	discogsToken string
	uploadsDir   string
	trackRepo    tracksrepo.TrackRepository
}

func NewDiscService(repo repository.DiscRepository, discogsToken string, uploadsDir string, trackRepo tracksrepo.TrackRepository) *DiscService {
	return &DiscService{
		repo:         repo,
		discogsToken: discogsToken,
		uploadsDir:   uploadsDir,
		trackRepo:    trackRepo,
	}
}

type DiscogsSearchResponse struct {
	Results []DiscogsResult `json:"results"`
}

type DiscogsResult struct {
	ID       int      `json:"id"`
	Title    string   `json:"title"`
	CoverURL string   `json:"cover_image"`
	Thumb    string   `json:"thumb"`
	Year     string   `json:"year"`
	Country  string   `json:"country"`
	Formats  []Format `json:"formats"`
	Genres   []string `json:"genres"`
	Labels   []Label  `json:"labels"`
}

type Format struct {
	Name         string   `json:"name"`
	Quantity     string   `json:"qty"`
	Descriptions []string `json:"descriptions"`
}

type Label struct {
	Name string `json:"name"`
}

type DiscogsPrice struct {
	LowestPrice  float64 `json:"lowest_price,omitempty"`
	MedianPrice  float64 `json:"median_price,omitempty"`
	HighestPrice float64 `json:"highest_price,omitempty"`
}

type TrackItem struct {
	Position string `json:"position"`
	Title    string `json:"title"`
	Duration string `json:"duration"`
}

type CoverPreview struct {
	CoverURL string          `json:"cover_url"`
	Title    string          `json:"title"`
	Artist   string          `json:"artist"`
	Year     string          `json:"year"`
	Genres   []string        `json:"genres"`
	Formats  []string        `json:"formats"`
	Country  string          `json:"country"`
	Label    string          `json:"label"`
	Found    bool            `json:"found"`
	Results  []DiscogsResult `json:"results,omitempty"`
	Prices   *DiscogsPrice   `json:"prices,omitempty"`
	Tracks   []TrackItem     `json:"tracklist,omitempty"`
}

// === Méthodes existantes inchangées jusqu’à getDiscogsPrices ===

func (s *DiscService) CheckBarcodeExists(ctx context.Context, barcode string, excludeID *int) (bool, *discs.DiscWithDetails, error) {
	if barcode == "" {
		return false, nil, nil
	}
	disc, err := s.repo.FindByBarcodeWithDetails(ctx, barcode)
	if err != nil {
		return false, nil, err
	}
	if disc == nil {
		fmt.Printf("ℹ️  Code-barres %s non trouvé\n", barcode)
		return false, nil, nil
	}
	if excludeID != nil && disc.ID == *excludeID {
		fmt.Printf("ℹ️  Code-barres %s exclu (même ID: %d)\n", barcode, disc.ID)
		return false, nil, nil
	}
	fmt.Printf("✅ Code-barres %s trouvé: ID=%d, titre=%s\n", barcode, disc.ID, disc.Title)
	return true, disc, nil
}

func (s *DiscService) downloadCover(imageURL string, discID int) (string, error) {
	if imageURL == "" {
		return "", nil
	}
	coversDir := filepath.Join(s.uploadsDir, "covers")
	if err := os.MkdirAll(coversDir, 0755); err != nil {
		return "", fmt.Errorf("erreur création dossier covers: %w", err)
	}
	ext := ".jpg"
	if strings.Contains(imageURL, ".png") {
		ext = ".png"
	}
	filename := fmt.Sprintf("cover_%d%s", discID, ext)
	filePath := filepath.Join(coversDir, filename)
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
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("erreur écriture fichier: %w", err)
	}
	return "/uploads/covers/" + filename, nil
}

func (s *DiscService) downloadCoverWithName(imageURL, title, artistName string, discID int) (string, error) {
	if imageURL == "" {
		fmt.Println("⚠️ downloadCoverWithName: imageURL vide")
		return "", nil
	}
	fmt.Printf("📸 downloadCoverWithName appelée - URL: %s, Titre: %s, Artiste: %s, ID: %d\n",
		imageURL, title, artistName, discID)
	coversDir := filepath.Join(s.uploadsDir, "covers")
	fmt.Printf("📁 Dossier covers: %s\n", coversDir)
	if err := os.MkdirAll(coversDir, 0755); err != nil {
		return "", fmt.Errorf("erreur création dossier covers: %w", err)
	}
	finalTitle := title
	finalArtist := artistName
	if finalArtist == "" {
		fmt.Printf("⚠️ Artiste vide, utilisation de 'Unknown'\n")
		finalArtist = "Unknown"
	}
	if finalTitle == "" {
		finalTitle = "Unknown"
	}
	cleanTitle := sanitizeFilename(finalTitle)
	cleanArtist := sanitizeFilename(finalArtist)
	if len(cleanTitle) > 50 {
		cleanTitle = cleanTitle[:50]
	}
	if len(cleanArtist) > 30 {
		cleanArtist = cleanArtist[:30]
	}
	ext := ".jpg"
	if strings.Contains(imageURL, ".png") {
		ext = ".png"
	} else if strings.Contains(imageURL, ".webp") {
		ext = ".webp"
	}
	filename := fmt.Sprintf("%s - %s (%d)%s", cleanArtist, cleanTitle, discID, ext)
	filePath := filepath.Join(coversDir, filename)
	fmt.Printf("📸 Téléchargement pochette: %s\n", filename)
	fmt.Printf("📂 Chemin complet: %s\n", filePath)
	if _, err := os.Stat(filePath); err == nil {
		fmt.Printf("✅ Fichier existe déjà, réutilisation: %s\n", filePath)
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
	fmt.Printf("✅ Image téléchargée avec succès (HTTP %d)\n", resp.StatusCode)
	out, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("erreur création fichier: %w", err)
	}
	defer out.Close()
	bytesWritten, err := io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("erreur écriture fichier: %w", err)
	}
	fmt.Printf("✅ Fichier créé avec succès (%d bytes)\n", bytesWritten)
	relativePath := "/uploads/covers/" + filename
	fmt.Printf("✅ Chemin relatif retourné: %s\n", relativePath)
	return relativePath, nil
}

func (s *DiscService) saveUploadedCover(imageData []byte, title, artistName string, discID int) (string, error) {
	fmt.Printf("💾 saveUploadedCover: Titre=%s, Artiste=%s, ID=%d\n", title, artistName, discID)
	coversDir := filepath.Join(s.uploadsDir, "covers")
	if err := os.MkdirAll(coversDir, 0755); err != nil {
		return "", fmt.Errorf("erreur création dossier covers: %w", err)
	}
	finalTitle := title
	finalArtist := artistName
	if finalArtist == "" {
		finalArtist = "Unknown"
	}
	if finalTitle == "" {
		finalTitle = "Unknown"
	}
	cleanTitle := sanitizeFilename(finalTitle)
	cleanArtist := sanitizeFilename(finalArtist)
	if len(cleanTitle) > 50 {
		cleanTitle = cleanTitle[:50]
	}
	if len(cleanArtist) > 30 {
		cleanArtist = cleanArtist[:30]
	}
	ext := ".jpg"
	if len(imageData) > 8 {
		if imageData[0] == 0x89 && imageData[1] == 0x50 && imageData[2] == 0x4E && imageData[3] == 0x47 {
			ext = ".png"
		} else if len(imageData) > 12 && string(imageData[0:4]) == "RIFF" && string(imageData[8:12]) == "WEBP" {
			ext = ".webp"
		}
	}
	filename := fmt.Sprintf("%s - %s (%d)%s", cleanArtist, cleanTitle, discID, ext)
	filePath := filepath.Join(coversDir, filename)
	fmt.Printf("💾 Sauvegarde fichier uploadé: %s\n", filePath)
	if _, err := os.Stat(filePath); err == nil {
		fmt.Printf("✅ Fichier existe déjà, réutilisation: %s\n", filePath)
	} else {
		if err := os.WriteFile(filePath, imageData, 0644); err != nil {
			return "", fmt.Errorf("erreur écriture fichier: %w", err)
		}
		fmt.Printf("✅ Fichier créé avec succès\n")
	}
	return "/uploads/covers/" + filename, nil
}

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
	cleaned = strings.ReplaceAll(cleaned, "\u2018", "")
	cleaned = strings.ReplaceAll(cleaned, "\u2019", "")
	cleaned = strings.ReplaceAll(cleaned, "\u201C", "")
	cleaned = strings.ReplaceAll(cleaned, "\u201D", "")
	cleaned = strings.Join(strings.Fields(cleaned), " ")
	cleaned = strings.ReplaceAll(cleaned, " ", "_")
	for strings.Contains(cleaned, "..") {
		cleaned = strings.ReplaceAll(cleaned, "..", ".")
	}
	return strings.TrimSpace(cleaned)
}

func (s *DiscService) searchDiscogs(title, artist string) (string, error) {
	if s.discogsToken == "" {
		return "", nil
	}
	title = strings.TrimSpace(title)
	artist = strings.TrimSpace(artist)

	// ✅ Construction avec paramètres dédiés pour meilleure pertinence
	url := fmt.Sprintf("https://api.discogs.com/database/search?type=release&format=Vinyl&token=%s", s.discogsToken)
	if title != "" {
		url += fmt.Sprintf("&release_title=%s", strings.ReplaceAll(title, " ", "+"))
	}
	if artist != "" {
		url += fmt.Sprintf("&artist=%s", strings.ReplaceAll(artist, " ", "+"))
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "RecordsManager/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("erreur API Discogs: %d", resp.StatusCode)
	}
	var searchResp DiscogsSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return "", err
	}
	if len(searchResp.Results) > 0 && searchResp.Results[0].CoverURL != "" {
		return searchResp.Results[0].CoverURL, nil
	}
	return "", nil
}

// findReleaseIDByTitleArtist cherche sur Discogs une release correspondant au
// titre/artiste et renvoie l'ID du premier résultat — même requête que
// searchDiscogs, mais on a besoin de l'ID (pour la tracklist) plutôt que de
// l'URL de pochette.
func (s *DiscService) findReleaseIDByTitleArtist(title, artist string) (int64, error) {
	if s.discogsToken == "" {
		return 0, fmt.Errorf("token Discogs manquant")
	}
	title = strings.TrimSpace(title)
	artist = strings.TrimSpace(artist)

	url := fmt.Sprintf("https://api.discogs.com/database/search?type=release&format=Vinyl&token=%s", s.discogsToken)
	if title != "" {
		url += fmt.Sprintf("&release_title=%s", strings.ReplaceAll(title, " ", "+"))
	}
	if artist != "" {
		url += fmt.Sprintf("&artist=%s", strings.ReplaceAll(artist, " ", "+"))
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", "RecordsManager/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("erreur API Discogs: %d", resp.StatusCode)
	}
	var searchResp DiscogsSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return 0, err
	}
	if len(searchResp.Results) == 0 {
		return 0, fmt.Errorf("aucune release Discogs trouvée")
	}
	return int64(searchResp.Results[0].ID), nil
}

// GetTracksForDisc renvoie les pistes déjà stockées pour un disque (peut
// être vide si jamais récupérées).
func (s *DiscService) GetTracksForDisc(ctx context.Context, discID int) ([]tracks.Track, error) {
	return s.trackRepo.FindByVinylID(ctx, discID)
}

// UpdateTracksForDisc remplace intégralement la tracklist d'un disque —
// utilisé par l'édition manuelle (ajout/suppression/réordonnancement de
// pistes) : on efface l'existant puis on réinsère la liste fournie, pour
// garder track_order cohérent avec l'ordre reçu.
func (s *DiscService) UpdateTracksForDisc(ctx context.Context, discID int, items []TrackItem) ([]tracks.Track, error) {
	if err := s.trackRepo.DeleteByVinylID(ctx, discID); err != nil {
		return nil, err
	}
	if len(items) > 0 {
		trackList := make([]tracks.Track, len(items))
		for i, t := range items {
			trackList[i] = tracks.Track{Position: t.Position, Title: t.Title, Duration: t.Duration}
		}
		if err := s.trackRepo.CreateBatch(ctx, discID, trackList); err != nil {
			return nil, err
		}
	}
	return s.trackRepo.FindByVinylID(ctx, discID)
}

// FetchTracklistForDisc récupère et stocke la tracklist d'un disque à la
// demande — filet de rattrapage pour les disques ajoutés avant cette
// fonctionnalité (pas de discogs_release_id connu, donc re-recherche par
// titre/artiste, avec le même risque de mauvais pressage que la recherche
// de pochette existante).
//
// force=false (usage normal) : ne fait rien si une tracklist existe déjà,
// renvoie l'existante telle quelle.
// force=true (rafraîchissement demandé explicitement par l'utilisateur) :
// re-recherche sur Discogs et remplace la tracklist existante, y compris
// si elle avait été saisie/modifiée manuellement.
func (s *DiscService) FetchTracklistForDisc(ctx context.Context, discID int, force bool) ([]tracks.Track, error) {
	if !force {
		existing, err := s.trackRepo.FindByVinylID(ctx, discID)
		if err != nil {
			return nil, err
		}
		if len(existing) > 0 {
			return existing, nil
		}
	}

	disc, err := s.repo.FindByID(ctx, discID)
	if err != nil {
		return nil, err
	}
	if disc == nil {
		return nil, fmt.Errorf("disque introuvable")
	}

	releaseID, err := s.findReleaseIDByTitleArtist(disc.Title, disc.ArtistName)
	if err != nil {
		return nil, err
	}

	preview, err := s.getDiscogsReleaseDetails(releaseID)
	if err != nil {
		return nil, err
	}
	if len(preview.Tracks) == 0 {
		return nil, fmt.Errorf("aucune piste trouvée sur Discogs")
	}

	trackList := make([]tracks.Track, len(preview.Tracks))
	for i, t := range preview.Tracks {
		trackList[i] = tracks.Track{Position: t.Position, Title: t.Title, Duration: t.Duration}
	}

	if force {
		if err := s.trackRepo.DeleteByVinylID(ctx, discID); err != nil {
			return nil, err
		}
	}
	if err := s.trackRepo.CreateBatch(ctx, discID, trackList); err != nil {
		return nil, err
	}

	return s.trackRepo.FindByVinylID(ctx, discID)
}

func (s *DiscService) searchDiscogsByBarcode(barcode string) (*CoverPreview, error) {
	if s.discogsToken == "" {
		return &CoverPreview{Found: false}, nil
	}
	barcode = strings.TrimSpace(barcode)
	if barcode == "" {
		return &CoverPreview{Found: false}, nil
	}
	fmt.Printf("🔍 [Service] Recherche Discogs par code-barres: %s\n", barcode)
	url := fmt.Sprintf("https://api.discogs.com/database/search?barcode=%s&token=%s",
		barcode,
		s.discogsToken,
	)
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "RecordsManager/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erreur API Discogs: %d", resp.StatusCode)
	}
	var searchResp DiscogsSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, err
	}
	fmt.Printf("📊 [Service] Discogs a retourné %d résultats\n", len(searchResp.Results))
	if len(searchResp.Results) == 0 {
		url = fmt.Sprintf("https://api.discogs.com/database/search?q=%s&type=release&token=%s",
			strings.ReplaceAll(barcode, " ", "+"),
			s.discogsToken,
		)
		req, err = http.NewRequest("GET", url, nil)
		if err != nil {
			return &CoverPreview{Found: false}, nil
		}
		req.Header.Set("User-Agent", "RecordsManager/1.0")
		resp, err = client.Do(req)
		if err != nil {
			return &CoverPreview{Found: false}, nil
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
				return &CoverPreview{Found: false}, nil
			}
		}
	}
	if len(searchResp.Results) > 0 {
		firstResult := searchResp.Results[0]
		coverURL := firstResult.CoverURL
		detailedPreview, err := s.getDiscogsReleaseDetails(int64(firstResult.ID))
		if err == nil && detailedPreview != nil {
			if detailedPreview.CoverURL == "" && coverURL != "" {
				detailedPreview.CoverURL = coverURL
			}
			detailedPreview.Results = searchResp.Results
			return detailedPreview, nil
		}
		parts := strings.SplitN(firstResult.Title, " - ", 2)
		artist := ""
		title := firstResult.Title
		if len(parts) == 2 {
			artist = parts[0]
			title = parts[1]
		}
		var formatNames []string
		for _, format := range firstResult.Formats {
			formatNames = append(formatNames, format.Name)
			for _, desc := range format.Descriptions {
				formatNames = append(formatNames, desc)
			}
		}
		label := ""
		if len(firstResult.Labels) > 0 {
			label = firstResult.Labels[0].Name
		}
		prices, _ := s.getDiscogsPrices(int64(firstResult.ID))
		return &CoverPreview{
			CoverURL: coverURL,
			Title:    title,
			Artist:   artist,
			Year:     firstResult.Year,
			Genres:   firstResult.Genres,
			Formats:  formatNames,
			Country:  firstResult.Country,
			Label:    label,
			Found:    true,
			Results:  searchResp.Results,
			Prices:   prices,
		}, nil
	}
	return &CoverPreview{Found: false}, nil
}

func (s *DiscService) searchDiscogsWithResults(title, artist, country, year string) (*CoverPreview, error) {
	if s.discogsToken == "" {
		return &CoverPreview{Found: false}, nil
	}
	title = strings.TrimSpace(title)
	artist = strings.TrimSpace(artist)
	if title == "" && artist == "" {
		return &CoverPreview{Found: false}, nil
	}
	fmt.Printf("🔍 [Service] Recherche Discogs manuelle: titre='%s', artiste='%s', pays='%s', année='%s'\n", title, artist, country, year)

	// ✅ Construction de l'URL avec paramètres dédiés pour une meilleure pertinence
	url := fmt.Sprintf("https://api.discogs.com/database/search?type=release&format=Vinyl&per_page=50&token=%s", s.discogsToken)

	// ✅ Utiliser release_title et artist séparément (meilleure correspondance)
	if title != "" {
		url += fmt.Sprintf("&release_title=%s", strings.ReplaceAll(title, " ", "+"))
	}
	if artist != "" {
		url += fmt.Sprintf("&artist=%s", strings.ReplaceAll(artist, " ", "+"))
	}
	if country != "" {
		url += fmt.Sprintf("&country=%s", strings.ReplaceAll(strings.TrimSpace(country), " ", "+"))
	}
	if year != "" {
		url += fmt.Sprintf("&year=%s", strings.TrimSpace(year))
	}
	fmt.Printf("🔗 [Service] URL Discogs: %s\n", url)
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "RecordsManager/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erreur API Discogs: %d", resp.StatusCode)
	}
	var searchResp DiscogsSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, err
	}
	fmt.Printf("📊 [Service] Discogs a retourné %d résultats\n", len(searchResp.Results))
	if len(searchResp.Results) > 0 {
		firstResult := searchResp.Results[0]
		coverURL := firstResult.CoverURL
		detailedPreview, err := s.getDiscogsReleaseDetails(int64(firstResult.ID))
		if err == nil && detailedPreview != nil {
			if detailedPreview.CoverURL == "" && coverURL != "" {
				detailedPreview.CoverURL = coverURL
			}
			detailedPreview.Results = searchResp.Results
			return detailedPreview, nil
		}
		parts := strings.SplitN(firstResult.Title, " - ", 2)
		artistName := artist
		titleName := title
		if len(parts) == 2 {
			artistName = parts[0]
			titleName = parts[1]
		}
		var formatNames []string
		for _, format := range firstResult.Formats {
			formatNames = append(formatNames, format.Name)
			for _, desc := range format.Descriptions {
				formatNames = append(formatNames, desc)
			}
		}
		label := ""
		if len(firstResult.Labels) > 0 {
			label = firstResult.Labels[0].Name
		}
		prices, _ := s.getDiscogsPrices(int64(firstResult.ID))
		return &CoverPreview{
			CoverURL: coverURL,
			Title:    titleName,
			Artist:   artistName,
			Year:     firstResult.Year,
			Genres:   firstResult.Genres,
			Formats:  formatNames,
			Country:  firstResult.Country,
			Label:    label,
			Found:    true,
			Results:  searchResp.Results,
			Prices:   prices,
		}, nil
	}
	return &CoverPreview{Found: false}, nil
}

// 🔧 CORRIGÉ : support int64 + parsing flexible EUR
func (s *DiscService) getDiscogsPrices(releaseID int64) (*DiscogsPrice, error) {
	if s.discogsToken == "" {
		fmt.Println("❌ [getDiscogsPrices] Token Discogs manquant")
		return nil, fmt.Errorf("token Discogs manquant")
	}
	url := fmt.Sprintf("https://api.discogs.com/marketplace/stats/%d?curr=EUR&token=%s", releaseID, s.discogsToken)
	fmt.Printf("💶 [getDiscogsPrices] Appel API pour release_id=%d\n", releaseID)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("❌ [getDiscogsPrices] Erreur création requête: %v\n", err)
		return nil, err
	}
	req.Header.Set("User-Agent", "RecordsManager/1.0")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ [getDiscogsPrices] Erreur HTTP: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ [getDiscogsPrices] Erreur HTTP status: %d\n", resp.StatusCode)
		return nil, fmt.Errorf("erreur API Discogs pour les prix: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ [getDiscogsPrices] Erreur lecture body: %v\n", err)
		return nil, err
	}

	var rawResponse map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &rawResponse); err != nil {
		fmt.Printf("❌ [getDiscogsPrices] Erreur décodage JSON: %v\n", err)
		return nil, err
	}

	extractPrice := func(key string) float64 {
		val, exists := rawResponse[key]
		if !exists || val == nil {
			return 0
		}
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case map[string]interface{}:
			if value, ok := v["value"].(float64); ok {
				return value
			}
		}
		return 0
	}

	result := &DiscogsPrice{
		LowestPrice:  extractPrice("lowest_price"),
		MedianPrice:  extractPrice("median_price"),
		HighestPrice: extractPrice("highest_price"),
	}

	if result.LowestPrice == 0 && result.MedianPrice == 0 && result.HighestPrice == 0 {
		fmt.Printf("⚠️ [getDiscogsPrices] Aucun prix disponible pour release_id=%d\n", releaseID)
		return nil, nil
	}

	fmt.Printf("✅ [getDiscogsPrices] Prix extraits: lowest=%.2f, median=%.2f, highest=%.2f\n",
		result.LowestPrice, result.MedianPrice, result.HighestPrice)
	return result, nil
}

// 🔧 CORRIGÉ : int64
func (s *DiscService) getDiscogsReleaseDetails(releaseID int64) (*CoverPreview, error) {
	if s.discogsToken == "" {
		return nil, fmt.Errorf("token Discogs manquant")
	}
	url := fmt.Sprintf("https://api.discogs.com/releases/%d?token=%s", releaseID, s.discogsToken)
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "RecordsManager/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erreur API Discogs: %d", resp.StatusCode)
	}
	var release struct {
		Title   string `json:"title"`
		Artists []struct {
			Name string `json:"name"`
		} `json:"artists"`
		Year     int      `json:"year"`
		CoverURL string   `json:"cover_image"`
		Genres   []string `json:"genres"`
		Styles   []string `json:"styles"`
		Formats  []struct {
			Name         string   `json:"name"`
			Descriptions []string `json:"descriptions"`
		} `json:"formats"`
		Country string `json:"country"`
		Labels  []struct {
			Name string `json:"name"`
		} `json:"labels"`
		Tracklist []struct {
			Position string `json:"position"`
			Title    string `json:"title"`
			Duration string `json:"duration"`
		} `json:"tracklist"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	artist := ""
	if len(release.Artists) > 0 {
		artist = release.Artists[0].Name
	}
	var formatNames []string
	for _, format := range release.Formats {
		formatNames = append(formatNames, format.Name)
		for _, desc := range format.Descriptions {
			formatNames = append(formatNames, desc)
		}
	}
	var allGenres []string
	allGenres = append(allGenres, release.Genres...)
	allGenres = append(allGenres, release.Styles...)
	label := ""
	if len(release.Labels) > 0 {
		label = release.Labels[0].Name
	}
	var tracks []TrackItem
	for _, t := range release.Tracklist {
		tracks = append(tracks, TrackItem{Position: t.Position, Title: t.Title, Duration: t.Duration})
	}
	prices, _ := s.getDiscogsPrices(releaseID)
	return &CoverPreview{
		CoverURL: release.CoverURL,
		Title:    release.Title,
		Artist:   artist,
		Year:     fmt.Sprintf("%d", release.Year),
		Genres:   allGenres,
		Formats:  formatNames,
		Country:  release.Country,
		Label:    label,
		Found:    true,
		Prices:   prices,
		Tracks:   tracks,
	}, nil
}

func (s *DiscService) handleDiscogsCoverImage(coverPreview *CoverPreview, title, artistName string, discID int) (*string, error) {
	if coverPreview == nil || !coverPreview.Found || coverPreview.CoverURL == "" {
		return nil, nil
	}
	fmt.Printf("🖼️ Gestion image Discogs: %s\n", coverPreview.CoverURL)
	if strings.HasPrefix(coverPreview.CoverURL, "/uploads/covers/") {
		fmt.Printf("📁 Image déjà locale: %s\n", coverPreview.CoverURL)
		return &coverPreview.CoverURL, nil
	}
	if strings.HasPrefix(coverPreview.CoverURL, "http") {
		fmt.Printf("🌐 Téléchargement image Discogs: %s\n", coverPreview.CoverURL)
		finalTitle := title
		if coverPreview.Title != "" {
			finalTitle = coverPreview.Title
		}
		finalArtist := artistName
		if coverPreview.Artist != "" {
			finalArtist = coverPreview.Artist
		}
		coverPath, err := s.downloadCoverWithName(coverPreview.CoverURL, finalTitle, finalArtist, discID)
		if err != nil {
			fmt.Printf("❌ Erreur téléchargement image Discogs: %v\n", err)
			return nil, err
		}
		fmt.Printf("✅ Image Discogs téléchargée: %s\n", coverPath)
		return &coverPath, nil
	}
	return nil, nil
}

func (s *DiscService) CreateDisc(
	ctx context.Context,
	title string,
	artistID int,
	artistName string,
	genreID *int,
	formatID *int,
	countryID *int,
	labelID *int,
	releaseYear *int,
	barcode *string,
	notes *string,
	price *float64,
	quantity *int,
	coverImage *string,
	// ✅ NOUVEAU: Paramètres de streaming
	appleMusicURL *string,
	spotifyURL *string,
	deezerURL *string,
	youtubeURL *string,
	isrc *string,
	// Tracklist Discogs, récupérée côté frontend au moment de la sélection
	// de la release (voir CoverPreview.Tracks) et transmise ici pour être
	// stockée en une fois avec le disque.
	discogsReleaseID *int64,
	trackItems []TrackItem,
) (*discs.DiscWithDetails, error) {
	if barcode != nil && *barcode != "" {
		exists, existingDisc, err := s.CheckBarcodeExists(ctx, *barcode, nil)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, fmt.Errorf("un disque avec ce code-barres existe déjà (ID: %d)", existingDisc.ID)
		}
	}
	disc := &discs.Disc{
		Title:       title,
		ArtistID:    artistID,
		GenreID:     genreID,
		FormatID:    formatID,
		CountryID:   countryID,
		LabelID:     labelID,
		ReleaseYear: releaseYear,
		Barcode:     barcode,
		Notes:       notes,
		Price:       price,
		Quantity:    quantity,
		CoverURL:    coverImage,
		// ✅ AJOUT: Champs de streaming
		AppleMusicURL: appleMusicURL,
		SpotifyURL:    spotifyURL,
		DeezerURL:     deezerURL,
		YoutubeURL:    youtubeURL,
		ISRC:          isrc,

		DiscogsReleaseID: discogsReleaseID,
	}
	if err := s.repo.Create(ctx, disc); err != nil {
		return nil, err
	}
	var created *discs.DiscWithDetails
	var err error
	if barcode != nil && *barcode != "" {
		created, err = s.repo.FindByBarcodeWithDetails(ctx, *barcode)
	} else {
		created, err = s.repo.FindByTitleWithDetails(ctx, title)
	}
	if err != nil {
		return nil, fmt.Errorf("échec de la récupération après création : %w", err)
	}
	if created == nil {
		return nil, fmt.Errorf("disque créé mais introuvable")
	}
	finalArtistName := artistName
	if finalArtistName == "" && created.ArtistName != "" {
		finalArtistName = created.ArtistName
		fmt.Printf("✅ [CreateDisc] Nom artiste récupéré depuis la base: %s (ID: %d)\n",
			finalArtistName, created.ArtistID)
	}
	// La résolution de pochette (téléchargement depuis Discogs, recherche
	// Discogs de secours) implique des appels réseau externes lents ; on la
	// détache en arrière-plan pour que l'enregistrement du disque réponde
	// immédiatement — la pochette apparaît quelques instants plus tard au
	// prochain rafraîchissement de la liste. On utilise un contexte détaché
	// (context.Background) car `ctx` sera annulé dès la fin de la requête
	// HTTP, bien avant que cette goroutine ait fini.
	go func() {
		bgCtx := context.Background()
		if coverImage != nil && *coverImage != "" {
			fmt.Printf("📁 Pochette fournie par le frontend: %s\n", *coverImage)
			if strings.HasPrefix(*coverImage, "data:image") {
				fmt.Printf("🔧 Détection image base64, conversion...\n")
				parts := strings.SplitN(*coverImage, ",", 2)
				if len(parts) == 2 {
					imageData, err := base64.StdEncoding.DecodeString(parts[1])
					if err == nil {
						newCoverPath, err := s.saveUploadedCover(imageData, title, finalArtistName, created.ID)
						if err == nil {
							created.CoverURL = &newCoverPath
							if err := s.repo.Update(bgCtx, &discs.Disc{
								ID:          &created.ID,
								Title:       created.Title,
								ArtistID:    created.ArtistID,
								GenreID:     created.GenreID,
								FormatID:    created.FormatID,
								CountryID:   created.CountryID,
								LabelID:     created.LabelID,
								ReleaseYear: created.ReleaseYear,
								Barcode:     created.Barcode,
								CoverURL:    &newCoverPath,
								Notes:       created.Notes,
								Price:       created.Price,
								Quantity:    created.Quantity,
							}); err != nil {
								fmt.Printf("❌ Erreur mise à jour cover_url: %v\n", err)
							}
						}
					}
				}
			} else if strings.HasPrefix(*coverImage, "http") {
				fmt.Printf("🌐 URL internet détectée, téléchargement...\n")
				coverPath, err := s.downloadCoverWithName(*coverImage, title, finalArtistName, created.ID)
				if err == nil && coverPath != "" {
					created.CoverURL = &coverPath
					if err := s.repo.Update(bgCtx, &discs.Disc{
						ID:          &created.ID,
						Title:       created.Title,
						ArtistID:    created.ArtistID,
						GenreID:     created.GenreID,
						FormatID:    created.FormatID,
						CountryID:   created.CountryID,
						LabelID:     created.LabelID,
						ReleaseYear: created.ReleaseYear,
						Barcode:     created.Barcode,
						CoverURL:    &coverPath,
						Notes:       created.Notes,
						Price:       created.Price,
						Quantity:    created.Quantity,
					}); err != nil {
						fmt.Printf("❌ Erreur mise à jour cover_url: %v\n", err)
					}
				}
			} else if strings.HasPrefix(*coverImage, "/uploads/covers/") {
				fmt.Printf("📂 Chemin local détecté: %s\n", *coverImage)
				filename := filepath.Base(*coverImage)
				expectedPattern := fmt.Sprintf(".* - .* \\(%d\\)\\.(jpg|png|webp)", created.ID)
				matched, _ := regexp.MatchString(expectedPattern, filename)
				if !matched {
					fmt.Printf("⚠️ Format de fichier incorrect, renommage...\n")
					oldPath := filepath.Join(s.uploadsDir, strings.TrimPrefix(*coverImage, "/uploads/"))
					newFilename := fmt.Sprintf("%s - %s (%d)%s",
						sanitizeFilename(finalArtistName),
						sanitizeFilename(title),
						created.ID,
						filepath.Ext(filename))
					newPath := filepath.Join(filepath.Dir(oldPath), newFilename)
					if err := os.Rename(oldPath, newPath); err == nil {
						newCoverPath := "/uploads/covers/" + newFilename
						created.CoverURL = &newCoverPath
						if err := s.repo.Update(bgCtx, &discs.Disc{
							ID:          &created.ID,
							Title:       created.Title,
							ArtistID:    created.ArtistID,
							GenreID:     created.GenreID,
							FormatID:    created.FormatID,
							CountryID:   created.CountryID,
							LabelID:     created.LabelID,
							ReleaseYear: created.ReleaseYear,
							Barcode:     created.Barcode,
							CoverURL:    &newCoverPath,
							Notes:       created.Notes,
							Price:       created.Price,
							Quantity:    created.Quantity,
						}); err != nil {
							fmt.Printf("❌ Erreur mise à jour cover_url: %v\n", err)
						}
					}
				}
			}
		} else {
			var discogsPreview *CoverPreview
			if barcode != nil && *barcode != "" {
				fmt.Printf("🔍 Recherche pochette par code-barres: %s\n", *barcode)
				preview, err := s.searchDiscogsByBarcode(*barcode)
				if err == nil && preview.Found {
					discogsPreview = preview
					fmt.Printf("✅ Pochette trouvée par code-barres: %s\n", preview.CoverURL)
				} else {
					fmt.Printf("⚠️ Pas de pochette trouvée par code-barres\n")
				}
			}
			if discogsPreview == nil && finalArtistName != "" {
				fmt.Printf("🔍 Recherche pochette par titre/artiste: %s - %s\n", finalArtistName, title)
				preview, err := s.searchDiscogsWithResults(title, finalArtistName, "", "")
				if err == nil && preview.Found {
					discogsPreview = preview
					fmt.Printf("✅ Pochette trouvée par titre/artiste: %s\n", preview.CoverURL)
				} else {
					fmt.Printf("⚠️ Pas de pochette trouvée par titre/artiste\n")
				}
			}
			if discogsPreview != nil && discogsPreview.Found {
				fmt.Printf("📥 Téléchargement de la pochette pour le disque ID: %d\n", created.ID)
				coverPath, err := s.handleDiscogsCoverImage(discogsPreview, title, finalArtistName, created.ID)
				if err != nil {
					fmt.Printf("❌ Erreur téléchargement pochette: %v\n", err)
				} else if coverPath != nil && *coverPath != "" {
					fmt.Printf("✅ Pochette téléchargée: %s\n", *coverPath)
					created.CoverURL = coverPath
					fmt.Printf("💾 Mise à jour du cover_url en base pour ID: %d\n", created.ID)
					if err := s.repo.Update(bgCtx, &discs.Disc{
						ID:          &created.ID,
						Title:       created.Title,
						ArtistID:    created.ArtistID,
						GenreID:     created.GenreID,
						FormatID:    created.FormatID,
						CountryID:   created.CountryID,
						LabelID:     created.LabelID,
						ReleaseYear: created.ReleaseYear,
						Barcode:     created.Barcode,
						CoverURL:    coverPath,
						Notes:       created.Notes,
						Price:       created.Price,
						Quantity:    created.Quantity,
					}); err != nil {
						fmt.Printf("❌ Erreur mise à jour cover_url: %v\n", err)
					} else {
						fmt.Printf("✅ cover_url mis à jour avec succès en base\n")
					}
				}
			} else {
				fmt.Printf("⚠️ Aucune URL de pochette trouvée\n")
			}
		}
	}()

	if len(trackItems) > 0 {
		trackList := make([]tracks.Track, len(trackItems))
		for i, t := range trackItems {
			trackList[i] = tracks.Track{Position: t.Position, Title: t.Title, Duration: t.Duration}
		}
		if err := s.trackRepo.CreateBatch(ctx, created.ID, trackList); err != nil {
			fmt.Printf("❌ Erreur enregistrement tracklist: %v\n", err)
		} else {
			fmt.Printf("✅ Tracklist enregistrée (%d pistes)\n", len(trackList))
		}
	}

	return s.repo.FindByID(ctx, created.ID)
}

func (s *DiscService) GetAllDiscs(ctx context.Context) ([]discs.DiscWithDetails, error) {
	return s.repo.FindAll(ctx)
}

func (s *DiscService) GetDiscByID(ctx context.Context, id int) (*discs.DiscWithDetails, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *DiscService) UpdateDisc(
	ctx context.Context,
	id int,
	title string,
	artistID int,
	artistName string,
	genreID *int,
	formatID *int,
	countryID *int,
	labelID *int,
	releaseYear *int,
	barcode *string,
	notes *string,
	price *float64,
	quantity *int,
	coverImage *string,
	// ✅ NOUVEAU: Paramètres de streaming
	appleMusicURL *string,
	spotifyURL *string,
	deezerURL *string,
	youtubeURL *string,
	isrc *string,
) (*discs.DiscWithDetails, error) {
	if barcode != nil && *barcode != "" {
		exists, existingDisc, err := s.CheckBarcodeExists(ctx, *barcode, &id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, fmt.Errorf("un autre disque avec ce code-barres existe déjà (ID: %d)", existingDisc.ID)
		}
	}
	currentDisc, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	idPtr := id
	finalArtistName := artistName
	if finalArtistName == "" && currentDisc.ArtistName != "" {
		finalArtistName = currentDisc.ArtistName
		fmt.Printf("✅ [UpdateDisc] Nom artiste récupéré depuis la base: %s (ID: %d)\n",
			finalArtistName, currentDisc.ArtistID)
	}
	var finalCoverURL *string
	if coverImage != nil {
		if *coverImage == "" {
			fmt.Println("🗑️ Suppression explicite de la pochette demandée (coverImage = \"\")")
			if currentDisc.CoverURL != nil && *currentDisc.CoverURL != "" {
				oldFilePath := *currentDisc.CoverURL
				if strings.HasPrefix(oldFilePath, "/uploads/") {
					relativePath := strings.TrimPrefix(oldFilePath, "/uploads/")
					diskPath := filepath.Join(s.uploadsDir, relativePath)
					fmt.Printf("🗑️ Suppression du fichier: %s\n", diskPath)
					if err := os.Remove(diskPath); err != nil {
						if !os.IsNotExist(err) {
							fmt.Printf("⚠️ Erreur lors de la suppression: %v\n", err)
						} else {
							fmt.Println("ℹ️  Fichier déjà absent (normal)")
						}
					} else {
						fmt.Println("✅ Fichier supprimé avec succès")
					}
				}
			}
			finalCoverURL = nil
			fmt.Println("✅ Couverture mise à NULL")
		} else {
			if currentDisc.CoverURL != nil && *currentDisc.CoverURL != "" && *currentDisc.CoverURL != *coverImage {
				oldFilePath := *currentDisc.CoverURL
				if strings.HasPrefix(oldFilePath, "/uploads/") {
					relativePath := strings.TrimPrefix(oldFilePath, "/uploads/")
					diskPath := filepath.Join(s.uploadsDir, relativePath)
					fmt.Printf("🔄 Remplacement: suppression de l'ancienne pochette: %s\n", diskPath)
					if err := os.Remove(diskPath); err != nil {
						if !os.IsNotExist(err) {
							fmt.Printf("⚠️ Impossible de supprimer l'ancien fichier: %v\n", err)
						} else {
							fmt.Println("ℹ️  Ancien fichier déjà absent")
						}
					} else {
						fmt.Println("✅ Ancienne pochette supprimée")
					}
				}
			}
			if strings.HasPrefix(*coverImage, "data:image") {
				fmt.Println("🔧 Image base64 détectée")
				parts := strings.SplitN(*coverImage, ",", 2)
				if len(parts) == 2 {
					imageData, err := base64.StdEncoding.DecodeString(parts[1])
					if err == nil {
						newCoverPath, err := s.saveUploadedCover(imageData, title, finalArtistName, id)
						if err == nil {
							finalCoverURL = &newCoverPath
							fmt.Printf("✅ Base64 convertie et sauvegardée: %s\n", newCoverPath)
						} else {
							fmt.Printf("❌ Erreur sauvegarde base64: %v\n", err)
						}
					}
				}
			} else if strings.HasPrefix(*coverImage, "http") {
				fmt.Println("🌐 URL détectée, téléchargement...")
				coverPath, err := s.downloadCoverWithName(*coverImage, title, finalArtistName, id)
				if err == nil && coverPath != "" {
					finalCoverURL = &coverPath
					fmt.Printf("✅ Image téléchargée: %s\n", coverPath)
				} else {
					fmt.Printf("❌ Échec du téléchargement: %v\n", err)
				}
			} else {
				finalCoverURL = coverImage
				fmt.Printf("✅ Chemin local utilisé: %s\n", *coverImage)
			}
		}
	} else {
		finalCoverURL = currentDisc.CoverURL
		fmt.Println("ℹ️  coverImage = nil → conservation de la pochette actuelle")
	}
	disc := &discs.Disc{
		ID:          &idPtr,
		Title:       title,
		ArtistID:    artistID,
		GenreID:     genreID,
		FormatID:    formatID,
		CountryID:   countryID,
		LabelID:     labelID,
		ReleaseYear: releaseYear,
		Barcode:     barcode,
		CoverURL:    finalCoverURL,
		Notes:       notes,
		Price:       price,
		Quantity:    quantity,
		// ✅ AJOUT: Champs de streaming
		AppleMusicURL: appleMusicURL,
		SpotifyURL:    spotifyURL,
		DeezerURL:     deezerURL,
		YoutubeURL:    youtubeURL,
		ISRC:          isrc,
	}
	if err := s.repo.Update(ctx, disc); err != nil {
		return nil, fmt.Errorf("échec de la mise à jour en base : %w", err)
	}
	return s.repo.FindByID(ctx, id)
}

func (s *DiscService) DeleteDisc(ctx context.Context, id int) error {
	disc, err := s.repo.FindByID(ctx, id)
	if err == nil && disc != nil && disc.CoverURL != nil && *disc.CoverURL != "" {
		imagePath := filepath.Join(s.uploadsDir, strings.TrimPrefix(*disc.CoverURL, "/uploads/"))
		if err := os.Remove(imagePath); err != nil && !os.IsNotExist(err) {
			fmt.Printf("⚠️ Erreur suppression image lors de deleteDisc: %v\n", err)
		} else {
			fmt.Printf("🗑️ Pochette supprimée lors de la suppression du disque: %s\n", imagePath)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *DiscService) DownloadCoverForDisc(ctx context.Context, discID int) error {
	disc, err := s.repo.FindByID(ctx, discID)
	if err != nil {
		return err
	}
	if disc.CoverURL != nil && *disc.CoverURL != "" {
		diskPath := filepath.Join(s.uploadsDir, strings.TrimPrefix(*disc.CoverURL, "/uploads/"))
		if _, statErr := os.Stat(diskPath); statErr == nil {
			return nil
		}
	}
	imageURL, err := s.searchDiscogs(disc.Title, disc.ArtistName)
	if err != nil {
		return err
	}
	if imageURL == "" {
		return fmt.Errorf("aucune pochette trouvée sur Discogs")
	}
	coverPath, err := s.downloadCoverWithName(imageURL, disc.Title, disc.ArtistName, discID)
	if err != nil {
		return err
	}
	discUpdate := &discs.Disc{
		ID:          &discID,
		Title:       disc.Title,
		ArtistID:    disc.ArtistID,
		GenreID:     disc.GenreID,
		FormatID:    disc.FormatID,
		CountryID:   disc.CountryID,
		LabelID:     disc.LabelID,
		ReleaseYear: disc.ReleaseYear,
		Barcode:     disc.Barcode,
		CoverURL:    &coverPath,
		Notes:       disc.Notes,
		Price:       disc.Price,
		Quantity:    disc.Quantity,
	}
	return s.repo.Update(ctx, discUpdate)
}

func (s *DiscService) PreviewCover(ctx context.Context, barcode, title, artist, country, year string) (*CoverPreview, error) {
	if barcode != "" {
		return s.searchDiscogsByBarcode(barcode)
	}
	if title != "" || artist != "" {
		return s.searchDiscogsWithResults(title, artist, country, year)
	}
	return &CoverPreview{Found: false}, nil
}

// 🔧 CORRIGÉ : signature int64
func (s *DiscService) SelectDiscogsResult(ctx context.Context, releaseID int64) (*CoverPreview, error) {
	preview, err := s.getDiscogsReleaseDetails(releaseID)
	if err != nil {
		return nil, err
	}
	prices, _ := s.getDiscogsPrices(releaseID)
	if prices != nil {
		preview.Prices = prices
	}
	return preview, nil
}
