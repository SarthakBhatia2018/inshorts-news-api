package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"

	"inshorts-news-api/config"
)

func main() {
	cfg := config.Load()

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("Loading news data from JSON file...")

	data, err := ioutil.ReadFile("/Users/sarthakbhatia/Developer/inshorts-news-api/scripts/news_data.json")
	if err != nil {
		log.Fatal("Failed to read JSON file:", err)
	}

	var rawArticles []map[string]interface{}
	if err := json.Unmarshal(data, &rawArticles); err != nil {
		log.Fatal("Failed to parse JSON:", err)
	}

	log.Printf("Inserting %d articles into database...", len(rawArticles))

	// Prepare statement
	stmt, err := db.Prepare(`
        INSERT INTO articles (
            id, title, description, url, publication_date, 
            source_name, category, relevance_score, latitude, longitude,
            created_at, updated_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
        ) ON CONFLICT (id) DO NOTHING
    `)
	if err != nil {
		log.Fatal("Failed to prepare statement:", err)
	}
	defer stmt.Close()

	// Insert articles
	successCount := 0
	for _, raw := range rawArticles {
		pubDate, _ := time.Parse(time.RFC3339, raw["publication_date"].(string))

		// Handle category array properly
		categories := []string{}
		if catArray, ok := raw["category"].([]interface{}); ok {
			for _, cat := range catArray {
				if catStr, ok := cat.(string); ok {
					categories = append(categories, catStr)
				}
			}
		}

		if len(categories) == 0 {
			categories = []string{"General"}
		}

		now := time.Now()

		_, err := stmt.Exec(
			raw["id"].(string),
			raw["title"].(string),
			raw["description"].(string),
			raw["url"].(string),
			pubDate,
			raw["source_name"].(string),
			pq.Array(categories), // Use pq.Array for proper PostgreSQL array handling
			raw["relevance_score"].(float64),
			raw["latitude"].(float64),
			raw["longitude"].(float64),
			now,
			now,
		)

		if err != nil {
			log.Printf("Warning: Failed to insert article %s: %v", raw["id"], err)
		} else {
			successCount++
		}
	}

	log.Printf("Successfully inserted %d articles!", successCount)

	// Generate sample events
	log.Println("Generating sample user events...")
	generateEvents(db, rawArticles)

	log.Println("All done!")
}

func generateEvents(db *sql.DB, articles []map[string]interface{}) {
	stmt, err := db.Prepare(`
        INSERT INTO user_events (article_id, event_type, latitude, longitude, timestamp, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `)
	if err != nil {
		log.Printf("Failed to prepare event statement: %v", err)
		return
	}
	defer stmt.Close()

	eventTypes := []string{"view", "click", "share"}
	successCount := 0

	for i := 0; i < 1000 && i < len(articles)*2; i++ {
		article := articles[i%len(articles)]
		eventType := eventTypes[i%len(eventTypes)]

		lat := article["latitude"].(float64) + (float64(i%10)-5)*0.01
		lon := article["longitude"].(float64) + (float64(i%10)-5)*0.01
		eventTime := time.Now().Add(-time.Duration(i) * time.Minute)

		_, err := stmt.Exec(
			article["id"].(string),
			eventType,
			lat,
			lon,
			eventTime,
			time.Now(),
		)

		if err == nil {
			successCount++
		}
	}

	log.Printf("Generated %d sample events", successCount)
}
