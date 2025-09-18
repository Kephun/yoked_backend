
package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"yoked_backend/internal/models"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	CreateUserProgram(ctx context.Context, userProgram *models.UserProgram) error
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id string) error
	UpdateUserPreferences(ctx context.Context, userID string, prefs *models.UserPreferences) error
	GetUserPreferences(ctx context.Context, userID string) (*models.UserPreferences, error)
	UpdateUserPassword(ctx context.Context, userID, passwordHash string) error
	UpdateUserStats(ctx context.Context, userID string)(*models.UserStats, error)
	GetUserStats(ctx context.Context, userID string) (*models.UserStats, error)
}

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{db: db}
}

// CreateUser inserts a new user into the database
func (r *userRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (email, password_hash, name, age, sex, height, weight, 
		                  activity_level, goal, program_id, weekly_budget, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id
	`

	err := r.db.QueryRow(ctx, query,
		user.Email, user.PasswordHash, user.Name, user.Age, user.Sex,
		user.Height, user.Weight, user.ActivityLevel, user.Goal, user.ProgramID, user.WeeklyBudget,
		time.Now(), time.Now(),
	).Scan(&user.ID)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}



func (r *userRepository) CreateUserProgram(ctx context.Context, userProgram *models.UserProgram) error {
	query := `
		INSERT INTO user_programs(user_id, program_id, start_date, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	err := r.db.QueryRow(ctx, query,
		userProgram.UserID,
		userProgram.ProgramID,
		userProgram.StartDate,
		userProgram.IsActive,
		time.Now(),
	).Scan(&userProgram.ID)

	if err != nil {
		return fmt.Errorf("Failed to create user program: %w", err)
	}

	return nil
}


// GetUserByID retrieves a user by their ID
func (r *userRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, name, age, sex, height, weight, 
		       activity_level, goal, program_id, weekly_budget, created_at, updated_at
		FROM users 
		WHERE id = $1 AND deleted_at IS NULL
	`

	var user models.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Age, &user.Sex,
		&user.Height, &user.Weight, &user.ActivityLevel, &user.Goal, &user.ProgramID, &user.WeeklyBudget,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by their email
func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, name, age, sex, height, weight, 
		       activity_level, goal, program_id, weekly_budget, created_at, updated_at
		FROM users 
		WHERE email = $1 AND deleted_at IS NULL
	`

	var user models.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Age, &user.Sex,
		&user.Height, &user.Weight, &user.ActivityLevel, &user.Goal, &user.ProgramID, &user.WeeklyBudget,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// UpdateUser updates an existing user's information
func (r *userRepository) UpdateUser(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users 
		SET email = $2, name = $3, age = $4, sex = $5, height = $6, weight = $7,
		    activity_level = $8, goal = $9, program_id = $10, weekly_budget = $11, updated_at = $12
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query,
		user.ID, user.Email, user.Name, user.Age, user.Sex, user.Height, user.Weight,
		user.ActivityLevel, user.Goal, user.ProgramID, user.WeeklyBudget, time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found or already deleted")
	}

	return nil
}

// DeleteUser soft deletes a user by setting deleted_at timestamp
func (r *userRepository) DeleteUser(ctx context.Context, id string) error {
	query := `
		UPDATE users 
		SET deleted_at = $2 
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found or already deleted")
	}

	return nil
}

// UpdateUserPreferences updates or creates user preferences
func (r *userRepository) UpdateUserPreferences(ctx context.Context, userID string, prefs *models.UserPreferences) error {
	// First check if preferences exist
	checkQuery := `SELECT COUNT(*) FROM user_preferences WHERE user_id = $1`
	var count int
	err := r.db.QueryRow(ctx, checkQuery, userID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing preferences: %w", err)
	}

	if count > 0 {
		// Update existing preferences
		query := `
			UPDATE user_preferences 
			SET preferences = $2, allergies = $3, dislikes = $4, updated_at = $5
			WHERE user_id = $1
		`
		_, err = r.db.Exec(ctx, query, userID, prefs.Preferences, prefs.Allergies, prefs.Dislikes, time.Now())
	} else {
		// Insert new preferences
		query := `
			INSERT INTO user_preferences (user_id, preferences, allergies, dislikes, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, err = r.db.Exec(ctx, query, userID, prefs.Preferences, prefs.Allergies, prefs.Dislikes, time.Now(), time.Now())
	}

	if err != nil {
		return fmt.Errorf("failed to update user preferences: %w", err)
	}

	return nil
}

// GetUserPreferences retrieves user preferences
func (r *userRepository) GetUserPreferences(ctx context.Context, userID string) (*models.UserPreferences, error) {
	query := `
		SELECT user_id, preferences, allergies, dislikes
		FROM user_preferences 
		WHERE user_id = $1
	`

	var prefs models.UserPreferences
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&prefs.UserID, &prefs.Preferences, &prefs.Allergies, &prefs.Dislikes,
	)

	if err != nil {
		// Return empty preferences if not found rather than error
		return &models.UserPreferences{
			UserID:      userID,
			Preferences: []string{},
			Allergies:   []string{},
			Dislikes:    []string{},
		}, nil
	}

	return &prefs, nil
}

// UpdateUserPassword updates a user's password hash
func (r *userRepository) UpdateUserPassword(ctx context.Context, userID, passwordHash string) error {
	query := `
		UPDATE users 
		SET password_hash = $2, updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, userID, passwordHash, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found or already deleted")
	}

	return nil
}

// UpdateUserStats updates and returns user statistics
func (r *userRepository) UpdateUserStats(ctx context.Context, userID string) (*models.UserStats, error) {
    // First get current stats to calculate updates
    currentStats, err := r.GetUserStats(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to get current stats: %w", err)
    }

    // Calculate new statistics based on recent activities
    // This is where you'd add your business logic
    query := `
        SELECT 
            COUNT(*) as completed_workouts,
            COALESCE(SUM(weight * sets * reps), 0) as total_weight
        FROM exercises e
        JOIN workout_sessions ws ON e.session_id = ws.id
        JOIN workouts w ON ws.workout_id = w.id
        WHERE w.user_id = $1 AND e.completed = true
        AND e.created_at > $2
    `

    var newWorkouts int
    var newWeight float64
    err = r.db.QueryRow(ctx, query, userID, currentStats.UpdatedAt).Scan(
        &newWorkouts, &newWeight,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to calculate new stats: %w", err)
    }

    // Update the statistics
    updateQuery := `
        INSERT INTO user_stats (user_id, completed_workouts, completed_sessions, 
                               total_weight_lifted, current_streak, longest_streak, 
                               last_workout_date, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        ON CONFLICT (user_id) 
        DO UPDATE SET 
            completed_workouts = user_stats.completed_workouts + EXCLUDED.completed_workouts,
            total_weight_lifted = user_stats.total_weight_lifted + EXCLUDED.total_weight_lifted,
            last_workout_date = EXCLUDED.last_workout_date,
            updated_at = EXCLUDED.updated_at
        RETURNING id, completed_workouts, completed_sessions, total_weight_lifted,
                  current_streak, longest_streak, last_workout_date, created_at, updated_at
    `

    var stats models.UserStats
    err = r.db.QueryRow(ctx, updateQuery,
        userID, newWorkouts, 0, newWeight, // completed_sessions would need similar logic
        currentStats.CurrentStreak, currentStats.LongestStreak,
        time.Now(), time.Now(),
    ).Scan(
        &stats.ID, &stats.CompletedWorkouts, &stats.CompletedSessions,
        &stats.TotalWeightLifted, &stats.CurrentStreak, &stats.LongestStreak,
        &stats.LastWorkoutDate, &stats.CreatedAt, &stats.UpdatedAt,
    )

    if err != nil {
        return nil, fmt.Errorf("failed to update user stats: %w", err)
    }

    stats.UserID = userID
    return &stats, nil
}


// GetUserStats retrieves user statistics
func (r *userRepository) GetUserStats(ctx context.Context, userID string) (*models.UserStats, error) {
    query := `
        SELECT id, user_id, completed_workouts, completed_sessions, 
               total_weight_lifted, current_streak, longest_streak, 
               last_workout_date, created_at, updated_at
        FROM user_stats 
        WHERE user_id = $1
    `

    var stats models.UserStats
    err := r.db.QueryRow(ctx, query, userID).Scan(
        &stats.ID, &stats.UserID, &stats.CompletedWorkouts, &stats.CompletedSessions,
        &stats.TotalWeightLifted, &stats.CurrentStreak, &stats.LongestStreak,
        &stats.LastWorkoutDate, &stats.CreatedAt, &stats.UpdatedAt,
    )

    if err != nil {
        // Return default stats if not found
        return &models.UserStats{
            UserID:            userID,
            CompletedWorkouts: 0,
            CompletedSessions: 0,
            TotalWeightLifted: 0,
            CurrentStreak:     0,
            LongestStreak:     0,
        }, nil
    }

    return &stats, nil
}
