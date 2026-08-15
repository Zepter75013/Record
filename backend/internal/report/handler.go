package report

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const maxReportUploadBytes = 30 << 20 // 30 Mo

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Upload reçoit le rapport généré côté client (PDF/XLSX/DOCX/CSV) et le
// stocke comme nouveau "dernier rapport", en remplaçant celui qui existait
// avant, quel que soit son format.
func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxReportUploadBytes)

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "fichier trop volumineux ou requête invalide", http.StatusBadRequest)
		return
	}

	uploaded, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "fichier manquant", http.StatusBadRequest)
		return
	}
	defer uploaded.Close()

	criteria := r.FormValue("criteria")
	format := r.FormValue("format")
	if format == "" {
		format = "pdf"
	}

	meta, err := h.service.SaveLatest(criteria, format, uploaded)
	if err != nil {
		if errors.Is(err, ErrUnsupportedFormat) {
			http.Error(w, "format de rapport non supporté", http.StatusBadRequest)
			return
		}
		http.Error(w, "échec de l'enregistrement du rapport", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(meta)
}

func (h *Handler) GetLatestMetadata(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	meta, err := h.service.GetLatestMetadata()
	if err != nil {
		if errors.Is(err, ErrNoReport) {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		http.Error(w, "échec de la lecture du rapport", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(meta)
}

func (h *Handler) DownloadLatestFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	file, meta, err := h.service.OpenLatestFile()
	if err != nil {
		if errors.Is(err, ErrNoReport) {
			http.Error(w, "aucun rapport généré pour l'instant", http.StatusNotFound)
			return
		}

		http.Error(w, "échec de la lecture du rapport", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	extension := formatInfo[meta.Format].Extension

	w.Header().Set("Content-Type", meta.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="rapport-disques-manager.%s"`, extension))
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, file)
}
