package repository

import (
	"database/sql"
)

type mysqlStatsRepository struct {
	db *sql.DB
}

func NewMySQLStatsRepository(db *sql.DB) StatsRepository {
	return &mysqlStatsRepository{db: db}
}

func (r *mysqlStatsRepository) GetCounts() (map[string]int, error) {
	stats := make(map[string]int)

	// Compter les artistes
	var artistCount int
	err := r.db.QueryRow("SELECT COUNT(*) FROM records_artists").Scan(&artistCount)
	if err != nil {
		return nil, err
	}
	stats["artists"] = artistCount

	// Compter les genres
	var genreCount int
	err = r.db.QueryRow("SELECT COUNT(*) FROM records_genres").Scan(&genreCount)
	if err != nil {
		return nil, err
	}
	stats["genres"] = genreCount

	// Compter les formats
	var formatCount int
	err = r.db.QueryRow("SELECT COUNT(*) FROM records_formats").Scan(&formatCount)
	if err != nil {
		return nil, err
	}
	stats["formats"] = formatCount

	// Compter les pays
	var countryCount int
	err = r.db.QueryRow("SELECT COUNT(*) FROM records_countries").Scan(&countryCount)
	if err != nil {
		return nil, err
	}
	stats["countries"] = countryCount

	// Compter les labels
	var labelCount int
	err = r.db.QueryRow("SELECT COUNT(*) FROM records_labels").Scan(&labelCount)
	if err != nil {
		return nil, err
	}
	stats["labels"] = labelCount

	// Compter les vinyls (si la table existe)
	var vinylCount int
	err = r.db.QueryRow("SELECT COUNT(*) FROM records_vinyls").Scan(&vinylCount)
	if err != nil {
		// Si la table n'existe pas encore, mettre 0
		stats["vinyls"] = 0
	} else {
		stats["vinyls"] = vinylCount
	}

	return stats, nil
}
