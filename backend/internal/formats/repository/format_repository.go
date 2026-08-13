package repository

import (
	"context"
	"records-manager/backend/internal/formats"
)

type FormatRepository interface {
	Create(ctx context.Context, format *formats.Format) error
	FindAll(ctx context.Context) ([]formats.Format, error)
	FindByID(ctx context.Context, id int) (*formats.Format, error)
	FindByName(ctx context.Context, name string) (*formats.Format, error)
	Update(ctx context.Context, format *formats.Format) error
	Delete(ctx context.Context, id int) error
	// CreateIfNotExists crée un format seulement s'il n'existe pas déjà
	CreateIfNotExists(ctx context.Context, name, description string) (*formats.Format, error)
}
