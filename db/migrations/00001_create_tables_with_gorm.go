package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"inshorts-news-api/models"
)

func init() {
	goose.AddMigrationContext(upCreateArticlesUserEventsTables, downCreateArticlesUserEventsTables)
}

func upCreateArticlesUserEventsTables(ctx context.Context, tx *sql.Tx) error {
	// Wrap the *sql.Tx in a GORM DB instance
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: tx,
	}), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to create gorm instance: %w", err)
	}

	// Auto-migrate using GORM models
	if err := gormDB.WithContext(ctx).AutoMigrate(
		&models.Article{},
		&models.UserEvent{},
	); err != nil {
		return fmt.Errorf("failed to auto-migrate: %w", err)
	}

	// Create custom indexes that GORM doesn't handle
	customIndexes := []string{
		// Full-text search indexes
		`CREATE INDEX IF NOT EXISTS idx_articles_title_fts 
         ON articles USING gin(to_tsvector('english', title))`,

		`CREATE INDEX IF NOT EXISTS idx_articles_description_fts 
         ON articles USING gin(to_tsvector('english', description))`,

		// Composite index for common queries
		`CREATE INDEX IF NOT EXISTS idx_articles_category_date 
         ON articles USING gin(category) WHERE deleted_at IS NULL`,
	}

	for i, query := range customIndexes {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("failed to create custom index %d: %w", i+1, err)
		}
	}

	return nil
}

func downCreateArticlesUserEventsTables(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, `DROP TABLE IF EXISTS user_events CASCADE`); err != nil {
		return fmt.Errorf("failed to drop user_events: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `DROP TABLE IF EXISTS articles CASCADE`); err != nil {
		return fmt.Errorf("failed to drop articles: %w", err)
	}

	return nil
}
