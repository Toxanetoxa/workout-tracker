package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Toxanetoxa/workout-tracker/internal/domain"
)

type StatisticsRepository struct {
	pool *pgxpool.Pool
}

func NewStatisticsRepository(pool *pgxpool.Pool) *StatisticsRepository {
	return &StatisticsRepository{pool: pool}
}

func (r *StatisticsRepository) GetByUser(ctx context.Context, userID string) (domain.UserStatistics, error) {
	const totalsQuery = `
		SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE performed_at >= CURRENT_DATE AND performed_at < CURRENT_DATE + INTERVAL '1 day') AS today
		FROM executions
		WHERE user_id = $1
	`

	stats := domain.UserStatistics{UserID: userID}
	if err := r.pool.QueryRow(ctx, totalsQuery, userID).Scan(&stats.Total, &stats.Today); err != nil {
		return domain.UserStatistics{}, fmt.Errorf("get totals: %w", err)
	}

	const last7DaysQuery = `
		WITH days AS (
			SELECT generate_series(CURRENT_DATE - INTERVAL '6 days', CURRENT_DATE, INTERVAL '1 day')::date AS day
		)
		SELECT
			days.day::text,
			COUNT(executions.id)
		FROM days
		LEFT JOIN executions
			ON executions.user_id = $1
			AND executions.performed_at >= days.day
			AND executions.performed_at < days.day + INTERVAL '1 day'
		GROUP BY days.day
		ORDER BY days.day
	`

	rows, err := r.pool.Query(ctx, last7DaysQuery, userID)
	if err != nil {
		return domain.UserStatistics{}, fmt.Errorf("get last 7 days: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item domain.DailyExecutionCount
		if err := rows.Scan(&item.Date, &item.Count); err != nil {
			return domain.UserStatistics{}, fmt.Errorf("scan daily count: %w", err)
		}
		stats.Last7Days = append(stats.Last7Days, item)
	}
	if err := rows.Err(); err != nil {
		return domain.UserStatistics{}, fmt.Errorf("iterate daily counts: %w", err)
	}

	return stats, nil
}
