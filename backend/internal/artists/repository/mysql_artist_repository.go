package repository

import (
	"context"
	"database/sql"
	"records-manager/backend/internal/artists"
)

type MySQLArtistRepository struct {
	db *sql.DB
}

func NewMySQLArtistRepository(db *sql.DB) *MySQLArtistRepository {
	return &MySQLArtistRepository{db: db}
}

func (r *MySQLArtistRepository) Create(ctx context.Context, artist *artists.Artist) error {
	query := `INSERT INTO artists (name, biography, created_at, updated_at) 
              VALUES (?, ?, NOW(), NOW())`

	result, err := r.db.ExecContext(ctx, query, artist.Name, artist.Biography)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	artist.ID = int(id)
	return nil
}

func (r *MySQLArtistRepository) FindAll(ctx context.Context) ([]artists.Artist, error) {
	query := `SELECT id, name, biography, created_at, updated_at FROM artists ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var artistList []artists.Artist
	for rows.Next() {
		var a artists.Artist
		err := rows.Scan(&a.ID, &a.Name, &a.Biography, &a.CreatedAt, &a.UpdatedAt)
		if err != nil {
			return nil, err
		}
		artistList = append(artistList, a)
	}

	return artistList, nil
}

func (r *MySQLArtistRepository) FindByID(ctx context.Context, id int) (*artists.Artist, error) {
	query := `SELECT id, name, biography, created_at, updated_at FROM artists WHERE id = ?`

	row := r.db.QueryRowContext(ctx, query, id)
	var a artists.Artist
	err := row.Scan(&a.ID, &a.Name, &a.Biography, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &a, nil
}

func (r *MySQLArtistRepository) Update(ctx context.Context, artist *artists.Artist) error {
	query := `UPDATE artists SET name = ?, biography = ?, updated_at = NOW() WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, artist.Name, artist.Biography, artist.ID)
	return err
}

func (r *MySQLArtistRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM artists WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *MySQLArtistRepository) FindByName(ctx context.Context, name string) (*artists.Artist, error) {
	query := `SELECT id, name, biography, created_at, updated_at FROM artists WHERE name = ?`

	row := r.db.QueryRowContext(ctx, query, name)
	var a artists.Artist
	err := row.Scan(&a.ID, &a.Name, &a.Biography, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &a, nil
}
