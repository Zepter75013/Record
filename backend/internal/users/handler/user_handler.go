package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"records-manager/backend/internal/users/service"

	"github.com/golang-jwt/jwt/v5"
)

type UserHandler struct {
	userService service.UserService
	uploadsDir  string
}

func NewUserHandler(userService service.UserService, uploadsDir string) *UserHandler {
	return &UserHandler{userService: userService, uploadsDir: uploadsDir}
}

// userIDFromRequest extrait le user_id du JWT injecté dans le contexte par
// authMiddlewareWithService (voir cmd/api/main.go).
func userIDFromRequest(r *http.Request) (uint, bool) {
	claims, ok := r.Context().Value("user").(*jwt.MapClaims)
	if !ok {
		return 0, false
	}
	userIDFloat, ok := (*claims)["user_id"].(float64)
	if !ok {
		return 0, false
	}
	return uint(userIDFloat), true
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
	userID, ok := userIDFromRequest(r)
	if !ok {
		http.Error(w, "Utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

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

type AvatarResponse struct {
	AvatarURL string `json:"avatar_url"`
	Message   string `json:"message"`
}

var validAvatarTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/jpg":  ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

func (h *UserHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(r)
	if !ok {
		http.Error(w, "Utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

	if err := r.ParseMultipartForm(5 << 20); err != nil {
		http.Error(w, "Fichier trop volumineux (max 5MB)", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		http.Error(w, "Fichier non trouvé dans la requête", http.StatusBadRequest)
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	ext, valid := validAvatarTypes[contentType]
	if !valid {
		http.Error(w, "Format de fichier non supporté. Utilisez JPG, PNG ou WebP.", http.StatusBadRequest)
		return
	}

	avatarsDir := filepath.Join(h.uploadsDir, "avatars")
	if err := os.MkdirAll(avatarsDir, 0755); err != nil {
		http.Error(w, "Erreur création du répertoire", http.StatusInternalServerError)
		return
	}

	// Nom déterministe par utilisateur : un nouvel upload remplace l'ancien
	// fichier sur disque (nettoyé via removeAvatarFile côté service) au lieu
	// d'en accumuler un par upload, contrairement aux pochettes de disques/jeux.
	filename := fmt.Sprintf("user-%d%s", userID, ext)
	filePath := filepath.Join(avatarsDir, filename)
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Erreur sauvegarde du fichier", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Erreur copie du fichier", http.StatusInternalServerError)
		return
	}

	avatarPath := fmt.Sprintf("/uploads/avatars/%s", filename)
	if _, err := h.userService.UpdateAvatar(r.Context(), userID, avatarPath); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AvatarResponse{
		AvatarURL: avatarPath,
		Message:   "Photo de profil mise à jour",
	})
}

func (h *UserHandler) RemoveAvatar(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(r)
	if !ok {
		http.Error(w, "Utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

	if err := h.userService.RemoveAvatar(r.Context(), userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
