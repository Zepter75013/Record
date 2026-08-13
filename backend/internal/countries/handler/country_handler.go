package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	countriesService "records-manager/backend/internal/countries/service"

	"github.com/gorilla/mux"
)

type CountryHandler struct {
	service *countriesService.CountryService
}

func NewCountryHandler(service *countriesService.CountryService) *CountryHandler {
	return &CountryHandler{service: service}
}

type CreateCountryRequest struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

type UpdateCountryRequest struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

func (h *CountryHandler) CreateCountry(w http.ResponseWriter, r *http.Request) {
	var req CreateCountryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	country, err := h.service.CreateCountry(r.Context(), req.Name, req.Code, req.Description)
	if err != nil {
		if err.Error() == "un pays avec ce nom existe déjà" || err.Error() == "un pays avec ce code existe déjà" {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(country)
}

func (h *CountryHandler) GetAllCountries(w http.ResponseWriter, r *http.Request) {
	countries, err := h.service.GetAllCountries(r.Context())
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des pays", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(countries)
}

func (h *CountryHandler) UpdateCountry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	var req UpdateCountryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	country, err := h.service.UpdateCountry(r.Context(), id, req.Name, req.Code, req.Description)
	if err != nil {
		if err.Error() == "un autre pays avec ce nom existe déjà" || err.Error() == "un autre pays avec ce code existe déjà" {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(country)
}

func (h *CountryHandler) DeleteCountry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteCountry(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CountryHandler) GetCountryByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	country, err := h.service.GetCountryByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Pays non trouvé", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(country)
}
