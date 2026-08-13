package config

import (
	"os"
)

type DBConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	Name     string
}

// LoadDBConfig charge les configurations depuis les variables d'environnement
func LoadDBConfig() DBConfig {
	return DBConfig{
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Name:     os.Getenv("DB_NAME"),
	}
}

// GetJWTSecret récupère la clé secrète JWT
func GetJWTSecret() string {
	return os.Getenv("JWT_SECRET")
}

// ✅ NOUVEAU: Récupérer le token Discogs pour téléchargement des pochettes
func GetDiscogsToken() string {
	return os.Getenv("DISCOGS_TOKEN")
}

// ✅ NOUVEAU: Récupérer le répertoire des uploads
func GetUploadsDir() string {
	uploadsDir := os.Getenv("UPLOADS_DIR")
	if uploadsDir == "" {
		uploadsDir = "./uploads"
	}
	return uploadsDir
}
