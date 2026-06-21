package domain

type DailyExecutionCount struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type UserStatistics struct {
	UserID    string                `json:"user_id"`
	Total     int64                 `json:"total"`
	Today     int64                 `json:"today"`
	Last7Days []DailyExecutionCount `json:"last_7_days"`
}
