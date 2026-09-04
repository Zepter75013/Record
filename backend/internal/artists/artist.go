package artists // ← CHANGÉ

import (
	"time"
)

type Artist struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Biography   string    `json:"biography"`
	CountryID   *int      `json:"country_id,omitempty"`
	CountryName *string   `json:"countryname,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
