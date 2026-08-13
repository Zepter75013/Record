package service

import (
	"context"
	"fmt"
	"time"

	"records-manager/backend/config"
	"records-manager/backend/internal/users/domain"
	"records-manager/backend/internal/users/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Authenticate(email, password string) (string, *domain.User, error)
	CreateTestUser(ctx context.Context, email, password string) error
}

type authService struct {
	userRepo  repository.UserRepository
	jwtSecret string
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{
		userRepo:  userRepo,
		jwtSecret: config.GetJWTSecret(),
	}
}

func (s *authService) Authenticate(email, password string) (string, *domain.User, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return "", nil, fmt.Errorf("erreur lors de la recherche de l'utilisateur: %w", err)
	}
	if user == nil {
		return "", nil, fmt.Errorf("identifiants invalides")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", nil, fmt.Errorf("identifiants invalides")
	}

	// 3. Générer le JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", nil, fmt.Errorf("erreur lors de la génération du token: %w", err)
	}

	user.PasswordHash = ""
	return tokenString, user, nil
}

func (s *authService) CreateTestUser(ctx context.Context, email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("erreur de hachage: %w", err)
	}

	user := &domain.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	return s.userRepo.AddUser(user)
}
