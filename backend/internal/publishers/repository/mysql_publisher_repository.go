package repository

import (
	"context"
	"database/sql"
	"records-manager/backend/internal/publishers"
)

type MySQLPublisherRepository struct {
	db *sql.DB
}

func NewMySQLPublisherRepository(db *sql.DB) *MySQLPublisherRepository {
	return &MySQLPublisherRepository{db: db}
}

func (r *MySQLPublisherRepository) Create(ctx context.Context, publisher *publishers.Publisher) error {
	query := `INSERT INTO publishers (name, description, created_at, updated_at)
              VALUES (?, ?, NOW(), NOW())`

	result, err := r.db.ExecContext(ctx, query, publisher.Name, publisher.Description)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	publisher.ID = int(id)
	return nil
}

func (r *MySQLPublisherRepository) FindAll(ctx context.Context) ([]publishers.Publisher, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM publishers ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []publishers.Publisher
	for rows.Next() {
		var p publishers.Publisher
		err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, p)
	}

	return list, nil
}

func (r *MySQLPublisherRepository) FindByID(ctx context.Context, id int) (*publishers.Publisher, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM publishers WHERE id = ?`

	row := r.db.QueryRowContext(ctx, query, id)
	var p publishers.Publisher
	err := row.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *MySQLPublisherRepository) Update(ctx context.Context, publisher *publishers.Publisher) error {
	query := `UPDATE publishers SET name = ?, description = ?, updated_at = NOW() WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, publisher.Name, publisher.Description, publisher.ID)
	return err
}

func (r *MySQLPublisherRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM publishers WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *MySQLPublisherRepository) FindByName(ctx context.Context, name string) (*publishers.Publisher, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM publishers WHERE name = ?`

	row := r.db.QueryRowContext(ctx, query, name)
	var p publishers.Publisher
	err := row.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &p, nil
}

// CreateIfNotExists crée un éditeur seulement s'il n'existe pas déjà
func (r *MySQLPublisherRepository) CreateIfNotExists(ctx context.Context, name, description string) (*publishers.Publisher, error) {
	existing, err := r.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		return existing, nil
	}

	publisher := &publishers.Publisher{
		Name:        name,
		Description: description,
	}

	err = r.Create(ctx, publisher)
	if err != nil {
		return nil, err
	}

	return publisher, nil
}
