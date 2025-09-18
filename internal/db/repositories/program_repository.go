// repository/program_repo.go
package repositories

import (
    "fmt"
    "context"
    "log"
    "github.com/jackc/pgx/v5/pgxpool"
    "yoked_backend/internal/models"
)

type ProgramRepository interface {
    // Program management
    GetProgramByID(ctx context.Context, programID int) (*models.Program, error)
    GetProgramsByGoal(ctx context.Context, goal string) ([]*models.Program, error)
    GetAllPrograms(ctx context.Context) ([]*models.Program, error)
    
    // Program structure
    GetProgramWorkouts(ctx context.Context, programID int) ([]*models.ProgramWorkout, error)
    GetProgramWorkoutExercises(ctx context.Context, workoutID int) ([]*models.ProgramWorkoutExercise, error)
    GetProgramWorkoutExercise(ctx context.Context, id int) (*models.ProgramWorkoutExercise, error)
    
    // Exercises
    GetAllExercises(ctx context.Context) ([]*models.Exercise, error)
    GetExerciseByID(ctx context.Context, id int) (*models.Exercise, error)
    
    // User program tracking
    CreateUserProgram(ctx context.Context, userProgram *models.UserProgram) error
    GetUserActiveProgram(ctx context.Context, userID string) (*models.UserProgram, error)
    GetUserProgramByID(ctx context.Context, id int) (*models.UserProgram, error)
    UpdateUserProgram(ctx context.Context, userProgram *models.UserProgram) error
    
    // Workout sessions
    CreateWorkoutSession(ctx context.Context, session *models.WorkoutSession) (int, error)
    GetWorkoutSessionByID(ctx context.Context, id int) (*models.WorkoutSession, error)
    GetWorkoutSessionsByUserProgram(ctx context.Context, userProgramID int, limit int) ([]*models.WorkoutSession, error)
    GetLastWorkoutSessionByType(ctx context.Context, userID string, programWorkoutID int) (*models.WorkoutSession, error)
    
    // Exercise logs
    CreateWorkoutExerciseLog(ctx context.Context, log *models.WorkoutExerciseLog) error
    GetExerciseLogsByWorkout(ctx context.Context, workoutID int) ([]*models.WorkoutExerciseLog, error)
    GetLastExerciseLog(ctx context.Context, userID string, programWorkoutExerciseID int) (*models.WorkoutExerciseLog, error)
    
}

type programRepository struct {
    pool *pgxpool.Pool
}

func NewProgramRepository(pool *pgxpool.Pool) ProgramRepository {
    return &programRepository{pool: pool}
}

// Implement all the interface methods below...
func (r *programRepository) GetProgramByID(ctx context.Context, programID int) (*models.Program, error) {
    query := `SELECT id, name, description, goal, estimated_weeks, created_at 
              FROM programs WHERE id = $1`
    
    var program models.Program
    err := r.pool.QueryRow(ctx, query, programID).Scan(
        &program.ID, &program.Name, &program.Description, 
        &program.Goal, &program.EstimatedWeeks, &program.CreatedAt,
    )
    if err != nil {
        return nil, err
    }
    return &program, nil
}

func (r *programRepository) GetProgramsByGoal(ctx context.Context, goal string) ([]*models.Program, error) {
    query := `SELECT id, name, description, goal, estimated_weeks, created_at 
              FROM programs WHERE goal = $1 ORDER BY name`
    
    rows, err := r.pool.Query(ctx, query, goal)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var programs []*models.Program
    for rows.Next() {
        var program models.Program
        if err := rows.Scan(
            &program.ID, &program.Name, &program.Description,
            &program.Goal, &program.EstimatedWeeks, &program.CreatedAt,
        ); err != nil {
            return nil, err
        }
        programs = append(programs, &program)
    }
    return programs, nil
}

