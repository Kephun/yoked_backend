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
    programRepo := repositories.NewProgramRepository(database.GetPool())

    // Initialize Services
    userService := services.NewUserService(userRepo)
    programService := services.NewProgramService(programRepo, userRepo)

    // Initialize Handlers
    userHandler := handlers.NewUserHandler(userService)
    programHandler := handlers.NewProgramHandler(programService, userService)
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
	public.GET("/programs?goal=hypertrophy",programHandler.GetProgramsByGoal)
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
	//Program Routes
	programs := authenticated.Group("/programs")
	{
		programs.POST("/assign", programHandler.AssignProgram)
		programs.GET("/user/:user_id", programHandler.GetUserProgram)

	}
	// Workout Routes
	workouts := authenticated.Group("/workouts")
	{
    		workouts.POST("/start", programHandler.StartWorkoutSession)
		workouts.POST("/:id/complete", programHandler.CompleteWorkoutSession)
		workouts.GET("/history/:user_id", programHandler.GetWorkoutHistory)
		workouts.GET("/next-weights", programHandler.GetNextWorkoutWeights)
    	}

    // Start Server - Listening to ALL MUST CHANGE BEFORE PRODUCTION
    if err := router.Run("0.0.0.0:8080"); err != nil {
        log.Fatal("Failed to start server: ", err)
        }
    }
}

func healthCheck(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}


