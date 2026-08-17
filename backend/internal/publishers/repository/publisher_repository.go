package repository

import (
	"context"
	"records-manager/backend/internal/publishers"
)

type PublisherRepository interface {
	Create(ctx context.Context, publisher *publishers.Publisher) error
	FindAll(ctx context.Context) ([]publishers.Publisher, error)
	FindByID(ctx context.Context, id int) (*publishers.Publisher, error)
	FindByName(ctx context.Context, name string) (*publishers.Publisher, error)
	Update(ctx context.Context, publisher *publishers.Publisher) error
	Delete(ctx context.Context, id int) error
	// CreateIfNotExists crée un éditeur seulement s'il n'existe pas déjà
	CreateIfNotExists(ctx context.Context, name, description string) (*publishers.Publisher, error)
}
