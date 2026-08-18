package repository

import (
	"database/sql"
	"time"
)

type PasswordResetRepository interface {
	CreateResetToken(userID uint, token string, expiresAt time.Time) error
	GetResetToken(token string) (userID uint, valid bool, err error)
	MarkTokenAsUsed(token string) error
	DeleteExpiredTokens() error
}

type passwordResetRepo struct {
	DB *sql.DB
}

func NewPasswordResetRepository(db *sql.DB) PasswordResetRepository {
	return &passwordResetRepo{DB: db}
}

func (r *passwordResetRepo) CreateResetToken(userID uint, token string, expiresAt time.Time) error {
	query := `INSERT INTO common_password_reset_tokens (user_id, token, expires_at) VALUES (?, ?, ?)`
	_, err := r.DB.Exec(query, userID, token, expiresAt)
	return err
}

func (r *passwordResetRepo) GetResetToken(token string) (uint, bool, error) {
	var userID uint
	var expiresAt time.Time
	var used bool

	query := `SELECT user_id, expires_at, used FROM common_password_reset_tokens WHERE token = ?`
	err := r.DB.QueryRow(query, token).Scan(&userID, &expiresAt, &used)

	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}

	// Vérifier si le token est valide (non utilisé et non expiré)
	if used || time.Now().After(expiresAt) {
		return userID, false, nil
	}

	return userID, true, nil
}

func (r *passwordResetRepo) MarkTokenAsUsed(token string) error {
	query := `UPDATE common_password_reset_tokens SET used = TRUE WHERE token = ?`
	_, err := r.DB.Exec(query, token)
	return err
}

func (r *passwordResetRepo) DeleteExpiredTokens() error {
	query := `DELETE FROM common_password_reset_tokens WHERE expires_at < NOW() OR used = TRUE`
	_, err := r.DB.Exec(query)
	return err
}
