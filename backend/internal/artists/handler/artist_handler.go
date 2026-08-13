package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	artistsService "records-manager/backend/internal/artists/service"

	"github.com/gorilla/mux"
)

type ArtistHandler struct {
	service *artistsService.ArtistService
}

func NewArtistHandler(service *artistsService.ArtistService) *ArtistHandler {
	return &ArtistHandler{service: service}
}

type CreateArtistRequest struct {
	Name      string `json:"name"`
	Biography string `json:"biography"`
}

type UpdateArtistRequest struct {
	Name      string `json:"name"`
	Biography string `json:"biography"`
}

func (h *ArtistHandler) CreateArtist(w http.ResponseWriter, r *http.Request) {
	var req CreateArtistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	artist, err := h.service.CreateArtist(r.Context(), req.Name, req.Biography)
	if err != nil {
		if err.Error() == "un artiste avec ce nom existe déjà" {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(artist)
}

func (h *ArtistHandler) GetAllArtists(w http.ResponseWriter, r *http.Request) {
	artists, err := h.service.GetAllArtists(r.Context())
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des artistes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(artists)
}

func (h *ArtistHandler) UpdateArtist(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	var req UpdateArtistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	artist, err := h.service.UpdateArtist(r.Context(), id, req.Name, req.Biography)
	if err != nil {
		if err.Error() == "un autre artiste avec ce nom existe déjà" {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(artist)
}

func (h *ArtistHandler) DeleteArtist(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteArtist(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetArtistByID - Récupère un artiste par son ID
func (h *ArtistHandler) GetArtistByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	artist, err := h.service.GetArtistByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Artiste non trouvé", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(artist)
}
