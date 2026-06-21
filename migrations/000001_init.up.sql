CREATE TABLE exercises (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE executions (
    id BIGSERIAL PRIMARY KEY,
    user_id TEXT NOT NULL,
    exercise_id BIGINT NOT NULL REFERENCES exercises(id) ON DELETE RESTRICT,
    performed_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX executions_user_performed_at_idx ON executions (user_id, performed_at);
CREATE INDEX executions_exercise_id_idx ON executions (exercise_id);
