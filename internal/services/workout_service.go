package services

import (
    "context"
    "fmt"
    "time"
    "yoked_backend/internal/models"
    "yoked_backend/internal/db/repositories"
)

type WorkoutService interface {
    CreateWorkout(ctx context.Context, workout *models.Workout) error
    GetWorkoutByID(ctx context.Context, workoutID string) (*models.Workout, error)
    GetUserWorkouts(ctx context.Context, userID string, page, limit int) ([]models.Workout, int, error)
    GetCurrentWorkout(ctx context.Context, userID string) (*models.Workout, error)
    UpdateWorkout(ctx context.Context, workout *models.Workout) error
    DeleteWorkout(ctx context.Context, workoutID string) error
    GenerateWorkout(ctx context.Context, userID, workoutType string, duration int) (*models.Workout, error)
    CreateSession(ctx context.Context, session *models.Session) error
    GetSessionByID(ctx context.Context, sessionID string) (*models.Session, error)
    UpdateSession(ctx context.Context, session *models.Session) error
    StartSession(ctx context.Context, sessionID string) error
    CompleteSession(ctx context.Context, sessionID string) error
    LogExercise(ctx context.Context, exercise *models.Exercise) error
    GetExerciseByID(ctx context.Context, exerciseID string) (*models.Exercise, error)
    UpdateExercise(ctx context.Context, exercise *models.Exercise) error
    GetWorkoutSessions(ctx context.Context, workoutID string) ([]models.Session, error)
    GetSessionExercises(ctx context.Context, sessionID string) ([]models.Exercise, error)
}

type workoutService struct {
    workoutRepo repositories.WorkoutRepository
    userRepo    repositories.UserRepository
}

func NewWorkoutService(workoutRepo repositories.WorkoutRepository, userRepo repositories.UserRepository) WorkoutService {
    return &workoutService{
        workoutRepo: workoutRepo,
        userRepo:    userRepo,
    }
}

func (s *workoutService) CreateWorkout(ctx context.Context, workout *models.Workout) error {
    _, err := s.userRepo.GetUserByID(ctx, workout.UserID)
    if err != nil {
        return fmt.Errorf("user not found: %w", err)
    }

    if err := s.validateWorkout(workout); err != nil {
        return err
    }

    return s.workoutRepo.CreateWorkout(ctx, workout)
}

func (s *workoutService) GetWorkoutByID(ctx context.Context, workoutID string) (*models.Workout, error) {
    workout, err := s.workoutRepo.GetWorkoutByID(ctx, workoutID)
    if err != nil {
        return nil, fmt.Errorf("failed to get workout: %w", err)
    }
    return workout, nil
}

func (s *workoutService) GetUserWorkouts(ctx context.Context, userID string, page, limit int) ([]models.Workout, int, error) {
    _, err := s.userRepo.GetUserByID(ctx, userID)
    if err != nil {
        return nil, 0, fmt.Errorf("user not found: %w", err)
    }

    if page < 1 {
        page = 1
    }
    if limit < 1 || limit > 100 {
        limit = 20
    }

    return s.workoutRepo.GetUserWorkouts(ctx, userID, page, limit)
}

func (s *workoutService) GetCurrentWorkout(ctx context.Context, userID string) (*models.Workout, error) {
    _, err := s.userRepo.GetUserByID(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("user not found: %w", err)
    }

    workout, err := s.workoutRepo.GetCurrentWorkout(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to get current workout: %w", err)
    }
    return workout, nil
}

func (s *workoutService) UpdateWorkout(ctx context.Context, workout *models.Workout) error {
    existingWorkout, err := s.workoutRepo.GetWorkoutByID(ctx, workout.ID)
    if err != nil {
        return fmt.Errorf("workout not found: %w", err)
    }

    if err := s.validateWorkout(workout); err != nil {
        return err
    }

    workout.UserID = existingWorkout.UserID
    return s.workoutRepo.UpdateWorkout(ctx, workout)
}

