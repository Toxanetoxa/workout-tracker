package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Toxanetoxa/workout-tracker/internal/domain"
)

type ExecutionRepository struct {
	pool *pgxpool.Pool
}

func NewExecutionRepository(pool *pgxpool.Pool) *ExecutionRepository {
	return &ExecutionRepository{pool: pool}
}

func (r *ExecutionRepository) Create(ctx context.Context, userID string, exerciseID int64, performedAt time.Time) (domain.Execution, error) {
	const query = `
		INSERT INTO executions (user_id, exercise_id, performed_at)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, exercise_id, performed_at, created_at
	`

	var execution domain.Execution
	if err := r.pool.QueryRow(ctx, query, userID, exerciseID, performedAt).
		Scan(&execution.ID, &execution.UserID, &execution.ExerciseID, &execution.PerformedAt, &execution.CreatedAt); err != nil {
		return domain.Execution{}, fmt.Errorf("create execution: %w", err)
	}

	return execution, nil
}
