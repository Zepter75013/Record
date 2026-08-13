package service

import (
	"context"
	"fmt"
	"unicode"

	"records-manager/backend/internal/users/repository"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
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
