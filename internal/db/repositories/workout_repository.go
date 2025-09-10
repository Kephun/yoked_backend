package repositories

import (
    "context"
    "fmt"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
    "yoked_backend/internal/models"
)

type WorkoutRepository interface {
    CreateWorkout(ctx context.Context, workout *models.Workout) error
    GetWorkoutByID(ctx context.Context, id string) (*models.Workout, error)
    GetUserWorkouts(ctx context.Context, userID string, page, limit int) ([]models.Workout, int, error)
    GetCurrentWorkout(ctx context.Context, userID string) (*models.Workout, error)
    UpdateWorkout(ctx context.Context, workout *models.Workout) error
    DeleteWorkout(ctx context.Context, id string) error
    CreateSession(ctx context.Context, session *models.Session) error
    GetSessionByID(ctx context.Context, id string) (*models.Session, error)
    UpdateSession(ctx context.Context, session *models.Session) error
    LogExercise(ctx context.Context, exercise *models.Exercise) error
    GetExerciseByID(ctx context.Context, id string) (*models.Exercise, error)
    UpdateExercise(ctx context.Context, exercise *models.Exercise) error
    GetWorkoutSessions(ctx context.Context, workoutID string) ([]models.Session, error)
    GetSessionExercises(ctx context.Context, sessionID string) ([]models.Exercise, error)
}

type workoutRepository struct {
    db *pgxpool.Pool
}

func NewWorkoutRepository(db *pgxpool.Pool) WorkoutRepository {
    return &workoutRepository{db: db}
}

// CreateWorkout inserts a new workout plan into the database
func (r *workoutRepository) CreateWorkout(ctx context.Context, workout *models.Workout) error {
    query := `
        INSERT INTO workouts (id, user_id, name, type, duration_weeks, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, created_at
    `

    err := r.db.QueryRow(ctx, query,
        workout.ID, workout.UserID, workout.Name, workout.Type, workout.DurationWeeks,
        time.Now(), time.Now(),
    ).Scan(&workout.ID, &workout.CreatedAt)

    if err != nil {
        return fmt.Errorf("failed to create workout: %w", err)
    }

    return nil
}

// GetWorkoutByID retrieves a workout by its ID
func (r *workoutRepository) GetWorkoutByID(ctx context.Context, id string) (*models.Workout, error) {
    query := `
        SELECT id, user_id, name, type, duration_weeks, created_at, updated_at, completed
        FROM workouts 
        WHERE id = $1 AND deleted_at IS NULL
    `

    var workout models.Workout
    err := r.db.QueryRow(ctx, query, id).Scan(
        &workout.ID, &workout.UserID, &workout.Name, &workout.Type,
        &workout.DurationWeeks, &workout.CreatedAt, &workout.UpdatedAt, &workout.Completed,
    )

    if err != nil {
        return nil, fmt.Errorf("failed to get workout by ID: %w", err)
    }

    return &workout, nil
}

// GetUserWorkouts retrieves paginated workout history for a user
func (r *workoutRepository) GetUserWorkouts(ctx context.Context, userID string, page, limit int) ([]models.Workout, int, error) {
    offset := (page - 1) * limit

    query := `
        SELECT id, user_id, name, type, duration_weeks, created_at, completed
        FROM workouts 
        WHERE user_id = $1 AND deleted_at IS NULL
        ORDER BY created_at DESC 
        LIMIT $2 OFFSET $3
    `

    rows, err := r.db.Query(ctx, query, userID, limit, offset)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to query workouts: %w", err)
    }
    defer rows.Close()

    var workouts []models.Workout
    for rows.Next() {
        var workout models.Workout
        err := rows.Scan(
            &workout.ID, &workout.UserID, &workout.Name, &workout.Type,
            &workout.DurationWeeks, &workout.CreatedAt, &workout.Completed,
        )
        if err != nil {
            return nil, 0, fmt.Errorf("failed to scan workout: %w", err)
        }
        workouts = append(workouts, workout)
    }

    // Get total count
    var total int
    countQuery := `SELECT COUNT(*) FROM workouts WHERE user_id = $1 AND deleted_at IS NULL`
    err = r.db.QueryRow(ctx, countQuery, userID).Scan(&total)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to get total count: %w", err)
    }

    return workouts, total, nil
}

