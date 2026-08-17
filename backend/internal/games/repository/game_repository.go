package repository

import (
	"context"
	"records-manager/backend/internal/games"
)

type GameRepository interface {
	Create(ctx context.Context, game *games.Game) error
	FindAll(ctx context.Context) ([]games.GameWithDetails, error)
	FindByID(ctx context.Context, id int) (*games.GameWithDetails, error)
	Update(ctx context.Context, game *games.Game) error
	Delete(ctx context.Context, id int) error
	FindByBarcode(ctx context.Context, barcode string) (*games.GameWithDetails, error)
}
