package main

import (
	"log"
	"net/http"
	"os"

	"backend-test/internal/database"
	"backend-test/internal/entity"
	"backend-test/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if available
	if err := godotenv.Load(); err != nil {
		log.Println("Notice: No .env file found or error loading, falling back to environment variables")
	}

	env := os.Getenv("ENV")
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database connection
	dbConfig := database.LoadConfigFromEnv()
	db, err := database.Connect(dbConfig)
	if err != nil {
		log.Fatalf("Fatal: Database initialization failed: %v", err)
	}

	// Initialize layers
	repo := entity.NewRepository(db)
	service := entity.NewService(repo)
	handler := entity.NewHandler(service)

	// Set up Gin engine
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())

	// Health check endpoint
	healthHandler := func(c *gin.Context) {
		c.JSON(http.StatusOK, entity.NewSuccessResponse("Geospatial Entity Management API is running", nil))
	}
	r.GET("/health", healthHandler)
	r.GET("/api/health", healthHandler)

	// API routes
	apiGroup := r.Group("/api")
	handler.RegisterRoutes(apiGroup)

	// Custom 404 route handling for consistent error response
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, entity.NewErrorResponse("Endpoint not found", nil))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server stopped unexpectedly: %v", err)
	}
}
