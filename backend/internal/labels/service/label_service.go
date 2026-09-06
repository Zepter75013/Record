// records-manager/backend/internal/labels/service/label_service.go
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"records-manager/backend/internal/countries"
	countriesRepo "records-manager/backend/internal/countries/repository"
	"records-manager/backend/internal/discogs"
	"records-manager/backend/internal/labels"
	"records-manager/backend/internal/labels/repository"
	"records-manager/backend/internal/musicbrainz"
	"records-manager/backend/internal/translate"
	"strings"
)

type LabelService struct {
	repo         repository.LabelRepository
	countryRepo  countriesRepo.CountryRepository
	discogsToken string
	deeplAPIKey  string
}

func NewLabelService(repo repository.LabelRepository, countryRepo countriesRepo.CountryRepository, discogsToken, deeplAPIKey string) *LabelService {
	return &LabelService{repo: repo, countryRepo: countryRepo, discogsToken: discogsToken, deeplAPIKey: deeplAPIKey}
}

func (s *LabelService) CreateLabel(ctx context.Context, name, description string, countryID *int) (*labels.Label, error) {
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
		CountryID:   countryID,
	}

	err = s.repo.Create(ctx, label)
	if err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, label.ID)
}

func (s *LabelService) GetAllLabels(ctx context.Context) ([]labels.Label, error) {
	return s.repo.FindAll(ctx)
}

func (s *LabelService) UpdateLabel(ctx context.Context, id int, name, description string, countryID *int) (*labels.Label, error) {
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
	existingLabel.CountryID = countryID

	err = s.repo.Update(ctx, existingLabel)
	if err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, id)
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
	if resp.StatusCode == http.StatusTooManyRequests {
		return "", fmt.Errorf("Discogs limite le débit, réessayez plus tard")
	}
	if resp.StatusCode != http.StatusOK {
		// Discogs a rejeté cette recherche précise (label introuvable, requête
		// mal formée pour ce nom...) — ce n'est pas une panne de l'app, on
		// traite ça comme "pas de donnée" plutôt que comme une erreur.
		return "", nil
	}

	var searchResult discogsLabelSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		return "", err
	}
	if len(searchResult.Results) == 0 {
		return "", nil
	}

	detailsURL := fmt.Sprintf("https://api.discogs.com/labels/%d?token=%s", searchResult.Results[0].ID, s.discogsToken)
	detailsResp, err := discogs.Request(detailsURL)
	if err != nil {
		return "", err
	}
	defer detailsResp.Body.Close()
	if detailsResp.StatusCode == http.StatusTooManyRequests {
		return "", fmt.Errorf("Discogs limite le débit, réessayez plus tard")
	}
	if detailsResp.StatusCode != http.StatusOK {
		return "", nil
	}

	var details discogsLabelDetails
	if err := json.NewDecoder(detailsResp.Body).Decode(&details); err != nil {
		return "", err
	}

	description := discogs.CleanProfile(details.Profile)
	if description == "" {
		return "", nil
	}
	// Tronqué avant traduction pour limiter la consommation de quota DeepL
	// à ce qui sera effectivement affiché, puis retronqué après (le
	// français est en général un peu plus long que l'anglais).
	description = discogs.TruncateAtWordBoundary(description, 11000)
	description = translate.ToFrench(s.deeplAPIKey, description)
	description = discogs.TruncateAtWordBoundary(description, 10000)

	return description, nil
}

// CountrySuggestion est la proposition de pays renvoyée à l'utilisateur pour
// validation — country_id est toujours renseigné (le pays est créé dans
// records_countries s'il n'y existait pas encore), mais rien n'est touché
// sur le label lui-même : c'est à l'utilisateur de valider en enregistrant
// le formulaire (bouton "Mettre à jour").
type CountrySuggestion struct {
	CountryID int    `json:"country_id"`
	Name      string `json:"name"`
	Code      string `json:"code"`
}

// SuggestCountryForLabel propose un pays pour un label à partir de
// MusicBrainz (Discogs, lui, ne structure pas de champ pays pour les
// labels). Ne modifie jamais le label lui-même — seul le pays est créé
// dans records_countries si besoin, comme une simple table de référence
// partagée.
func (s *LabelService) SuggestCountryForLabel(ctx context.Context, id int) (*CountrySuggestion, error) {
	label, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("label non trouvé")
	}
	if label == nil {
		return nil, errors.New("label non trouvé")
	}

	code, name, err := musicbrainz.SearchCountry("label", label.Name)
	if err != nil {
		return nil, err
	}
	if alias, ok := musicbrainz.CountryAliases[code]; ok {
		code = alias
	}

	existing, err := s.countryRepo.FindByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return &CountrySuggestion{CountryID: existing.ID, Name: existing.Name, Code: existing.Code}, nil
	}

	newCountry := &countries.Country{Name: name, Code: code}
	if err := s.countryRepo.Create(ctx, newCountry); err != nil {
		return nil, err
	}
	return &CountrySuggestion{CountryID: newCountry.ID, Name: newCountry.Name, Code: newCountry.Code}, nil
}
