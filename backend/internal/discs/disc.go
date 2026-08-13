// backend/internal/discs/disc.go

package discs

import "time"

type Disc struct {
	ID          *int     `json:"id,omitempty"`
	Title       string   `json:"title"`
	ArtistID    int      `json:"artist_id"`
	GenreID     *int     `json:"genre_id,omitempty"`
	FormatID    *int     `json:"format_id,omitempty"`
	CountryID   *int     `json:"country_id,omitempty"`
	LabelID     *int     `json:"label_id,omitempty"`
	ReleaseYear *int     `json:"release_year,omitempty"`
	Barcode     *string  `json:"barcode,omitempty"`
	CoverURL    *string  `json:"cover_url,omitempty"`
	Notes       *string  `json:"notes,omitempty"`
	Price       *float64 `json:"price,omitempty"`
	Quantity    *int     `json:"quantity,omitempty"` // ✅ NOUVEAU: Nombre de disques

	// ✅ NOUVEAU: URLs de streaming
	AppleMusicURL *string `json:"apple_music_url,omitempty"`
	SpotifyURL    *string `json:"spotify_url,omitempty"`
	DeezerURL     *string `json:"deezer_url,omitempty"`
	YoutubeURL    *string `json:"youtube_url,omitempty"`
	ISRC          *string `json:"isrc,omitempty"` // Code ISRC pour identification

	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

type DiscWithDetails struct {
	ID          int      `json:"id"`
	Title       string   `json:"title"`
	ArtistID    int      `json:"artist_id"`
	ArtistName  string   `json:"artist_name"`
	GenreID     *int     `json:"genre_id,omitempty"`
	GenreName   *string  `json:"genre_name,omitempty"`
	FormatID    *int     `json:"format_id,omitempty"`
	FormatName  *string  `json:"format_name,omitempty"`
	CountryID   *int     `json:"country_id,omitempty"`
	CountryName *string  `json:"country_name,omitempty"`
	LabelID     *int     `json:"label_id,omitempty"`
	LabelName   *string  `json:"label_name,omitempty"`
	ReleaseYear *int     `json:"release_year,omitempty"`
	Barcode     *string  `json:"barcode,omitempty"`
	CoverURL    *string  `json:"cover_url,omitempty"`
	Notes       *string  `json:"notes,omitempty"`
	Price       *float64 `json:"price,omitempty"`
	Quantity    *int     `json:"quantity,omitempty"` // ✅ NOUVEAU: Nombre de disques

	// ✅ NOUVEAU: URLs de streaming
	AppleMusicURL *string `json:"apple_music_url,omitempty"`
	SpotifyURL    *string `json:"spotify_url,omitempty"`
	DeezerURL     *string `json:"deezer_url,omitempty"`
	YoutubeURL    *string `json:"youtube_url,omitempty"`
	ISRC          *string `json:"isrc,omitempty"` // Code ISRC pour identification

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
