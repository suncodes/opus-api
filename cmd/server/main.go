package main

import (
	"fmt"
	"log"
	"opus-api/internal/handler"
	"opus-api/internal/logger"
	"opus-api/internal/middleware"
	"opus-api/internal/model"
	"opus-api/internal/service"
	"opus-api/internal/tokenizer"
	"opus-api/internal/types"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("[INFO] No .env file found or error loading it: %v", err)
	} else {
		log.Printf("[INFO] .env file loaded successfully")
	}

	// Create logs directory
	if err := os.MkdirAll(types.LogDir, 0755); err != nil {
		log.Fatalf("Failed to create logs directory: %v", err)
	}

	// Cleanup old logs on startup
	if types.DebugMode {
		logger.CleanupOldLogs()
	}

	// Initialize tokenizer for token counting
	if err := tokenizer.Init(); err != nil {
		log.Printf("[WARN] Failed to initialize tokenizer: %v (will use fallback)", err)
	}

	// Initialize database
	if err := model.InitDB(); err != nil {
		log.Printf("[WARN] Failed to initialize database: %v (running without database)", err)
	} else {
		// Create default admin user
		if err := model.CreateDefaultAdmin(model.DB); err != nil {
			log.Printf("[WARN] Failed to create default admin: %v", err)
		}
	}

	// Initialize services
	var authService *service.AuthService
	var cookieService *service.CookieService
	var cookieValidator *service.CookieValidator
	var cookieRotator *service.CookieRotator

	if model.DB != nil {
		authService = service.NewAuthService(model.DB)
		cookieService = service.NewCookieService(model.DB)
		cookieValidator = service.NewCookieValidator(cookieService)
		cookieRotator = service.NewCookieRotator(cookieService, service.StrategyRoundRobin)

		// Store rotator in types for use in messages handler
		types.CookieRotatorInstance = cookieRotator
	}

	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Create Gin router
	router := gin.Default()

	// Serve static files
	router.Static("/static", "./web/static")

	// Redirect root to login page or dashboard
	router.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/static/index.html")
	})

	// Dashboard redirect
	router.GET("/dashboard", func(c *gin.Context) {
		c.Redirect(302, "/static/dashboard.html")
	})

	// Register API routes
	router.POST("/v1/messages", handler.HandleMessages)
	router.GET("/health", handler.HandleHealth)

	// Auth routes (only if database is available)
	if authService != nil {
		authHandler := handler.NewAuthHandler(authService)
		router.POST("/api/auth/login", authHandler.Login)
		router.POST("/api/auth/logout", authHandler.Logout)

		// Protected routes
		authGroup := router.Group("/api")
		authGroup.Use(middleware.AuthMiddleware(authService))
		{
			authGroup.GET("/auth/me", authHandler.Me)
			authGroup.PUT("/auth/password", authHandler.ChangePassword)

			// Cookie management routes
			if cookieService != nil && cookieValidator != nil {
				cookieHandler := handler.NewCookieHandler(cookieService, cookieValidator)
				authGroup.GET("/cookies", cookieHandler.ListCookies)
				authGroup.GET("/cookies/stats", cookieHandler.GetStats)
				authGroup.POST("/cookies", cookieHandler.CreateCookie)
				authGroup.GET("/cookies/:id", cookieHandler.GetCookie)
				authGroup.PUT("/cookies/:id", cookieHandler.UpdateCookie)
				authGroup.DELETE("/cookies/:id", cookieHandler.DeleteCookie)
				authGroup.POST("/cookies/:id/validate", cookieHandler.ValidateCookie)
				authGroup.POST("/cookies/validate/all", cookieHandler.ValidateAllCookies)
			}
		}
	}

	// Start server
	// Hugging Face Spaces uses port 7860
	port := 7860
	if envPort := os.Getenv("PORT"); envPort != "" {
		fmt.Sscanf(envPort, "%d", &port)
	}
	addr := fmt.Sprintf("0.0.0.0:%d", port)
	log.Printf("Server running on http://0.0.0.0:%d", port)
	log.Printf("Debug mode: %v", types.DebugMode)
	log.Printf("Log directory: %s", types.LogDir)
	log.Printf("Database connected: %v", model.DB != nil)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