func (s *workoutService) DeleteWorkout(ctx context.Context, workoutID string) error {
    _, err := s.workoutRepo.GetWorkoutByID(ctx, workoutID)
    if err != nil {
        return fmt.Errorf("workout not found: %w", err)
    }

    return s.workoutRepo.DeleteWorkout(ctx, workoutID)
}

func (s *workoutService) GenerateWorkout(ctx context.Context, userID, workoutType string, duration int) (*models.Workout, error) {
    user, err := s.userRepo.GetUserByID(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("user not found: %w", err)
    }

    validTypes := map[string]bool{
        "hypertrophy": true, "strength": true, "endurance": true,
        "powerlifting": true, "cardio": true, "flexibility": true,
    }
    if !validTypes[workoutType] {
        return nil, fmt.Errorf("invalid workout type: %s", workoutType)
    }
    if duration < 1 || duration > 52 {
        return nil, fmt.Errorf("duration must be between 1 and 52 weeks")
    }

    workout := &models.Workout{
        UserID:        userID,
        Name:          s.generateWorkoutName(workoutType, duration, user),
        Type:          workoutType,
        DurationWeeks: duration,
        Completed:     false,
    }

    workout.Sessions = s.generateWorkoutSessions(workoutType, duration, user)

    err = s.workoutRepo.CreateWorkout(ctx, workout)
    if err != nil {
        return nil, fmt.Errorf("failed to create workout: %w", err)
    }

    return workout, nil
}

func (s *workoutService) CreateSession(ctx context.Context, session *models.Session) error {
    _, err := s.workoutRepo.GetWorkoutByID(ctx, session.WorkoutID)
    if err != nil {
        return fmt.Errorf("workout not found: %w", err)
    }

    if session.Day < 1 || session.Day > 7 {
        return fmt.Errorf("day must be between 1 and 7")
    }

    return s.workoutRepo.CreateSession(ctx, session)
}

func (s *workoutService) GetSessionByID(ctx context.Context, sessionID string) (*models.Session, error) {
    session, err := s.workoutRepo.GetSessionByID(ctx, sessionID)
    if err != nil {
        return nil, fmt.Errorf("failed to get session: %w", err)
    }
    return session, nil
}

func (s *workoutService) UpdateSession(ctx context.Context, session *models.Session) error {
    existingSession, err := s.workoutRepo.GetSessionByID(ctx, session.ID)
    if err != nil {
        return fmt.Errorf("session not found: %w", err)
    }

    if session.Day < 1 || session.Day > 7 {
        return fmt.Errorf("day must be between 1 and 7")
    }

    session.WorkoutID = existingSession.WorkoutID
    return s.workoutRepo.UpdateSession(ctx, session)
}

func (s *workoutService) StartSession(ctx context.Context, sessionID string) error {
    session, err := s.workoutRepo.GetSessionByID(ctx, sessionID)
    if err != nil {
        return fmt.Errorf("session not found: %w", err)
    }

    if session.Completed {
        return fmt.Errorf("session already completed")
    }

    return nil
}

func (s *workoutService) CompleteSession(ctx context.Context, sessionID string) error {
    session, err := s.workoutRepo.GetSessionByID(ctx, sessionID)
    if err != nil {
        return fmt.Errorf("session not found: %w", err)
    }

    session.Completed = true
    session.UpdatedAt = time.Now()

    return s.workoutRepo.UpdateSession(ctx, session)
}

func (s *workoutService) LogExercise(ctx context.Context, exercise *models.Exercise) error {
    _, err := s.workoutRepo.GetSessionByID(ctx, exercise.SessionID)
    if err != nil {
        return fmt.Errorf("session not found: %w", err)
    }

    if exercise.Sets < 1 || exercise.Sets > 20 {
        return fmt.Errorf("sets must be between 1 and 20")
    }
    if exercise.Reps < 1 || exercise.Reps > 100 {
        return fmt.Errorf("reps must be between 1 and 100")
    }
    if exercise.Weight < 0 || exercise.Weight > 1000 {
        return fmt.Errorf("weight must be between 0 and 1000")
    }

    return s.workoutRepo.LogExercise(ctx, exercise)
}

