package service

import (
	"context"
	"testing"
	"time"

	"github.com/Toxanetoxa/workout-tracker/internal/domain"
)

func TestExecutionServiceCreateUsesProvidedPerformedAt(t *testing.T) {
	t.Parallel()

	wantPerformedAt := time.Date(2026, 6, 21, 10, 0, 0, 0, time.UTC)
	store := executionStoreFunc(func(ctx context.Context, userID string, exerciseID int64, performedAt time.Time) (domain.Execution, error) {
		if !performedAt.Equal(wantPerformedAt) {
			t.Fatalf("expected performedAt %s, got %s", wantPerformedAt, performedAt)
		}

		return domain.Execution{UserID: userID, ExerciseID: exerciseID, PerformedAt: performedAt}, nil
	})

	svc := NewExecutionService(store)

	execution, err := svc.Create(context.Background(), "user-1", 1, wantPerformedAt)
	if err != nil {
		t.Fatalf("create execution: %v", err)
	}

	if !execution.PerformedAt.Equal(wantPerformedAt) {
		t.Fatalf("expected execution performedAt %s, got %s", wantPerformedAt, execution.PerformedAt)
	}
}

func TestExecutionServiceCreateDefaultsPerformedAt(t *testing.T) {
	t.Parallel()

	store := executionStoreFunc(func(ctx context.Context, userID string, exerciseID int64, performedAt time.Time) (domain.Execution, error) {
		if performedAt.IsZero() {
			t.Fatal("expected default performedAt to be set")
		}

		if performedAt.Location() != time.UTC {
			t.Fatalf("expected UTC location, got %s", performedAt.Location())
		}

		return domain.Execution{UserID: userID, ExerciseID: exerciseID, PerformedAt: performedAt}, nil
	})

	svc := NewExecutionService(store)

	execution, err := svc.Create(context.Background(), "user-1", 1, time.Time{})
	if err != nil {
		t.Fatalf("create execution: %v", err)
	}

	if execution.PerformedAt.IsZero() {
		t.Fatal("expected execution performedAt to be set")
	}
}

type executionStoreFunc func(ctx context.Context, userID string, exerciseID int64, performedAt time.Time) (domain.Execution, error)

func (f executionStoreFunc) Create(ctx context.Context, userID string, exerciseID int64, performedAt time.Time) (domain.Execution, error) {
	return f(ctx, userID, exerciseID, performedAt)
}
