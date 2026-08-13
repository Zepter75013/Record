package repository

import (
	"context"
	"records-manager/backend/internal/artists"
)

// ArtistRepository définit le contrat que doit respecter tout repository d'artistes
type ArtistRepository interface {
	Create(ctx context.Context, artist *artists.Artist) error
	FindAll(ctx context.Context) ([]artists.Artist, error)
	FindByID(ctx context.Context, id int) (*artists.Artist, error)
	FindByName(ctx context.Context, name string) (*artists.Artist, error)
	Update(ctx context.Context, artist *artists.Artist) error
	Delete(ctx context.Context, id int) error
}
