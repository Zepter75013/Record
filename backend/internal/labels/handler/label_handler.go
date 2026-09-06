// records-manager/backend/internal/labels/handler/label_handler.go
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	labelsService "records-manager/backend/internal/labels/service"

	"github.com/gorilla/mux"
)

type LabelHandler struct {
	service *labelsService.LabelService
}

func NewLabelHandler(service *labelsService.LabelService) *LabelHandler {
	return &LabelHandler{service: service}
}

type CreateLabelRequest struct {
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	CountryID    *int    `json:"country_id"`
	FoundingYear *int    `json:"founding_year"`
	Website      *string `json:"website"`
}

type UpdateLabelRequest struct {
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	CountryID    *int    `json:"country_id"`
	FoundingYear *int    `json:"founding_year"`
	Website      *string `json:"website"`
}

func (h *LabelHandler) CreateLabel(w http.ResponseWriter, r *http.Request) {
	var req CreateLabelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	label, err := h.service.CreateLabel(r.Context(), req.Name, req.Description, req.CountryID, req.FoundingYear, req.Website)
	if err != nil {
		if err.Error() == "un label avec ce nom existe déjà" {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(label)
}

func (h *LabelHandler) GetAllLabels(w http.ResponseWriter, r *http.Request) {
	labels, err := h.service.GetAllLabels(r.Context())
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des labels", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(labels)
}

func (h *LabelHandler) UpdateLabel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	var req UpdateLabelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	label, err := h.service.UpdateLabel(r.Context(), id, req.Name, req.Description, req.CountryID, req.FoundingYear, req.Website)
	if err != nil {
		if err.Error() == "un autre label avec ce nom existe déjà" {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(label)
}

func (h *LabelHandler) DeleteLabel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteLabel(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetLabelByID - Récupère un label par son ID
func (h *LabelHandler) GetLabelByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	label, err := h.service.GetLabelByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Label non trouvé", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(label)
}

// SuggestLabelDescription - Propose une description pour un label à partir de Discogs
func (h *LabelHandler) SuggestLabelDescription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	description, err := h.service.SuggestDescriptionForLabel(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"description": description})
}

// SuggestLabelInfo - Propose pays/année de fondation/site web pour le label
// via MusicBrainz (ne modifie pas le label, voir LabelService.SuggestLabelInfo).
func (h *LabelHandler) SuggestLabelInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	suggestion, err := h.service.SuggestLabelInfo(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suggestion)
}
