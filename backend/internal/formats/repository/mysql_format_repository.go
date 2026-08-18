package repository

import (
	"context"
	"database/sql"
	"records-manager/backend/internal/formats"
)

type MySQLFormatRepository struct {
	db *sql.DB
}

func NewMySQLFormatRepository(db *sql.DB) *MySQLFormatRepository {
	return &MySQLFormatRepository{db: db}
}

func (r *MySQLFormatRepository) Create(ctx context.Context, format *formats.Format) error {
	query := `INSERT INTO records_formats (name, description, created_at, updated_at) 
              VALUES (?, ?, NOW(), NOW())`

	result, err := r.db.ExecContext(ctx, query, format.Name, format.Description)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	format.ID = int(id)
	return nil
}

func (r *MySQLFormatRepository) FindAll(ctx context.Context) ([]formats.Format, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM records_formats ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var formatList []formats.Format
	for rows.Next() {
		var f formats.Format
		err := rows.Scan(&f.ID, &f.Name, &f.Description, &f.CreatedAt, &f.UpdatedAt)
		if err != nil {
			return nil, err
		}
		formatList = append(formatList, f)
	}

	return formatList, nil
}

func (r *MySQLFormatRepository) FindByID(ctx context.Context, id int) (*formats.Format, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM records_formats WHERE id = ?`

	row := r.db.QueryRowContext(ctx, query, id)
	var f formats.Format
	err := row.Scan(&f.ID, &f.Name, &f.Description, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &f, nil
}

func (r *MySQLFormatRepository) Update(ctx context.Context, format *formats.Format) error {
	query := `UPDATE records_formats SET name = ?, description = ?, updated_at = NOW() WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, format.Name, format.Description, format.ID)
	return err
}

func (r *MySQLFormatRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM records_formats WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *MySQLFormatRepository) FindByName(ctx context.Context, name string) (*formats.Format, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM records_formats WHERE name = ?`

	row := r.db.QueryRowContext(ctx, query, name)
	var f formats.Format
	err := row.Scan(&f.ID, &f.Name, &f.Description, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &f, nil
}

// CreateIfNotExists crée un format seulement s'il n'existe pas déjà
func (r *MySQLFormatRepository) CreateIfNotExists(ctx context.Context, name, description string) (*formats.Format, error) {
	// Vérifier si le format existe déjà
	existing, err := r.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}

	// S'il existe, le retourner
	if existing != nil {
		return existing, nil
	}

	// Sinon, le créer
	format := &formats.Format{
		Name:        name,
		Description: description,
	}

	err = r.Create(ctx, format)
	if err != nil {
		return nil, err
	}

	return format, nil
}
