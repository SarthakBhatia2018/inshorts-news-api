package repositories

import (
	"math"
	"strings"
	"time"

	"gorm.io/gorm"

	"inshorts-news-api/models"
)

type ArticleRepository struct {
	db *gorm.DB
}

func NewArticleRepository(db *gorm.DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

func (r *ArticleRepository) Create(article *models.Article) error {
	return r.db.Create(article).Error
}

func (r *ArticleRepository) BulkCreate(articles []models.Article) error {
	return r.db.CreateInBatches(articles, 100).Error
}

func (r *ArticleRepository) GetByCategory(category string, limit int) ([]models.Article, error) {
	var articles []models.Article
	err := r.db.Where("? = ANY(category)", category).
		Order("publication_date DESC").
		Limit(limit).
		Find(&articles).Error
	return articles, err
}

func (r *ArticleRepository) GetBySource(source string, limit int) ([]models.Article, error) {
	var articles []models.Article
	err := r.db.Where("LOWER(source_name) LIKE LOWER(?)", "%"+source+"%").
		Order("publication_date DESC").
		Limit(limit).
		Find(&articles).Error
	return articles, err
}

func (r *ArticleRepository) GetByScore(minScore float64, limit int) ([]models.Article, error) {
	var articles []models.Article
	err := r.db.Where("relevance_score >= ?", minScore).
		Order("relevance_score DESC").
		Limit(limit).
		Find(&articles).Error
	return articles, err
}

func (r *ArticleRepository) SearchByText(query string, limit int) ([]models.Article, error) {
	var articles []models.Article
	searchQuery := "%" + strings.ToLower(query) + "%"

	err := r.db.Where("LOWER(title) LIKE ? OR LOWER(description) LIKE ?", searchQuery, searchQuery).
		Order("relevance_score DESC, publication_date DESC").
		Limit(limit).
		Find(&articles).Error

	return articles, err
}

func (r *ArticleRepository) GetNearby(lat, lon, radiusKm float64, limit int) ([]models.Article, error) {
	var articles []models.Article

	// Using subquery to properly handle the distance calculation
	query := `
        SELECT * FROM (
            SELECT *, (
                6371 * acos(
                    cos(radians(?)) * cos(radians(latitude)) * 
                    cos(radians(longitude) - radians(?)) + 
                    sin(radians(?)) * sin(radians(latitude))
                )
            ) AS distance
            FROM articles
            WHERE deleted_at IS NULL
        ) AS articles_with_distance
        WHERE distance < ?
        ORDER BY distance
        LIMIT ?
    `

	err := r.db.Raw(query, lat, lon, lat, radiusKm, limit).Scan(&articles).Error
	return articles, err
}

func (r *ArticleRepository) CreateUserEvent(event *models.UserEvent) error {
	return r.db.Create(event).Error
}

func (r *ArticleRepository) GetTrendingByLocation(lat, lon, radiusKm float64, limit int, hoursBack int) ([]models.TrendingArticle, error) {
	timeThreshold := time.Now().Add(-time.Duration(hoursBack) * time.Hour)

	latDelta := radiusKm / 111.0
	lonDelta := radiusKm / (111.0 * math.Cos(lat*math.Pi/180.0))

	minLat := lat - latDelta
	maxLat := lat + latDelta
	minLon := lon - lonDelta
	maxLon := lon + lonDelta

	type articleWithScore struct {
		models.Article
		WeightedScore    float64
		HoursSinceLast   float64
		InteractionCount int64
	}

	var articlesWithScores []articleWithScore

	query := `
        WITH nearby_articles AS (
            SELECT id, title, description, url, publication_date, source_name, 
                   category, relevance_score, latitude, longitude
            FROM articles
            WHERE deleted_at IS NULL
              AND latitude BETWEEN ? AND ?
              AND longitude BETWEEN ? AND ?
        ),
        trending_scores AS (
            SELECT 
                ue.article_id,
                COUNT(*) AS interaction_count,
                SUM(CASE 
                    WHEN ue.event_type = 'share' THEN 3.0
                    WHEN ue.event_type = 'click' THEN 2.0
                    WHEN ue.event_type = 'view' THEN 1.0
                    ELSE 0.0
                END) AS weighted_score,
                EXTRACT(EPOCH FROM (NOW() - MAX(ue.timestamp))) / 3600.0 AS hours_since_last
            FROM user_events ue
            WHERE ue.timestamp > ?
            GROUP BY ue.article_id
        )
        SELECT 
            na.id, na.title, na.description, na.url, na.publication_date,
            na.source_name, na.category, na.relevance_score, na.latitude, na.longitude,
            COALESCE(ts.weighted_score, 0) AS weighted_score,
            COALESCE(ts.hours_since_last, 999999) AS hours_since_last,
            COALESCE(ts.interaction_count, 0) AS interaction_count
        FROM nearby_articles na
        LEFT JOIN trending_scores ts ON na.id = ts.article_id
        ORDER BY 
            CASE 
                WHEN ts.weighted_score IS NOT NULL 
                THEN ts.weighted_score / (1 + ts.hours_since_last)
                ELSE 0 
            END DESC,
            na.publication_date DESC
        LIMIT ?
    `

	err := r.db.Raw(query, minLat, maxLat, minLon, maxLon, timeThreshold, limit).Scan(&articlesWithScores).Error
	if err != nil {
		return nil, err
	}

	results := make([]models.TrendingArticle, len(articlesWithScores))
	for i, aws := range articlesWithScores {
		var trendingScore float64
		if aws.WeightedScore > 0 {
			trendingScore = (aws.WeightedScore * 100) / (1 + aws.HoursSinceLast)
		}

		results[i] = models.TrendingArticle{
			Article: models.Article{
				ID:              aws.ID,
				Title:           aws.Title,
				Description:     aws.Description,
				URL:             aws.URL,
				PublicationDate: aws.PublicationDate,
				SourceName:      aws.SourceName,
				Category:        aws.Category,
				RelevanceScore:  aws.RelevanceScore,
				Latitude:        aws.Latitude,
				Longitude:       aws.Longitude,
			},
			TrendingScore: trendingScore,
		}
	}

	return results, nil
}

func (r *ArticleRepository) GetAllCategories() ([]string, error) {
	var categories []string
	err := r.db.Raw("SELECT DISTINCT unnest(category) FROM articles ORDER BY 1").Scan(&categories).Error
	return categories, err
}

func (r *ArticleRepository) GetAllSources() ([]string, error) {
	var sources []string
	err := r.db.Model(&models.Article{}).Distinct("source_name").Pluck("source_name", &sources).Error
	return sources, err
}

func (r *ArticleRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.Article{}).Count(&count).Error
	return count, err
}
