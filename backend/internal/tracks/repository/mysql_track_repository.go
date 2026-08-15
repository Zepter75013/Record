package repository

import (
	"context"
	"database/sql"
	"records-manager/backend/internal/tracks"
)

type MySQLTrackRepository struct {
	db *sql.DB
}

func NewMySQLTrackRepository(db *sql.DB) *MySQLTrackRepository {
	return &MySQLTrackRepository{db: db}
}

func (r *MySQLTrackRepository) CreateBatch(ctx context.Context, vinylID int, list []tracks.Track) error {
	if len(list) == 0 {
		return nil
	}

	query := `INSERT INTO tracks (vinyl_id, track_order, position, title, duration) VALUES (?, ?, ?, ?, ?)`

	for i, t := range list {
		if _, err := r.db.ExecContext(ctx, query, vinylID, i, t.Position, t.Title, t.Duration); err != nil {
			return err
		}
	}

	return nil
}

func (r *MySQLTrackRepository) FindByVinylID(ctx context.Context, vinylID int) ([]tracks.Track, error) {
	query := `SELECT id, vinyl_id, track_order, position, title, duration FROM tracks WHERE vinyl_id = ? ORDER BY track_order`

	rows, err := r.db.QueryContext(ctx, query, vinylID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []tracks.Track
	for rows.Next() {
		var t tracks.Track
		if err := rows.Scan(&t.ID, &t.VinylID, &t.TrackOrder, &t.Position, &t.Title, &t.Duration); err != nil {
			return nil, err
		}
		list = append(list, t)
	}

	return list, nil
}

func (r *MySQLTrackRepository) DeleteByVinylID(ctx context.Context, vinylID int) error {
	query := `DELETE FROM tracks WHERE vinyl_id = ?`

	_, err := r.db.ExecContext(ctx, query, vinylID)
	return err
}
