package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	discsService "records-manager/backend/internal/discs/service"

	"github.com/gorilla/mux"
)

type DiscHandler struct {
	service *discsService.DiscService
}

func NewDiscHandler(service *discsService.DiscService) *DiscHandler {
	return &DiscHandler{service: service}
}

type CreateDiscRequest struct {
	Title       string   `json:"title"`
	ArtistID    int      `json:"artist_id"`
	ArtistName  string   `json:"artist_name"`
	GenreID     *int     `json:"genre_id,omitempty"`
	FormatID    *int     `json:"format_id,omitempty"`
	CountryID   *int     `json:"country_id,omitempty"`
	LabelID     *int     `json:"label_id,omitempty"`
	ReleaseYear *int     `json:"release_year,omitempty"`
	Barcode     *string  `json:"barcode,omitempty"`
	Notes       *string  `json:"notes,omitempty"`
	Price       *float64 `json:"price,omitempty"`
	Quantity    *int     `json:"quantity,omitempty"`
	CoverImage  *string  `json:"cover_image,omitempty"`
	// NOUVEAU: Champs de streaming
	AppleMusicURL *string `json:"apple_music_url,omitempty"`
	SpotifyURL    *string `json:"spotify_url,omitempty"`
	DeezerURL     *string `json:"deezer_url,omitempty"`
	YoutubeURL    *string `json:"youtube_url,omitempty"`
	ISRC          *string `json:"isrc,omitempty"`
	// NOUVEAU: Tracklist Discogs capturée à la sélection de la release
	DiscogsReleaseID *int64                   `json:"discogs_release_id,omitempty"`
	Tracks           []discsService.TrackItem `json:"tracks,omitempty"`
}

type UpdateDiscRequest struct {
	Title       string   `json:"title"`
	ArtistID    int      `json:"artist_id"`
	ArtistName  string   `json:"artist_name"`
	GenreID     *int     `json:"genre_id,omitempty"`
	FormatID    *int     `json:"format_id,omitempty"`
	CountryID   *int     `json:"country_id,omitempty"`
	LabelID     *int     `json:"label_id,omitempty"`
	ReleaseYear *int     `json:"release_year,omitempty"`
	Barcode     *string  `json:"barcode,omitempty"`
	Notes       *string  `json:"notes,omitempty"`
	Price       *float64 `json:"price,omitempty"`
	Quantity    *int     `json:"quantity,omitempty"`
	CoverImage  *string  `json:"cover_image,omitempty"`
	// NOUVEAU: Champs de streaming
	AppleMusicURL *string `json:"apple_music_url,omitempty"`
	SpotifyURL    *string `json:"spotify_url,omitempty"`
	DeezerURL     *string `json:"deezer_url,omitempty"`
	YoutubeURL    *string `json:"youtube_url,omitempty"`
	ISRC          *string `json:"isrc,omitempty"`
}

type PreviewCoverRequest struct {
	Barcode string `json:"barcode"`
	Title   string `json:"title"`
	Artist  string `json:"artist"`
	Country string `json:"country"`
	Year    string `json:"year"`
}

type PreviewCoverResponse struct {
	CoverURL  string                   `json:"cover_url"`
	Title     string                   `json:"title"`
	Artist    string                   `json:"artist"`
	Year      string                   `json:"year"`
	Genres    []string                 `json:"genres"`
	Formats   []string                 `json:"formats"`
	Country   string                   `json:"country"`
	Label     string                   `json:"label"`
	Found     bool                     `json:"found"`
	Results   []PreviewResult          `json:"results,omitempty"`
	Prices    *DiscogsPriceResponse    `json:"prices,omitempty"`
	Tracklist []discsService.TrackItem `json:"tracklist,omitempty"`
}

type PreviewResult struct {
	ID       int                   `json:"id"`
	Title    string                `json:"title"`
	CoverURL string                `json:"cover_url"`
	Year     string                `json:"year"`
	Country  string                `json:"country"`
	Formats  []string              `json:"formats"`
	Genres   []string              `json:"genres"`
	Label    string                `json:"label"`
	Prices   *DiscogsPriceResponse `json:"prices,omitempty"`
}

