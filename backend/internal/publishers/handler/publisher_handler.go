package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	publishersService "records-manager/backend/internal/publishers/service"

	"github.com/gorilla/mux"
)

type PublisherHandler struct {
	service *publishersService.PublisherService
}

func NewPublisherHandler(service *publishersService.PublisherService) *PublisherHandler {
	return &PublisherHandler{service: service}
}

type CreatePublisherRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdatePublisherRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreatePublisherIfNotExistsRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *PublisherHandler) CreatePublisher(w http.ResponseWriter, r *http.Request) {
	var req CreatePublisherRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	publisher, err := h.service.CreatePublisher(r.Context(), req.Name, req.Description)
	if err != nil {
		if err.Error() == "un éditeur avec ce nom existe déjà" {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(publisher)
}

func (h *PublisherHandler) GetAllPublishers(w http.ResponseWriter, r *http.Request) {
	publishers, err := h.service.GetAllPublishers(r.Context())
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des éditeurs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(publishers)
}

func (h *PublisherHandler) UpdatePublisher(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	var req UpdatePublisherRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	publisher, err := h.service.UpdatePublisher(r.Context(), id, req.Name, req.Description)
	if err != nil {
		if err.Error() == "un autre éditeur avec ce nom existe déjà" {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(publisher)
}

func (h *PublisherHandler) DeletePublisher(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	err = h.service.DeletePublisher(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *PublisherHandler) GetPublisherByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	publisher, err := h.service.GetPublisherByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Éditeur non trouvé", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(publisher)
}

// CreatePublisherIfNotExists - Crée un éditeur seulement s'il n'existe pas déjà
func (h *PublisherHandler) CreatePublisherIfNotExists(w http.ResponseWriter, r *http.Request) {
	var req CreatePublisherIfNotExistsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	publisher, err := h.service.CreatePublisherIfNotExists(r.Context(), req.Name, req.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(publisher)
}
