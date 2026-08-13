package repository

import (
	"context"
	"records-manager/backend/internal/genres"
)

type GenreRepository interface {
	Create(ctx context.Context, genre *genres.Genre) error
	FindAll(ctx context.Context) ([]genres.Genre, error)
	FindByID(ctx context.Context, id int) (*genres.Genre, error)
	FindByName(ctx context.Context, name string) (*genres.Genre, error)
	Update(ctx context.Context, genre *genres.Genre) error
	Delete(ctx context.Context, id int) error
}