type DiscogsPriceResponse struct {
	LowestPrice  float64 `json:"lowest_price,omitempty"`
	MedianPrice  float64 `json:"median_price,omitempty"`
	HighestPrice float64 `json:"highest_price,omitempty"`
}

// 🔧 CORRIGÉ : int → int64
type SelectDiscogsResultRequest struct {
	ReleaseID int64 `json:"release_id"`
}

type CheckBarcodeRequest struct {
	Barcode   string `json:"barcode"`
	ExcludeID *int   `json:"exclude_id,omitempty"`
}

type CheckBarcodeResponse struct {
	Exists bool                `json:"exists"`
	Disc   *BarcodeDiscDetails `json:"disc,omitempty"`
}

type BarcodeDiscDetails struct {
	ID         int     `json:"id"`
	Title      string  `json:"title"`
	ArtistName string  `json:"artist_name"`
	Barcode    *string `json:"barcode"`
}

func (h *DiscHandler) CheckBarcodeExists(w http.ResponseWriter, r *http.Request) {
	var req CheckBarcodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}
	exists, disc, err := h.service.CheckBarcodeExists(r.Context(), req.Barcode, req.ExcludeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := CheckBarcodeResponse{Exists: exists}
	if exists && disc != nil {
		response.Disc = &BarcodeDiscDetails{
			ID:         disc.ID,
			Title:      disc.Title,
			ArtistName: disc.ArtistName,
			Barcode:    disc.Barcode,
		}
		fmt.Printf("✅ Réponse check-barcode: disque trouvé ID=%d, titre=%s\n", disc.ID, disc.Title)
	} else {
		fmt.Printf("ℹ️  Réponse check-barcode: code-barres %s non trouvé\n", req.Barcode)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *DiscHandler) CheckBarcodeExistsGET(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	barcode := vars["barcode"]
	if barcode == "" {
		http.Error(w, "Code-barres manquant", http.StatusBadRequest)
		return
	}
	fmt.Printf("🔍 Vérification GET code-barres: %s\n", barcode)
	exists, disc, err := h.service.CheckBarcodeExists(r.Context(), barcode, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := CheckBarcodeResponse{Exists: exists}
	if exists && disc != nil {
		response.Disc = &BarcodeDiscDetails{
			ID:         disc.ID,
			Title:      disc.Title,
			ArtistName: disc.ArtistName,
			Barcode:    disc.Barcode,
		}
		fmt.Printf("✅ Réponse check-barcode GET: disque trouvé ID=%d, titre=%s\n", disc.ID, disc.Title)
	} else {
		fmt.Printf("ℹ️  Réponse check-barcode GET: code-barres %s non trouvé\n", barcode)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *DiscHandler) CreateDisc(w http.ResponseWriter, r *http.Request) {
	var req CreateDiscRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}
	if req.CoverImage != nil {
		fmt.Printf("✅ Go reçoit cover_image pour création: %s\n", *req.CoverImage)
		if strings.HasPrefix(*req.CoverImage, "https://i.discogs.com/") ||
			strings.Contains(*req.CoverImage, "discogs.com") {
			fmt.Printf("🌐 URL Discogs détectée, le service va la télécharger\n")
		} else if strings.HasPrefix(*req.CoverImage, "http") {
			fmt.Printf("⚠️ URL internet non-Discogs: %s\n", *req.CoverImage)
		} else if strings.HasPrefix(*req.CoverImage, "data:image") {
			fmt.Printf("🔧 Image base64 détectée\n")
		} else if strings.HasPrefix(*req.CoverImage, "/uploads/covers/") {
			fmt.Printf("📂 Chemin local détecté: %s\n", *req.CoverImage)
		}
	} else {
		fmt.Printf("⚠️  Go reçoit cover_image pour création: nil\n")
	}
	if req.Price != nil {
		fmt.Printf("💰 Go reçoit price pour création: %.2f\n", *req.Price)
	} else {
		fmt.Printf("ℹ️  Go reçoit price pour création: nil\n")
	}
	if req.Quantity != nil {
		fmt.Printf("📦 Go reçoit quantity pour création: %d\n", *req.Quantity)
	} else {
		fmt.Printf("ℹ️  Go reçoit quantity pour création: nil (défaut = 1)\n")
	}
	disc, err := h.service.CreateDisc(
		r.Context(),
		req.Title,
		req.ArtistID,
		req.ArtistName,
		req.GenreID,
		req.FormatID,
		req.CountryID,
		req.LabelID,
		req.ReleaseYear,
		req.Barcode,
		req.Notes,
		req.Price,
		req.Quantity,
		req.CoverImage,
		// NOUVEAU: Paramètres de streaming
		req.AppleMusicURL,
		req.SpotifyURL,
		req.DeezerURL,
		req.YoutubeURL,
		req.ISRC,
		// NOUVEAU: Tracklist Discogs
		req.DiscogsReleaseID,
		req.Tracks,
	)
	if err != nil {
		if err.Error() == "un disque avec ce code-barres existe déjà" {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(disc)
}

func (h *DiscHandler) GetAllDiscs(w http.ResponseWriter, r *http.Request) {
	discs, err := h.service.GetAllDiscs(r.Context())
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des disques", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(discs)
}

// ✅ FONCTION MODIFIÉE
func (h *DiscHandler) UpdateDisc(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}
	var req UpdateDiscRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}
	if req.CoverImage != nil {
		fmt.Printf("✅ Go reçoit cover_image pour modification: %s\n", *req.CoverImage)
		if strings.HasPrefix(*req.CoverImage, "https://i.discogs.com/") ||
			strings.Contains(*req.CoverImage, "discogs.com") {
			fmt.Printf("🌐 URL Discogs détectée, le service va la télécharger\n")
		} else if strings.HasPrefix(*req.CoverImage, "http") {
			fmt.Printf("⚠️ URL internet non-Discogs: %s\n", *req.CoverImage)
		} else if strings.HasPrefix(*req.CoverImage, "data:image") {
			fmt.Printf("🔧 Image base64 détectée\n")
		} else if strings.HasPrefix(*req.CoverImage, "/uploads/covers/") {
			fmt.Printf("📂 Chemin local détecté: %s\n", *req.CoverImage)
		}
	} else {
		fmt.Printf("⚠️  Go reçoit cover_image pour modification: nil\n")
	}
	if req.Price != nil {
		fmt.Printf("💰 Go reçoit price pour modification: %.2f\n", *req.Price)
	} else {
		fmt.Printf("ℹ️  Go reçoit price pour modification: nil\n")
	}
	if req.Quantity != nil {
		fmt.Printf("📦 Go reçoit quantity pour modification: %d\n", *req.Quantity)
	} else {
		fmt.Printf("ℹ️  Go reçoit quantity pour modification: nil\n")
	}

	// ✅ Effectuer la mise à jour
	_, err = h.service.UpdateDisc(
		r.Context(),
		id,
		req.Title,
		req.ArtistID,
		req.ArtistName,
		req.GenreID,
		req.FormatID,
		req.CountryID,
		req.LabelID,
		req.ReleaseYear,
		req.Barcode,
		req.Notes,
		req.Price,
		req.Quantity,
		req.CoverImage,
		// NOUVEAU: Paramètres de streaming
		req.AppleMusicURL,
		req.SpotifyURL,
		req.DeezerURL,
		req.YoutubeURL,
		req.ISRC,
	)

	if err != nil {
		if err.Error() == "un autre disque avec ce code-barres existe déjà" {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	// ✅ NOUVEAU: Récupérer le disque complet après mise à jour
	disc, err := h.service.GetDiscByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération du disque mis à jour", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(disc)
}

func (h *DiscHandler) DeleteDisc(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}
	err = h.service.DeleteDisc(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *DiscHandler) GetDiscByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}
	disc, err := h.service.GetDiscByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Disque non trouvé", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(disc)
}

func (h *DiscHandler) DownloadCover(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}
	err = h.service.DownloadCoverForDisc(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Pochette téléchargée avec succès"})
}

func (h *DiscHandler) GetTracks(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}
	list, err := h.service.GetTracksForDisc(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *DiscHandler) FetchTracks(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}
	list, err := h.service.FetchTracklistForDisc(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *DiscHandler) UpdateTracks(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}
	var items []discsService.TrackItem
	if err := json.NewDecoder(r.Body).Decode(&items); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}
	list, err := h.service.UpdateTracksForDisc(r.Context(), id, items)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *DiscHandler) PreviewCover(w http.ResponseWriter, r *http.Request) {
	var req PreviewCoverRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}
	preview, err := h.service.PreviewCover(r.Context(), req.Barcode, req.Title, req.Artist, req.Country, req.Year)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := PreviewCoverResponse{
		CoverURL: preview.CoverURL,
		Title:    preview.Title,
		Artist:   preview.Artist,
		Year:     preview.Year,
		Genres:   preview.Genres,
		Formats:  preview.Formats,
		Country:  preview.Country,
		Label:    preview.Label,
		Found:    preview.Found,
	}
	if preview.Prices != nil {
		response.Prices = &DiscogsPriceResponse{
			LowestPrice:  preview.Prices.LowestPrice,
			MedianPrice:  preview.Prices.MedianPrice,
			HighestPrice: preview.Prices.HighestPrice,
		}
	}
	if len(preview.Results) > 0 {
		response.Results = make([]PreviewResult, len(preview.Results))
		for i, result := range preview.Results {
			var formatNames []string
			for _, format := range result.Formats {
				formatNames = append(formatNames, format.Name)
				for _, desc := range format.Descriptions {
					formatNames = append(formatNames, desc)
				}
			}
			label := ""
			if len(result.Labels) > 0 {
				label = result.Labels[0].Name
			}
			response.Results[i] = PreviewResult{
				ID:       result.ID,
				Title:    result.Title,
				CoverURL: result.CoverURL,
				Year:     result.Year,
				Country:  result.Country,
				Formats:  formatNames,
				Genres:   result.Genres,
				Label:    label,
			}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 🔧 CORRIGÉ : gestion int64 + validation
func (h *DiscHandler) SelectDiscogsResult(w http.ResponseWriter, r *http.Request) {
	var req SelectDiscogsResultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	if req.ReleaseID <= 0 {
		http.Error(w, "release_id invalide (doit être > 0)", http.StatusBadRequest)
		return
	}

	fmt.Printf("🔍 [Handler] SelectDiscogsResult appelé avec release_id=%d\n", req.ReleaseID)

	preview, err := h.service.SelectDiscogsResult(r.Context(), req.ReleaseID)
	if err != nil {
		fmt.Printf("❌ [Handler] Erreur service SelectDiscogsResult: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := PreviewCoverResponse{
		CoverURL:  preview.CoverURL,
		Title:     preview.Title,
		Artist:    preview.Artist,
		Year:      preview.Year,
		Genres:    preview.Genres,
		Formats:   preview.Formats,
		Country:   preview.Country,
		Label:     preview.Label,
		Found:     preview.Found,
		Tracklist: preview.Tracks,
	}

	if preview.Prices != nil {
		response.Prices = &DiscogsPriceResponse{
			LowestPrice:  preview.Prices.LowestPrice,
			MedianPrice:  preview.Prices.MedianPrice,
			HighestPrice: preview.Prices.HighestPrice,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *DiscHandler) SearchDiscogs(w http.ResponseWriter, r *http.Request) {
	var req PreviewCoverRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}
	preview, err := h.service.PreviewCover(r.Context(), req.Barcode, req.Title, req.Artist, req.Country, req.Year)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := PreviewCoverResponse{
		CoverURL: preview.CoverURL,
		Title:    preview.Title,
		Artist:   preview.Artist,
		Year:     preview.Year,
		Genres:   preview.Genres,
		Formats:  preview.Formats,
		Country:  preview.Country,
		Label:    preview.Label,
		Found:    preview.Found,
	}
	if preview.Prices != nil {
		response.Prices = &DiscogsPriceResponse{
			LowestPrice:  preview.Prices.LowestPrice,
			MedianPrice:  preview.Prices.MedianPrice,
			HighestPrice: preview.Prices.HighestPrice,
		}
	}
	if len(preview.Results) > 0 {
		response.Results = make([]PreviewResult, len(preview.Results))
		for i, result := range preview.Results {
			var formatNames []string
			for _, format := range result.Formats {
				formatNames = append(formatNames, format.Name)
				for _, desc := range format.Descriptions {
					formatNames = append(formatNames, desc)
				}
			}
			label := ""
			if len(result.Labels) > 0 {
				label = result.Labels[0].Name
			}
			response.Results[i] = PreviewResult{
				ID:       result.ID,
				Title:    result.Title,
				CoverURL: result.CoverURL,
				Year:     result.Year,
				Country:  result.Country,
				Formats:  formatNames,
				Genres:   result.Genres,
				Label:    label,
			}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *DiscHandler) UploadCover(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Fichier trop volumineux (max 10MB)", http.StatusBadRequest)
		return
	}
	discIDStr := r.FormValue("disc_id")
	var discID int = 0
	if discIDStr != "" {
		discID, err = strconv.Atoi(discIDStr)
		if err != nil {
			http.Error(w, "disc_id invalide", http.StatusBadRequest)
			return
		}
	}
	title := r.FormValue("title")
	artist := r.FormValue("artist")
	fmt.Printf("📎 [UploadCover] Paramètres reçus: disc_id=%s, title=%s, artist=%s\n",
		discIDStr, title, artist)

	var discTitle, discArtist string
	if discID > 0 {
		disc, err := h.service.GetDiscByID(r.Context(), discID)
		if err != nil {
			http.Error(w, "Disque non trouvé", http.StatusNotFound)
			return
		}
		discTitle = disc.Title
		discArtist = disc.ArtistName
		fmt.Printf("📎 Infos depuis base: discTitle=%s, discArtist=%s\n", discTitle, discArtist)
	} else {
		discTitle = title
		discArtist = artist
		if discArtist == "" {
			discArtist = "Unknown"
		}
		fmt.Printf("📎 Infos depuis formulaire: discTitle=%s, discArtist=%s\n", discTitle, discArtist)
	}
	if discTitle == "" {
		discTitle = "Unknown"
	}
	file, header, err := r.FormFile("cover")
	if err != nil {
		http.Error(w, "Fichier non trouvé dans la requête", http.StatusBadRequest)
		return
	}
	defer file.Close()
	contentType := header.Header.Get("Content-Type")
	validTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/webp": true,
	}
	if !validTypes[contentType] {
		http.Error(w, "Format de fichier non supporté. Utilisez JPG, PNG ou WebP.", http.StatusBadRequest)
		return
	}
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		switch contentType {
		case "image/jpeg", "image/jpg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/webp":
			ext = ".webp"
		default:
			ext = ".jpg"
		}
	}
	cleanTitle := sanitizeFilename(discTitle)
	cleanArtist := sanitizeFilename(discArtist)
	if cleanArtist == "" {
		cleanArtist = "Unknown"
	}
	if len(cleanTitle) > 50 {
		cleanTitle = cleanTitle[:50]
	}
	if len(cleanArtist) > 30 {
		cleanArtist = cleanArtist[:30]
	}
	fileDiscID := discID
	if fileDiscID == 0 {
		// ✅ CORRECTION: Utiliser un ID temporaire plus unique (nanoseconds + partie aléatoire)
		// pour éviter les collisions avec des IDs de disques existants
		fileDiscID = int(time.Now().UnixNano()%9000000000) + 1000000000 // ID entre 1 milliard et 10 milliards
		fmt.Printf("📎 ID temporaire généré: %d\n", fileDiscID)
	}
	filename := fmt.Sprintf("%s - %s (%d)%s", cleanArtist, cleanTitle, fileDiscID, ext)
	fmt.Printf("📎 Nom généré: %s\n", filename)
	uploadsDir := "./uploads/covers"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		http.Error(w, "Erreur création du répertoire", http.StatusInternalServerError)
		return
	}
	filePath := filepath.Join(uploadsDir, filename)
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
	relativePath := fmt.Sprintf("/uploads/covers/%s", filename)
	fmt.Printf("✅ Pochette uploadée avec succès: %s\n", relativePath)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"filePath": relativePath,
		"message":  "Fichier uploadé avec succès",
		"filename": filename,
	})
}

