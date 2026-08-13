package repository

import (
	"context"
	"database/sql"
	"records-manager/backend/internal/countries"
)

type MySQLCountryRepository struct {
	db *sql.DB
}

func NewMySQLCountryRepository(db *sql.DB) *MySQLCountryRepository {
	return &MySQLCountryRepository{db: db}
}

func (r *MySQLCountryRepository) Create(ctx context.Context, country *countries.Country) error {
	query := `INSERT INTO countries (name, code, description, created_at, updated_at) 
              VALUES (?, ?, ?, NOW(), NOW())`

	result, err := r.db.ExecContext(ctx, query, country.Name, country.Code, country.Description)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	country.ID = int(id)
	return nil
}

func (r *MySQLCountryRepository) FindAll(ctx context.Context) ([]countries.Country, error) {
	query := `SELECT id, name, code, description, created_at, updated_at FROM countries ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var countryList []countries.Country
	for rows.Next() {
		var c countries.Country
		err := rows.Scan(&c.ID, &c.Name, &c.Code, &c.Description, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, err
		}
		countryList = append(countryList, c)
	}

	return countryList, nil
}

func (r *MySQLCountryRepository) FindByID(ctx context.Context, id int) (*countries.Country, error) {
	query := `SELECT id, name, code, description, created_at, updated_at FROM countries WHERE id = ?`

	row := r.db.QueryRowContext(ctx, query, id)
	var c countries.Country
	err := row.Scan(&c.ID, &c.Name, &c.Code, &c.Description, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (r *MySQLCountryRepository) FindByName(ctx context.Context, name string) (*countries.Country, error) {
	query := `SELECT id, name, code, description, created_at, updated_at FROM countries WHERE name = ?`

	row := r.db.QueryRowContext(ctx, query, name)
	var c countries.Country
	err := row.Scan(&c.ID, &c.Name, &c.Code, &c.Description, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &c, nil
}

func (r *MySQLCountryRepository) FindByCode(ctx context.Context, code string) (*countries.Country, error) {
	query := `SELECT id, name, code, description, created_at, updated_at FROM countries WHERE code = ?`

	row := r.db.QueryRowContext(ctx, query, code)
	var c countries.Country
	err := row.Scan(&c.ID, &c.Name, &c.Code, &c.Description, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &c, nil
}

func (r *MySQLCountryRepository) Update(ctx context.Context, country *countries.Country) error {
	query := `UPDATE countries SET name = ?, code = ?, description = ?, updated_at = NOW() WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, country.Name, country.Code, country.Description, country.ID)
	return err
}

func (r *MySQLCountryRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM countries WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
