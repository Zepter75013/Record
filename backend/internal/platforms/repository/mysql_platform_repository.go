package repository

import (
	"context"
	"database/sql"
	"records-manager/backend/internal/platforms"
)

type MySQLPlatformRepository struct {
	db *sql.DB
}

func NewMySQLPlatformRepository(db *sql.DB) *MySQLPlatformRepository {
	return &MySQLPlatformRepository{db: db}
}

func (r *MySQLPlatformRepository) Create(ctx context.Context, platform *platforms.Platform) error {
	query := `INSERT INTO games_platforms (name, description, created_at, updated_at)
              VALUES (?, ?, NOW(), NOW())`

	result, err := r.db.ExecContext(ctx, query, platform.Name, platform.Description)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	platform.ID = int(id)
	return nil
}

func (r *MySQLPlatformRepository) FindAll(ctx context.Context) ([]platforms.Platform, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM games_platforms ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var platformList []platforms.Platform
	for rows.Next() {
		var p platforms.Platform
		err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		platformList = append(platformList, p)
	}

	return platformList, nil
}

func (r *MySQLPlatformRepository) FindByID(ctx context.Context, id int) (*platforms.Platform, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM games_platforms WHERE id = ?`

	row := r.db.QueryRowContext(ctx, query, id)
	var p platforms.Platform
	err := row.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *MySQLPlatformRepository) Update(ctx context.Context, platform *platforms.Platform) error {
	query := `UPDATE games_platforms SET name = ?, description = ?, updated_at = NOW() WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, platform.Name, platform.Description, platform.ID)
	return err
}

func (r *MySQLPlatformRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM games_platforms WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *MySQLPlatformRepository) FindByName(ctx context.Context, name string) (*platforms.Platform, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM games_platforms WHERE name = ?`

	row := r.db.QueryRowContext(ctx, query, name)
	var p platforms.Platform
	err := row.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &p, nil
}

// CreateIfNotExists crée une plateforme seulement si elle n'existe pas déjà
func (r *MySQLPlatformRepository) CreateIfNotExists(ctx context.Context, name, description string) (*platforms.Platform, error) {
	existing, err := r.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		return existing, nil
	}

	platform := &platforms.Platform{
		Name:        name,
		Description: description,
	}

	err = r.Create(ctx, platform)
	if err != nil {
		return nil, err
	}

	return platform, nil
}
