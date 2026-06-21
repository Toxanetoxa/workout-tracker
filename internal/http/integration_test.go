package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"

	"github.com/Toxanetoxa/workout-tracker/internal/database"
	"github.com/Toxanetoxa/workout-tracker/internal/domain"
	"github.com/Toxanetoxa/workout-tracker/internal/http/handlers"
	"github.com/Toxanetoxa/workout-tracker/internal/repository"
	"github.com/Toxanetoxa/workout-tracker/internal/service"
)

func TestHTTPFlowWithPostgres(t *testing.T) {
	pool := newHTTPIntegrationPool(t)
	t.Cleanup(pool.Close)

	exerciseRepo := repository.NewExerciseRepository(pool)
	executionRepo := repository.NewExecutionRepository(pool)
	statsRepo := repository.NewStatisticsRepository(pool)

	apiHandlers := handlers.New(
		validator.New(),
		service.NewExerciseService(exerciseRepo),
		service.NewExecutionService(executionRepo),
		service.NewStatisticsService(statsRepo),
	)
	router := NewRouter(slog.New(slog.NewTextHandler(os.Stdout, nil)), apiHandlers)

	exerciseID := createExerciseViaAPI(t, router, "Bench Press")
	createExecutionViaAPI(t, router, "user-1", exerciseID, "2026-06-21T10:00:00Z")

	stats := getStatisticsViaAPI(t, router, "user-1")

	if stats.UserID != "user-1" {
		t.Fatalf("unexpected user_id: %s", stats.UserID)
	}
	if stats.Total != 1 || stats.Today != 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
	if len(stats.Last7Days) != 7 {
		t.Fatalf("unexpected last_7_days len: %d", len(stats.Last7Days))
	}
}

func newHTTPIntegrationPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = os.Getenv("DATABASE_URL")
	}
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL or DATABASE_URL is required for integration tests")
	}

	pool, err := database.NewPostgres(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("connect database: %v", err)
	}

	if err := resetHTTPSchema(context.Background(), pool); err != nil {
		t.Fatalf("reset schema: %v", err)
	}
	if err := applyHTTPMigration(context.Background(), pool); err != nil {
		t.Fatalf("apply migration: %v", err)
	}

	return pool
}

func resetHTTPSchema(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, `DROP SCHEMA public CASCADE;`); err != nil {
		return err
	}

	_, err := pool.Exec(ctx, `CREATE SCHEMA public;`)
	return err
}

func applyHTTPMigration(ctx context.Context, pool *pgxpool.Pool) error {
	migrationPath := filepath.Join("..", "..", "migrations", "000001_init.up.sql")
	sqlBytes, err := os.ReadFile(migrationPath)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, string(sqlBytes))
	return err
}

func createExerciseViaAPI(t *testing.T, router http.Handler, name string) int64 {
	t.Helper()

	body := []byte(`{"name":"` + name + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/exercises", bytesReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var exercise domain.Exercise
	if err := json.NewDecoder(rec.Body).Decode(&exercise); err != nil {
		t.Fatalf("decode exercise: %v", err)
	}

	return exercise.ID
}

func createExecutionViaAPI(t *testing.T, router http.Handler, userID string, exerciseID int64, performedAt string) {
	t.Helper()

	payload := map[string]any{
		"user_id":      userID,
		"exercise_id":  exerciseID,
		"performed_at": performedAt,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/executions", bytesReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func getStatisticsViaAPI(t *testing.T, router http.Handler, userID string) domain.UserStatistics {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/users/"+userID+"/statistics", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var stats domain.UserStatistics
	if err := json.NewDecoder(rec.Body).Decode(&stats); err != nil {
		t.Fatalf("decode statistics: %v", err)
	}

	return stats
}

func bytesReader(b []byte) *bytes.Reader {
	return bytes.NewReader(b)
}
