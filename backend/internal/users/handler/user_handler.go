package handler

import (
	"encoding/json"
	"net/http"

	"records-manager/backend/internal/users/service"

	"github.com/golang-jwt/jwt/v5"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type ChangePasswordResponse struct {
	Message string `json:"message"`
}

func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	// 1. Extraire le user_id du JWT
	claims, ok := r.Context().Value("user").(*jwt.MapClaims)
	if !ok {
		http.Error(w, "Utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

	userIDFloat, ok := (*claims)["user_id"].(float64)
	if !ok {
		http.Error(w, "ID utilisateur invalide", http.StatusUnauthorized)
		return
	}
	userID := uint(userIDFloat)

	// 2. Décoder la requête
	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	// 3. Valider les champs
	if req.OldPassword == "" || req.NewPassword == "" {
		http.Error(w, "Tous les champs sont requis", http.StatusBadRequest)
		return
	}

	// 4. Changer le mot de passe
	if err := h.userService.ChangePassword(r.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 5. Réponse
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ChangePasswordResponse{
		Message: "Mot de passe modifié avec succès",
	})
}
