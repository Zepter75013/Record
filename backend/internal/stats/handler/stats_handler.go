package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"records-manager/backend/internal/stats/service"
)

type StatsHandler struct {
	statsService service.StatsService
}

func NewStatsHandler(statsService service.StatsService) *StatsHandler {
	return &StatsHandler{
		statsService: statsService,
	}
}

func (h *StatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	log.Println("📊 Récupération des statistiques...")

	stats, err := h.statsService.GetStats()
	if err != nil {
		log.Printf("❌ Erreur lors de la récupération des statistiques: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Erreur lors de la récupération des statistiques"})
		return
	}

	log.Printf("✅ Statistiques récupérées: %+v", stats)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(stats); err != nil {
		log.Printf("❌ Erreur d'encodage JSON: %v", err)
		http.Error(w, "Erreur lors de l'encodage des données", http.StatusInternalServerError)
		return
	}
}
