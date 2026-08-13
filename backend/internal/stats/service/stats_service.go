package service

import "records-manager/backend/internal/stats/repository"

type StatsService interface {
	GetStats() (map[string]int, error)
}

type statsService struct {
	repo repository.StatsRepository
}

func NewStatsService(repo repository.StatsRepository) StatsService {
	return &statsService{repo: repo}
}

func (s *statsService) GetStats() (map[string]int, error) {
	return s.repo.GetCounts()
}
