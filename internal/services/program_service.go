package services

import (
	"context"
	"fmt"
	"math"
	"time"
	"log"

	"yoked_backend/internal/models"
	"yoked_backend/internal/db/repositories"
)

type ProgramService interface {
	GetProgramByID(ctx context.Context, programID int) (*models.Program, error)
	GetProgramsByGoal(ctx context.Context, goal string) ([]*models.Program, error)
	GetAllExercises(ctx context.Context) ([]*models.Exercise, error)
	GetExerciseByID(ctx context.Context, id int) (*models.Exercise, error)
	AssignProgramToUser(ctx context.Context, userID string, programID int) error
	GetUserProgramWithWorkouts(ctx context.Context, userID string) (*UserProgramDetail, error)
	CalculateInitialWeights(ctx context.Context, user *models.User, programID int) (map[int]int, error)
	StartWorkoutSession(ctx context.Context, userID string, programWorkoutID int) (*models.WorkoutSession, error)
	CompleteWorkoutSession(ctx context.Context, sessionID int, exercises []ExerciseLogRequest) error
	GetWorkoutHistory(ctx context.Context, userID string, limit int) ([]*models.WorkoutSession, error)
	CalculateNextWorkoutWeights(ctx context.Context, userID string, programWorkoutID int) (map[int]int, error)
}

type programService struct {
	programRepo repositories.ProgramRepository
	userRepo    repositories.UserRepository
}

func NewProgramService(programRepo repositories.ProgramRepository, userRepo repositories.UserRepository) ProgramService {
	return &programService{programRepo: programRepo, userRepo: userRepo}
}

// Request/Response structures
type ExerciseLogRequest struct {
	ProgramWorkoutExerciseID int   `json:"program_workout_exercise_id"`
	ActualReps               []int `json:"actual_reps"`
	ActualRIR                []int `json:"actual_rir"`
}

type UserProgramDetail struct {
	UserProgram *models.UserProgram   `json:"user_program"`
	Program     *models.Program       `json:"program"`
	Workouts    []*WorkoutDetail      `json:"workouts"`
}

type WorkoutDetail struct {
	ProgramWorkout *models.ProgramWorkout  `json:"program_workout"`
	Exercises      []*ExerciseWithWeight   `json:"exercises"`
}

type ExerciseWithWeight struct {
	ProgramExercise *models.ProgramWorkoutExercise `json:"program_exercise"`
	SuggestedWeight float64                        `json:"suggested_weight"`
}

// Service implementations
func (s *programService) GetProgramByID(ctx context.Context, programID int) (*models.Program, error) {
	return s.programRepo.GetProgramByID(ctx, programID)
}

func (s *programService) GetProgramsByGoal(ctx context.Context, goal string) ([]*models.Program, error) {
	return s.programRepo.GetProgramsByGoal(ctx, goal)
}

func (s *programService) GetAllExercises(ctx context.Context) ([]*models.Exercise, error) {
	return s.programRepo.GetAllExercises(ctx)
}

func (s *programService) GetExerciseByID(ctx context.Context, id int) (*models.Exercise, error) {
	return s.programRepo.GetExerciseByID(ctx, id)
}

func (s *programService) AssignProgramToUser(ctx context.Context, userID string, programID int) error {
	// Deactivate any current active program
	currentProgram, err := s.programRepo.GetUserActiveProgram(ctx, userID)
	if err != nil && err.Error() != "sql: no rows in result set" {
		return err
	}

	if currentProgram != nil {
		currentProgram.IsActive = false
		if err := s.programRepo.UpdateUserProgram(ctx, currentProgram); err != nil {
			return err
		}
	}

	// Create new user program
	userProgram := &models.UserProgram{
		UserID:    userID,
		ProgramID: programID,
		StartDate: time.Now(),
		IsActive:  true,
	}

	return s.programRepo.CreateUserProgram(ctx, userProgram)
}

func (s *programService) GetUserProgramWithWorkouts(ctx context.Context, userID string) (*UserProgramDetail, error) {
	// Get user's active program
	userProgram, err := s.programRepo.GetUserActiveProgram(ctx, userID)
	if err != nil {
		return nil, err
	}
	if userProgram == nil {
		return nil, fmt.Errorf("no active program found for user")
	}

	// Get program details
	program, err := s.programRepo.GetProgramByID(ctx, userProgram.ProgramID)
	if err != nil {
		return nil, err
	}

	// Get program workouts
	workouts, err := s.programRepo.GetProgramWorkouts(ctx, userProgram.ProgramID)
	if err != nil {
		return nil, err
	}

	// Get user for weight calculation
	//user, err := s.userRepo.GetUserByID(ctx, userID)
	//if err != nil {
	//	return nil, err
	//}

	// Calculate initial weights
	//weights, err := s.CalculateInitialWeights(ctx, user, userProgram.ProgramID)
	//if err != nil {
	//	return nil, err
	//}

	// Build workout details with weight
	log.Printf("Starting the looop")
	workoutDetails := make([]*WorkoutDetail, len(workouts))
	for i, workout := range workouts {
		exercises, err := s.programRepo.GetProgramWorkoutExercises(ctx, workout.ID)
		if err != nil {
			return nil, err
		}

		exerciseDetails := make([]*ExerciseWithWeight, len(exercises))
		for j, exercise := range exercises {
			exerciseDetails[j] = &ExerciseWithWeight{
				ProgramExercise: exercise,
				//SuggestedWeight: weights[exercise.ID],
			}
		}

		workoutDetails[i] = &WorkoutDetail{
			ProgramWorkout: workout,
			Exercises:      exerciseDetails,
		}
	}

	return &UserProgramDetail{
		UserProgram: userProgram,
		Program:     program,
		Workouts:    workoutDetails,
	}, nil
}

