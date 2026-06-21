package handlers

import (
	"encoding/json"
	"net/http"
)

type CreateExerciseRequest struct {
	Name string `json:"name" validate:"required,min=2,max=120"`
}

// CreateExercise godoc
// @Summary Создать упражнение
// @Tags exercises
// @Accept json
// @Produce json
// @Param request body CreateExerciseRequest true "Данные упражнения"
// @Success 201 {object} domain.Exercise
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /exercises [post]
func (h *Handler) CreateExercise(w http.ResponseWriter, r *http.Request) {
	var req CreateExerciseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "validation failed")
		return
	}

	exercise, err := h.exerciseService.Create(r.Context(), req.Name)
	if err != nil {
		if isUniqueExerciseNameError(err) {
			writeError(w, http.StatusConflict, "exercise already exists")
			return
		}

		writeError(w, http.StatusInternalServerError, "create exercise failed")
		return
	}

	writeJSON(w, http.StatusCreated, exercise)
}
