package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"records-manager/backend/internal/email"
	"records-manager/backend/internal/users/repository"

	"golang.org/x/crypto/bcrypt"
)

type PasswordResetService struct {
	userRepo       repository.UserRepository
	resetTokenRepo repository.PasswordResetRepository
	emailService   *email.EmailService
}

func NewPasswordResetService(
	userRepo repository.UserRepository,
	resetTokenRepo repository.PasswordResetRepository,
	emailService *email.EmailService,
) *PasswordResetService {
	return &PasswordResetService{
		userRepo:       userRepo,
		resetTokenRepo: resetTokenRepo,
		emailService:   emailService,
	}
}

// RequestPasswordReset génère un token et envoie l'email
func (s *PasswordResetService) RequestPasswordReset(email string) error {
	// 1. Trouver l'utilisateur
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return fmt.Errorf("erreur recherche utilisateur: %w", err)
	}
	if user == nil {
		// Ne pas révéler si l'email existe ou non (sécurité)
		return nil
	}

	// 2. Générer un token sécurisé (64 caractères hexadécimaux)
	token, err := generateSecureToken()
	if err != nil {
		return fmt.Errorf("erreur génération token: %w", err)
	}

	// 3. Stocker le token (valable 1h)
	expiresAt := time.Now().Add(1 * time.Hour)
	err = s.resetTokenRepo.CreateResetToken(user.ID, token, expiresAt) // user.ID est déjà uint
	if err != nil {
		return fmt.Errorf("erreur stockage token: %w", err)
	}

	// 4. Envoyer l'email
	err = s.emailService.SendPasswordResetEmail(user.Email, token)
	if err != nil {
		return fmt.Errorf("erreur envoi email: %w", err)
	}

	return nil
}

// ResetPassword valide le token et change le mot de passe
func (s *PasswordResetService) ResetPassword(token, newPassword string) error {
	// 1. Vérifier le token
	userID, valid, err := s.resetTokenRepo.GetResetToken(token)
	if err != nil {
		return fmt.Errorf("erreur vérification token: %w", err)
	}
	if !valid {
		return fmt.Errorf("token invalide ou expiré")
	}

	// 2. Hasher le nouveau mot de passe
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("erreur hashage mot de passe: %w", err)
	}

	// 3. Mettre à jour le mot de passe (userID est déjà uint)
	err = s.userRepo.UpdatePassword(userID, string(hashedPassword))
	if err != nil {
		return fmt.Errorf("erreur mise à jour mot de passe: %w", err)
	}

	// 4. Marquer le token comme utilisé
	err = s.resetTokenRepo.MarkTokenAsUsed(token)
	if err != nil {
		return fmt.Errorf("erreur marquage token: %w", err)
	}

	return nil
}

// generateSecureToken génère un token cryptographiquement sécurisé
func generateSecureToken() (string, error) {
	bytes := make([]byte, 32) // 32 bytes = 64 caractères hex
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
