package games

import "time"

type Game struct {
	ID          *int     `json:"id,omitempty"`
	Title       string   `json:"title"`
	PlatformID  int      `json:"platform_id"`
	GenreID     *int     `json:"genre_id,omitempty"`
	PublisherID *int     `json:"publisher_id,omitempty"`
	ReleaseYear *int     `json:"release_year,omitempty"`
	Barcode     *string  `json:"barcode,omitempty"`
	CoverURL    *string  `json:"cover_url,omitempty"`
	Notes       *string  `json:"notes,omitempty"`
	Price       *float64 `json:"price,omitempty"`
	Quantity    *int     `json:"quantity,omitempty"`

	RAWGID *int64 `json:"rawg_id,omitempty"`

	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

type GameWithDetails struct {
	ID            int      `json:"id"`
	Title         string   `json:"title"`
	PlatformID    int      `json:"platform_id"`
	PlatformName  string   `json:"platform_name"`
	GenreID       *int     `json:"genre_id,omitempty"`
	GenreName     *string  `json:"genre_name,omitempty"`
	PublisherID   *int     `json:"publisher_id,omitempty"`
	PublisherName *string  `json:"publisher_name,omitempty"`
	ReleaseYear   *int     `json:"release_year,omitempty"`
	Barcode       *string  `json:"barcode,omitempty"`
	CoverURL      *string  `json:"cover_url,omitempty"`
	Notes         *string  `json:"notes,omitempty"`
	Price         *float64 `json:"price,omitempty"`
	Quantity      *int     `json:"quantity,omitempty"`

	RAWGID *int64 `json:"rawg_id,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
