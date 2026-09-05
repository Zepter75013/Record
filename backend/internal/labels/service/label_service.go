// records-manager/backend/internal/labels/service/label_service.go
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"records-manager/backend/internal/discogs"
	"records-manager/backend/internal/labels"
	"records-manager/backend/internal/labels/repository"
	"strings"
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
	resp, err := discogs.Request(searchURL)
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
	detailsResp, err := discogs.Request(detailsURL)
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

	description := discogs.CleanProfile(details.Profile)
	if description == "" {
		return "", fmt.Errorf("aucune description disponible sur Discogs pour ce label")
	}
	description = discogs.TruncateAtWordBoundary(description, 500)

	return description, nil
}
