package handlers

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "yoked_backend/internal/api/middleware"
    "yoked_backend/internal/models"
    "yoked_backend/internal/services"
)

type AuthHandler struct {
    userService services.UserService
}

func NewAuthHandler(userService services.UserService) *AuthHandler {
    return &AuthHandler{
        userService: userService,
    }
}

// RegisterRequest represents the registration request body
type RegisterRequest struct {
    Email           string  `json:"email" binding:"required,email"`
    PasswordHash    string  `json:"password_hash" binding:"required,min=8"`
    Name            string  `json:"name" binding:"required"`
    Age             int     `json:"age" binding:"required,min=13,max=120"`
    Sex             string  `json:"sex" binding:"required,oneof=male female other"`
    Height          float64 `json:"height" binding:"required,min=30,max=250"`
    Weight          float64 `json:"weight" binding:"required,min=20,max=500"`
    ActivityLevel   string  `json:"activity_level" binding:"required,oneof=sedentary lightly_active moderately_active very_active extra_active"`
    Goal            string  `json:"goal" binding:"required,oneof=weight_loss muscle_gain maintenance endurance"`
    ProgramID       int     `json:"program_id" binding:"required"`
    WeeklyBudget    float64 `json:"weekly_budget" binding:"min=0"`
}

// LoginRequest represents the login request body
type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
    Token     string        `json:"token"`
    ExpiresAt time.Time     `json:"expires_at"`
    User      *models.User  `json:"user"`
}

func (h *AuthHandler) Register(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
        return
    }

    // Check if user already exists
    existingUser, err := h.userService.GetUserByEmail(c.Request.Context(), req.Email)
    if existingUser != nil && err == nil {
        c.JSON(http.StatusConflict, gin.H{"error": "User already exists with this email"})
        return
    }

    // Hash password
    hashedPassword, err := middleware.HashPassword(req.PasswordHash)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
        return
    }

    // Create user
    user := &models.User{
        Email:         req.Email,
        PasswordHash:  hashedPassword,
        Name:          req.Name,
        Age:           req.Age,
        Sex:           req.Sex,
        Height:        req.Height,
        Weight:        req.Weight,
        ActivityLevel: req.ActivityLevel,
        Goal:          req.Goal,
	ProgramID:     req.ProgramID,
        WeeklyBudget:  req.WeeklyBudget,
    }

    if err := h.userService.CreateUser(c.Request.Context(), user); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user: " + err.Error()})
        return
    }


    userProgram := &models.UserProgram{
        UserID:    user.ID,
        ProgramID: req.ProgramID,
        StartDate: time.Now(),
        IsActive:  true,
    }


    if err := h.userService.CreateUserProgram(c.Request.Context(), userProgram); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user program: " + err.Error()})
        return
    }





    // Generate JWT token
    token, err := middleware.GenerateJWT(user.ID, user.Email)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }

    expiresAt := time.Now().Add(24 * time.Hour * 7)

    // Clear password hash from response
    user.PasswordHash = ""

    c.JSON(http.StatusCreated, AuthResponse{
        Token:     token,
        ExpiresAt: expiresAt,
        User:      user,
    })
}

func (h *AuthHandler) Login(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
        return
    }

    // Get user by email
    user, err := h.userService.GetUserByEmail(c.Request.Context(), req.Email)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
        return
    }

    // Check password
    if err := middleware.CheckPassword(req.Password, user.PasswordHash); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
        return
    }

    // Generate JWT token
    token, err := middleware.GenerateJWT(user.ID, user.Email)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }

    expiresAt := time.Now().Add(24 * time.Hour * 7)

    // Clear password hash from response
    user.PasswordHash = ""

    c.JSON(http.StatusOK, AuthResponse{
        Token:     token,
        ExpiresAt: expiresAt,
        User:      user,
    })
}

func (h *AuthHandler) Logout(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
    userID, err := middleware.GetUserIDFromContext(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
        return
    }

    user, err := h.userService.GetUserByID(c.Request.Context(), userID)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
        return
    }

    // Generate new JWT token
    token, err := middleware.GenerateJWT(user.ID, user.Email)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }

    expiresAt := time.Now().Add(24 * time.Hour * 7)

    // Clear password hash from response
    user.PasswordHash = ""

    c.JSON(http.StatusOK, AuthResponse{
        Token:     token,
        ExpiresAt: expiresAt,
        User:      user,
    })
}
