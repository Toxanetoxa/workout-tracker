WITH inserted_exercises AS (
    INSERT INTO exercises (name)
    VALUES
        ('Bench Press'),
        ('Squat'),
        ('Pull-up'),
        ('Plank'),
        ('Running')
    ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
    RETURNING id, name
),
all_exercises AS (
    SELECT id, name FROM inserted_exercises
    UNION
    SELECT id, name
    FROM exercises
    WHERE name IN ('Bench Press', 'Squat', 'Pull-up', 'Plank', 'Running')
),
seed_executions AS (
    SELECT 'demo-user-1' AS user_id, 'Bench Press' AS exercise_name, CURRENT_DATE + TIME '09:00' AS performed_at
    UNION ALL
    SELECT 'demo-user-1', 'Squat', CURRENT_DATE + TIME '10:00'
    UNION ALL
    SELECT 'demo-user-1', 'Pull-up', CURRENT_DATE - INTERVAL '1 day' + TIME '18:30'
    UNION ALL
    SELECT 'demo-user-1', 'Plank', CURRENT_DATE - INTERVAL '2 days' + TIME '07:45'
    UNION ALL
    SELECT 'demo-user-1', 'Running', CURRENT_DATE - INTERVAL '3 days' + TIME '20:00'
    UNION ALL
    SELECT 'demo-user-1', 'Bench Press', CURRENT_DATE - INTERVAL '6 days' + TIME '12:00'
    UNION ALL
    SELECT 'demo-user-2', 'Running', CURRENT_DATE + TIME '08:15'
    UNION ALL
    SELECT 'demo-user-2', 'Plank', CURRENT_DATE - INTERVAL '1 day' + TIME '19:00'
)
INSERT INTO executions (user_id, exercise_id, performed_at)
SELECT seed_executions.user_id, all_exercises.id, seed_executions.performed_at
FROM seed_executions
JOIN all_exercises ON all_exercises.name = seed_executions.exercise_name
WHERE NOT EXISTS (
    SELECT 1
    FROM executions
    WHERE executions.user_id = seed_executions.user_id
        AND executions.exercise_id = all_exercises.id
        AND executions.performed_at = seed_executions.performed_at
);
