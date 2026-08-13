// records-manager/backend/internal/labels/repository/mysql_label_repository.go
package repository

import (
	"context"
	"database/sql"
	"records-manager/backend/internal/labels"
)

type MySQLLabelRepository struct {
	db *sql.DB
}

func NewMySQLLabelRepository(db *sql.DB) *MySQLLabelRepository {
	return &MySQLLabelRepository{db: db}
}

func (r *MySQLLabelRepository) Create(ctx context.Context, label *labels.Label) error {
	query := `INSERT INTO labels (name, description, created_at, updated_at) 
              VALUES (?, ?, NOW(), NOW())`

	result, err := r.db.ExecContext(ctx, query, label.Name, label.Description)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	label.ID = int(id)
	return nil
}

func (r *MySQLLabelRepository) FindAll(ctx context.Context) ([]labels.Label, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM labels ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var labelList []labels.Label
	for rows.Next() {
		var l labels.Label
		err := rows.Scan(&l.ID, &l.Name, &l.Description, &l.CreatedAt, &l.UpdatedAt)
		if err != nil {
			return nil, err
		}
		labelList = append(labelList, l)
	}

	return labelList, nil
}

func (r *MySQLLabelRepository) FindByID(ctx context.Context, id int) (*labels.Label, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM labels WHERE id = ?`

	row := r.db.QueryRowContext(ctx, query, id)
	var l labels.Label
	err := row.Scan(&l.ID, &l.Name, &l.Description, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &l, nil
}

func (r *MySQLLabelRepository) Update(ctx context.Context, label *labels.Label) error {
	query := `UPDATE labels SET name = ?, description = ?, updated_at = NOW() WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, label.Name, label.Description, label.ID)
	return err
}

func (r *MySQLLabelRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM labels WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *MySQLLabelRepository) FindByName(ctx context.Context, name string) (*labels.Label, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM labels WHERE name = ?`

	row := r.db.QueryRowContext(ctx, query, name)
	var l labels.Label
	err := row.Scan(&l.ID, &l.Name, &l.Description, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &l, nil
}
