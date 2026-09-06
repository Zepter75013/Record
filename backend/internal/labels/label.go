// records-manager/backend/internal/labels/model/label.go
package labels

import (
	"time"
)

type Label struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CountryID   *int      `json:"country_id,omitempty"`
	CountryName *string   `json:"countryname,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
