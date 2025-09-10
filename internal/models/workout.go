package models

import "time"

type Workout struct {
    ID            string    `json:"id"`
    UserID        string    `json:"user_id"`
    Name          string    `json:"name"`
    Type          string    `json:"type"` // hypertrophy, strength, endurance, etc.
    DurationWeeks int       `json:"duration_weeks"`
    Completed     bool      `json:"completed"`
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
    Sessions      []Session `json:"sessions,omitempty"` // Optional: populated when needed
}

type Session struct {
    ID         string     `json:"id"`
    WorkoutID  string     `json:"workout_id"`
    Day        int        `json:"day"` // Day 1, 2, 3, etc.
    Completed  bool       `json:"completed"`
    CreatedAt  time.Time  `json:"created_at"`
    UpdatedAt  time.Time  `json:"updated_at"`
    Exercises  []Exercise `json:"exercises,omitempty"` // Optional: populated when needed
}

type Exercise struct {
    ID         string    `json:"id"`
    SessionID  string    `json:"session_id"`
    Name       string    `json:"name"`
    Sets       int       `json:"sets"`
    Reps       int       `json:"reps"`
    Weight     float64   `json:"weight"`
    Completed  bool      `json:"completed"`
    CreatedAt  time.Time `json:"created_at"`
}
