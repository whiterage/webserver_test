package service

import (
	"context"
	"geo-alert-core/internal/domain"
	"geo-alert-core/internal/repository"
)

// business logic for stats
type StatsService struct {
	repo repository.IncidentRepository
}

func NewStatsService(repo repository.IncidentRepository) *StatsService {
	return &StatsService{repo: repo}
}

// get stats for incidents for last N minutes
func (s *StatsService) GetStats(ctx context.Context, minutes int) ([]*domain.IncidentStats, error) {
	if minutes <= 0 {
		minutes = 60 // default
	}

	return s.repo.GetStats(ctx, minutes)
}
