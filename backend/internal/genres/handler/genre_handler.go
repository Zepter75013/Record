package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	genresService "records-manager/backend/internal/genres/service"

	"github.com/gorilla/mux"
)

type GenreHandler struct {
	service *genresService.GenreService
}

func NewGenreHandler(service *genresService.GenreService) *GenreHandler {
	return &GenreHandler{service: service}
}

type CreateGenreRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateGenreRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *GenreHandler) CreateGenre(w http.ResponseWriter, r *http.Request) {
	var req CreateGenreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	genre, err := h.service.CreateGenre(r.Context(), req.Name, req.Description)
	if err != nil {
		if err.Error() == "un genre avec ce nom existe déjà" {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(genre)
}

// CreateGenreIfNotExists - Crée un genre s'il n'existe pas déjà
func (h *GenreHandler) CreateGenreIfNotExists(w http.ResponseWriter, r *http.Request) {
	var req CreateGenreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	genre, err := h.service.CreateGenreIfNotExists(r.Context(), req.Name, req.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(genre)
}

func (h *GenreHandler) GetAllGenres(w http.ResponseWriter, r *http.Request) {
	genres, err := h.service.GetAllGenres(r.Context())
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des genres", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(genres)
}

func (h *GenreHandler) UpdateGenre(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	var req UpdateGenreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	genre, err := h.service.UpdateGenre(r.Context(), id, req.Name, req.Description)
	if err != nil {
		if err.Error() == "un autre genre avec ce nom existe déjà" {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(genre)
}

func (h *GenreHandler) DeleteGenre(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteGenre(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetGenreByID - Récupère un genre par son ID
func (h *GenreHandler) GetGenreByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	genre, err := h.service.GetGenreByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Genre non trouvé", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(genre)
}
