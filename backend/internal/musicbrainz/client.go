// records-manager/backend/internal/musicbrainz/client.go
//
// Client HTTP partagé pour interroger l'API publique MusicBrainz (gratuite,
// sans clé) — utilisé pour retrouver le pays d'un artiste ou d'un label,
// que Discogs, lui, ne structure pas de façon fiable.
package musicbrainz

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// CountryAliases : Discogs/l'app existante utilisent parfois des codes pays
// différents de l'ISO 3166-1 strict renvoyé par MusicBrainz (ex: "UK" au
// lieu de "GB").
var CountryAliases = map[string]string{
	"GB": "UK",
}

// ErrNotFound signale qu'aucune entité ne correspond à la recherche — un cas
// normal (l'artiste/le label n'est simplement pas dans MusicBrainz), pas une
// panne. errors.Is permet aux appelants de le distinguer d'une vraie erreur
// réseau/débit limité pour l'afficher comme "aucune donnée" plutôt qu'un
// échec.
var ErrNotFound = errors.New("aucun résultat trouvé sur MusicBrainz")

// interCallDelay espace les appels MusicBrainz enchaînés au sein d'une même
// résolution (recherche -> remontée de zone -> site officiel), pour rester
// sous la limite de débit (~1 req/s en usage anonyme) sans attendre de se
// faire limiter puis retenter.
const interCallDelay = 400 * time.Millisecond

// Table minimale ISO 3166-1 alpha-2 -> nom anglais. Sert uniquement de repli
// quand MusicBrainz ne donne pas de nom de pays lisible directement — ce nom
// ne sert qu'à créer l'entrée dans records_countries si elle n'existe pas
// encore localement.
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

// CountryName renvoie un nom anglais lisible pour un code pays ISO
// 3166-1 alpha-2, ou le code lui-même si absent de la table.
func CountryName(code string) string {
	if n, ok := isoCountryNames[code]; ok {
		return n
	}
	return code
}

// Request effectue une requête GET vers l'API MusicBrainz. En cas de débit
// limité (429/503 — MusicBrainz est strict, ~1 requête/seconde en usage
// anonyme, et une résolution de pays peut enchaîner plusieurs appels), elle
// patiente puis retente au lieu d'échouer immédiatement.
func Request(path string) (*http.Response, error) {
	client := &http.Client{Timeout: 20 * time.Second}

	const maxAttempts = 3
	backoff := 2 * time.Second

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
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
				lastErr = fmt.Errorf("MusicBrainz a mis trop de temps à répondre, réessayez")
			} else {
				lastErr = fmt.Errorf("impossible de joindre MusicBrainz : %w", err)
			}
			// Une lenteur ou une coupure ponctuelle vaut la peine d'être
			// retentée, comme un 429/503 — MusicBrainz répond parfois très
			// lentement sous charge sans pour autant renvoyer d'erreur HTTP.
			if attempt < maxAttempts {
				time.Sleep(backoff)
				backoff *= 2
				continue
			}
			return nil, lastErr
		}

		isRateLimited := resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable
		if isRateLimited && attempt < maxAttempts {
			resp.Body.Close()
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("erreur API MusicBrainz : %d", resp.StatusCode)
		}
		return resp, nil
	}

	// Inatteignable : la boucle ci-dessus retourne toujours, mais le
	// compilateur exige un retour explicite en fin de fonction.
	return nil, fmt.Errorf("échec de la requête MusicBrainz après plusieurs tentatives")
}

type areaRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type searchResult struct {
	ID       string  `json:"id"`
	Country  string  `json:"country"`
	Area     areaRef `json:"area"`
	LifeSpan struct {
		Begin string `json:"begin"`
	} `json:"life-span"`
}

type searchResponse struct {
	Artists []searchResult `json:"artists"`
	Labels  []searchResult `json:"labels"`
}

type areaDetail struct {
	Name          string   `json:"name"`
	ISO31661Codes []string `json:"iso-3166-1-codes"`
	ISO31662Codes []string `json:"iso-3166-2-codes"`
	Relations     []struct {
		Direction string  `json:"direction"`
		Type      string  `json:"type"`
		Area      areaRef `json:"area"`
	} `json:"relations"`
}

// resolveCountryFromArea remonte la hiérarchie administrative MusicBrainz
// (ville -> subdivision -> pays) via les relations "part of" jusqu'à
// trouver un code pays ISO — beaucoup de villes/quartiers n'ont pas de code
// ISO propre, seul un ancêtre (souvent la subdivision de premier niveau,
// parfois le pays lui-même) en porte un. Bornée à quelques niveaux pour
// éviter une remontée infinie en cas de données MusicBrainz inhabituelles.
func resolveCountryFromArea(areaID string) (code, name string, err error) {
	currentID := areaID
	for depth := 0; depth < 6; depth++ {
		resp, err := Request("area/" + currentID + "?inc=area-rels&fmt=json")
		if err != nil {
			return "", "", err
		}
		var detail areaDetail
		decodeErr := json.NewDecoder(resp.Body).Decode(&detail)
		resp.Body.Close()
		if decodeErr != nil {
			return "", "", decodeErr
		}

		if len(detail.ISO31661Codes) > 0 {
			c := detail.ISO31661Codes[0]
			return c, CountryName(c), nil
		}
		if len(detail.ISO31662Codes) > 0 {
			c := strings.SplitN(detail.ISO31662Codes[0], "-", 2)[0]
			if c != "" {
				return c, CountryName(c), nil
			}
		}

		var parentID string
		for _, rel := range detail.Relations {
			if rel.Direction == "backward" && rel.Type == "part of" {
				parentID = rel.Area.ID
				break
			}
		}
		if parentID == "" {
			break
		}
		currentID = parentID
		time.Sleep(interCallDelay)
	}
	return "", "", fmt.Errorf("pays inconnu sur MusicBrainz")
}

