package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"records-manager/backend/internal/artists"
	"records-manager/backend/internal/artists/repository"
	"records-manager/backend/internal/countries"
	countriesRepo "records-manager/backend/internal/countries/repository"
	"records-manager/backend/internal/discogs"
	"records-manager/backend/internal/musicbrainz"
	"records-manager/backend/internal/translate"
)

type ArtistService struct {
	repo         repository.ArtistRepository
	countryRepo  countriesRepo.CountryRepository
	discogsToken string
	deeplAPIKey  string
}

func NewArtistService(repo repository.ArtistRepository, countryRepo countriesRepo.CountryRepository, discogsToken, deeplAPIKey string) *ArtistService {
	return &ArtistService{repo: repo, countryRepo: countryRepo, discogsToken: discogsToken, deeplAPIKey: deeplAPIKey}
}

func (s *ArtistService) CreateArtist(ctx context.Context, name, biography string, countryID *int) (*artists.Artist, error) {
	// Vérifier si l'artiste existe déjà
	existing, err := s.repo.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("un artiste avec ce nom existe déjà")
	}

	artist := &artists.Artist{
		Name:      name,
		Biography: biography,
		CountryID: countryID,
	}

	if err := s.repo.Create(ctx, artist); err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, artist.ID)
}

func (s *ArtistService) GetAllArtists(ctx context.Context) ([]artists.Artist, error) {
	return s.repo.FindAll(ctx)
}

func (s *ArtistService) GetArtistByID(ctx context.Context, id int) (*artists.Artist, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *ArtistService) UpdateArtist(ctx context.Context, id int, name, biography string, countryID *int) (*artists.Artist, error) {
	// Vérifier si un autre artiste avec ce nom existe
	existing, err := s.repo.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if existing != nil && existing.ID != id {
		return nil, fmt.Errorf("un autre artiste avec ce nom existe déjà")
	}

	artist, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	artist.Name = name
	artist.Biography = biography
	artist.CountryID = countryID

	if err := s.repo.Update(ctx, artist); err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, id)
}

func (s *ArtistService) DeleteArtist(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

// CountrySuggestion est la proposition de pays renvoyée à l'utilisateur pour
// validation — country_id est toujours renseigné (le pays est créé dans
// records_countries s'il n'y existait pas encore), mais rien n'est touché
// sur l'artiste lui-même : c'est à l'utilisateur de valider en enregistrant
// le formulaire (bouton "Mettre à jour").
type CountrySuggestion struct {
	CountryID int    `json:"country_id"`
	Name      string `json:"name"`
	Code      string `json:"code"`
}

// SuggestCountryForArtist propose un pays pour un artiste à partir de
// MusicBrainz. Ne modifie jamais l'artiste (pour ne pas écraser des
// modifications en cours de saisie dans le formulaire) — seul le pays
// lui-même est créé dans records_countries si besoin, comme une simple
// table de référence partagée.
func (s *ArtistService) SuggestCountryForArtist(ctx context.Context, artistID int) (*CountrySuggestion, error) {
	artist, err := s.repo.FindByID(ctx, artistID)
	if err != nil {
		return nil, err
	}
	if artist == nil {
		return nil, fmt.Errorf("artiste introuvable")
	}

	code, name, err := musicbrainz.SearchCountry("artist", artist.Name)
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

type discogsArtistSearchResponse struct {
	Results []struct {
		ID int `json:"id"`
	} `json:"results"`
}

type discogsArtistDetails struct {
	Profile string `json:"profile"`
}

// SuggestBiographyForArtist propose une biographie pour un artiste à partir
// de son profil Discogs (contrairement au pays, mal structuré côté Discogs
// pour les artistes — d'où l'usage de MusicBrainz pour SuggestCountryForArtist
// — le profil texte, lui, est disponible et adapté à une biographie). Ne
// modifie jamais l'artiste lui-même — c'est à l'utilisateur de valider en
// enregistrant le formulaire.
func (s *ArtistService) SuggestBiographyForArtist(ctx context.Context, artistID int) (string, error) {
	if s.discogsToken == "" {
		return "", fmt.Errorf("le token Discogs n'est pas configuré")
	}

	artist, err := s.repo.FindByID(ctx, artistID)
	if err != nil {
		return "", fmt.Errorf("artiste introuvable")
	}
	if artist == nil {
		return "", fmt.Errorf("artiste introuvable")
	}

	searchURL := fmt.Sprintf("https://api.discogs.com/database/search?type=artist&q=%s&token=%s",
		strings.ReplaceAll(artist.Name, " ", "+"), s.discogsToken)
	resp, err := discogs.Request(searchURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusTooManyRequests {
		return "", fmt.Errorf("Discogs limite le débit, réessayez plus tard")
	}
	if resp.StatusCode != http.StatusOK {
		// Discogs a rejeté cette recherche précise (artiste introuvable,
		// requête mal formée pour ce nom...) — ce n'est pas une panne de
		// l'app, on traite ça comme "pas de donnée" plutôt qu'une erreur.
		return "", nil
	}

	var searchResult discogsArtistSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		return "", err
	}
	if len(searchResult.Results) == 0 {
		return "", nil
	}

	detailsURL := fmt.Sprintf("https://api.discogs.com/artists/%d?token=%s", searchResult.Results[0].ID, s.discogsToken)
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

	var details discogsArtistDetails
	if err := json.NewDecoder(detailsResp.Body).Decode(&details); err != nil {
		return "", err
	}

	biography := discogs.CleanProfile(details.Profile)
	if biography == "" {
		return "", nil
	}
	// Tronqué avant traduction pour limiter la consommation de quota DeepL
	// à ce qui sera effectivement affiché, puis retronqué après (le
	// français est en général un peu plus long que l'anglais).
	biography = discogs.TruncateAtWordBoundary(biography, 11000)
	biography = translate.ToFrench(s.deeplAPIKey, biography)
	biography = discogs.TruncateAtWordBoundary(biography, 10000)

	return biography, nil
}
