package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Toxanetoxa/workout-tracker/internal/domain"
)

type ExerciseRepository struct {
	pool *pgxpool.Pool
}

func NewExerciseRepository(pool *pgxpool.Pool) *ExerciseRepository {
	return &ExerciseRepository{pool: pool}
}

func (r *ExerciseRepository) Create(ctx context.Context, name string) (domain.Exercise, error) {
	const query = `
		INSERT INTO exercises (name)
		VALUES ($1)
		RETURNING id, name, created_at
	`

	var exercise domain.Exercise
	if err := r.pool.QueryRow(ctx, query, name).Scan(&exercise.ID, &exercise.Name, &exercise.CreatedAt); err != nil {
		return domain.Exercise{}, fmt.Errorf("create exercise: %w", err)
	}

	return exercise, nil
}
