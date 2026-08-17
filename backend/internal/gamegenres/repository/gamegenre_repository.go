package repository

import (
	"context"
	"records-manager/backend/internal/gamegenres"
)

type GameGenreRepository interface {
	Create(ctx context.Context, genre *gamegenres.GameGenre) error
	FindAll(ctx context.Context) ([]gamegenres.GameGenre, error)
	FindByID(ctx context.Context, id int) (*gamegenres.GameGenre, error)
	FindByName(ctx context.Context, name string) (*gamegenres.GameGenre, error)
	Update(ctx context.Context, genre *gamegenres.GameGenre) error
	Delete(ctx context.Context, id int) error
	// CreateIfNotExists crée un genre seulement s'il n'existe pas déjà
	CreateIfNotExists(ctx context.Context, name, description string) (*gamegenres.GameGenre, error)
}
