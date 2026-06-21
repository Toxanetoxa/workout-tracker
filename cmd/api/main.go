package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"

	_ "github.com/Toxanetoxa/workout-tracker/docs"
	"github.com/Toxanetoxa/workout-tracker/internal/config"
	"github.com/Toxanetoxa/workout-tracker/internal/database"
	httpapi "github.com/Toxanetoxa/workout-tracker/internal/http"
	"github.com/Toxanetoxa/workout-tracker/internal/http/handlers"
	"github.com/Toxanetoxa/workout-tracker/internal/repository"
	"github.com/Toxanetoxa/workout-tracker/internal/service"
)

// @title Workout Tracker API
// @version 1.0
// @description API для учета выполненных упражнений и получения статистики по пользователю.
// @host localhost:3000
// @BasePath /
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := database.NewPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("connect database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	validate := validator.New()

	exerciseRepo := repository.NewExerciseRepository(pool)
	executionRepo := repository.NewExecutionRepository(pool)
	statisticsRepo := repository.NewStatisticsRepository(pool)

	exerciseService := service.NewExerciseService(exerciseRepo)
	executionService := service.NewExecutionService(executionRepo)
	statisticsService := service.NewStatisticsService(statisticsRepo)

	apiHandlers := handlers.New(validate, exerciseService, executionService, statisticsService)
	router := httpapi.NewRouter(logger, apiHandlers)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("api started", slog.String("addr", cfg.HTTPAddr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("listen and serve", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown server", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("api stopped")
}
