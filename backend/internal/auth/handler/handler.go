package handler

import (
	"encoding/json"
	"net/http"
	"records-manager/backend/internal/auth/service"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  struct {
		ID         uint    `json:"id"`
		Email      string  `json:"email"`
		AvatarPath *string `json:"avatar_path"`
	} `json:"user"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Requête invalide", http.StatusBadRequest)
		return
	}

	token, user, err := h.authService.Authenticate(req.Email, req.Password)
	if err != nil {
		http.Error(w, "Email ou mot de passe invalide", http.StatusUnauthorized)
		return
	}

	resp := LoginResponse{Token: token}
	resp.User.ID = user.ID
	resp.User.Email = user.Email
	resp.User.AvatarPath = user.AvatarPath

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("API is running and healthy!"))
}
