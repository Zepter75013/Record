package repository

import (
	"context"
	"database/sql"
	"records-manager/backend/internal/genres"
)

type MySQLGenreRepository struct {
	db *sql.DB
}

func NewMySQLGenreRepository(db *sql.DB) *MySQLGenreRepository {
	return &MySQLGenreRepository{db: db}
}

func (r *MySQLGenreRepository) Create(ctx context.Context, genre *genres.Genre) error {
	query := `INSERT INTO genres (name, description, created_at, updated_at) 
              VALUES (?, ?, NOW(), NOW())`

	result, err := r.db.ExecContext(ctx, query, genre.Name, genre.Description)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	genre.ID = int(id)
	return nil
}

func (r *MySQLGenreRepository) FindAll(ctx context.Context) ([]genres.Genre, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM genres ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genreList []genres.Genre
	for rows.Next() {
		var g genres.Genre
		err := rows.Scan(&g.ID, &g.Name, &g.Description, &g.CreatedAt, &g.UpdatedAt)
		if err != nil {
			return nil, err
		}
		genreList = append(genreList, g)
	}

	return genreList, nil
}

func (r *MySQLGenreRepository) FindByID(ctx context.Context, id int) (*genres.Genre, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM genres WHERE id = ?`

	row := r.db.QueryRowContext(ctx, query, id)
	var g genres.Genre
	err := row.Scan(&g.ID, &g.Name, &g.Description, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &g, nil
}

func (r *MySQLGenreRepository) Update(ctx context.Context, genre *genres.Genre) error {
	query := `UPDATE genres SET name = ?, description = ?, updated_at = NOW() WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, genre.Name, genre.Description, genre.ID)
	return err
}

func (r *MySQLGenreRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM genres WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *MySQLGenreRepository) FindByName(ctx context.Context, name string) (*genres.Genre, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM genres WHERE name = ?`

	row := r.db.QueryRowContext(ctx, query, name)
	var g genres.Genre
	err := row.Scan(&g.ID, &g.Name, &g.Description, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &g, nil
}
