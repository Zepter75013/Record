package repository

import (
	"context"
	"database/sql"
	"records-manager/backend/internal/gamegenres"
)

type MySQLGameGenreRepository struct {
	db *sql.DB
}

func NewMySQLGameGenreRepository(db *sql.DB) *MySQLGameGenreRepository {
	return &MySQLGameGenreRepository{db: db}
}

func (r *MySQLGameGenreRepository) Create(ctx context.Context, genre *gamegenres.GameGenre) error {
	query := `INSERT INTO game_genres (name, description, created_at, updated_at)
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

func (r *MySQLGameGenreRepository) FindAll(ctx context.Context) ([]gamegenres.GameGenre, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM game_genres ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []gamegenres.GameGenre
	for rows.Next() {
		var g gamegenres.GameGenre
		err := rows.Scan(&g.ID, &g.Name, &g.Description, &g.CreatedAt, &g.UpdatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, g)
	}

	return list, nil
}

func (r *MySQLGameGenreRepository) FindByID(ctx context.Context, id int) (*gamegenres.GameGenre, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM game_genres WHERE id = ?`

	row := r.db.QueryRowContext(ctx, query, id)
	var g gamegenres.GameGenre
	err := row.Scan(&g.ID, &g.Name, &g.Description, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &g, nil
}

func (r *MySQLGameGenreRepository) Update(ctx context.Context, genre *gamegenres.GameGenre) error {
	query := `UPDATE game_genres SET name = ?, description = ?, updated_at = NOW() WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, genre.Name, genre.Description, genre.ID)
	return err
}

func (r *MySQLGameGenreRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM game_genres WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *MySQLGameGenreRepository) FindByName(ctx context.Context, name string) (*gamegenres.GameGenre, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM game_genres WHERE name = ?`

	row := r.db.QueryRowContext(ctx, query, name)
	var g gamegenres.GameGenre
	err := row.Scan(&g.ID, &g.Name, &g.Description, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &g, nil
}

// CreateIfNotExists crée un genre seulement s'il n'existe pas déjà
func (r *MySQLGameGenreRepository) CreateIfNotExists(ctx context.Context, name, description string) (*gamegenres.GameGenre, error) {
	existing, err := r.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		return existing, nil
	}

	genre := &gamegenres.GameGenre{
		Name:        name,
		Description: description,
	}

	err = r.Create(ctx, genre)
	if err != nil {
		return nil, err
	}

	return genre, nil
}
