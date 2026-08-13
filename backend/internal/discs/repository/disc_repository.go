package repository

import (
	"context"
	"records-manager/backend/internal/discs"
)

// DiscRepository définit le contrat que doit respecter tout repository de disques
type DiscRepository interface {
	Create(ctx context.Context, disc *discs.Disc) error
	FindAll(ctx context.Context) ([]discs.DiscWithDetails, error)
	FindByID(ctx context.Context, id int) (*discs.DiscWithDetails, error)
	FindByTitle(ctx context.Context, title string) (*discs.Disc, error)
	FindByBarcode(ctx context.Context, barcode string) (*discs.Disc, error)
	FindByTitleWithDetails(ctx context.Context, title string) (*discs.DiscWithDetails, error)
	FindByBarcodeWithDetails(ctx context.Context, barcode string) (*discs.DiscWithDetails, error)
	Update(ctx context.Context, disc *discs.Disc) error
	Delete(ctx context.Context, id int) error
}
