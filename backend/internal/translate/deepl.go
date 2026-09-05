// records-manager/backend/internal/translate/deepl.go
//
// Traduction anglais -> français via l'API DeepL (offre gratuite). Utilisé
// pour traduire les descriptions/biographies récupérées depuis Discogs
// (toujours en anglais) avant de les proposer à l'utilisateur.
package translate

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type deeplResponse struct {
	Translations []struct {
		Text string `json:"text"`
	} `json:"translations"`
}

// ToFrench traduit un texte anglais en français via DeepL. Si la clé API
// n'est pas configurée, ou que la traduction échoue pour une raison
// quelconque (quota dépassé, réseau, réponse invalide...), elle renvoie le
// texte original tel quel plutôt que de faire échouer l'appelant : la
// traduction est un agrément, jamais une condition requise pour la
// suggestion de description/biographie elle-même.
func ToFrench(apiKey, text string) string {
	if apiKey == "" || strings.TrimSpace(text) == "" {
		return text
	}

	form := url.Values{}
	form.Set("text", text)
	form.Set("target_lang", "FR")
	form.Set("source_lang", "EN")

	req, err := http.NewRequest("POST", "https://api-free.deepl.com/v2/translate", strings.NewReader(form.Encode()))
	if err != nil {
		return text
	}
	req.Header.Set("Authorization", "DeepL-Auth-Key "+apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return text
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return text
	}

	var result deeplResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return text
	}
	if len(result.Translations) == 0 || result.Translations[0].Text == "" {
		return text
	}

	return result.Translations[0].Text
}