// GetCurrentWorkout retrieves the user's current active workout plan
func (r *workoutRepository) GetCurrentWorkout(ctx context.Context, userID string) (*models.Workout, error) {
    query := `
        SELECT id, user_id, name, type, duration_weeks, created_at, completed
        FROM workouts 
        WHERE user_id = $1 AND completed = false AND deleted_at IS NULL
        ORDER BY created_at DESC 
        LIMIT 1
    `

    var workout models.Workout
    err := r.db.QueryRow(ctx, query, userID).Scan(
        &workout.ID, &workout.UserID, &workout.Name, &workout.Type,
        &workout.DurationWeeks, &workout.CreatedAt, &workout.Completed,
    )

    if err != nil {
        return nil, fmt.Errorf("failed to get current workout: %w", err)
    }

    return &workout, nil
}

// UpdateWorkout updates an existing workout plan
func (r *workoutRepository) UpdateWorkout(ctx context.Context, workout *models.Workout) error {
    query := `
        UPDATE workouts 
        SET name = $2, type = $3, duration_weeks = $4, completed = $5, updated_at = $6
        WHERE id = $1 AND deleted_at IS NULL
    `

    result, err := r.db.Exec(ctx, query,
        workout.ID, workout.Name, workout.Type, workout.DurationWeeks,
        workout.Completed, time.Now(),
    )

    if err != nil {
        return fmt.Errorf("failed to update workout: %w", err)
    }

    if result.RowsAffected() == 0 {
        return fmt.Errorf("workout not found or already deleted")
    }

    return nil
}

// DeleteWorkout soft deletes a workout plan
func (r *workoutRepository) DeleteWorkout(ctx context.Context, id string) error {
    query := `
        UPDATE workouts 
        SET deleted_at = $2 
        WHERE id = $1 AND deleted_at IS NULL
    `

    result, err := r.db.Exec(ctx, query, id, time.Now())
    if err != nil {
        return fmt.Errorf("failed to delete workout: %w", err)
    }

    if result.RowsAffected() == 0 {
        return fmt.Errorf("workout not found or already deleted")
    }

    return nil
}

// CreateSession inserts a new workout session into the database
func (r *workoutRepository) CreateSession(ctx context.Context, session *models.Session) error {
    query := `
        INSERT INTO workout_sessions (id, workout_id, day, completed, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, created_at
    `

    err := r.db.QueryRow(ctx, query,
        session.ID, session.WorkoutID, session.Day, session.Completed,
        time.Now(), time.Now(),
    ).Scan(&session.ID, &session.CreatedAt)

    if err != nil {
        return fmt.Errorf("failed to create session: %w", err)
    }

    return nil
}

// GetSessionByID retrieves a session by its ID
func (r *workoutRepository) GetSessionByID(ctx context.Context, id string) (*models.Session, error) {
    query := `
        SELECT id, workout_id, day, completed, created_at, updated_at
        FROM workout_sessions 
        WHERE id = $1 AND deleted_at IS NULL
    `

    var session models.Session
    err := r.db.QueryRow(ctx, query, id).Scan(
        &session.ID, &session.WorkoutID, &session.Day,
        &session.Completed, &session.CreatedAt, &session.UpdatedAt,
    )

    if err != nil {
        return nil, fmt.Errorf("failed to get session by ID: %w", err)
    }

    return &session, nil
}

// UpdateSession updates an existing workout session
func (r *workoutRepository) UpdateSession(ctx context.Context, session *models.Session) error {
    query := `
        UPDATE workout_sessions 
        SET day = $2, completed = $3, updated_at = $4
        WHERE id = $1 AND deleted_at IS NULL
    `

    result, err := r.db.Exec(ctx, query,
        session.ID, session.Day, session.Completed, time.Now(),
    )

    if err != nil {
        return fmt.Errorf("failed to update session: %w", err)
    }

    if result.RowsAffected() == 0 {
        return fmt.Errorf("session not found or already deleted")
    }

    return nil
}

