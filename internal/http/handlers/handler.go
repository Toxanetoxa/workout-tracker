package handlers

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/Toxanetoxa/workout-tracker/internal/domain"
)

type ExerciseService interface {
	Create(ctx context.Context, name string) (domain.Exercise, error)
}

type ExecutionService interface {
	Create(ctx context.Context, userID string, exerciseID int64, performedAt time.Time) (domain.Execution, error)
}

type StatisticsService interface {
	GetByUser(ctx context.Context, userID string) (domain.UserStatistics, error)
}

type Handler struct {
	validate          *validator.Validate
	exerciseService   ExerciseService
	executionService  ExecutionService
	statisticsService StatisticsService
}

func New(validate *validator.Validate, exerciseService ExerciseService, executionService ExecutionService, statisticsService StatisticsService) *Handler {
	registerValidationRules(validate)

	return &Handler{
		validate:          validate,
		exerciseService:   exerciseService,
		executionService:  executionService,
		statisticsService: statisticsService,
	}
}
