package routes

import (
	"github.com/gin-gonic/gin"

	"inshorts-news-api/handlers"
	"inshorts-news-api/middleware"
)

func SetupRoutes(r *gin.Engine, handler *handlers.ArticleHandler) {
	r.Use(middleware.ErrorHandler())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")
	{
		news := v1.Group("/news")
		{
			news.GET("/query", handler.QueryNews)
			news.GET("/category", handler.GetByCategory)
			news.GET("/source", handler.GetBySource)
			news.GET("/score", handler.GetByScore)
			news.GET("/search", handler.Search)
			news.GET("/nearby", handler.GetNearby)
			news.GET("/trending", handler.GetTrending)
		}

		v1.POST("/events", handler.RecordEvent)
	}
}
