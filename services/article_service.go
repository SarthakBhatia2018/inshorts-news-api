package services

import (
    "fmt"
    
    "inshorts-news-api/models"
    "inshorts-news-api/repositories"
)

type ArticleService struct {
    repo       *repositories.ArticleRepository
    llmService *LLMService
}

func NewArticleService(repo *repositories.ArticleRepository, llmService *LLMService) *ArticleService {
    return &ArticleService{
        repo:       repo,
        llmService: llmService,
    }
}

func (s *ArticleService) GetArticlesByIntent(intent *models.QueryIntent, params map[string]interface{}) ([]models.ArticleResponse, error) {
    var articles []models.Article
    var err error
    limit := 5

    switch intent.Intent {
    case "category":
        category := params["category"].(string)
        articles, err = s.repo.GetByCategory(category, limit)
    case "source":
        source := params["source"].(string)
        articles, err = s.repo.GetBySource(source, limit)
    case "score":
        minScore := params["min_score"].(float64)
        articles, err = s.repo.GetByScore(minScore, limit)
    case "search":
        query := params["query"].(string)
        articles, err = s.repo.SearchByText(query, limit)
    case "nearby":
        lat := params["lat"].(float64)
        lon := params["lon"].(float64)
        radius := params["radius"].(float64)
        articles, err = s.repo.GetNearby(lat, lon, radius, limit)
    default:
        return nil, fmt.Errorf("unknown intent: %s", intent.Intent)
    }

    if err != nil {
        return nil, err
    }

    return s.enrichArticles(articles)
}

func (s *ArticleService) enrichArticles(articles []models.Article) ([]models.ArticleResponse, error) {
    summaries, err := s.llmService.BatchGenerateSummaries(articles)
    if err != nil {
        return nil, err
    }

    responses := make([]models.ArticleResponse, len(articles))
    for i, article := range articles {
        responses[i] = models.ArticleResponse{
            Title:           article.Title,
            Description:     article.Description,
            URL:             article.URL,
            PublicationDate: article.PublicationDate,
            SourceName:      article.SourceName,
            Category:        article.Category,
            RelevanceScore:  article.RelevanceScore,
            LLMSummary:      summaries[article.ID],
            Latitude:        article.Latitude,
            Longitude:       article.Longitude,
        }
    }

    return responses, nil
}

func (s *ArticleService) GetTrending(lat, lon, radius float64, limit, hoursBack int) ([]models.ArticleResponse, error) {
    trendingArticles, err := s.repo.GetTrendingByLocation(lat, lon, radius, limit, hoursBack)
    if err != nil {
        return nil, err
    }

    articles := make([]models.Article, len(trendingArticles))
    for i, ta := range trendingArticles {
        articles[i] = ta.Article
    }

    return s.enrichArticles(articles)
}

func (s *ArticleService) RecordUserEvent(articleID string, eventType string, lat, lon float64) error {
    event := &models.UserEvent{
        ArticleID: articleID,
        EventType: eventType,
        Latitude:  lat,
        Longitude: lon,
    }
    return s.repo.CreateUserEvent(event)
}
