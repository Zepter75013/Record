package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	platformsService "records-manager/backend/internal/platforms/service"

	"github.com/gorilla/mux"
)

type PlatformHandler struct {
	service *platformsService.PlatformService
}

func NewPlatformHandler(service *platformsService.PlatformService) *PlatformHandler {
	return &PlatformHandler{service: service}
}

type CreatePlatformRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdatePlatformRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreatePlatformIfNotExistsRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *PlatformHandler) CreatePlatform(w http.ResponseWriter, r *http.Request) {
	var req CreatePlatformRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	platform, err := h.service.CreatePlatform(r.Context(), req.Name, req.Description)
	if err != nil {
		if err.Error() == "une plateforme avec ce nom existe déjà" {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(platform)
}

func (h *PlatformHandler) GetAllPlatforms(w http.ResponseWriter, r *http.Request) {
	platforms, err := h.service.GetAllPlatforms(r.Context())
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des plateformes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(platforms)
}

func (h *PlatformHandler) UpdatePlatform(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	var req UpdatePlatformRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	platform, err := h.service.UpdatePlatform(r.Context(), id, req.Name, req.Description)
	if err != nil {
		if err.Error() == "une autre plateforme avec ce nom existe déjà" {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(platform)
}

func (h *PlatformHandler) DeletePlatform(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	err = h.service.DeletePlatform(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *PlatformHandler) GetPlatformByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	platform, err := h.service.GetPlatformByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Plateforme non trouvée", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(platform)
}

// CreatePlatformIfNotExists - Crée une plateforme seulement si elle n'existe pas déjà
func (h *PlatformHandler) CreatePlatformIfNotExists(w http.ResponseWriter, r *http.Request) {
	var req CreatePlatformIfNotExistsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	platform, err := h.service.CreatePlatformIfNotExists(r.Context(), req.Name, req.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(platform)
}
