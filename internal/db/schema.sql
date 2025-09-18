-- Enable UUID generation --
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table --
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    age INTEGER NOT NULL CHECK (age >= 13 AND age <= 120),
    sex VARCHAR(10) NOT NULL CHECK (sex IN ('male', 'female', 'other')),
    height DECIMAL(5,2) NOT NULL CHECK (height >= 30 AND height <= 250),
    weight DECIMAL(5,2) NOT NULL CHECK (weight >= 20 AND weight <= 500),
    activity_level VARCHAR(20) NOT NULL CHECK (activity_level IN ('sedentary', 'lightly_active', 'moderately_active', 'very_active', 'extra_active')),
    goal VARCHAR(20) NOT NULL CHECK (goal IN ('weight_loss', 'muscle_gain', 'maintenance', 'endurance')),
    program_id INTEGER NOT NULL CHECK (program_id > 0),
    weekly_budget DECIMAL(10,2) DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Program Tables --

-- This stores the high-level program definition (e.g., "Jeff Nippard Hypertrophy")
CREATE TABLE programs (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    goal VARCHAR(100), -- e.g., 'hypertrophy', 'strength', 'endurance'
    estimated_weeks INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Table: program_workouts
-- This breaks a program down into individual workouts (e.g., "Day 1: Chest")
CREATE TABLE program_workouts (
    id SERIAL PRIMARY KEY,
    program_id INTEGER NOT NULL REFERENCES programs(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL, -- e.g., "Push Day", "Chest & Back"
    day_of_week INTEGER, -- Order within the program week (e.g., 1, 2, 3...)
    description TEXT,
    -- A unique constraint to prevent duplicate day names within the same program
    CONSTRAINT unique_workout_per_program UNIQUE (program_id, day_of_week)
);

-- Table: program_workout_exercises
-- This defines each exercise within a workout, including sets, reps, and target RIR.
CREATE TABLE program_workout_exercises (
    id SERIAL PRIMARY KEY,
    program_workout_id INTEGER NOT NULL REFERENCES program_workouts(id) ON DELETE CASCADE,
    exercise_id INTEGER NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    sets INTEGER NOT NULL CHECK (sets > 0),
    reps INTEGER NOT NULL CHECK (reps > 0),
    target_rir INTEGER CHECK (target_rir >= 0), -- Reps in Reserve target
    prescribed_weight REAL CHECK (prescribed_weight >= 0), -- Optional starting weight
    exercise_order INTEGER NOT NULL CHECK (exercise_order > 0), -- Order of the exercise in the workout
    notes TEXT,
    -- A unique constraint to prevent an exercise from being added twice to the same workout
    CONSTRAINT unique_exercise_in_workout UNIQUE (program_workout_id, exercise_id),
    -- A unique constraint to maintain consistent ordering within a workout
    CONSTRAINT unique_order_in_workout UNIQUE (program_workout_id, exercise_order)
);

-- Exercises -- 
CREATE TABLE exercises (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    primary_muscle_group VARCHAR(100),
    equipment VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Logs for Dynamic Changes -- 
CREATE TABLE user_programs (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE, -- Matches users.id type (UUID)
    program_id INTEGER NOT NULL REFERENCES programs(id) ON DELETE CASCADE,
    start_date DATE NOT NULL DEFAULT CURRENT_DATE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Table: workouts
-- A record of a user completing a specific workout from their program
CREATE TABLE workouts (
    id SERIAL PRIMARY KEY,
    user_program_id INTEGER NOT NULL REFERENCES user_programs(id) ON DELETE CASCADE,
    program_workout_id INTEGER NOT NULL REFERENCES program_workouts(id) ON DELETE CASCADE,
    completed_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Table: workout_exercises
-- The core log of a user's performance for each exercise in a session
CREATE TABLE workout_exercises (
    id SERIAL PRIMARY KEY,
    workout_id INTEGER NOT NULL REFERENCES workouts(id) ON DELETE CASCADE,
    program_workout_exercise_id INTEGER NOT NULL REFERENCES program_workout_exercises(id) ON DELETE CASCADE,
    -- Using PostgreSQL arrays to store the sequence of reps and RIR for each set
    actual_reps INTEGER[] NOT NULL CHECK (
        array_length(actual_reps, 1) IS NOT NULL AND
        array_length(actual_reps, 1) > 0
    ), -- e.g., {8, 8, 10}
    actual_rir INTEGER[] NOT NULL CHECK (
        array_length(actual_rir, 1) IS NOT NULL AND
        array_length(actual_rir, 1) > 0
    ), -- e.g., {2, 2, 0}
    -- Ensure the reps and RIR arrays are the same length (same number of sets)
    CONSTRAINT same_number_of_sets CHECK (
        array_length(actual_reps, 1) = array_length(actual_rir, 1)
    ),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


-- Indexes for better performance --
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_deleted ON users(deleted_at);
CREATE INDEX IF NOT EXISTS idx_user_stats_user_id ON user_stats(user_id);
CREATE INDEX idx_program_workouts_program_id ON program_workouts(program_id);
CREATE INDEX idx_pwe_workout_id ON program_workout_exercises(program_workout_id);
CREATE INDEX idx_pwe_exercise_id ON program_workout_exercises(exercise_id);
CREATE INDEX idx_exercises_muscle_group ON exercises(primary_muscle_group);
CREATE INDEX idx_programs_goal ON programs(goal);
CREATE INDEX idx_user_programs_user_id ON user_programs(user_id);
CREATE INDEX idx_user_programs_program_id ON user_programs(program_id);
CREATE INDEX idx_user_programs_is_active ON user_programs(is_active) WHERE is_active = true;
-- Helps find active programs quickly

CREATE INDEX idx_workouts_user_program_id ON workouts(user_program_id);
CREATE INDEX idx_workouts_program_workout_id ON workouts(program_workout_id);
CREATE INDEX idx_workouts_completed_date ON workouts(completed_date);

CREATE INDEX idx_workout_exercises_workout_id ON workout_exercises(workout_id);
CREATE INDEX idx_workout_exercises_pwe_id ON workout_exercises(program_workout_exercise_id);



-- Insert sample workout types (optional)
INSERT INTO workouts (id, user_id, name, type, duration_weeks, completed) 
VALUES 
    ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', '4-Week Strength Program', 'strength', 4, false),
    ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001', '8-Week Hypertrophy Program', 'hypertrophy', 8, false)
ON CONFLICT (id) DO NOTHING;

-- Insert sample exercises (optional)
INSERT INTO exercises (id, session_id, name, sets, reps, weight, completed) 
VALUES 
    ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'Bench Press', 4, 8, 185.5, false),
    ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001', 'Squats', 4, 10, 225.0, false)
ON CONFLICT (id) DO NOTHING;

-- Create a test user for CI/CD (optional)
INSERT INTO users (id, email, password_hash, name, age, sex, height, weight, activity_level, goal, weekly_budget) 
VALUES 
    ('00000000-0000-0000-0000-000000000001', 'test@example.com', '$2a$10$examplehashedpassword', 'Test User', 25, 'male', 175.5, 70.0, 'moderately_active', 'muscle_gain', 150.0)
ON CONFLICT (id) DO NOTHING;
