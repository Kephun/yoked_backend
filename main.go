package main

import (
    "log"
    "net/http"

    "github.com/gin-gonic/gin"
    "yoked_backend/internal/api/handlers"
    "yoked_backend/internal/api/middleware"
    "yoked_backend/internal/db"
    "yoked_backend/internal/db/repositories"
    "yoked_backend/internal/services"
)

func main() {
    // Initialize DB
    database, err := db.NewDatabase()
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer database.Close()

    // Initialize Repositories
    userRepo := repositories.NewUserRepository(database.GetPool())
    workoutRepo := repositories.NewWorkoutRepository(database.GetPool())

    // Initialize Services
    userService := services.NewUserService(userRepo)
    workoutService := services.NewWorkoutService(workoutRepo, userRepo)

    // Initialize Handlers
    userHandler := handlers.NewUserHandler(userService)
    workoutHandler := handlers.NewWorkoutHandler(workoutService)
    authHandler := handlers.NewAuthHandler(userService)

    router := gin.Default()
    
    // Middleware
    router.Use(gin.Logger())
    router.Use(gin.Recovery())

    // Auth Routes
    public := router.Group("/")
    {
        public.POST("/auth/register", authHandler.Register)
        public.POST("/auth/login", authHandler.Login)
        public.GET("/health", healthCheck)
    }

    // Authenticated Routes - Requires JWT
    authenticated := router.Group("/")
    authenticated.Use(middleware.AuthMiddleware())
    {
        // Auth
        authenticated.POST("/auth/logout", authHandler.Logout)
        authenticated.POST("/auth/refresh", authHandler.RefreshToken)

        // User Routes
	user := authenticated.Group("/users")
	{
    		user.GET("/me", userHandler.GetCurrentUser)
    		user.PUT("/me", userHandler.UpdateUser)
    		user.GET("/me/preferences", userHandler.GetUserPreferences)
    		user.PUT("/me/preferences", userHandler.UpdateUserPreferences)
    		user.PUT("/me/password", userHandler.UpdatePassword)
    		user.GET("/me/stats", userHandler.GetUserStats)
	}

	// Workout Routes
	workouts := authenticated.Group("/workouts")
	{
    		workouts.GET("", workoutHandler.GetWorkouts)
    		workouts.POST("", workoutHandler.CreateWorkout)
    		workouts.GET("/current", workoutHandler.GetCurrentWorkout)
    		workouts.POST("/generate", workoutHandler.GenerateWorkout)
    		workouts.GET("/:workout_id", workoutHandler.GetWorkoutDetails)
    		workouts.GET("/:workout_id/sessions", workoutHandler.GetWorkoutSessions)
    	}
    	// Session routes
    	sessions := workouts.Group("/sessions")
    	{
        	sessions.GET("/:session_id", workoutHandler.GetWorkoutSession)
        	sessions.POST("/:session_id/start", workoutHandler.StartSession)
        	sessions.POST("/:session_id/complete", workoutHandler.CompleteSession)
        	sessions.GET("/:session_id/exercises", workoutHandler.GetSessionExercises)
        	sessions.POST("/:session_id/exercises", workoutHandler.LogExercise)
    	}
    }

    // Start Server - Listening to ALL MUST CHANGE BEFORE PRODUCTION
    if err := router.Run("0.0.0.0:8080"); err != nil {
        log.Fatal("Failed to start server: ", err)
    }
}

func healthCheck(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}
