package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Article struct {
	ID              string         `gorm:"primaryKey" json:"id"`
	Title           string         `gorm:"type:text;index:idx_title" json:"title"`
	Description     string         `gorm:"type:text" json:"description"`
	URL             string         `gorm:"type:text" json:"url"`
	PublicationDate time.Time      `gorm:"index:idx_pub_date" json:"publication_date"`
	SourceName      string         `gorm:"index:idx_source" json:"source_name"`
	Category        pq.StringArray `gorm:"type:text[]" json:"category"` // Changed from []string
	RelevanceScore  float64        `gorm:"index:idx_relevance" json:"relevance_score"`
	Latitude        float64        `gorm:"index:idx_location" json:"latitude"`
	Longitude       float64        `gorm:"index:idx_location" json:"longitude"`
	CreatedAt       time.Time      `json:"-"`
	UpdatedAt       time.Time      `json:"-"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

type ArticleResponse struct {
	Title           string         `json:"title"`
	Description     string         `json:"description"`
	URL             string         `json:"url"`
	PublicationDate time.Time      `json:"publication_date"`
	SourceName      string         `json:"source_name"`
	Category        pq.StringArray `json:"category"` // Changed from []string
	RelevanceScore  float64        `json:"relevance_score"`
	LLMSummary      string         `json:"llm_summary"`
	Latitude        float64        `json:"latitude"`
	Longitude       float64        `json:"longitude"`
}

type QueryIntent struct {
	Intent   string   `json:"intent"`
	Entities []string `json:"entities"`
	Concepts []string `json:"concepts"`
}

type UserEvent struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ArticleID string    `gorm:"index:idx_article" json:"article_id"`
	EventType string    `gorm:"index:idx_event_type" json:"event_type"` // view, click, share
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Timestamp time.Time `gorm:"index:idx_timestamp" json:"timestamp"`
	CreatedAt time.Time `json:"-"`
}

type TrendingArticle struct {
	Article
	TrendingScore float64 `json:"trending_score"`
}