func (s *workoutService) GetExerciseByID(ctx context.Context, exerciseID string) (*models.Exercise, error) {
    exercise, err := s.workoutRepo.GetExerciseByID(ctx, exerciseID)
    if err != nil {
        return nil, fmt.Errorf("failed to get exercise: %w", err)
    }
    return exercise, nil
}

func (s *workoutService) UpdateExercise(ctx context.Context, exercise *models.Exercise) error {
    existingExercise, err := s.workoutRepo.GetExerciseByID(ctx, exercise.ID)
    if err != nil {
        return fmt.Errorf("exercise not found: %w", err)
    }

    if exercise.Sets < 1 || exercise.Sets > 20 {
        return fmt.Errorf("sets must be between 1 and 20")
    }
    if exercise.Reps < 1 || exercise.Reps > 100 {
        return fmt.Errorf("reps must be between 1 and 100")
    }
    if exercise.Weight < 0 || exercise.Weight > 1000 {
        return fmt.Errorf("weight must be between 0 and 1000")
    }

    exercise.SessionID = existingExercise.SessionID
    return s.workoutRepo.UpdateExercise(ctx, exercise)
}

func (s *workoutService) GetWorkoutSessions(ctx context.Context, workoutID string) ([]models.Session, error) {
    _, err := s.workoutRepo.GetWorkoutByID(ctx, workoutID)
    if err != nil {
        return nil, fmt.Errorf("workout not found: %w", err)
    }

    sessions, err := s.workoutRepo.GetWorkoutSessions(ctx, workoutID)
    if err != nil {
        return nil, fmt.Errorf("failed to get workout sessions: %w", err)
    }
    return sessions, nil
}

func (s *workoutService) GetSessionExercises(ctx context.Context, sessionID string) ([]models.Exercise, error) {
    _, err := s.workoutRepo.GetSessionByID(ctx, sessionID)
    if err != nil {
        return nil, fmt.Errorf("session not found: %w", err)
    }

    exercises, err := s.workoutRepo.GetSessionExercises(ctx, sessionID)
    if err != nil {
        return nil, fmt.Errorf("failed to get session exercises: %w", err)
    }
    return exercises, nil
}

func (s *workoutService) validateWorkout(workout *models.Workout) error {
    validTypes := map[string]bool{
        "hypertrophy": true, "strength": true, "endurance": true,
        "powerlifting": true, "cardio": true, "flexibility": true,
    }

    if workout.Name == "" {
        return fmt.Errorf("workout name is required")
    }
    if !validTypes[workout.Type] {
        return fmt.Errorf("invalid workout type: %s", workout.Type)
    }
    if workout.DurationWeeks < 1 || workout.DurationWeeks > 52 {
        return fmt.Errorf("duration must be between 1 and 52 weeks")
    }

    return nil
}

func (s *workoutService) generateWorkoutName(workoutType string, duration int, user *models.User) string {
    return fmt.Sprintf("%d-Week %s Program for %s", duration, workoutType, user.Name)
}

func (s *workoutService) generateWorkoutSessions(workoutType string, duration int, user *models.User) []models.Session {
    var sessions []models.Session
    sessionsPerWeek := 3

    switch workoutType {
    case "hypertrophy":
        sessionsPerWeek = 4
    case "strength":
        sessionsPerWeek = 3
    case "endurance":
        sessionsPerWeek = 5
    case "powerlifting":
        sessionsPerWeek = 3
    case "cardio":
        sessionsPerWeek = 5
    case "flexibility":
        sessionsPerWeek = 6
    }

    totalSessions := duration * sessionsPerWeek

    for i := 1; i <= totalSessions; i++ {
        sessions = append(sessions, models.Session{
            Day:       (i-1)%7 + 1,
            Completed: false,
        })
    }

    return sessions
}