func sanitizeFilename(name string) string {
	cleaned := name
	replacements := map[string]string{
		"à": "a", "á": "a", "â": "a", "ã": "a", "ä": "a", "å": "a",
		"è": "e", "é": "e", "ê": "e", "ë": "e",
		"ì": "i", "í": "i", "î": "i", "ï": "i",
		"ò": "o", "ó": "o", "ô": "o", "õ": "o", "ö": "o", "ø": "o",
		"ù": "u", "ú": "u", "û": "u", "ü": "u",
		"ý": "y", "ÿ": "y",
		"ñ": "n", "ç": "c",
		"À": "A", "Á": "A", "Â": "A", "Ã": "A", "Ä": "A", "Å": "A",
		"È": "E", "É": "E", "Ê": "E", "Ë": "E",
		"Ì": "I", "Í": "I", "Î": "I", "Ï": "I",
		"Ò": "O", "Ó": "O", "Ô": "O", "Õ": "O", "Ö": "O", "Ø": "O",
		"Ù": "U", "Ú": "U", "Û": "U", "Ü": "U",
		"Ý": "Y", "Ÿ": "Y",
		"Ñ": "N", "Ç": "C",
		"æ": "ae", "œ": "oe",
		"Æ": "AE", "Œ": "OE",
	}
	for old, new := range replacements {
		cleaned = strings.ReplaceAll(cleaned, old, new)
	}
	cleaned = strings.ReplaceAll(cleaned, "/", "-")
	cleaned = strings.ReplaceAll(cleaned, "\\", "-")
	cleaned = strings.ReplaceAll(cleaned, ":", "-")
	cleaned = strings.ReplaceAll(cleaned, "*", "")
	cleaned = strings.ReplaceAll(cleaned, "?", "")
	cleaned = strings.ReplaceAll(cleaned, "\"", "")
	cleaned = strings.ReplaceAll(cleaned, "<", "")
	cleaned = strings.ReplaceAll(cleaned, ">", "")
	cleaned = strings.ReplaceAll(cleaned, "|", "")
	cleaned = strings.ReplaceAll(cleaned, "&", "and")
	cleaned = strings.ReplaceAll(cleaned, "'", "")
	cleaned = strings.ReplaceAll(cleaned, "`", "")
	cleaned = strings.ReplaceAll(cleaned, "\u2018", "")
	cleaned = strings.ReplaceAll(cleaned, "\u2019", "")
	cleaned = strings.ReplaceAll(cleaned, "\u201C", "")
	cleaned = strings.ReplaceAll(cleaned, "\u201D", "")
	cleaned = strings.Join(strings.Fields(cleaned), " ")
	cleaned = strings.ReplaceAll(cleaned, " ", "_")
	for strings.Contains(cleaned, "..") {
		cleaned = strings.ReplaceAll(cleaned, "..", ".")
	}
	return strings.TrimSpace(cleaned)
}
