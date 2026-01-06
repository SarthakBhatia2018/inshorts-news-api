package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"inshorts-news-api/models"
	"inshorts-news-api/services"
	"inshorts-news-api/utils"
)

type ArticleHandler struct {
	articleService *services.ArticleService
	llmService     *services.LLMService
}

func NewArticleHandler(articleService *services.ArticleService, llmService *services.LLMService) *ArticleHandler {
	return &ArticleHandler{
		articleService: articleService,
		llmService:     llmService,
	}
}

// GET /api/v1/news/query
func (h *ArticleHandler) QueryNews(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Query parameter 'q' is required")
		return
	}

	location := c.Query("location")
	lat, _ := strconv.ParseFloat(c.Query("lat"), 64)
	lon, _ := strconv.ParseFloat(c.Query("lon"), 64)
	radius, _ := strconv.ParseFloat(c.DefaultQuery("radius", "50"), 64)

	// Analyze query using LLM
	intent, err := h.llmService.AnalyzeQuery(query, location)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to analyze query: "+err.Error())
		return
	}

	// Build parameters based on intent
	params := make(map[string]interface{})
	params["query"] = query

	switch intent.Intent {
	case "category":
		if len(intent.Entities) > 0 {
			params["category"] = intent.Entities
		} else {
			params["category"] = "General"
		}
	case "source":
		if len(intent.Entities) > 0 {
			params["source"] = intent.Entities
		}
	case "score":
		params["min_score"] = 0.7
	case "nearby":
		params["lat"] = lat
		params["lon"] = lon
		params["radius"] = radius
	}

	articles, err := h.articleService.GetArticlesByIntent(intent, params)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch articles: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"intent":   intent,
		"articles": articles,
		"count":    len(articles),
	})
}

// GET /api/v1/news/category
func (h *ArticleHandler) GetByCategory(c *gin.Context) {
	category := c.Query("category")
	if category == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Query parameter 'category' is required")
		return
	}

	intent := &models.QueryIntent{Intent: "category"}
	params := map[string]interface{}{"category": category}

	articles, err := h.articleService.GetArticlesByIntent(intent, params)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"articles": articles})
}

// GET /api/v1/news/source
func (h *ArticleHandler) GetBySource(c *gin.Context) {
	source := c.Query("source")
	if source == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Query parameter 'source' is required")
		return
	}

	intent := &models.QueryIntent{Intent: "source"}
	params := map[string]interface{}{"source": source}

	articles, err := h.articleService.GetArticlesByIntent(intent, params)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"articles": articles})
}

// GET /api/v1/news/score
func (h *ArticleHandler) GetByScore(c *gin.Context) {
	minScore, _ := strconv.ParseFloat(c.DefaultQuery("min_score", "0.7"), 64)

	intent := &models.QueryIntent{Intent: "score"}
	params := map[string]interface{}{"min_score": minScore}

	articles, err := h.articleService.GetArticlesByIntent(intent, params)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"articles": articles})
}

// GET /api/v1/news/search
func (h *ArticleHandler) Search(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Query parameter 'query' is required")
		return
	}

	intent := &models.QueryIntent{Intent: "search"}
	params := map[string]interface{}{"query": query}

	articles, err := h.articleService.GetArticlesByIntent(intent, params)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"articles": articles})
}

// GET /api/v1/news/nearby
func (h *ArticleHandler) GetNearby(c *gin.Context) {
	lat, err := strconv.ParseFloat(c.Query("lat"), 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid latitude")
		return
	}

	lon, err := strconv.ParseFloat(c.Query("lon"), 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid longitude")
		return
	}

	radius, _ := strconv.ParseFloat(c.DefaultQuery("radius", "10"), 64)

	intent := &models.QueryIntent{Intent: "nearby"}
	params := map[string]interface{}{
		"lat":    lat,
		"lon":    lon,
		"radius": radius,
	}

	articles, err := h.articleService.GetArticlesByIntent(intent, params)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"articles": articles})
}

// GET /api/v1/news/trending
func (h *ArticleHandler) GetTrending(c *gin.Context) {
	lat, err := strconv.ParseFloat(c.Query("lat"), 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid latitude")
		return
	}

	lon, err := strconv.ParseFloat(c.Query("lon"), 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid longitude")
		return
	}

	radius, _ := strconv.ParseFloat(c.DefaultQuery("radius", "50"), 64)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))
	hoursBack, _ := strconv.Atoi(c.DefaultQuery("hours_back", "24"))

	articles, err := h.articleService.GetTrending(lat, lon, radius, limit, hoursBack)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"articles": articles})
}

// POST /api/v1/events
func (h *ArticleHandler) RecordEvent(c *gin.Context) {
	var req struct {
		ArticleID string  `json:"article_id" binding:"required"`
		EventType string  `json:"event_type" binding:"required"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	err := h.articleService.RecordUserEvent(req.ArticleID, req.EventType, req.Latitude, req.Longitude)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Event recorded successfully"})
}
