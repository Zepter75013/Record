package repository

import (
	"database/sql"
	"fmt"
	"records-manager/backend/internal/users/domain"
)

type UserRepository interface {
	FindByEmail(email string) (*domain.User, error)
	FindByID(userID uint) (*domain.User, error)
	AddUser(user *domain.User) error
	UpdatePassword(userID uint, passwordHash string) error
	UpdateAvatar(userID uint, avatarPath *string) error
}

type mysqlUserRepo struct {
	DB *sql.DB
}

func NewMySQLUserRepository(db *sql.DB) UserRepository {
	return &mysqlUserRepo{DB: db}
}

func (r *mysqlUserRepo) FindByEmail(email string) (*domain.User, error) {
	user := &domain.User{}
	query := "SELECT id, email, password_hash, avatar_path FROM common_users WHERE email = ?"
	err := r.DB.QueryRow(query, email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.AvatarPath)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erreur de requête FindByEmail: %w", err)
	}
	return user, nil
}

func (r *mysqlUserRepo) FindByID(userID uint) (*domain.User, error) {
	user := &domain.User{}
	query := "SELECT id, email, password_hash, avatar_path FROM common_users WHERE id = ?"
	err := r.DB.QueryRow(query, userID).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.AvatarPath)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erreur de requête FindByID: %w", err)
	}
	return user, nil
}

func (r *mysqlUserRepo) AddUser(user *domain.User) error {
	query := "INSERT INTO common_users (email, password_hash) VALUES (?, ?)"
	_, err := r.DB.Exec(query, user.Email, user.PasswordHash)
	return err
}

func (r *mysqlUserRepo) UpdatePassword(userID uint, passwordHash string) error {
	query := "UPDATE common_users SET password_hash = ?, updated_at = NOW() WHERE id = ?"
	result, err := r.DB.Exec(query, passwordHash, userID)
	if err != nil {
		return fmt.Errorf("erreur lors de la mise à jour du mot de passe: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur lors de la vérification de la mise à jour: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("aucun utilisateur trouvé avec cet ID")
	}

	return nil
}

func (r *mysqlUserRepo) UpdateAvatar(userID uint, avatarPath *string) error {
	query := "UPDATE common_users SET avatar_path = ?, updated_at = NOW() WHERE id = ?"
	result, err := r.DB.Exec(query, avatarPath, userID)
	if err != nil {
		return fmt.Errorf("erreur lors de la mise à jour de l'avatar: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur lors de la vérification de la mise à jour: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("aucun utilisateur trouvé avec cet ID")
	}

	return nil
}
