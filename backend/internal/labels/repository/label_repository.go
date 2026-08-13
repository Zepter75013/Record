// records-manager/backend/internal/labels/repository/label_repository.go
package repository

import (
	"context"
	"records-manager/backend/internal/labels"
)

type LabelRepository interface {
	Create(ctx context.Context, label *labels.Label) error
	FindAll(ctx context.Context) ([]labels.Label, error)
	FindByID(ctx context.Context, id int) (*labels.Label, error)
	FindByName(ctx context.Context, name string) (*labels.Label, error)
	Update(ctx context.Context, label *labels.Label) error
	Delete(ctx context.Context, id int) error
}
