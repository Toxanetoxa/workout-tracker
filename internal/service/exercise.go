package service

import (
	"context"
	"strings"

	"github.com/Toxanetoxa/workout-tracker/internal/domain"
)

type ExerciseStore interface {
	Create(ctx context.Context, name string) (domain.Exercise, error)
}

type ExerciseService struct {
	store ExerciseStore
}

func NewExerciseService(store ExerciseStore) *ExerciseService {
	return &ExerciseService{store: store}
}

func (s *ExerciseService) Create(ctx context.Context, name string) (domain.Exercise, error) {
	return s.store.Create(ctx, strings.TrimSpace(name))
}
