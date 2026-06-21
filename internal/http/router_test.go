package httpapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/Toxanetoxa/workout-tracker/internal/domain"
	"github.com/Toxanetoxa/workout-tracker/internal/http/handlers"
)

func TestHealth(t *testing.T) {
	t.Parallel()

	handler := handlers.New(validator.New(), fakeExercises{}, fakeExecutions{}, fakeStatistics{})
	router := NewRouter(slog.Default(), handler)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body["status"] != "ok" {
		t.Fatalf("expected ok status, got %q", body["status"])
	}
}

type fakeExercises struct{}

func (fakeExercises) Create(ctx context.Context, name string) (domain.Exercise, error) {
	return domain.Exercise{Name: name}, nil
}

type fakeExecutions struct{}

func (fakeExecutions) Create(ctx context.Context, userID string, exerciseID int64, performedAt time.Time) (domain.Execution, error) {
	return domain.Execution{UserID: userID, ExerciseID: exerciseID, PerformedAt: performedAt}, nil
}

type fakeStatistics struct{}

func (fakeStatistics) GetByUser(ctx context.Context, userID string) (domain.UserStatistics, error) {
	return domain.UserStatistics{UserID: userID}, nil
}
