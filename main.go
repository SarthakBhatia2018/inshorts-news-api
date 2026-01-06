package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"inshorts-news-api/config"
	"inshorts-news-api/db"
	"inshorts-news-api/handlers"
	"inshorts-news-api/repositories"
	"inshorts-news-api/routes"
	"inshorts-news-api/services"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to database
	if err := db.Connect(cfg); err != nil {
		log.Fatal("Database connection failed:", err)
	}

	// Auto-migrate models
	if err := db.AutoMigrate(); err != nil {
		log.Fatal("Migration failed:", err)
	}

	// Initialize dependencies
	articleRepo := repositories.NewArticleRepository(db.GetDB())
	llmService := services.NewLLMService(cfg.OpenAIKey)
	articleService := services.NewArticleService(articleRepo, llmService)
	articleHandler := handlers.NewArticleHandler(articleService, llmService)

	// Setup Gin router
	r := gin.Default()
	routes.SetupRoutes(r, articleHandler)

	// Start server
	log.Printf("Server starting on port %s...", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
