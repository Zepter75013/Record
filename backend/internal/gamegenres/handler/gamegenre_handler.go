package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	gameGenresService "records-manager/backend/internal/gamegenres/service"

	"github.com/gorilla/mux"
)

type GameGenreHandler struct {
	service *gameGenresService.GameGenreService
}

func NewGameGenreHandler(service *gameGenresService.GameGenreService) *GameGenreHandler {
	return &GameGenreHandler{service: service}
}

type CreateGameGenreRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateGameGenreRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreateGameGenreIfNotExistsRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *GameGenreHandler) CreateGameGenre(w http.ResponseWriter, r *http.Request) {
	var req CreateGameGenreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	genre, err := h.service.CreateGameGenre(r.Context(), req.Name, req.Description)
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

func (h *GameGenreHandler) GetAllGameGenres(w http.ResponseWriter, r *http.Request) {
	genres, err := h.service.GetAllGameGenres(r.Context())
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des genres", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(genres)
}

func (h *GameGenreHandler) UpdateGameGenre(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	var req UpdateGameGenreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	genre, err := h.service.UpdateGameGenre(r.Context(), id, req.Name, req.Description)
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

func (h *GameGenreHandler) DeleteGameGenre(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteGameGenre(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *GameGenreHandler) GetGameGenreByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	genre, err := h.service.GetGameGenreByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Genre non trouvé", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(genre)
}

// CreateGameGenreIfNotExists - Crée un genre seulement s'il n'existe pas déjà
func (h *GameGenreHandler) CreateGameGenreIfNotExists(w http.ResponseWriter, r *http.Request) {
	var req CreateGameGenreIfNotExistsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	genre, err := h.service.CreateGameGenreIfNotExists(r.Context(), req.Name, req.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(genre)
}
