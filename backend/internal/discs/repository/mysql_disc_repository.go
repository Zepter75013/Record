package repository

import (
	"context"
	"database/sql"
	"records-manager/backend/internal/discs"
)

// Helpers
func intPtrToNullInt64(ptr *int) sql.NullInt64 {
	if ptr == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*ptr), Valid: true}
}

func int64PtrToNullInt64(ptr *int64) sql.NullInt64 {
	if ptr == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *ptr, Valid: true}
}

func stringPtrToNullString(ptr *string) sql.NullString {
	if ptr == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *ptr, Valid: true}
}

func float64PtrToNullFloat64(ptr *float64) sql.NullFloat64 {
	if ptr == nil {
		return sql.NullFloat64{}
	}
	return sql.NullFloat64{Float64: *ptr, Valid: true}
}

type MySQLDiscRepository struct {
	db *sql.DB
}

func NewMySQLDiscRepository(db *sql.DB) *MySQLDiscRepository {
	return &MySQLDiscRepository{db: db}
}

func (r *MySQLDiscRepository) Create(ctx context.Context, disc *discs.Disc) error {
	query := `
		INSERT INTO records_vinyls (
			title, artist_id, genre_id, format_id, country_id, label_id, release_year, barcode, cover_url, notes, price, quantity,
			apple_music_url, spotify_url, deezer_url, youtube_url, isrc, discogs_release_id,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`

	result, err := r.db.ExecContext(ctx, query,
		disc.Title,
		disc.ArtistID,
		intPtrToNullInt64(disc.GenreID),
		intPtrToNullInt64(disc.FormatID),
		intPtrToNullInt64(disc.CountryID),
		intPtrToNullInt64(disc.LabelID),
		intPtrToNullInt64(disc.ReleaseYear),
		stringPtrToNullString(disc.Barcode),
		stringPtrToNullString(disc.CoverURL),
		stringPtrToNullString(disc.Notes),
		float64PtrToNullFloat64(disc.Price),
		intPtrToNullInt64(disc.Quantity),
		// ✅ NOUVEAU: Champs de streaming
		stringPtrToNullString(disc.AppleMusicURL),
		stringPtrToNullString(disc.SpotifyURL),
		stringPtrToNullString(disc.DeezerURL),
		stringPtrToNullString(disc.YoutubeURL),
		stringPtrToNullString(disc.ISRC),
		int64PtrToNullInt64(disc.DiscogsReleaseID),
	)

	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	idInt := int(id)
	disc.ID = &idInt
	return nil
}