// searchTop recherche une entité MusicBrainz (artist ou label) par son nom
// et renvoie son meilleur résultat. entityType vaut "artist" ou "label".
func searchTop(entityType, name string) (*searchResult, error) {
	query := url.QueryEscape(fmt.Sprintf(`%s:"%s"`, entityType, name))
	resp, err := Request(entityType + "/?query=" + query + "&fmt=json&limit=1")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	switch entityType {
	case "label":
		if len(result.Labels) > 0 {
			return &result.Labels[0], nil
		}
	case "artist":
		if len(result.Artists) > 0 {
			return &result.Artists[0], nil
		}
	}
	return nil, fmt.Errorf("%w pour %q", ErrNotFound, name)
}

// countryFromResult déduit le pays d'un résultat de recherche MusicBrainz :
// directement s'il porte un code pays, sinon en remontant depuis sa zone
// (ville, quartier...) via resolveCountryFromArea.
func countryFromResult(top *searchResult) (code, countryName string, err error) {
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

	return "", "", fmt.Errorf("pays inconnu sur MusicBrainz")
}

// SearchCountry recherche une entité MusicBrainz (artist ou label) par son
// nom et renvoie le code pays ISO du meilleur résultat ainsi que son nom
// lisible. entityType vaut "artist" ou "label". Certaines entités ne sont
// rattachées qu'à une zone (ville) sans pays renseigné directement ; on
// remonte alors vers le pays via resolveCountryFromArea.
func SearchCountry(entityType, name string) (code, countryName string, err error) {
	top, err := searchTop(entityType, name)
	if err != nil {
		return "", "", err
	}
	return countryFromResult(top)
}

// LabelInfo regroupe les informations récupérées sur un label via
// MusicBrainz — pays, année de fondation et site officiel. Chaque champ
// peut manquer indépendamment des autres (une valeur absente n'est pas une
// erreur : ce n'est pas rare que MusicBrainz n'ait que certaines de ces
// informations pour un label donné).
type LabelInfo struct {
	CountryCode  string
	CountryName  string
	FoundingYear *int
	Website      *string
}

// officialSite recherche le lien "site officiel" d'une entité MusicBrainz à
// partir de son identifiant — contrairement au pays/à l'année de fondation,
// ce n'est renvoyé que par la fiche détaillée (avec inc=url-rels), pas par
// la recherche elle-même.
func officialSite(mbid, entityType string) (string, error) {
	resp, err := Request(entityType + "/" + mbid + "?inc=url-rels&fmt=json")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Relations []struct {
			Type string `json:"type"`
			URL  struct {
				Resource string `json:"resource"`
			} `json:"url"`
		} `json:"relations"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	for _, rel := range result.Relations {
		if rel.Type == "official site" || rel.Type == "official homepage" {
			return rel.URL.Resource, nil
		}
	}
	return "", nil
}

// SearchLabelInfo recherche un label par son nom sur MusicBrainz et renvoie
// tout ce qui a pu être trouvé (pays, année de fondation, site officiel) —
// des informations que Discogs, lui, ne structure pas pour les labels.
// Chaque champ de LabelInfo peut rester vide indépendamment des autres. Si
// le label lui-même est introuvable sur MusicBrainz, ce n'est pas traité
// comme une erreur : un LabelInfo vide est renvoyé (cas normal, pas une
// panne — voir ErrNotFound).
func SearchLabelInfo(name string) (*LabelInfo, error) {
	top, err := searchTop("label", name)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return &LabelInfo{}, nil
		}
		return nil, err
	}

	info := &LabelInfo{}

	if code, countryName, err := countryFromResult(top); err == nil {
		info.CountryCode = code
		info.CountryName = countryName
	}

	// L'année de fondation est parfois une date complète ("1980-05-12"),
	// on ne garde que les 4 premiers chiffres.
	if len(top.LifeSpan.Begin) >= 4 {
		if year, err := strconv.Atoi(top.LifeSpan.Begin[:4]); err == nil {
			info.FoundingYear = &year
		}
	}

	if top.ID != "" {
		time.Sleep(interCallDelay)
		if site, err := officialSite(top.ID, "label"); err == nil && site != "" {
			info.Website = &site
		}
	}

	return info, nil
}
