package handlers

import (
    "net/http"
    "strconv"
    //"log"
    "github.com/gin-gonic/gin"

    "yoked_backend/internal/services"
)


type ProgramHandler struct {
	programService services.ProgramService
	userService services.UserService
}

func NewProgramHandler(programService services.ProgramService, userService services.UserService) *ProgramHandler {
	return &ProgramHandler{programService: programService, userService: userService}
}

// GetProgramsByGoal returns programs filtered by goal
// GET /api/programs?goal=hypertrophy
func (h *ProgramHandler) GetProgramsByGoal(c *gin.Context) {
	goal := c.Query("goal")
	if goal == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Goal parameter is required"})
		return
	}

	programs, err := h.programService.GetProgramsByGoal(c.Request.Context(), goal)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, programs)
}

// GetAllExercises returns all available exercises
// GET /api/exercises
func (h *ProgramHandler) GetAllExercises(c *gin.Context) {
	exercises, err := h.programService.GetAllExercises(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, exercises)
}

// GetExerciseByID returns a specific exercise
// GET /api/exercises/{id}
func (h *ProgramHandler) GetExerciseByID(c *gin.Context) {
	exerciseID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid exercise ID"})
		return
	}

	exercise, err := h.programService.GetExerciseByID(c.Request.Context(), exerciseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, exercise)
}

// AssignProgram assigns a program to a user
// POST /api/programs/assign
func (h *ProgramHandler) AssignProgram(c *gin.Context) {
	var request struct {
		UserID    string `json:"user_id"`
		ProgramID int    `json:"program_id"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err := h.programService.AssignProgramToUser(c.Request.Context(), request.UserID, request.ProgramID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Program assigned successfully"})
}

// GetUserProgram returns user's current program with workouts
// GET /api/programs/user/{user_id}
func (h *ProgramHandler) GetUserProgram(c *gin.Context) {
	userID := c.Param("user_id")

	// Get user details first
	//user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	//	return
	//}

	// Get program details
	programDetail, err := h.programService.GetUserProgramWithWorkouts(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Calculate weights on demand
	//weights, err := h.programService.CalculateInitialWeights(c.Request.Context(), user, programDetail.UserProgram.ProgramID)
	//if err != nil {
		// Log but don't fail the request
	//	log.Printf("Warning: Failed to calculate weights: %v", err)
	//}

	// Add weights to the response
	//programDetail.SuggestedWeights = weights

	c.JSON(http.StatusOK, programDetail)
}


// StartWorkoutSession starts a new workout session
// POST /api/workouts/start
func (h *ProgramHandler) StartWorkoutSession(c *gin.Context) {
	var request struct {
		UserID           string `json:"user_id"`
		ProgramWorkoutID int    `json:"program_workout_id"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	session, err := h.programService.StartWorkoutSession(c.Request.Context(), request.UserID, request.ProgramWorkoutID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, session)
}

// CompleteWorkoutSession completes a workout with exercise logs
// POST /api/workouts/{id}/complete
func (h *ProgramHandler) CompleteWorkoutSession(c *gin.Context) {
	sessionID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	var request struct {
		Exercises []struct {
			ProgramWorkoutExerciseID int   `json:"program_workout_exercise_id"`
			ActualReps               []int `json:"actual_reps"`
			ActualRIR                []int `json:"actual_rir"`
		} `json:"exercises"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Convert to service layer format
	exerciseLogs := make([]services.ExerciseLogRequest, len(request.Exercises))
	for i, ex := range request.Exercises {
		exerciseLogs[i] = services.ExerciseLogRequest{
			ProgramWorkoutExerciseID: ex.ProgramWorkoutExerciseID,
			ActualReps:               ex.ActualReps,
			ActualRIR:                ex.ActualRIR,
		}
	}

	err = h.programService.CompleteWorkoutSession(c.Request.Context(), sessionID, exerciseLogs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Workout completed successfully"})
}

// GetWorkoutHistory returns user's workout history
// GET /api/workouts/history/{user_id}
func (h *ProgramHandler) GetWorkoutHistory(c *gin.Context) {
	userID := c.Param("user_id")

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	history, err := h.programService.GetWorkoutHistory(c.Request.Context(), userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, history)
}

// GetNextWorkoutWeights calculates weights for next workout based on previous performance
// GET /api/workouts/next-weights?user_id=xxx&program_workout_id=123
func (h *ProgramHandler) GetNextWorkoutWeights(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	programWorkoutID, err := strconv.Atoi(c.Query("program_workout_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid program workout ID"})
		return
	}

	weights, err := h.programService.CalculateNextWorkoutWeights(c.Request.Context(), userID, programWorkoutID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, weights)
}
