package handlers

import "github.com/jackc/pgx/v5/pgconn"

const (
	pgUniqueViolation     = "23505"
	pgForeignKeyViolation = "23503"
)

func isUniqueExerciseNameError(err error) bool {
	var pgErr *pgconn.PgError
	if !asPgError(err, &pgErr) {
		return false
	}

	return pgErr.Code == pgUniqueViolation && pgErr.ConstraintName == "exercises_name_key"
}

func isExerciseForeignKeyError(err error) bool {
	var pgErr *pgconn.PgError
	if !asPgError(err, &pgErr) {
		return false
	}

	return pgErr.Code == pgForeignKeyViolation && pgErr.ConstraintName == "executions_exercise_id_fkey"
}

func asPgError(err error, target **pgconn.PgError) bool {
	if err == nil {
		return false
	}

	if e, ok := err.(*pgconn.PgError); ok {
		*target = e
		return true
	}

	if unwrapped, ok := err.(interface{ Unwrap() error }); ok {
		return asPgError(unwrapped.Unwrap(), target)
	}

	return false
}
