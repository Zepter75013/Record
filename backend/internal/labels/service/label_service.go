// records-manager/backend/internal/labels/service/label_service.go
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"records-manager/backend/internal/labels"
	"records-manager/backend/internal/labels/repository"
	"regexp"
	"strings"
	"time"
)

type LabelService struct {
	repo         repository.LabelRepository
	discogsToken string
}

func NewLabelService(repo repository.LabelRepository, discogsToken string) *LabelService {
	return &LabelService{repo: repo, discogsToken: discogsToken}
}

func (s *LabelService) CreateLabel(ctx context.Context, name, description string) (*labels.Label, error) {
	if name == "" {
		return nil, errors.New("le nom du label est requis")
	}

	existingLabel, err := s.repo.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}

	if existingLabel != nil {
		return nil, errors.New("un label avec ce nom existe déjà")
	}

	label := &labels.Label{
		Name:        name,
		Description: description,
	}

	err = s.repo.Create(ctx, label)
	if err != nil {
		return nil, err
	}

	return label, nil
}

func (s *LabelService) GetAllLabels(ctx context.Context) ([]labels.Label, error) {
	return s.repo.FindAll(ctx)
}

func (s *LabelService) UpdateLabel(ctx context.Context, id int, name, description string) (*labels.Label, error) {
	existingLabel, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if existingLabel == nil {
		return nil, errors.New("label non trouvé")
	}

	if name != existingLabel.Name {
		labelWithSameName, err := s.repo.FindByName(ctx, name)
		if err != nil {
			return nil, err
		}

		if labelWithSameName != nil && labelWithSameName.ID != id {
			return nil, errors.New("un autre label avec ce nom existe déjà")
		}
	}

	existingLabel.Name = name
	existingLabel.Description = description

	err = s.repo.Update(ctx, existingLabel)
	if err != nil {
		return nil, err
	}

	return existingLabel, nil
}

func (s *LabelService) DeleteLabel(ctx context.Context, id int) error {
	existingLabel, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if existingLabel == nil {
		return errors.New("label non trouvé")
	}

	return s.repo.Delete(ctx, id)
}

// GetLabelByID - Récupère un label par son ID
func (s *LabelService) GetLabelByID(ctx context.Context, id int) (*labels.Label, error) {
	label, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if label == nil {
		return nil, errors.New("label non trouvé")
	}

	return label, nil
}

type discogsLabelSearchResponse struct {
	Results []struct {
		ID int `json:"id"`
	} `json:"results"`
}

type discogsLabelDetails struct {
	Profile string `json:"profile"`
}

// discogsMarkupRegex retire les balises de mise en forme propres à Discogs
// dans les textes de profil (ex: "[a=Depeche Mode]", "[l=Mute Records]",
// "[r123456]") pour ne garder que le texte lisible.
var discogsMarkupRegex = regexp.MustCompile(`\[/?[a-z]=?[^\]]*\]`)

func discogsRequest(url string) (*http.Response, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "RecordsManager/1.0")

	resp, err := client.Do(req)
	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return nil, fmt.Errorf("Discogs a mis trop de temps à répondre, réessayez")
		}
		return nil, fmt.Errorf("impossible de joindre Discogs : %w", err)
	}
	return resp, nil
}

// multiSpaceRegex recolle les doubles espaces laissés par le retrait des
// balises Discogs (ex: "par [a=Ivo Watts-Russell] et" -> "par  et").
var multiSpaceRegex = regexp.MustCompile(`[ \t]{2,}`)

func cleanDiscogsProfile(profile string) string {
	cleaned := discogsMarkupRegex.ReplaceAllString(profile, "")
	cleaned = strings.ReplaceAll(cleaned, "\r\n", "\n")
	cleaned = multiSpaceRegex.ReplaceAllString(cleaned, " ")
	cleaned = strings.TrimSpace(cleaned)
	return cleaned
}

// truncateAtWordBoundary coupe une chaîne à au plus maxRunes runes, en
// reculant jusqu'au dernier espace pour éviter de couper un mot en deux.
func truncateAtWordBoundary(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	truncated := string(runes[:maxRunes])
	if idx := strings.LastIndexAny(truncated, " \n"); idx > 0 {
		truncated = truncated[:idx]
	}
	return strings.TrimSpace(truncated)
}

// SuggestDescriptionForLabel propose une description pour un label à partir
// de son profil Discogs. Ne modifie jamais le label lui-même — c'est à
// l'utilisateur de valider en enregistrant le formulaire (bouton "Mettre à
// jour"), pour ne pas écraser une description déjà en cours de saisie.
func (s *LabelService) SuggestDescriptionForLabel(ctx context.Context, id int) (string, error) {
	if s.discogsToken == "" {
		return "", fmt.Errorf("le token Discogs n'est pas configuré")
	}

	label, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return "", errors.New("label non trouvé")
	}
	if label == nil {
		return "", errors.New("label non trouvé")
	}

	searchURL := fmt.Sprintf("https://api.discogs.com/database/search?type=label&q=%s&token=%s",
		strings.ReplaceAll(label.Name, " ", "+"), s.discogsToken)
	resp, err := discogsRequest(searchURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("erreur API Discogs : %d", resp.StatusCode)
	}

	var searchResult discogsLabelSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		return "", err
	}
	if len(searchResult.Results) == 0 {
		return "", fmt.Errorf("aucun label trouvé sur Discogs pour %q", label.Name)
	}

	detailsURL := fmt.Sprintf("https://api.discogs.com/labels/%d?token=%s", searchResult.Results[0].ID, s.discogsToken)
	detailsResp, err := discogsRequest(detailsURL)
	if err != nil {
		return "", err
	}
	defer detailsResp.Body.Close()
	if detailsResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("erreur API Discogs : %d", detailsResp.StatusCode)
	}

	var details discogsLabelDetails
	if err := json.NewDecoder(detailsResp.Body).Decode(&details); err != nil {
		return "", err
	}

	description := cleanDiscogsProfile(details.Profile)
	if description == "" {
		return "", fmt.Errorf("aucune description disponible sur Discogs pour ce label")
	}
	description = truncateAtWordBoundary(description, 500)

	return description, nil
}
