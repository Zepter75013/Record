// records-manager/backend/internal/discogs/client.go
//
// Client HTTP partagé pour interroger l'API publique Discogs (recherche de
// labels, d'artistes, etc.) — factorisé ici car plusieurs modules (labels,
// artistes) en ont besoin et doivent bénéficier du même traitement du
// débit limité (429) sans dupliquer la logique.
package discogs

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// Request effectue une requête GET vers Discogs. En cas de 429 (débit
// limité — chaque suggestion fait plusieurs appels Discogs, ce qui arrive
// vite en mise à jour groupée sur de nombreux éléments), elle patiente puis
// retente au lieu d'échouer immédiatement.
func Request(url string) (*http.Response, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	const maxAttempts = 3
	backoff := 3 * time.Second

	for attempt := 1; attempt <= maxAttempts; attempt++ {
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

		if resp.StatusCode == http.StatusTooManyRequests && attempt < maxAttempts {
			resp.Body.Close()
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		return resp, nil
	}

	// Inatteignable : la boucle ci-dessus retourne toujours, mais le
	// compilateur exige un retour explicite en fin de fonction.
	return nil, fmt.Errorf("échec de la requête Discogs après plusieurs tentatives")
}

// markupRegex retire les balises de mise en forme propres à Discogs dans
// les textes de profil (ex: "[a=Depeche Mode]", "[l=Mute Records]",
// "[r123456]") pour ne garder que le texte lisible.
var markupRegex = regexp.MustCompile(`\[/?[a-z]=?[^\]]*\]`)

// multiSpaceRegex recolle les doubles espaces laissés par le retrait des
// balises Discogs (ex: "par [a=Ivo Watts-Russell] et" -> "par  et").
var multiSpaceRegex = regexp.MustCompile(`[ \t]{2,}`)

// CleanProfile nettoie un texte de profil Discogs (label ou artiste) pour
// l'utiliser comme description lisible.
func CleanProfile(profile string) string {
	cleaned := markupRegex.ReplaceAllString(profile, "")
	cleaned = strings.ReplaceAll(cleaned, "\r\n", "\n")
	cleaned = multiSpaceRegex.ReplaceAllString(cleaned, " ")
	cleaned = strings.TrimSpace(cleaned)
	return cleaned
}

// TruncateAtWordBoundary coupe une chaîne à au plus maxRunes runes, en
// reculant jusqu'au dernier espace pour éviter de couper un mot en deux.
func TruncateAtWordBoundary(s string, maxRunes int) string {
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
