package repository

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Toxanetoxa/workout-tracker/internal/domain"
)

func newIntegrationPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = os.Getenv("DATABASE_URL")
	}
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL or DATABASE_URL is required for integration tests")
	}

	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatalf("parse database config: %v", err)
	}
	cfg.MaxConns = 4

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		t.Fatalf("connect database: %v", err)
	}

	t.Cleanup(pool.Close)

	ctx := context.Background()

	if err := resetSchema(ctx, pool); err != nil {
		t.Fatalf("reset schema: %v", err)
	}

	if err := applyMigration(ctx, pool); err != nil {
		t.Fatalf("apply migration: %v", err)
	}

	return pool
}

func resetSchema(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, `DROP SCHEMA public CASCADE;`); err != nil {
		return err
	}

	_, err := pool.Exec(ctx, `CREATE SCHEMA public;`)
	return err
}

func applyMigration(ctx context.Context, pool *pgxpool.Pool) error {
	migrationPath := filepath.Join("..", "..", "migrations", "000001_init.up.sql")
	sqlBytes, err := os.ReadFile(migrationPath) // #nosec G304 -- test-only migration file path is fixed in repo
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, string(sqlBytes))
	return err
}

func seedExercise(t *testing.T, pool *pgxpool.Pool, name string) domain.Exercise {
	t.Helper()

	var exercise domain.Exercise
	err := pool.QueryRow(context.Background(), `
		INSERT INTO exercises (name)
		VALUES ($1)
		RETURNING id, name, created_at
	`, name).Scan(&exercise.ID, &exercise.Name, &exercise.CreatedAt)
	if err != nil {
		t.Fatalf("seed exercise: %v", err)
	}

	return exercise
}

func TestExerciseRepositoryCreate(t *testing.T) {
	pool := newIntegrationPool(t)
	repo := NewExerciseRepository(pool)

	exercise, err := repo.Create(context.Background(), "Bench Press")
	if err != nil {
		t.Fatalf("create exercise: %v", err)
	}

	if exercise.ID == 0 || exercise.Name != "Bench Press" || exercise.CreatedAt.IsZero() {
		t.Fatalf("unexpected exercise: %+v", exercise)
	}
}

func TestExecutionRepositoryCreate(t *testing.T) {
	pool := newIntegrationPool(t)
	exercises := NewExerciseRepository(pool)
	exercise, err := exercises.Create(context.Background(), "Squat")
	if err != nil {
		t.Fatalf("create exercise: %v", err)
	}

	repo := NewExecutionRepository(pool)
	performedAt := time.Date(2026, 6, 21, 10, 0, 0, 0, time.UTC)
	execution, err := repo.Create(context.Background(), "user-1", exercise.ID, performedAt)
	if err != nil {
		t.Fatalf("create execution: %v", err)
	}

	if execution.ID == 0 || execution.UserID != "user-1" || execution.ExerciseID != exercise.ID {
		t.Fatalf("unexpected execution: %+v", execution)
	}
	if !execution.PerformedAt.Equal(performedAt) {
		t.Fatalf("unexpected performed_at: got %s want %s", execution.PerformedAt, performedAt)
	}
}

func TestStatisticsRepositoryGetByUser(t *testing.T) {
	pool := newIntegrationPool(t)
	exercise := seedExercise(t, pool, "Pull-up")

	var dbToday time.Time
	if err := pool.QueryRow(context.Background(), `SELECT CURRENT_DATE`).Scan(&dbToday); err != nil {
		t.Fatalf("load current date: %v", err)
	}
	todayMorning := time.Date(dbToday.Year(), dbToday.Month(), dbToday.Day(), 9, 0, 0, 0, time.UTC)
	yesterday := todayMorning.Add(-24 * time.Hour)
	threeDaysAgo := todayMorning.Add(-72 * time.Hour)

	_, err := pool.Exec(context.Background(), `
		INSERT INTO executions (user_id, exercise_id, performed_at)
		VALUES
			($1, $2, $3),
			($1, $2, $4),
			($1, $2, $5),
			($6, $2, $3)
	`, "user-1", exercise.ID, todayMorning, yesterday, threeDaysAgo, "other-user")
	if err != nil {
		t.Fatalf("seed executions: %v", err)
	}

	repo := NewStatisticsRepository(pool)
	stats, err := repo.GetByUser(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("get statistics: %v", err)
	}

	if stats.UserID != "user-1" {
		t.Fatalf("unexpected user id: %s", stats.UserID)
	}
	if stats.Total != 3 {
		t.Fatalf("unexpected total: %d", stats.Total)
	}
	if stats.Today != 1 {
		t.Fatalf("unexpected today: %d", stats.Today)
	}
	if len(stats.Last7Days) != 7 {
		t.Fatalf("unexpected last 7 days len: %d", len(stats.Last7Days))
	}

	want := map[string]int64{
		todayMorning.Format("2006-01-02"): 1,
		yesterday.Format("2006-01-02"):    1,
		threeDaysAgo.Format("2006-01-02"): 1,
	}
	for _, item := range stats.Last7Days {
		if count, ok := want[item.Date]; ok {
			if item.Count != count {
				t.Fatalf("unexpected count for %s: got %d want %d", item.Date, item.Count, count)
			}
			delete(want, item.Date)
		}
	}
	if len(want) != 0 {
		t.Fatalf("missing dates in statistics: %v", want)
	}
}
