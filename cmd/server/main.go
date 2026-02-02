package main

import (
	"fmt"
	"log"
	"opus-api/internal/handler"
	"opus-api/internal/logger"
	"opus-api/internal/tokenizer"
	"opus-api/internal/types"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
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

	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Create Gin router
	router := gin.Default()

	// Register routes
	router.POST("/v1/messages", handler.HandleMessages)
	router.GET("/health", handler.HandleHealth)

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

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
