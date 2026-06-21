package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startedAt := time.Now()

			next.ServeHTTP(w, r)

			logger.Info("http request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Duration("duration", time.Since(startedAt)),
			)
		})
	}
}
