package service

import (
	"context"

	"github.com/Toxanetoxa/workout-tracker/internal/domain"
)

type StatisticsStore interface {
	GetByUser(ctx context.Context, userID string) (domain.UserStatistics, error)
}

type StatisticsService struct {
	store StatisticsStore
}

func NewStatisticsService(store StatisticsStore) *StatisticsService {
	return &StatisticsService{store: store}
}

func (s *StatisticsService) GetByUser(ctx context.Context, userID string) (domain.UserStatistics, error) {
	return s.store.GetByUser(ctx, userID)
}
