package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	gamesService "records-manager/backend/internal/games/service"

	"github.com/gorilla/mux"
)

type GameHandler struct {
	service *gamesService.GameService
}

func NewGameHandler(service *gamesService.GameService) *GameHandler {
	return &GameHandler{service: service}
}

type CreateGameRequest struct {
	Title       string   `json:"title"`
	PlatformID  int      `json:"platform_id"`
	GenreID     *int     `json:"genre_id,omitempty"`
	PublisherID *int     `json:"publisher_id,omitempty"`
	ReleaseYear *int     `json:"release_year,omitempty"`
	Barcode     *string  `json:"barcode,omitempty"`
	Notes       *string  `json:"notes,omitempty"`
	Price       *float64 `json:"price,omitempty"`
	Quantity    *int     `json:"quantity,omitempty"`
	CoverImage  *string  `json:"cover_image,omitempty"`
	RAWGID      *int64   `json:"rawg_id,omitempty"`
}

type UpdateGameRequest struct {
	Title       string   `json:"title"`
	PlatformID  int      `json:"platform_id"`
	GenreID     *int     `json:"genre_id,omitempty"`
	PublisherID *int     `json:"publisher_id,omitempty"`
	ReleaseYear *int     `json:"release_year,omitempty"`
	Barcode     *string  `json:"barcode,omitempty"`
	Notes       *string  `json:"notes,omitempty"`
	Price       *float64 `json:"price,omitempty"`
	Quantity    *int     `json:"quantity,omitempty"`
	CoverImage  *string  `json:"cover_image,omitempty"`
}

type SearchRAWGRequest struct {
	Title string `json:"title"`
}

type SelectRAWGResultRequest struct {
	RAWGID int64 `json:"rawg_id"`
}

type CheckGameBarcodeRequest struct {
	Barcode   string `json:"barcode"`
	ExcludeID *int   `json:"exclude_id,omitempty"`
}

type CheckGameBarcodeResponse struct {
	Exists bool               `json:"exists"`
	Game   *BarcodeGameDetails `json:"game,omitempty"`
}

type BarcodeGameDetails struct {
	ID      int     `json:"id"`
	Title   string  `json:"title"`
	Barcode *string `json:"barcode"`
}

func (h *GameHandler) CreateGame(w http.ResponseWriter, r *http.Request) {
	var req CreateGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}
	if req.CoverImage != nil {
		fmt.Printf("✅ Go reçoit cover_image jeu pour création: %s\n", *req.CoverImage)
	}

	game, err := h.service.CreateGame(
		r.Context(),
		req.Title,
		req.PlatformID,
		req.GenreID,
		req.PublisherID,
		req.ReleaseYear,
		req.Barcode,
		req.Notes,
		req.Price,
		req.Quantity,
		req.CoverImage,
		req.RAWGID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(game)
}

func (h *GameHandler) GetAllGames(w http.ResponseWriter, r *http.Request) {
	games, err := h.service.GetAllGames(r.Context())
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des jeux", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(games)
}

func (h *GameHandler) GetGameByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	game, err := h.service.GetGameByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération du jeu", http.StatusInternalServerError)
		return
	}
	if game == nil {
		http.Error(w, "Jeu non trouvé", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(game)
}

func (h *GameHandler) UpdateGame(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	var req UpdateGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	game, err := h.service.UpdateGame(
		r.Context(),
		id,
		req.Title,
		req.PlatformID,
		req.GenreID,
		req.PublisherID,
		req.ReleaseYear,
		req.Barcode,
		req.Notes,
		req.Price,
		req.Quantity,
		req.CoverImage,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(game)
}

func (h *GameHandler) DeleteGame(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteGame(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *GameHandler) SearchRAWG(w http.ResponseWriter, r *http.Request) {
	var req SearchRAWGRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	preview, err := h.service.PreviewGame(strings.TrimSpace(req.Title))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(preview)
}

func (h *GameHandler) SelectRAWGResult(w http.ResponseWriter, r *http.Request) {
	var req SelectRAWGResultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	preview, err := h.service.SelectRAWGResult(req.RAWGID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(preview)
}

func (h *GameHandler) CheckBarcodeExists(w http.ResponseWriter, r *http.Request) {
	var req CheckGameBarcodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	exists, game, err := h.service.CheckBarcodeExists(r.Context(), req.Barcode, req.ExcludeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := CheckGameBarcodeResponse{Exists: exists}
	if exists && game != nil {
		response.Game = &BarcodeGameDetails{
			ID:      game.ID,
			Title:   game.Title,
			Barcode: game.Barcode,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