func (r *programRepository) GetAllPrograms(ctx context.Context) ([]*models.Program, error) {
    query := `SELECT id, name, description, goal, estimated_weeks, created_at 
              FROM programs ORDER BY name`
    
    rows, err := r.pool.Query(ctx, query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var programs []*models.Program
    for rows.Next() {
        var program models.Program
        if err := rows.Scan(
            &program.ID, &program.Name, &program.Description,
            &program.Goal, &program.EstimatedWeeks, &program.CreatedAt,
        ); err != nil {
            return nil, err
        }
        programs = append(programs, &program)
    }
    return programs, nil
}

func (r *programRepository) GetProgramWorkouts(ctx context.Context, programID int) ([]*models.ProgramWorkout, error) {
    query := `SELECT id, program_id, name, day_of_week, description 
              FROM program_workouts WHERE program_id = $1 ORDER BY day_of_week`
    
    rows, err := r.pool.Query(ctx, query, programID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var workouts []*models.ProgramWorkout
    for rows.Next() {
        var workout models.ProgramWorkout
        if err := rows.Scan(
            &workout.ID, &workout.ProgramID, &workout.Name,
            &workout.DayOfWeek, &workout.Description,
        ); err != nil {
            return nil, err
        }
        workouts = append(workouts, &workout)
    }
    return workouts, nil
}

func (r *programRepository) GetProgramWorkoutExercises(ctx context.Context, workoutID int) ([]*models.ProgramWorkoutExercise, error) {
    query := `SELECT id, program_workout_id, exercise_id, sets, reps, target_rir, 
                     prescribed_weight, exercise_order, notes
              FROM program_workout_exercises 
              WHERE program_workout_id = $1 ORDER BY exercise_order`
    
    rows, err := r.pool.Query(ctx, query, workoutID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var exercises []*models.ProgramWorkoutExercise
    for rows.Next() {
        var exercise models.ProgramWorkoutExercise
        if err := rows.Scan(
            &exercise.ID, &exercise.ProgramWorkoutID, &exercise.ExerciseID,
            &exercise.Sets, &exercise.Reps, &exercise.TargetRIR,
            &exercise.PrescribedWeight, &exercise.ExerciseOrder, &exercise.Notes,
        ); err != nil {
            return nil, err
        }
        exercises = append(exercises, &exercise)
    }
    return exercises, nil
}

func (r *programRepository) GetProgramWorkoutExercise(ctx context.Context, id int) (*models.ProgramWorkoutExercise, error) {
    query := `SELECT id, program_workout_id, exercise_id, sets, reps, target_rir, 
                     prescribed_weight, exercise_order, notes
              FROM program_workout_exercises WHERE id = $1`
    
    var exercise models.ProgramWorkoutExercise
    err := r.pool.QueryRow(ctx, query, id).Scan(
        &exercise.ID, &exercise.ProgramWorkoutID, &exercise.ExerciseID,
        &exercise.Sets, &exercise.Reps, &exercise.TargetRIR,
        &exercise.PrescribedWeight, &exercise.ExerciseOrder, &exercise.Notes,
    )
    if err != nil {
        return nil, err
    }
    return &exercise, nil
}

func (r *programRepository) GetAllExercises(ctx context.Context) ([]*models.Exercise, error) {
    query := `SELECT id, name, description, primary_muscle_group, equipment, created_at 
              FROM exercises ORDER BY name`
    
    rows, err := r.pool.Query(ctx, query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var exercises []*models.Exercise
    for rows.Next() {
        var exercise models.Exercise
        if err := rows.Scan(
            &exercise.ID, &exercise.Name, &exercise.Description,
            &exercise.PrimaryMuscleGroup, &exercise.Equipment, &exercise.CreatedAt,
        ); err != nil {
            return nil, err
        }
        exercises = append(exercises, &exercise)
    }
    return exercises, nil
}

func (r *programRepository) GetExerciseByID(ctx context.Context, id int) (*models.Exercise, error) {
    query := `SELECT id, name, description, primary_muscle_group, equipment, created_at 
              FROM exercises WHERE id = $1`
    
    var exercise models.Exercise
    err := r.pool.QueryRow(ctx, query, id).Scan(
        &exercise.ID, &exercise.Name, &exercise.Description,
        &exercise.PrimaryMuscleGroup, &exercise.Equipment, &exercise.CreatedAt,
    )
    if err != nil {
        return nil, err
    }
    return &exercise, nil
}

func (r *programRepository) CreateUserProgram(ctx context.Context, userProgram *models.UserProgram) error {
    query := `INSERT INTO user_programs (user_id, program_id, start_date, is_active) 
              VALUES ($1, $2, $3, $4) RETURNING id, created_at`
    
    return r.pool.QueryRow(ctx, query, 
        userProgram.UserID, userProgram.ProgramID, userProgram.StartDate, userProgram.IsActive,
    ).Scan(&userProgram.ID, &userProgram.CreatedAt)
}

func (r *programRepository) GetUserActiveProgram(ctx context.Context, userID string) (*models.UserProgram, error) {
    query := `SELECT id, user_id, program_id, start_date, is_active, created_at 
              FROM user_programs WHERE user_id = $1 AND is_active = true`
    
    var userProgram models.UserProgram
    err := r.pool.QueryRow(ctx, query, userID).Scan(
        &userProgram.ID, &userProgram.UserID, &userProgram.ProgramID,
        &userProgram.StartDate, &userProgram.IsActive, &userProgram.CreatedAt,
    )
    if err != nil {
	log.Printf("Error in GetUserActiveProgram: %v", err)
        log.Printf("Query: %s", query)
        log.Printf("UserID: %s", userID)
        return nil, err
    }
    return &userProgram, nil
}

func (r *programRepository) GetUserProgramByID(ctx context.Context, id int) (*models.UserProgram, error) {
    query := `SELECT id, user_id, program_id, start_date, is_active, created_at 
              FROM user_programs WHERE id = $1`
    
    var userProgram models.UserProgram
    err := r.pool.QueryRow(ctx, query, id).Scan(
        &userProgram.ID, &userProgram.UserID, &userProgram.ProgramID,
        &userProgram.StartDate, &userProgram.IsActive, &userProgram.CreatedAt,
    )
    if err != nil {
        return nil, err
    }
    return &userProgram, nil
}

func (r *programRepository) UpdateUserProgram(ctx context.Context, userProgram *models.UserProgram) error {
    query := `UPDATE user_programs SET is_active = $1 WHERE id = $2`
    
    result, err := r.pool.Exec(ctx, query, userProgram.IsActive, userProgram.ID)
    if err != nil {
        return err
    }
    
    rows := result.RowsAffected()
    if rows == 0 {
        return fmt.Errorf("user program not found")
    }
    return nil
}

func (r *programRepository) CreateWorkoutSession(ctx context.Context, session *models.WorkoutSession) (int, error) {
    query := `INSERT INTO workouts (user_program_id, program_workout_id, completed_date, notes) 
              VALUES ($1, $2, $3, $4) RETURNING id, created_at`
    
    err := r.pool.QueryRow(ctx, query,
        session.UserProgramID, session.ProgramWorkoutID, session.CompletedDate, session.Notes,
    ).Scan(&session.ID, &session.CreatedAt)
    
    return session.ID, err
}

func (r *programRepository) GetWorkoutSessionByID(ctx context.Context, id int) (*models.WorkoutSession, error) {
    query := `SELECT id, user_program_id, program_workout_id, completed_date, notes, created_at 
              FROM workouts WHERE id = $1`
    
    var session models.WorkoutSession
    err := r.pool.QueryRow(ctx, query, id).Scan(
        &session.ID, &session.UserProgramID, &session.ProgramWorkoutID,
        &session.CompletedDate, &session.Notes, &session.CreatedAt,
    )
    if err != nil {
        return nil, err
    }
    return &session, nil
}

func (r *programRepository) GetWorkoutSessionsByUserProgram(ctx context.Context, userProgramID int, limit int) ([]*models.WorkoutSession, error) {
    query := `SELECT id, user_program_id, program_workout_id, completed_date, notes, created_at 
              FROM workouts WHERE user_program_id = $1 
              ORDER BY completed_date DESC LIMIT $2`
    
    rows, err := r.pool.Query(ctx, query, userProgramID, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var sessions []*models.WorkoutSession
    for rows.Next() {
        var session models.WorkoutSession
        if err := rows.Scan(
            &session.ID, &session.UserProgramID, &session.ProgramWorkoutID,
            &session.CompletedDate, &session.Notes, &session.CreatedAt,
        ); err != nil {
            return nil, err
        }
        sessions = append(sessions, &session)
    }
    return sessions, nil
}

func (r *programRepository) GetLastWorkoutSessionByType(ctx context.Context, userID string, programWorkoutID int) (*models.WorkoutSession, error) {
    query := `SELECT w.id, w.user_program_id, w.program_workout_id, w.completed_date, w.notes, w.created_at
              FROM workouts w
              JOIN user_programs up ON w.user_program_id = up.id
              WHERE up.user_id = $1 AND w.program_workout_id = $2
              ORDER BY w.completed_date DESC LIMIT 1`
    
    var session models.WorkoutSession
    err := r.pool.QueryRow(ctx, query, userID, programWorkoutID).Scan(
        &session.ID, &session.UserProgramID, &session.ProgramWorkoutID,
        &session.CompletedDate, &session.Notes, &session.CreatedAt,
    )
    if err != nil {
        return nil, err
    }
    return &session, nil
}

func (r *programRepository) CreateWorkoutExerciseLog(ctx context.Context, log *models.WorkoutExerciseLog) error {
    query := `INSERT INTO workout_exercises (workout_id, program_workout_exercise_id, actual_reps, actual_rir) 
              VALUES ($1, $2, $3, $4) RETURNING id, created_at`
    
    return r.pool.QueryRow(ctx, query,
        log.WorkoutID, log.ProgramWorkoutExerciseID,
        log.ActualReps, log.ActualRIR,
    ).Scan(&log.ID, &log.CreatedAt)
}

func (r *programRepository) GetExerciseLogsByWorkout(ctx context.Context, workoutID int) ([]*models.WorkoutExerciseLog, error) {
    query := `SELECT id, workout_id, program_workout_exercise_id, actual_reps, actual_rir, created_at
              FROM workout_exercises WHERE workout_id = $1 ORDER BY id`
    
    rows, err := r.pool.Query(ctx, query, workoutID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var logs []*models.WorkoutExerciseLog
    for rows.Next() {
        var log models.WorkoutExerciseLog
        if err := rows.Scan(
            &log.ID, &log.WorkoutID, &log.ProgramWorkoutExerciseID,
            &log.ActualReps, &log.ActualRIR, &log.CreatedAt,
        ); err != nil {
            return nil, err
        }
        logs = append(logs, &log)
    }
    return logs, nil
}

func (r *programRepository) GetLastExerciseLog(ctx context.Context, userID string, programWorkoutExerciseID int) (*models.WorkoutExerciseLog, error) {
    query := `SELECT we.id, we.workout_id, we.program_workout_exercise_id, we.actual_reps, we.actual_rir, we.created_at
              FROM workout_exercises we
              JOIN workouts w ON we.workout_id = w.id
              JOIN user_programs up ON w.user_program_id = up.id
              WHERE up.user_id = $1 AND we.program_workout_exercise_id = $2
              ORDER BY w.completed_date DESC LIMIT 1`
    
    var log models.WorkoutExerciseLog
    err := r.pool.QueryRow(ctx, query, userID, programWorkoutExerciseID).Scan(
        &log.ID, &log.WorkoutID, &log.ProgramWorkoutExerciseID,
        &log.ActualReps, &log.ActualRIR, &log.CreatedAt,
    )
    if err != nil {
        return nil, err
    }
    return &log, nil
}
