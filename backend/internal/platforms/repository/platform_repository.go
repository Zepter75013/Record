package repository

import (
	"context"
	"records-manager/backend/internal/platforms"
)

type PlatformRepository interface {
	Create(ctx context.Context, platform *platforms.Platform) error
	FindAll(ctx context.Context) ([]platforms.Platform, error)
	FindByID(ctx context.Context, id int) (*platforms.Platform, error)
	FindByName(ctx context.Context, name string) (*platforms.Platform, error)
	Update(ctx context.Context, platform *platforms.Platform) error
	Delete(ctx context.Context, id int) error
	// CreateIfNotExists crée une plateforme seulement si elle n'existe pas déjà
	CreateIfNotExists(ctx context.Context, name, description string) (*platforms.Platform, error)
}
