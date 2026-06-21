package service

import (
	"context"
	"time"

	"github.com/Toxanetoxa/workout-tracker/internal/domain"
)

type ExecutionStore interface {
	Create(ctx context.Context, userID string, exerciseID int64, performedAt time.Time) (domain.Execution, error)
}

type ExecutionService struct {
	store ExecutionStore
}

func NewExecutionService(store ExecutionStore) *ExecutionService {
	return &ExecutionService{store: store}
}

func (s *ExecutionService) Create(ctx context.Context, userID string, exerciseID int64, performedAt time.Time) (domain.Execution, error) {
	if performedAt.IsZero() {
		performedAt = time.Now().UTC()
	}

	return s.store.Create(ctx, userID, exerciseID, performedAt)
}
