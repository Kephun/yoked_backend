package handlers

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "yoked_backend/internal/api/middleware"
    "yoked_backend/internal/models"
    "yoked_backend/internal/services"
)

type UserHandler struct {
    userService services.UserService
}

func NewUserHandler(userService services.UserService) *UserHandler {
    return &UserHandler{userService: userService}
}

func (h *UserHandler) GetCurrentUser(c *gin.Context) {
    userID := c.MustGet("userID").(string)
    
    user, err := h.userService.GetUserProfile(c.Request.Context(), userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, user)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
    userID := c.MustGet("userID").(string)
    
    var updates map[string]interface{}
    if err := c.ShouldBindJSON(&updates); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
        return
    }
    
    user, err := h.userService.UpdateUserProfile(c.Request.Context(), userID, updates)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, user)
}

func (h *UserHandler) GetUserPreferences(c *gin.Context) {
    userID := c.MustGet("userID").(string)
    
    preferences, err := h.userService.GetUserPreferences(c.Request.Context(), userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, preferences)
}

func (h *UserHandler) UpdateUserPreferences(c *gin.Context) {
    userID := c.MustGet("userID").(string)
    
    var prefs models.UserPreferences
    if err := c.ShouldBindJSON(&prefs); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
        return
    }
    
    err := h.userService.UpdateUserPreferences(c.Request.Context(), userID, &prefs)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"message": "Preferences updated successfully"})
}

func (h *UserHandler) UpdatePassword(c *gin.Context) {
    userID := c.MustGet("userID").(string)
    
    var request struct {
        OldPassword string `json:"old_password" binding:"required"`
        NewPassword string `json:"new_password" binding:"required,min=8"`
    }
    
    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
        return
    }
    
    // Get user to verify old password
    user, err := h.userService.GetUserByID(c.Request.Context(), userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
        return
    }
    
    // Verify old password
    if err := middleware.CheckPassword(request.OldPassword, user.PasswordHash); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid old password"})
        return
    }
    
    // Hash new password
    hashedPassword, err := middleware.HashPassword(request.NewPassword)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
        return
    }
    
    // Update password
    err = h.userService.UpdateUserPassword(c.Request.Context(), userID, hashedPassword)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

func (h *UserHandler) GetUserStats(c *gin.Context) {
    userID := c.MustGet("userID").(string)
    
    // This would use a stats service if you implement it
    // For now, just return a placeholder response
    c.JSON(http.StatusOK, gin.H{
        "user_id": userID,
        "message": "User stats endpoint - implement stats service",
    })
}
