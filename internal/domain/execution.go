package domain

import "time"

type Execution struct {
	ID          int64     `json:"id"`
	UserID      string    `json:"user_id"`
	ExerciseID  int64     `json:"exercise_id"`
	PerformedAt time.Time `json:"performed_at"`
	CreatedAt   time.Time `json:"created_at"`
}
