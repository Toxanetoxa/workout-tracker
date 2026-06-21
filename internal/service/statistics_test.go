package service

import (
	"context"
	"testing"

	"github.com/Toxanetoxa/workout-tracker/internal/domain"
)

func TestStatisticsServiceGetByUser(t *testing.T) {
	t.Parallel()

	store := statisticsStoreFunc(func(ctx context.Context, userID string) (domain.UserStatistics, error) {
		if userID != "user-1" {
			t.Fatalf("expected userID %q, got %q", "user-1", userID)
		}

		return domain.UserStatistics{UserID: userID, Total: 7, Today: 1}, nil
	})

	svc := NewStatisticsService(store)

	stats, err := svc.GetByUser(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("get statistics: %v", err)
	}

	if stats.UserID != "user-1" || stats.Total != 7 || stats.Today != 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
}

type statisticsStoreFunc func(ctx context.Context, userID string) (domain.UserStatistics, error)

func (f statisticsStoreFunc) GetByUser(ctx context.Context, userID string) (domain.UserStatistics, error) {
	return f(ctx, userID)
}
