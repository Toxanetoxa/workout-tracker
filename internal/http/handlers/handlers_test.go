package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/Toxanetoxa/workout-tracker/internal/domain"
)

func TestCreateExercise(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		body           string
		serviceErr     error
		wantStatus     int
		wantValidation []ValidationErrorItem
	}{
		{
			name:       "created",
			body:       `{"name":"Bench Press"}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "invalid json",
			body:       `{`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:           "validation failed",
			body:           `{"name":""}`,
			wantStatus:     http.StatusUnprocessableEntity,
			wantValidation: []ValidationErrorItem{{Field: "name", Rule: "required", Message: "name is required"}},
		},
		{
			name:           "multiple validation errors",
			body:           `{"name":" "}`,
			wantStatus:     http.StatusUnprocessableEntity,
			wantValidation: []ValidationErrorItem{{Field: "name", Rule: "trimmed", Message: "name must not contain leading or trailing spaces"}},
		},
		{
			name:       "unknown field",
			body:       `{"name":"Bench Press","extra":true}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:           "trimmed validation failed",
			body:           `{"name":" Bench Press "}`,
			wantStatus:     http.StatusUnprocessableEntity,
			wantValidation: []ValidationErrorItem{{Field: "name", Rule: "trimmed", Message: "name must not contain leading or trailing spaces"}},
		},
		{
			name:       "duplicate exercise",
			body:       `{"name":"Bench Press"}`,
			serviceErr: &pgconn.PgError{Code: "23505", ConstraintName: "exercises_name_key"},
			wantStatus: http.StatusConflict,
		},
		{
			name:       "different unique violation",
			body:       `{"name":"Bench Press"}`,
			serviceErr: &pgconn.PgError{Code: "23505", ConstraintName: "other_unique_key"},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "service error",
			body:       `{"name":"Bench Press"}`,
			serviceErr: errors.New("database unavailable"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := newTestHandler(testExerciseService{
				create: func(ctx context.Context, name string) (domain.Exercise, error) {
					return domain.Exercise{ID: 1, Name: name}, tt.serviceErr
				},
			}, testExecutionService{}, testStatisticsService{})

			req := httptest.NewRequest(http.MethodPost, "/exercises", bytes.NewBufferString(tt.body))
			rec := httptest.NewRecorder()

			h.CreateExercise(rec, req)

			assertStatus(t, rec.Code, tt.wantStatus)
			assertValidationResponse(t, rec.Body.Bytes(), tt.wantValidation)
		})
	}
}

