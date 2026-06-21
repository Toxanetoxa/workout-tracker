package handlers

import "net/http"

// Health godoc
// @Summary Проверка состояния API
// @Tags service
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