func (r *MySQLDiscRepository) FindAll(ctx context.Context) ([]discs.DiscWithDetails, error) {
	query := `
		SELECT
			v.id,
			v.title,
			v.artist_id,
			a.name as artist_name,
			v.genre_id,
			g.name as genre_name,
			v.format_id,
			f.name as format_name,
			v.country_id,
			c.name as country_name,
			v.label_id,
			l.name as label_name,
			v.release_year,
			v.barcode,
			v.cover_url,
			v.notes,
			v.price,
			v.quantity,
			v.apple_music_url,
			v.spotify_url,
			v.deezer_url,
			v.youtube_url,
			v.isrc,
			EXISTS(SELECT 1 FROM records_tracks t WHERE t.vinyl_id = v.id) as has_tracks,
			v.created_at,
			v.updated_at
		FROM records_vinyls v
		LEFT JOIN records_artists a ON v.artist_id = a.id
		LEFT JOIN records_genres g ON v.genre_id = g.id
		LEFT JOIN records_formats f ON v.format_id = f.id
		LEFT JOIN records_countries c ON v.country_id = c.id
		LEFT JOIN records_labels l ON v.label_id = l.id
		ORDER BY v.title`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var discList []discs.DiscWithDetails
	for rows.Next() {
		var d discs.DiscWithDetails
		err := rows.Scan(
			&d.ID,
			&d.Title,
			&d.ArtistID,
			&d.ArtistName,
			&d.GenreID,
			&d.GenreName,
			&d.FormatID,
			&d.FormatName,
			&d.CountryID,
			&d.CountryName,
			&d.LabelID,
			&d.LabelName,
			&d.ReleaseYear,
			&d.Barcode,
			&d.CoverURL,
			&d.Notes,
			&d.Price,
			&d.Quantity,
			// ✅ NOUVEAU: Champs de streaming
			&d.AppleMusicURL,
			&d.SpotifyURL,
			&d.DeezerURL,
			&d.YoutubeURL,
			&d.ISRC,
			&d.HasTracks,
			&d.CreatedAt,
			&d.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		discList = append(discList, d)
	}

	return discList, nil
}

func (r *MySQLDiscRepository) FindByID(ctx context.Context, id int) (*discs.DiscWithDetails, error) {
	query := `
		SELECT
			v.id,
			v.title,
			v.artist_id,
			a.name as artist_name,
			v.genre_id,
			g.name as genre_name,
			v.format_id,
			f.name as format_name,
			v.country_id,
			c.name as country_name,
			v.label_id,
			l.name as label_name,
			v.release_year,
			v.barcode,
			v.cover_url,
			v.notes,
			v.price,
			v.quantity,
			v.apple_music_url,
			v.spotify_url,
			v.deezer_url,
			v.youtube_url,
			v.isrc,
			EXISTS(SELECT 1 FROM records_tracks t WHERE t.vinyl_id = v.id) as has_tracks,
			v.created_at,
			v.updated_at
		FROM records_vinyls v
		LEFT JOIN records_artists a ON v.artist_id = a.id
		LEFT JOIN records_genres g ON v.genre_id = g.id
		LEFT JOIN records_formats f ON v.format_id = f.id
		LEFT JOIN records_countries c ON v.country_id = c.id
		LEFT JOIN records_labels l ON v.label_id = l.id
		WHERE v.id = ?`

	row := r.db.QueryRowContext(ctx, query, id)

	var d discs.DiscWithDetails
	err := row.Scan(
		&d.ID,
		&d.Title,
		&d.ArtistID,
		&d.ArtistName,
		&d.GenreID,
		&d.GenreName,
		&d.FormatID,
		&d.FormatName,
		&d.CountryID,
		&d.CountryName,
		&d.LabelID,
		&d.LabelName,
		&d.ReleaseYear,
		&d.Barcode,
		&d.CoverURL,
		&d.Notes,
		&d.Price,
		&d.Quantity,
		// ✅ NOUVEAU: Champs de streaming
		&d.AppleMusicURL,
		&d.SpotifyURL,
		&d.DeezerURL,
		&d.YoutubeURL,
		&d.ISRC,
		&d.HasTracks,
		&d.CreatedAt,
		&d.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &d, nil
}

func (r *MySQLDiscRepository) Update(ctx context.Context, disc *discs.Disc) error {
	query := `
		UPDATE records_vinyls SET
			title = ?,
			artist_id = ?,
			genre_id = ?,
			format_id = ?,
			country_id = ?,
			label_id = ?,
			release_year = ?,
			barcode = ?,
			cover_url = ?,
			notes = ?,
			price = ?,
			quantity = ?,
			apple_music_url = ?,
			spotify_url = ?,
			deezer_url = ?,
			youtube_url = ?,
			isrc = ?,
			updated_at = NOW()
		WHERE id = ?`

	idVal := *disc.ID

	_, err := r.db.ExecContext(ctx, query,
		disc.Title,
		disc.ArtistID,
		intPtrToNullInt64(disc.GenreID),
		intPtrToNullInt64(disc.FormatID),
		intPtrToNullInt64(disc.CountryID),
		intPtrToNullInt64(disc.LabelID),
		intPtrToNullInt64(disc.ReleaseYear),
		stringPtrToNullString(disc.Barcode),
		stringPtrToNullString(disc.CoverURL),
		stringPtrToNullString(disc.Notes),
		float64PtrToNullFloat64(disc.Price),
		intPtrToNullInt64(disc.Quantity),
		// ✅ NOUVEAU: Champs de streaming
		stringPtrToNullString(disc.AppleMusicURL),
		stringPtrToNullString(disc.SpotifyURL),
		stringPtrToNullString(disc.DeezerURL),
		stringPtrToNullString(disc.YoutubeURL),
		stringPtrToNullString(disc.ISRC),
		idVal,
	)

	return err
}

func (r *MySQLDiscRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM records_vinyls WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *MySQLDiscRepository) FindByTitle(ctx context.Context, title string) (*discs.Disc, error) {
	query := `
		SELECT id, title, artist_id, genre_id, format_id, country_id, label_id, release_year, barcode, cover_url, notes, price, quantity,
		       apple_music_url, spotify_url, deezer_url, youtube_url, isrc, created_at, updated_at
		FROM records_vinyls WHERE title = ?`

	row := r.db.QueryRowContext(ctx, query, title)

	var d discs.Disc
	err := row.Scan(
		&d.ID,
		&d.Title,
		&d.ArtistID,
		&d.GenreID,
		&d.FormatID,
		&d.CountryID,
		&d.LabelID,
		&d.ReleaseYear,
		&d.Barcode,
		&d.CoverURL,
		&d.Notes,
		&d.Price,
		&d.Quantity,
		// ✅ NOUVEAU: Champs de streaming
		&d.AppleMusicURL,
		&d.SpotifyURL,
		&d.DeezerURL,
		&d.YoutubeURL,
		&d.ISRC,
		&d.CreatedAt,
		&d.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &d, nil
}

func (r *MySQLDiscRepository) FindByBarcode(ctx context.Context, barcode string) (*discs.Disc, error) {
	query := `
		SELECT id, title, artist_id, genre_id, format_id, country_id, label_id, release_year, barcode, cover_url, notes, price, quantity,
		       apple_music_url, spotify_url, deezer_url, youtube_url, isrc, created_at, updated_at
		FROM records_vinyls WHERE barcode = ?`

	row := r.db.QueryRowContext(ctx, query, barcode)

	var d discs.Disc
	err := row.Scan(
		&d.ID,
		&d.Title,
		&d.ArtistID,
		&d.GenreID,
		&d.FormatID,
		&d.CountryID,
		&d.LabelID,
		&d.ReleaseYear,
		&d.Barcode,
		&d.CoverURL,
		&d.Notes,
		&d.Price,
		&d.Quantity,
		// ✅ NOUVEAU: Champs de streaming
		&d.AppleMusicURL,
		&d.SpotifyURL,
		&d.DeezerURL,
		&d.YoutubeURL,
		&d.ISRC,
		&d.CreatedAt,
		&d.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &d, nil
}

func (r *MySQLDiscRepository) FindByTitleWithDetails(ctx context.Context, title string) (*discs.DiscWithDetails, error) {
	query := `
		SELECT
			v.id,
			v.title,
			v.artist_id,
			a.name as artist_name,
			v.genre_id,
			g.name as genre_name,
			v.format_id,
			f.name as format_name,
			v.country_id,
			c.name as country_name,
			v.label_id,
			l.name as label_name,
			v.release_year,
			v.barcode,
			v.cover_url,
			v.notes,
			v.price,
			v.quantity,
			v.apple_music_url,
			v.spotify_url,
			v.deezer_url,
			v.youtube_url,
			v.isrc,
			v.created_at,
			v.updated_at
		FROM records_vinyls v
		LEFT JOIN records_artists a ON v.artist_id = a.id
		LEFT JOIN records_genres g ON v.genre_id = g.id
		LEFT JOIN records_formats f ON v.format_id = f.id
		LEFT JOIN records_countries c ON v.country_id = c.id
		LEFT JOIN records_labels l ON v.label_id = l.id
		WHERE v.title = ?`

	row := r.db.QueryRowContext(ctx, query, title)

	var d discs.DiscWithDetails
	err := row.Scan(
		&d.ID,
		&d.Title,
		&d.ArtistID,
		&d.ArtistName,
		&d.GenreID,
		&d.GenreName,
		&d.FormatID,
		&d.FormatName,
		&d.CountryID,
		&d.CountryName,
		&d.LabelID,
		&d.LabelName,
		&d.ReleaseYear,
		&d.Barcode,
		&d.CoverURL,
		&d.Notes,
		&d.Price,
		&d.Quantity,
		// ✅ NOUVEAU: Champs de streaming
		&d.AppleMusicURL,
		&d.SpotifyURL,
		&d.DeezerURL,
		&d.YoutubeURL,
		&d.ISRC,
		&d.CreatedAt,
		&d.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &d, nil
}

func (r *MySQLDiscRepository) FindByBarcodeWithDetails(ctx context.Context, barcode string) (*discs.DiscWithDetails, error) {
	query := `
		SELECT
			v.id,
			v.title,
			v.artist_id,
			a.name as artist_name,
			v.genre_id,
			g.name as genre_name,
			v.format_id,
			f.name as format_name,
			v.country_id,
			c.name as country_name,
			v.label_id,
			l.name as label_name,
			v.release_year,
			v.barcode,
			v.cover_url,
			v.notes,
			v.price,
			v.quantity,
			v.apple_music_url,
			v.spotify_url,
			v.deezer_url,
			v.youtube_url,
			v.isrc,
			v.created_at,
			v.updated_at
		FROM records_vinyls v
		LEFT JOIN records_artists a ON v.artist_id = a.id
		LEFT JOIN records_genres g ON v.genre_id = g.id
		LEFT JOIN records_formats f ON v.format_id = f.id
		LEFT JOIN records_countries c ON v.country_id = c.id
		LEFT JOIN records_labels l ON v.label_id = l.id
		WHERE v.barcode = ?`

	row := r.db.QueryRowContext(ctx, query, barcode)

	var d discs.DiscWithDetails
	err := row.Scan(
		&d.ID,
		&d.Title,
		&d.ArtistID,
		&d.ArtistName,
		&d.GenreID,
		&d.GenreName,
		&d.FormatID,
		&d.FormatName,
		&d.CountryID,
		&d.CountryName,
		&d.LabelID,
		&d.LabelName,
		&d.ReleaseYear,
		&d.Barcode,
		&d.CoverURL,
		&d.Notes,
		&d.Price,
		&d.Quantity,
		// ✅ NOUVEAU: Champs de streaming
		&d.AppleMusicURL,
		&d.SpotifyURL,
		&d.DeezerURL,
		&d.YoutubeURL,
		&d.ISRC,
		&d.CreatedAt,
		&d.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &d, nil
}
