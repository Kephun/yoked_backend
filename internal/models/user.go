package models

import "time"

type User struct {
    ID            string    `json:"id"`
    Email         string    `json:"email"`
    PasswordHash  string    `json:"-"` // Hidden from JSON
    Name          string    `json:"name"`
    Age           int       `json:"age"`
    Sex           string    `json:"sex"`
    Height        float64   `json:"height"`
    Weight        float64   `json:"weight"`
    ActivityLevel string    `json:"activity_level"`
    Goal          string    `json:"goal"`
    ProgramID     int       `json:"program_id"`
    WeeklyBudget  float64   `json:"weekly_budget,omitempty"`
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
}


type UserProgram struct {
    ID        int       `json:"id"`
    UserID    string    `json:"user_id"`
    ProgramID int       `json:"program_id"`
    StartDate time.Time `json:"start_date"`
    IsActive  bool      `json:"is_active"`
    CreatedAt time.Time `json:"created_at"`
}

type UserPreferences struct {
    UserID      string   `json:"user_id"`
    Preferences []string `json:"preferences"`
    Allergies   []string `json:"allergies"`
    Dislikes    []string `json:"dislikes"`
}

type UserStats struct {
    ID                 string    `json:"id"`
    UserID             string    `json:"user_id"`
    CompletedWorkouts  int       `json:"completed_workouts"`
    CompletedSessions  int       `json:"completed_sessions"`
    TotalWeightLifted  float64   `json:"total_weight_lifted"`
    CurrentStreak      int       `json:"current_streak"`
    LongestStreak      int       `json:"longest_streak"`
    LastWorkoutDate    time.Time `json:"last_workout_date"`
    CreatedAt          time.Time `json:"created_at"`
    UpdatedAt          time.Time `json:"updated_at"`
}