func (s *programService) CalculateInitialWeights(ctx context.Context, user *models.User, programID int) (map[int]int, error) {
	workouts, err := s.programRepo.GetProgramWorkouts(ctx, programID)
	if err != nil {
		return nil, err
	}

	weightMap := make(map[int]int)

	for _, workout := range workouts {
		exercises, err := s.programRepo.GetProgramWorkoutExercises(ctx, workout.ID)
		if err != nil {
			return nil, err
		}

		for _, exercise := range exercises {
			if exercise.PrescribedWeight > 0 {
				weightMap[exercise.ID] = exercise.PrescribedWeight
				continue
			}

			baseStrength := s.calculateBaseStrength(user)
			exerciseModifier := s.getExerciseModifier(exercise.ExerciseID)
			
			suggestedWeight := baseStrength * exerciseModifier
			weightMap[exercise.ID] = math.Round(suggestedWeight/2.5)*2.5 // Round to nearest 2.5kg
		}
	}

	return weightMap, nil
}

func (s *programService) calculateBaseStrength(user *models.User) int {
	base := user.Weight * 0.6

	if user.Sex == "male" {
		base *= 1.2
	}

	if user.Age < 25 {
		base *= 0.9 + (user.Age-13)/120
	} else if user.Age > 35 {
		base *= 1.1 - (user.Age-35)/100
	}

	return base
}

func (s *programService) getExerciseModifier(exerciseID int) int {
	modifiers := map[int]int{
		1: 1.0,  // Bench press
		2: 0.8,  // Shoulder press
		3: 1.2,  // Squat
		4: 1.1,  // Deadlift
		5: 0.5,  // Bicep curl
	}
	return modifiers[exerciseID]
}

func (s *programService) StartWorkoutSession(ctx context.Context, userID string, programWorkoutID int) (*models.WorkoutSession, error) {
	userProgram, err := s.programRepo.GetUserActiveProgram(ctx, userID)
	if err != nil {
		return nil, err
	}
	if userProgram == nil {
		return nil, fmt.Errorf("no active program found for user")
	}

	session := &models.WorkoutSession{
		UserProgramID:    userProgram.ID,
		ProgramWorkoutID: programWorkoutID,
		CompletedDate:    time.Now(),
	}

	sessionID, err := s.programRepo.CreateWorkoutSession(ctx, session)
	if err != nil {
		return nil, err
	}

	session.ID = sessionID
	return session, nil
}

func (s *programService) CompleteWorkoutSession(ctx context.Context, sessionID int, exercises []ExerciseLogRequest) error {
	for _, exercise := range exercises {
		log := &models.WorkoutExerciseLog{
			WorkoutID:               sessionID,
			ProgramWorkoutExerciseID: exercise.ProgramWorkoutExerciseID,
			ActualReps:              exercise.ActualReps,
			ActualRIR:               exercise.ActualRIR,
		}

		if err := s.programRepo.CreateWorkoutExerciseLog(ctx, log); err != nil {
			return err
		}
	}
	return nil
}

func (s *programService) GetWorkoutHistory(ctx context.Context, userID string, limit int) ([]*models.WorkoutSession, error) {
	userProgram, err := s.programRepo.GetUserActiveProgram(ctx, userID)
	if err != nil {
		return nil, err
	}
	if userProgram == nil {
		return nil, fmt.Errorf("no active program found for user")
	}

	return s.programRepo.GetWorkoutSessionsByUserProgram(ctx, userProgram.ID, limit)
}

func (s *programService) CalculateNextWorkoutWeights(ctx context.Context, userID string, programWorkoutID int) (map[int]float64, error) {
	// Get the last workout session of this type
	lastSession, err := s.programRepo.GetLastWorkoutSessionByType(ctx, userID, programWorkoutID)
	if err != nil {
		return nil, err
	}
	if lastSession == nil {
		return nil, fmt.Errorf("no previous workout session found")
	}

	// Get exercise logs for the last session
	exerciseLogs, err := s.programRepo.GetExerciseLogsByWorkout(ctx, lastSession.ID)
	if err != nil {
		return nil, err
	}

	weightMap := make(map[int]float64)

	for _, exerciseLog := range exerciseLogs {
		programExercise, err := s.programRepo.GetProgramWorkoutExercise(ctx, exerciseLog.ProgramWorkoutExerciseID)
		if err != nil {
			return nil, err
		}

		avgRIR := calculateAverageRIR(exerciseLog.ActualRIR)
		weightAdjustment := s.calculateWeightAdjustment(avgRIR, float64(programExercise.TargetRIR))

		// Get current weight (this would need to be stored somewhere)
		currentWeight := 50.0 // Default, should come from previous session or calculation
		newWeight := currentWeight * weightAdjustment

		weightMap[exerciseLog.ProgramWorkoutExerciseID] = math.Round(newWeight/2.5)*2.5
	}

	return weightMap, nil
}

func calculateAverageRIR(rirArray []int) float64 {
	sum := 0
	for _, rir := range rirArray {
		sum += rir
	}
	return float64(sum) / float64(len(rirArray))
}

func (s *programService) calculateWeightAdjustment(actualRIR, targetRIR float64) float64 {
	difference := targetRIR - actualRIR

	switch {
	case difference >= 2:
		return 0.9 // 10% decrease
	case difference <= -2:
		return 1.1 // 10% increase
	case difference >= 1:
		return 0.95 // 5% decrease
	case difference <= -1:
		return 1.05 // 5% increase
	default:
		return 1.0 // Keep same weight
	}
}