func TestCreateExecution(t *testing.T) {
	t.Parallel()

	performedAt := "2026-06-21T10:00:00Z"

	tests := []struct {
		name           string
		body           string
		serviceErr     error
		wantStatus     int
		wantValidation []ValidationErrorItem
	}{
		{
			name:       "created",
			body:       `{"user_id":"user-1","exercise_id":1,"performed_at":"` + performedAt + `"}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "invalid json",
			body:       `{`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "validation failed",
			body:       `{"user_id":"","exercise_id":0}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantValidation: []ValidationErrorItem{
				{Field: "exercise_id", Rule: "required", Message: "exercise_id is required"},
				{Field: "user_id", Rule: "required", Message: "user_id is required"},
			},
		},
		{
			name:       "invalid user id format",
			body:       `{"user_id":"user 1","exercise_id":1}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantValidation: []ValidationErrorItem{
				{Field: "user_id", Rule: "alphanumdash", Message: "user_id must contain only letters, digits, dashes and underscores"},
			},
		},
		{
			name:       "performed at in the future",
			body:       `{"user_id":"user-1","exercise_id":1,"performed_at":"2099-01-01T00:00:00Z"}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantValidation: []ValidationErrorItem{
				{Field: "performed_at", Rule: "future", Message: "performed_at must not be in the future"},
			},
		},
		{
			name:       "unknown field",
			body:       `{"user_id":"user-1","exercise_id":1,"extra":true}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "exercise not found",
			body:       `{"user_id":"user-1","exercise_id":99}`,
			serviceErr: &pgconn.PgError{Code: "23503", ConstraintName: "executions_exercise_id_fkey"},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "different foreign key violation",
			body:       `{"user_id":"user-1","exercise_id":99}`,
			serviceErr: &pgconn.PgError{Code: "23503", ConstraintName: "other_fk"},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "service error",
			body:       `{"user_id":"user-1","exercise_id":1}`,
			serviceErr: errors.New("database unavailable"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := newTestHandler(testExerciseService{}, testExecutionService{
				create: func(ctx context.Context, userID string, exerciseID int64, performedAt time.Time) (domain.Execution, error) {
					return domain.Execution{ID: 1, UserID: userID, ExerciseID: exerciseID, PerformedAt: performedAt}, tt.serviceErr
				},
			}, testStatisticsService{})

			req := httptest.NewRequest(http.MethodPost, "/executions", bytes.NewBufferString(tt.body))
			rec := httptest.NewRecorder()

			h.CreateExecution(rec, req)

			assertStatus(t, rec.Code, tt.wantStatus)
			assertValidationResponse(t, rec.Body.Bytes(), tt.wantValidation)
		})
	}
}

func TestGetUserStatistics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		serviceErr     error
		wantStatus     int
		wantValidation []ValidationErrorItem
	}{
		{
			name:       "ok",
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid user id",
			wantStatus: http.StatusUnprocessableEntity,
			wantValidation: []ValidationErrorItem{
				{Field: "user_id", Rule: "alphanumdash", Message: "user_id must contain only letters, digits, dashes and underscores"},
			},
		},
		{
			name:       "service error",
			serviceErr: errors.New("database unavailable"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := newTestHandler(testExerciseService{}, testExecutionService{}, testStatisticsService{
				getByUser: func(ctx context.Context, userID string) (domain.UserStatistics, error) {
					return domain.UserStatistics{UserID: userID, Total: 3, Today: 1}, tt.serviceErr
				},
			})

			router := chi.NewRouter()
			router.Get("/users/{userID}/statistics", h.GetUserStatistics)

			path := "/users/user-1/statistics"
			if tt.name == "invalid user id" {
				path = "/users/user.1/statistics"
			}
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assertStatus(t, rec.Code, tt.wantStatus)
			assertValidationResponse(t, rec.Body.Bytes(), tt.wantValidation)
		})
	}
}

func TestGetUserStatisticsResponse(t *testing.T) {
	t.Parallel()

	h := newTestHandler(testExerciseService{}, testExecutionService{}, testStatisticsService{
		getByUser: func(ctx context.Context, userID string) (domain.UserStatistics, error) {
			return domain.UserStatistics{
				UserID: userID,
				Total:  10,
				Today:  2,
				Last7Days: []domain.DailyExecutionCount{
					{Date: "2026-06-21", Count: 2},
				},
			}, nil
		},
	})

	router := chi.NewRouter()
	router.Get("/users/{userID}/statistics", h.GetUserStatistics)

	req := httptest.NewRequest(http.MethodGet, "/users/user-1/statistics", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assertStatus(t, rec.Code, http.StatusOK)

	var response domain.UserStatistics
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.UserID != "user-1" || response.Total != 10 || response.Today != 2 {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func newTestHandler(exercises ExerciseService, executions ExecutionService, statistics StatisticsService) *Handler {
	return New(validator.New(), exercises, executions, statistics)
}

func assertStatus(t *testing.T, got, want int) {
	t.Helper()

	if got != want {
		t.Fatalf("expected status %d, got %d", want, got)
	}
}

func assertValidationResponse(t *testing.T, body []byte, want []ValidationErrorItem) {
	t.Helper()

	if len(want) == 0 {
		return
	}

	var response ValidationErrorResponse
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("decode validation response: %v", err)
	}

	if len(response.Errors) != len(want) {
		t.Fatalf("expected %d validation errors, got %d: %s", len(want), len(response.Errors), string(body))
	}

	for i, expected := range want {
		if response.Errors[i] != expected {
			t.Fatalf("unexpected validation error at %d: got %+v want %+v", i, response.Errors[i], expected)
		}
	}
}

type testExerciseService struct {
	create func(ctx context.Context, name string) (domain.Exercise, error)
}

func (s testExerciseService) Create(ctx context.Context, name string) (domain.Exercise, error) {
	if s.create == nil {
		return domain.Exercise{}, nil
	}

	return s.create(ctx, name)
}

type testExecutionService struct {
	create func(ctx context.Context, userID string, exerciseID int64, performedAt time.Time) (domain.Execution, error)
}

func (s testExecutionService) Create(ctx context.Context, userID string, exerciseID int64, performedAt time.Time) (domain.Execution, error) {
	if s.create == nil {
		return domain.Execution{}, nil
	}

	return s.create(ctx, userID, exerciseID, performedAt)
}

type testStatisticsService struct {
	getByUser func(ctx context.Context, userID string) (domain.UserStatistics, error)
}

func (s testStatisticsService) GetByUser(ctx context.Context, userID string) (domain.UserStatistics, error) {
	if s.getByUser == nil {
		return domain.UserStatistics{}, nil
	}

	return s.getByUser(ctx, userID)
}
