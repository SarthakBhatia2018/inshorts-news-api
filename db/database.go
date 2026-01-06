package db

import (
    "fmt"
    "log"
    
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
    
    "inshorts-news-api/config"
    "inshorts-news-api/models"
)

var DB *gorm.DB

func Connect(cfg *config.Config) error {
    dsn := fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
    )

    var err error
    DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    
    if err != nil {
        return fmt.Errorf("failed to connect to database: %w", err)
    }

    log.Println("Database connection established")
    return nil
}

func AutoMigrate() error {
    return DB.AutoMigrate(
        &models.Article{},
        &models.UserEvent{},
    )
}

func GetDB() *gorm.DB {
    return DB
}
