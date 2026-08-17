package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"records-manager/backend/internal/users/service"
)

type PasswordResetHandler struct {
	resetService *service.PasswordResetService
}

func NewPasswordResetHandler(resetService *service.PasswordResetService) *PasswordResetHandler {
	return &PasswordResetHandler{resetService: resetService}
}

// RequestPasswordReset - Demande de réinitialisation
func (h *PasswordResetHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Requête invalide", http.StatusBadRequest)
		return
	}

	err := h.resetService.RequestPasswordReset(req.Email)
	if err != nil {
		// On journalise côté serveur (SMTP mal configuré, etc.) sans jamais
		// révéler d'erreur spécifique au client — sinon une réponse
		// différente selon que l'email existe ou non permettrait d'énumérer
		// les comptes.
		log.Printf("⚠️  Échec de la demande de réinitialisation pour %s: %v", req.Email, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Si cet email existe, un lien de réinitialisation a été envoyé",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Si cet email existe, un lien de réinitialisation a été envoyé",
	})
}

// ResetPassword - Réinitialisation avec token
func (h *PasswordResetHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Requête invalide", http.StatusBadRequest)
		return
	}

	err := h.resetService.ResetPassword(req.Token, req.NewPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Mot de passe réinitialisé avec succès",
	})
}
