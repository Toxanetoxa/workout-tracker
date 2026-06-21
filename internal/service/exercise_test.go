package service

import (
	"context"
	"testing"

	"github.com/Toxanetoxa/workout-tracker/internal/domain"
)

func TestExerciseServiceCreateTrimsName(t *testing.T) {
	t.Parallel()

	store := exerciseStoreFunc(func(ctx context.Context, name string) (domain.Exercise, error) {
		if name != "Bench Press" {
			t.Fatalf("expected trimmed name, got %q", name)
		}

		return domain.Exercise{Name: name}, nil
	})

	svc := NewExerciseService(store)

	exercise, err := svc.Create(context.Background(), "  Bench Press  ")
	if err != nil {
		t.Fatalf("create exercise: %v", err)
	}

	if exercise.Name != "Bench Press" {
		t.Fatalf("expected exercise name %q, got %q", "Bench Press", exercise.Name)
	}
}

type exerciseStoreFunc func(ctx context.Context, name string) (domain.Exercise, error)

func (f exerciseStoreFunc) Create(ctx context.Context, name string) (domain.Exercise, error) {
	return f(ctx, name)
}
