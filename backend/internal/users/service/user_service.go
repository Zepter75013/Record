package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"records-manager/backend/internal/users/domain"
	"records-manager/backend/internal/users/repository"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error
	UpdateAvatar(ctx context.Context, userID uint, avatarPath string) (*domain.User, error)
	RemoveAvatar(ctx context.Context, userID uint) error
}

type userService struct {
	userRepo   repository.UserRepository
	uploadsDir string
}

func NewUserService(userRepo repository.UserRepository, uploadsDir string) UserService {
	return &userService{userRepo: userRepo, uploadsDir: uploadsDir}
}

// removeAvatarFile supprime l'ancien fichier avatar sur disque, uniquement
// s'il s'agit bien d'un fichier stocké localement sous /uploads/avatars/
// (jamais une URL externe, qui n'existe pas sur ce disque).
func (s *userService) removeAvatarFile(avatarPath *string) {
	if avatarPath == nil || *avatarPath == "" || !strings.HasPrefix(*avatarPath, "/uploads/avatars/") {
		return
	}
	relative := strings.TrimPrefix(*avatarPath, "/uploads/")
	_ = os.Remove(filepath.Join(s.uploadsDir, relative))
}

func (s *userService) UpdateAvatar(ctx context.Context, userID uint, avatarPath string) (*domain.User, error) {
	currentUser, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération de l'utilisateur: %w", err)
	}
	if currentUser == nil {
		return nil, fmt.Errorf("utilisateur non trouvé")
	}

	if err := s.userRepo.UpdateAvatar(userID, &avatarPath); err != nil {
		return nil, fmt.Errorf("erreur lors de la mise à jour de l'avatar: %w", err)
	}
	s.removeAvatarFile(currentUser.AvatarPath)

	currentUser.AvatarPath = &avatarPath
	return currentUser, nil
}

func (s *userService) RemoveAvatar(ctx context.Context, userID uint) error {
	currentUser, err := s.userRepo.FindByID(userID)
	if err != nil {
		return fmt.Errorf("erreur lors de la récupération de l'utilisateur: %w", err)
	}
	if currentUser == nil {
		return fmt.Errorf("utilisateur non trouvé")
	}

	if err := s.userRepo.UpdateAvatar(userID, nil); err != nil {
		return fmt.Errorf("erreur lors de la suppression de l'avatar: %w", err)
	}
	s.removeAvatarFile(currentUser.AvatarPath)

	return nil
}

// ValidatePasswordStrength valide la complexité du mot de passe
func (s *userService) ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("le mot de passe doit contenir au moins 8 caractères")
	}
	if len(password) > 72 {
		return fmt.Errorf("le mot de passe ne peut pas dépasser 72 caractères")
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("le mot de passe doit contenir au moins une majuscule")
	}
	if !hasLower {
		return fmt.Errorf("le mot de passe doit contenir au moins une minuscule")
	}
	if !hasNumber {
		return fmt.Errorf("le mot de passe doit contenir au moins un chiffre")
	}
	if !hasSpecial {
		return fmt.Errorf("le mot de passe doit contenir au moins un caractère spécial")
	}

	return nil
}

func (s *userService) ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error {
	// 1. Récupérer l'utilisateur par ID
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return fmt.Errorf("erreur lors de la récupération de l'utilisateur: %w", err)
	}
	if user == nil {
		return fmt.Errorf("utilisateur non trouvé")
	}

	// 2. Vérifier l'ancien mot de passe
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword))
	if err != nil {
		return fmt.Errorf("ancien mot de passe incorrect")
	}

	// 3. Valider la force du nouveau mot de passe
	if err := s.ValidatePasswordStrength(newPassword); err != nil {
		return err
	}

	// 4. Vérifier que le nouveau mot de passe est différent de l'ancien
	if oldPassword == newPassword {
		return fmt.Errorf("le nouveau mot de passe doit être différent de l'ancien")
	}

	// 5. Hasher le nouveau mot de passe
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("erreur lors du hashage du mot de passe: %w", err)
	}

	// 6. Mettre à jour en base de données
	if err := s.userRepo.UpdatePassword(userID, string(hashedPassword)); err != nil {
		return fmt.Errorf("erreur lors de la mise à jour du mot de passe: %w", err)
	}

	// TODO: Envoyer email de confirmation
	// Implémentation à ajouter selon votre service d'email

	return nil
}