// LogExercise logs an exercise completion with sets, reps, and weight
func (r *workoutRepository) LogExercise(ctx context.Context, exercise *models.Exercise) error {
    query := `
        INSERT INTO exercises (id, session_id, name, sets, reps, weight, completed, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, created_at
    `

    err := r.db.QueryRow(ctx, query,
        exercise.ID, exercise.SessionID, exercise.Name, exercise.Sets,
        exercise.Reps, exercise.Weight, exercise.Completed, time.Now(),
    ).Scan(&exercise.ID, &exercise.CreatedAt)

    if err != nil {
        return fmt.Errorf("failed to log exercise: %w", err)
    }

    return nil
}

// GetExerciseByID retrieves an exercise by its ID
func (r *workoutRepository) GetExerciseByID(ctx context.Context, id string) (*models.Exercise, error) {
    query := `
        SELECT id, session_id, name, sets, reps, weight, completed, created_at
        FROM exercises 
        WHERE id = $1 AND deleted_at IS NULL
    `

    var exercise models.Exercise
    err := r.db.QueryRow(ctx, query, id).Scan(
        &exercise.ID, &exercise.SessionID, &exercise.Name, &exercise.Sets,
        &exercise.Reps, &exercise.Weight, &exercise.Completed, &exercise.CreatedAt,
    )

    if err != nil {
        return nil, fmt.Errorf("failed to get exercise by ID: %w", err)
    }

    return &exercise, nil
}

// UpdateExercise updates an existing exercise record
func (r *workoutRepository) UpdateExercise(ctx context.Context, exercise *models.Exercise) error {
    query := `
        UPDATE exercises 
        SET sets = $2, reps = $3, weight = $4, completed = $5, updated_at = $6
        WHERE id = $1 AND deleted_at IS NULL
    `

    result, err := r.db.Exec(ctx, query,
        exercise.ID, exercise.Sets, exercise.Reps, exercise.Weight,
        exercise.Completed, time.Now(),
    )

    if err != nil {
        return fmt.Errorf("failed to update exercise: %w", err)
    }

    if result.RowsAffected() == 0 {
        return fmt.Errorf("exercise not found or already deleted")
    }

    return nil
}

// GetWorkoutSessions retrieves all sessions for a specific workout
func (r *workoutRepository) GetWorkoutSessions(ctx context.Context, workoutID string) ([]models.Session, error) {
    query := `
        SELECT id, workout_id, day, completed, created_at, updated_at
        FROM workout_sessions 
        WHERE workout_id = $1 AND deleted_at IS NULL
        ORDER BY day ASC
    `

    rows, err := r.db.Query(ctx, query, workoutID)
    if err != nil {
        return nil, fmt.Errorf("failed to query sessions: %w", err)
    }
    defer rows.Close()

    var sessions []models.Session
    for rows.Next() {
        var session models.Session
        err := rows.Scan(
            &session.ID, &session.WorkoutID, &session.Day,
            &session.Completed, &session.CreatedAt, &session.UpdatedAt,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan session: %w", err)
        }
        sessions = append(sessions, session)
    }

    return sessions, nil
}

// GetSessionExercises retrieves all exercises for a specific session
func (r *workoutRepository) GetSessionExercises(ctx context.Context, sessionID string) ([]models.Exercise, error) {
    query := `
        SELECT id, session_id, name, sets, reps, weight, completed, created_at
        FROM exercises 
        WHERE session_id = $1 AND deleted_at IS NULL
        ORDER BY created_at ASC
    `

    rows, err := r.db.Query(ctx, query, sessionID)
    if err != nil {
        return nil, fmt.Errorf("failed to query exercises: %w", err)
    }
    defer rows.Close()

    var exercises []models.Exercise
    for rows.Next() {
        var exercise models.Exercise
        err := rows.Scan(
            &exercise.ID, &exercise.SessionID, &exercise.Name, &exercise.Sets,
            &exercise.Reps, &exercise.Weight, &exercise.Completed, &exercise.CreatedAt,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan exercise: %w", err)
        }
        exercises = append(exercises, exercise)
    }

    return exercises, nil
}
