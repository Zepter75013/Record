package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	formatsService "records-manager/backend/internal/formats/service"

	"github.com/gorilla/mux"
)

type FormatHandler struct {
	service *formatsService.FormatService
}

func NewFormatHandler(service *formatsService.FormatService) *FormatHandler {
	return &FormatHandler{service: service}
}

type CreateFormatRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateFormatRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreateFormatIfNotExistsRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *FormatHandler) CreateFormat(w http.ResponseWriter, r *http.Request) {
	var req CreateFormatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	format, err := h.service.CreateFormat(r.Context(), req.Name, req.Description)
	if err != nil {
		if err.Error() == "un format avec ce nom existe déjà" {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(format)
}

func (h *FormatHandler) GetAllFormats(w http.ResponseWriter, r *http.Request) {
	formats, err := h.service.GetAllFormats(r.Context())
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des formats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(formats)
}

func (h *FormatHandler) UpdateFormat(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	var req UpdateFormatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	format, err := h.service.UpdateFormat(r.Context(), id, req.Name, req.Description)
	if err != nil {
		if err.Error() == "un autre format avec ce nom existe déjà" {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(format)
}

func (h *FormatHandler) DeleteFormat(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteFormat(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetFormatByID - Récupère un format par son ID
func (h *FormatHandler) GetFormatByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	format, err := h.service.GetFormatByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Format non trouvé", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(format)
}

// CreateFormatIfNotExists - Crée un format seulement s'il n'existe pas déjà
func (h *FormatHandler) CreateFormatIfNotExists(w http.ResponseWriter, r *http.Request) {
	var req CreateFormatIfNotExistsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	format, err := h.service.CreateFormatIfNotExists(r.Context(), req.Name, req.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(format)
}
