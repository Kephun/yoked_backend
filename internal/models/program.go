package models

import (
    "time"
)

type Program struct {
    ID             int       `json:"id"`
    Name           string    `json:"name"`
    Description    string    `json:"description"`
    Goal           string    `json:"goal"`
    EstimatedWeeks int       `json:"estimated_weeks"`
    CreatedAt      time.Time `json:"created_at"`
}

type ProgramWorkout struct {
    ID          int       `json:"id"`
    ProgramID   int       `json:"program_id"`
    Name        string    `json:"name"`
    DayOfWeek   int       `json:"day_of_week"`
    Description string    `json:"description"`
    CreatedAt   time.Time `json:"created_at"`
}

type ProgramWorkoutExercise struct {
    ID                 int     `json:"id"`
    ProgramWorkoutID   int     `json:"program_workout_id"`
    ExerciseID         int     `json:"exercise_id"`
    Sets               int     `json:"sets"`
    Reps               int     `json:"reps"`
    TargetRIR          int     `json:"target_rir"`
    PrescribedWeight   int     `json:"prescribed_weight,omitempty"`
    ExerciseOrder      int     `json:"exercise_order"`
    Notes              string  `json:"notes"`
}

type WorkoutSession struct {
    ID                int       `json:"id"`
    UserProgramID     int       `json:"user_program_id"`
    ProgramWorkoutID  int       `json:"program_workout_id"`
    CompletedDate     time.Time `json:"completed_date"`
    Notes             string    `json:"notes"`
    CreatedAt         time.Time `json:"created_at"`
}

type WorkoutExerciseLog struct {
    ID                      int   `json:"id"`
    WorkoutID               int   `json:"workout_id"`
    ProgramWorkoutExerciseID int   `json:"program_workout_exercise_id"`
    ActualReps              []int `json:"actual_reps"`
    ActualRIR               []int `json:"actual_rir"`
    CreatedAt               time.Time `json:"created_at"`
}
