package handlers

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "yoked_backend/internal/models"
    "yoked_backend/internal/services"
)

type WorkoutHandler struct {
    workoutService services.WorkoutService
}

func NewWorkoutHandler(workoutService services.WorkoutService) *WorkoutHandler {
    return &WorkoutHandler{workoutService: workoutService}
}

func (h *WorkoutHandler) GetWorkouts(c *gin.Context) {
    userID := c.MustGet("userID").(string)
    
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
    
    workouts, total, err := h.workoutService.GetUserWorkouts(c.Request.Context(), userID, page, limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "data": workouts,
        "pagination": gin.H{
            "page":  page,
            "limit": limit,
            "total": total,
        },
    })
}

func (h *WorkoutHandler) GetCurrentWorkout(c *gin.Context) {
    userID := c.MustGet("userID").(string)
    
    workout, err := h.workoutService.GetCurrentWorkout(c.Request.Context(), userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, workout)
}

func (h *WorkoutHandler) GenerateWorkout(c *gin.Context) {
    userID := c.MustGet("userID").(string)
    
    var request struct {
        Type     string `json:"type" binding:"required"`
        Duration int    `json:"duration" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    workout, err := h.workoutService.GenerateWorkout(c.Request.Context(), userID, request.Type, request.Duration)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusCreated, workout)
}

func (h *WorkoutHandler) GetWorkoutDetails(c *gin.Context) {
    workoutID := c.Param("workout_id")
    
    workout, err := h.workoutService.GetWorkoutByID(c.Request.Context(), workoutID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Workout not found"})
        return
    }
    
    c.JSON(http.StatusOK, workout)
}

func (h *WorkoutHandler) GetWorkoutSession(c *gin.Context) {
    sessionID := c.Param("session_id")
    
    session, err := h.workoutService.GetSessionByID(c.Request.Context(), sessionID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
        return
    }
    
    c.JSON(http.StatusOK, session)
}

func (h *WorkoutHandler) StartSession(c *gin.Context) {
    sessionID := c.Param("session_id")
    
    err := h.workoutService.StartSession(c.Request.Context(), sessionID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"message": "Session started successfully"})
}

func (h *WorkoutHandler) CompleteSession(c *gin.Context) {
    sessionID := c.Param("session_id")
    
    err := h.workoutService.CompleteSession(c.Request.Context(), sessionID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"message": "Session completed successfully"})
}

func (h *WorkoutHandler) LogExercise(c *gin.Context) {
    sessionID := c.Param("session_id")
    
    var exercise models.Exercise
    if err := c.ShouldBindJSON(&exercise); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
        return
    }
    
    exercise.SessionID = sessionID
    
    err := h.workoutService.LogExercise(c.Request.Context(), &exercise)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"message": "Exercise logged successfully"})
}

func (h *WorkoutHandler) CreateWorkout(c *gin.Context) {
    userID := c.MustGet("userID").(string)
    
    var workout models.Workout
    if err := c.ShouldBindJSON(&workout); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
        return
    }
    
    workout.UserID = userID
    
    err := h.workoutService.CreateWorkout(c.Request.Context(), &workout)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusCreated, workout)
}

func (h *WorkoutHandler) GetWorkoutSessions(c *gin.Context) {
    workoutID := c.Param("workout_id")
    
    sessions, err := h.workoutService.GetWorkoutSessions(c.Request.Context(), workoutID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, sessions)
}

func (h *WorkoutHandler) GetSessionExercises(c *gin.Context) {
    sessionID := c.Param("session_id")
    
    exercises, err := h.workoutService.GetSessionExercises(c.Request.Context(), sessionID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, exercises)
}
