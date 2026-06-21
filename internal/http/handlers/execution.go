package handlers

import (
	"net/http"
	"time"
)

type CreateExecutionRequest struct {
	UserID      string     `json:"user_id" validate:"required,min=3,max=120,alphanumdash"`
	ExerciseID  int64      `json:"exercise_id" validate:"required,gt=0"`
	PerformedAt *time.Time `json:"performed_at"`
}

// CreateExecution godoc
// @Summary Зафиксировать выполнение упражнения
// @Tags executions
// @Accept json
// @Produce json
// @Param request body CreateExecutionRequest true "Данные выполнения"
// @Success 201 {object} domain.Execution
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /executions [post]
func (h *Handler) CreateExecution(w http.ResponseWriter, r *http.Request) {
	var req CreateExecutionRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeValidationError(w, http.StatusUnprocessableEntity, validationErrors(err))
		return
	}

	var performedAt time.Time
	if req.PerformedAt != nil {
		performedAt = req.PerformedAt.UTC()
		if performedAt.After(time.Now().UTC()) {
			writeValidationError(w, http.StatusUnprocessableEntity, []ValidationErrorItem{
				{Field: "performed_at", Rule: "future", Message: "performed_at must not be in the future"},
			})
			return
		}
	}

	execution, err := h.executionService.Create(r.Context(), req.UserID, req.ExerciseID, performedAt)
	if err != nil {
		if isExerciseForeignKeyError(err) {
			writeError(w, http.StatusNotFound, "exercise not found")
			return
		}

		writeError(w, http.StatusInternalServerError, "create execution failed")
		return
	}

	writeJSON(w, http.StatusCreated, execution)
}
