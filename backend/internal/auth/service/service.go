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

// AuthService est l'interface pour la logique d'authentification
type AuthService interface {
	Authenticate(email, password string) (string, *domain.User, error)
	CreateTestUser(ctx context.Context, email, password string) error
	ValidateToken(tokenString string) (bool, error)
	ValidateTokenAndGetClaims(tokenString string) (*jwt.MapClaims, error) // 🆕 AJOUTÉ
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
	// 1. Trouver l'utilisateur
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return "", nil, fmt.Errorf("erreur lors de la recherche de l'utilisateur: %w", err)
	}

	if user == nil {
		return "", nil, fmt.Errorf("identifiants invalides")
	}

	// 2. Vérifier le mot de passe
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", nil, fmt.Errorf("identifiants invalides")
	}

	// 3. Générer le JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Expire dans 24h
	})

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", nil, fmt.Errorf("erreur lors de la génération du token: %w", err)
	}

	// Renvoyer l'utilisateur sans le hash du mot de passe
	user.PasswordHash = ""
	return tokenString, user, nil
}

// CreateTestUser : Fonction pour créer le premier utilisateur si la table est vide
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

// ValidateToken valide un token JWT
func (s *authService) ValidateToken(tokenString string) (bool, error) {
	if tokenString == "" {
		return false, fmt.Errorf("token vide")
	}

	// Parser et valider le token JWT
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Vérifier la méthode de signature
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("méthode de signature inattendue: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return false, fmt.Errorf("erreur lors du parsing du token: %w", err)
	}

	// Vérifier si le token est valide
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Vérifier l'expiration
		if exp, ok := claims["exp"].(float64); ok {
			expirationTime := time.Unix(int64(exp), 0)
			if time.Now().After(expirationTime) {
				return false, fmt.Errorf("token expiré")
			}
		}

		// Vérifier la présence des claims requis
		if _, ok := claims["user_id"]; !ok {
			return false, fmt.Errorf("claim user_id manquant")
		}

		if _, ok := claims["email"]; !ok {
			return false, fmt.Errorf("claim email manquant")
		}

		return true, nil
	}

	return false, fmt.Errorf("token invalide")
}

// 🆕 ValidateTokenAndGetClaims valide un token JWT et retourne les claims
func (s *authService) ValidateTokenAndGetClaims(tokenString string) (*jwt.MapClaims, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("token vide")
	}

	// Parser et valider le token JWT
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Vérifier la méthode de signature
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("méthode de signature inattendue: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("erreur lors du parsing du token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token invalide")
	}

	// Extraire les claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("impossible d'extraire les claims")
	}

	// Vérifier l'expiration
	if exp, ok := claims["exp"].(float64); ok {
		expirationTime := time.Unix(int64(exp), 0)
		if time.Now().After(expirationTime) {
			return nil, fmt.Errorf("token expiré")
		}
	}

	// Vérifier la présence des claims requis
	if _, ok := claims["user_id"]; !ok {
		return nil, fmt.Errorf("claim user_id manquant")
	}

	if _, ok := claims["email"]; !ok {
		return nil, fmt.Errorf("claim email manquant")
	}

	return &claims, nil
}
