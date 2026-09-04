package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"records-manager/backend/internal/artists"
	"records-manager/backend/internal/artists/repository"
	"records-manager/backend/internal/countries"
	countriesRepo "records-manager/backend/internal/countries/repository"
)

type ArtistService struct {
	repo        repository.ArtistRepository
	countryRepo countriesRepo.CountryRepository
}

func NewArtistService(repo repository.ArtistRepository, countryRepo countriesRepo.CountryRepository) *ArtistService {
	return &ArtistService{repo: repo, countryRepo: countryRepo}
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

// Discogs/l'app existante utilisent parfois des codes pays différents de
// l'ISO 3166-1 strict renvoyé par MusicBrainz (ex: "UK" au lieu de "GB").
var countryCodeAliases = map[string]string{
	"GB": "UK",
}

// Table minimale ISO 3166-1 alpha-2 -> nom anglais. Sert uniquement de repli
// quand MusicBrainz ne donne pas de nom de pays lisible directement (voir
// resolveCountryFromArea ci-dessous) — le code, lui, vient toujours de
// l'API ; ce nom ne sert qu'à créer l'entrée dans records_countries si elle
// n'existe pas encore localement.
var isoCountryNames = map[string]string{
	"US": "United States", "GB": "United Kingdom", "FR": "France", "DE": "Germany",
	"IT": "Italy", "ES": "Spain", "NL": "Netherlands", "BE": "Belgium",
	"CH": "Switzerland", "AT": "Austria", "SE": "Sweden", "NO": "Norway",
	"DK": "Denmark", "FI": "Finland", "PT": "Portugal", "IE": "Ireland",
	"PL": "Poland", "CA": "Canada", "AU": "Australia", "NZ": "New Zealand",
	"JP": "Japan", "KR": "South Korea", "CN": "China", "BR": "Brazil",
	"MX": "Mexico", "AR": "Argentina", "RU": "Russia", "GR": "Greece",
	"CZ": "Czech Republic", "HU": "Hungary", "IS": "Iceland", "ZA": "South Africa",
}

type musicBrainzArtistSearchResponse struct {
	Artists []struct {
		Name    string `json:"name"`
		Country string `json:"country"`
		Area    struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"area"`
	} `json:"artists"`
}

type musicBrainzAreaRelationsResponse struct {
	Relations []struct {
		Area struct {
			ISO31662Codes []string `json:"iso-3166-2-codes"`
		} `json:"area"`
	} `json:"relations"`
}

func musicBrainzRequest(path string) (*http.Response, error) {
	client := &http.Client{Timeout: 20 * time.Second}
	req, err := http.NewRequest("GET", "https://musicbrainz.org/ws/2/"+path, nil)
	if err != nil {
		return nil, err
	}
	// MusicBrainz exige un User-Agent explicite, sous peine de limitation.
	req.Header.Set("User-Agent", "RecordsManager/1.0 (self-hosted collection manager)")

	resp, err := client.Do(req)
	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return nil, fmt.Errorf("MusicBrainz a mis trop de temps à répondre, réessayez")
		}
		return nil, fmt.Errorf("impossible de joindre MusicBrainz : %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("erreur API MusicBrainz: %d", resp.StatusCode)
	}
	return resp, nil
}

// resolveCountryFromArea remonte de la zone (souvent une ville, ex: Londres
// pour un artiste UK) vers son pays via les relations "part of" de
// MusicBrainz — toute subdivision (ex: Angleterre, code ISO "GB-ENG") porte
// le code pays en préfixe de son code ISO 3166-2, ce qui évite d'avoir à
// remonter toute la hiérarchie administrative à la main.
func resolveCountryFromArea(areaID string) (code string, name string, err error) {
	resp, err := musicBrainzRequest("area/" + areaID + "?inc=area-rels&fmt=json")
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var result musicBrainzAreaRelationsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", err
	}

	for _, rel := range result.Relations {
		if len(rel.Area.ISO31662Codes) == 0 {
			continue
		}
		iso := rel.Area.ISO31662Codes[0]
		countryCode := strings.SplitN(iso, "-", 2)[0]
		if countryCode == "" {
			continue
		}
		countryName := countryCode
		if n, ok := isoCountryNames[countryCode]; ok {
			countryName = n
		}
		return countryCode, countryName, nil
	}
	return "", "", fmt.Errorf("pays inconnu sur MusicBrainz pour cet artiste")
}

// searchMusicBrainzCountry recherche un artiste par son nom sur MusicBrainz
// (API publique, gratuite, sans clé — contrairement à Discogs qui n'a pas de
// champ pays structuré pour les artistes) et renvoie le code pays ISO du
// meilleur résultat ainsi que son nom lisible (anglais). Certains artistes
// (ex: Placebo) n'ont qu'une zone (ville) sans pays renseigné directement ;
// on remonte alors vers le pays via resolveCountryFromArea.
func searchMusicBrainzCountry(artistName string) (code string, name string, err error) {
	query := url.QueryEscape(fmt.Sprintf(`artist:"%s"`, artistName))
	resp, err := musicBrainzRequest("artist/?query=" + query + "&fmt=json&limit=1")
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var result musicBrainzArtistSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", err
	}
	if len(result.Artists) == 0 {
		return "", "", fmt.Errorf("aucun artiste trouvé sur MusicBrainz")
	}

	top := result.Artists[0]
	if top.Country != "" {
		countryName := top.Area.Name
		if countryName == "" || top.Area.Type != "Country" {
			countryName = top.Country
		}
		return top.Country, countryName, nil
	}

	if top.Area.ID != "" {
		return resolveCountryFromArea(top.Area.ID)
	}

	return "", "", fmt.Errorf("pays inconnu sur MusicBrainz pour cet artiste")
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

	code, name, err := searchMusicBrainzCountry(artist.Name)
	if err != nil {
		return nil, err
	}
	if alias, ok := countryCodeAliases[code]; ok {
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
