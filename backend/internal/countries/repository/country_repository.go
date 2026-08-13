package repository

import (
	"context"
	"records-manager/backend/internal/countries"
)

type CountryRepository interface {
	Create(ctx context.Context, country *countries.Country) error
	FindAll(ctx context.Context) ([]countries.Country, error)
	FindByID(ctx context.Context, id int) (*countries.Country, error)
	FindByName(ctx context.Context, name string) (*countries.Country, error)
	FindByCode(ctx context.Context, code string) (*countries.Country, error)
	Update(ctx context.Context, country *countries.Country) error
	Delete(ctx context.Context, id int) error
}
