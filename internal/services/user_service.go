package services

import (
    "fmt"    
    "context"
    "strings"

    "yoked_backend/internal/models"
    "yoked_backend/internal/db/repositories"
)

type UserService interface {
    CreateUser(ctx context.Context, user *models.User) error
    CreateUserProgram(ctx context.Context, userProgram *models.UserProgram) error
    GetUserByID(ctx context.Context, userID string) (*models.User, error)
    GetUserByEmail(ctx context.Context, email string) (*models.User, error)
    UpdateUser(ctx context.Context, user *models.User) error
    GetUserProfile(ctx context.Context, userID string) (*models.User, error)
    UpdateUserProfile(ctx context.Context, userID string, updates map[string]interface{}) (*models.User, error)
    UpdateUserPreferences(ctx context.Context, userID string, prefs *models.UserPreferences) error
    GetUserPreferences(ctx context.Context, userID string) (*models.UserPreferences, error)
    UpdateUserPassword(ctx context.Context, userID, passwordHash string) error
}

type userService struct {
    userRepo repositories.UserRepository
}

func NewUserService(userRepo repositories.UserRepository) UserService {
    return &userService{userRepo: userRepo}
}

func (s *userService) CreateUser(ctx context.Context, user *models.User) error {
    return s.userRepo.CreateUser(ctx, user)
}

func (s *userService) CreateUserProgram(ctx context.Context, userProgram *models.UserProgram) error {
	return s.userRepo.CreateUserProgram(ctx, userProgram)
}

func (s *userService) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
    user, err := s.userRepo.GetUserByID(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    return user, nil
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
    user, err := s.userRepo.GetUserByEmail(ctx, email)
    if err != nil {
        return nil, fmt.Errorf("failed to get user by email: %w", err)
    }
    return user, nil
}

func (s *userService) UpdateUser(ctx context.Context, user *models.User) error {
    return s.userRepo.UpdateUser(ctx, user)
}

func (s *userService) GetUserProfile(ctx context.Context, userID string) (*models.User, error) {
    user, err := s.userRepo.GetUserByID(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to get user profile: %w", err)
    }

    // Clear sensitive data before returning
    user.PasswordHash = ""
    return user, nil
}

func (s *userService) UpdateUserProfile(ctx context.Context, userID string, updates map[string]interface{}) (*models.User, error) {
    user, err := s.userRepo.GetUserByID(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("user not found: %w", err)
    }

    // Validate and apply updates
    if err := s.validateAndApplyProfileUpdates(user, updates); err != nil {
        return nil, err
    }

    err = s.userRepo.UpdateUser(ctx, user)
    if err != nil {
        return nil, fmt.Errorf("failed to update user profile: %w", err)
    }

    // Clear sensitive data before returning
    user.PasswordHash = ""
    return user, nil
}

func (s *userService) UpdateUserPassword(ctx context.Context, userID, passwordHash string) error {
    return s.userRepo.UpdateUserPassword(ctx, userID, passwordHash)
}

// validateAndApplyProfileUpdates validates and applies profile updates
func (s *userService) validateAndApplyProfileUpdates(user *models.User, updates map[string]interface{}) error {
    validGoals := map[string]bool{
        "weight_loss": true, "muscle_gain": true, "maintenance": true, "endurance": true,
    }

    validActivityLevels := map[string]bool{
        "sedentary": true, "lightly_active": true, "moderately_active": true,
        "very_active": true, "extra_active": true,
    }

    for key, value := range updates {
        switch key {
        case "weight":
            if weight, ok := value.(float64); ok && weight >= 20 && weight <= 500 {
                user.Weight = weight
            }
        case "height":
            if height, ok := value.(float64); ok && height >= 30 && height <= 250 {
                user.Height = height
            }
        case "age":
            if age, ok := value.(int); ok && age >= 13 && age <= 120 {
                user.Age = age
            }
        case "goal":
            if goal, ok := value.(string); ok && validGoals[goal] {
                user.Goal = goal
            }
	case "program_id":
	    if program_id, ok := value.(int); ok && program_id > 0 {
		user.ProgramID = program_id
            }
        case "activity_level":
            if activityLevel, ok := value.(string); ok && validActivityLevels[activityLevel] {
                user.ActivityLevel = activityLevel
            }
        case "weekly_budget":
            if budget, ok := value.(float64); ok && budget >= 0 {
                user.WeeklyBudget = budget
            }
        case "name":
            if name, ok := value.(string); ok && strings.TrimSpace(name) != "" {
                user.Name = strings.TrimSpace(name)
            }
        }
    }

    return nil
}

func (s *userService) UpdateUserPreferences(ctx context.Context, userID string, prefs *models.UserPreferences) error {
    if _, err := s.userRepo.GetUserByID(ctx, userID); err != nil {
        return fmt.Errorf("user not found: %w", err)
    }

    if prefs == nil {
        return fmt.Errorf("preferences cannot be nil")
    }

    // Validate preferences
    if err := s.validatePreferences(prefs); err != nil {
        return err
    }

    prefs.UserID = userID

    if err := s.userRepo.UpdateUserPreferences(ctx, userID, prefs); err != nil {
        return fmt.Errorf("failed to update user preferences: %w", err)
    }

    return nil
}

// validatePreferences validates user preferences
func (s *userService) validatePreferences(prefs *models.UserPreferences) error {
    if len(prefs.Preferences) > 50 {
        return fmt.Errorf("too many preferences (max 50)")
    }

    if len(prefs.Allergies) > 20 {
        return fmt.Errorf("too many allergies (max 20)")
    }

    if len(prefs.Dislikes) > 30 {
        return fmt.Errorf("too many dislikes (max 30)")
    }

    return nil
}

func (s *userService) GetUserPreferences(ctx context.Context, userID string) (*models.UserPreferences, error) {
    if _, err := s.userRepo.GetUserByID(ctx, userID); err != nil {
        return nil, fmt.Errorf("user not found: %w", err)
    }

    preferences, err := s.userRepo.GetUserPreferences(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to get user preferences: %w", err)
    }

    return preferences, nil
}
