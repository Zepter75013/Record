package repository

import (
	"context"
	"records-manager/backend/internal/tracks"
)

type TrackRepository interface {
	CreateBatch(ctx context.Context, vinylID int, list []tracks.Track) error
	FindByVinylID(ctx context.Context, vinylID int) ([]tracks.Track, error)
	DeleteByVinylID(ctx context.Context, vinylID int) error
}
