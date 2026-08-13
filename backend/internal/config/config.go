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
