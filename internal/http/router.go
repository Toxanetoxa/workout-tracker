package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"github.com/Toxanetoxa/workout-tracker/internal/http/handlers"
	appmiddleware "github.com/Toxanetoxa/workout-tracker/internal/http/middleware"
)

func NewRouter(logger *slog.Logger, h *handlers.Handler) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(appmiddleware.RequestLogger(logger))
	r.Use(chimiddleware.Recoverer)

	r.Get("/health", h.Health)
	r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))
	r.Post("/exercises", h.CreateExercise)
	r.Post("/executions", h.CreateExecution)
	r.Get("/users/{userID}/statistics", h.GetUserStatistics)

	return r
}
