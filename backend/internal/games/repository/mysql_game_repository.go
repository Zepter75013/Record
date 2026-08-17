package repository

import (
	"context"
	"database/sql"
	"records-manager/backend/internal/games"
)

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

type MySQLGameRepository struct {
	db *sql.DB
}

func NewMySQLGameRepository(db *sql.DB) *MySQLGameRepository {
	return &MySQLGameRepository{db: db}
}

const gameSelectWithDetails = `
	SELECT
		g.id,
		g.title,
		g.platform_id,
		p.name as platform_name,
		g.genre_id,
		gg.name as genre_name,
		g.publisher_id,
		pub.name as publisher_name,
		g.release_year,
		g.barcode,
		g.cover_url,
		g.notes,
		g.price,
		g.quantity,
		g.rawg_id,
		g.created_at,
		g.updated_at
	FROM games g
	LEFT JOIN platforms p ON g.platform_id = p.id
	LEFT JOIN game_genres gg ON g.genre_id = gg.id
	LEFT JOIN publishers pub ON g.publisher_id = pub.id`

func scanGameWithDetails(row interface{ Scan(dest ...any) error }) (*games.GameWithDetails, error) {
	var g games.GameWithDetails
	err := row.Scan(
		&g.ID,
		&g.Title,
		&g.PlatformID,
		&g.PlatformName,
		&g.GenreID,
		&g.GenreName,
		&g.PublisherID,
		&g.PublisherName,
		&g.ReleaseYear,
		&g.Barcode,
		&g.CoverURL,
		&g.Notes,
		&g.Price,
		&g.Quantity,
		&g.RAWGID,
		&g.CreatedAt,
		&g.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *MySQLGameRepository) Create(ctx context.Context, game *games.Game) error {
	query := `
		INSERT INTO games (
			title, platform_id, genre_id, publisher_id, release_year, barcode, cover_url, notes, price, quantity, rawg_id,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`

	result, err := r.db.ExecContext(ctx, query,
		game.Title,
		game.PlatformID,
		intPtrToNullInt64(game.GenreID),
		intPtrToNullInt64(game.PublisherID),
		intPtrToNullInt64(game.ReleaseYear),
		stringPtrToNullString(game.Barcode),
		stringPtrToNullString(game.CoverURL),
		stringPtrToNullString(game.Notes),
		float64PtrToNullFloat64(game.Price),
		intPtrToNullInt64(game.Quantity),
		int64PtrToNullInt64(game.RAWGID),
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	idInt := int(id)
	game.ID = &idInt
	return nil
}

func (r *MySQLGameRepository) FindAll(ctx context.Context) ([]games.GameWithDetails, error) {
	query := gameSelectWithDetails + ` ORDER BY g.title`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []games.GameWithDetails
	for rows.Next() {
		g, err := scanGameWithDetails(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *g)
	}

	return list, nil
}

func (r *MySQLGameRepository) FindByID(ctx context.Context, id int) (*games.GameWithDetails, error) {
	query := gameSelectWithDetails + ` WHERE g.id = ?`

	row := r.db.QueryRowContext(ctx, query, id)
	g, err := scanGameWithDetails(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return g, nil
}

func (r *MySQLGameRepository) FindByBarcode(ctx context.Context, barcode string) (*games.GameWithDetails, error) {
	query := gameSelectWithDetails + ` WHERE g.barcode = ?`

	row := r.db.QueryRowContext(ctx, query, barcode)
	g, err := scanGameWithDetails(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return g, nil
}

func (r *MySQLGameRepository) Update(ctx context.Context, game *games.Game) error {
	query := `
		UPDATE games SET
			title = ?,
			platform_id = ?,
			genre_id = ?,
			publisher_id = ?,
			release_year = ?,
			barcode = ?,
			cover_url = ?,
			notes = ?,
			price = ?,
			quantity = ?,
			updated_at = NOW()
		WHERE id = ?`

	idVal := *game.ID

	_, err := r.db.ExecContext(ctx, query,
		game.Title,
		game.PlatformID,
		intPtrToNullInt64(game.GenreID),
		intPtrToNullInt64(game.PublisherID),
		intPtrToNullInt64(game.ReleaseYear),
		stringPtrToNullString(game.Barcode),
		stringPtrToNullString(game.CoverURL),
		stringPtrToNullString(game.Notes),
		float64PtrToNullFloat64(game.Price),
		intPtrToNullInt64(game.Quantity),
		idVal,
	)

	return err
}

func (r *MySQLGameRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM games WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
