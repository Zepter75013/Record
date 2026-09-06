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
	"strings"
	"time"
)

// CountryAliases : Discogs/l'app existante utilisent parfois des codes pays
// différents de l'ISO 3166-1 strict renvoyé par MusicBrainz (ex: "UK" au
// lieu de "GB").
var CountryAliases = map[string]string{
	"GB": "UK",
}

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
				return nil, fmt.Errorf("MusicBrainz a mis trop de temps à répondre, réessayez")
			}
			return nil, fmt.Errorf("impossible de joindre MusicBrainz : %w", err)
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
	Country string  `json:"country"`
	Area    areaRef `json:"area"`
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
	}
	return "", "", fmt.Errorf("pays inconnu sur MusicBrainz")
}

// SearchCountry recherche une entité MusicBrainz (artist ou label) par son
// nom et renvoie le code pays ISO du meilleur résultat ainsi que son nom
// lisible. entityType vaut "artist" ou "label". Certaines entités ne sont
// rattachées qu'à une zone (ville) sans pays renseigné directement ; on
// remonte alors vers le pays via resolveCountryFromArea.
func SearchCountry(entityType, name string) (code, countryName string, err error) {
	query := url.QueryEscape(fmt.Sprintf(`%s:"%s"`, entityType, name))
	resp, err := Request(entityType + "/?query=" + query + "&fmt=json&limit=1")
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var result searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", err
	}

	var top *searchResult
	switch entityType {
	case "label":
		if len(result.Labels) > 0 {
			top = &result.Labels[0]
		}
	case "artist":
		if len(result.Artists) > 0 {
			top = &result.Artists[0]
		}
	}
	if top == nil {
		return "", "", fmt.Errorf("aucun résultat trouvé sur MusicBrainz pour %q", name)
	}

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
