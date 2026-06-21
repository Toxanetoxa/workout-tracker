package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// GetUserStatistics godoc
// @Summary Получить статистику пользователя
// @Tags statistics
// @Produce json
// @Param userID path string true "ID пользователя"
// @Success 200 {object} domain.UserStatistics
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{userID}/statistics [get]
func (h *Handler) GetUserStatistics(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "user id is required")
		return
	}

	stats, err := h.statisticsService.GetByUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "get statistics failed")
		return
	}

	writeJSON(w, http.StatusOK, stats)
}
