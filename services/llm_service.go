package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"

	"inshorts-news-api/models"
)

type LLMService struct {
	client *openai.Client
}

func NewLLMService(apiKey string) *LLMService {
	if apiKey == "" {
		return &LLMService{client: nil}
	}
	return &LLMService{
		client: openai.NewClient(apiKey),
	}
}

func (s *LLMService) AnalyzeQuery(query string, userLocation string) (*models.QueryIntent, error) {
	if s.client == nil {
		// Fallback: simple keyword-based intent detection
		return s.fallbackAnalyzeQuery(query), nil
	}

	prompt := fmt.Sprintf(`Analyze the following news query and extract:
1. Intent: Choose ONE from [category, source, search, nearby, score]
2. Entities: Key people, organizations, locations, events
3. Concepts: Main topics or themes

Query: "%s"
User Location Context: %s

Return JSON only in this exact format:
{
  "intent": "<one of: category, source, search, nearby, score>",
  "entities": ["entity1", "entity2"],
  "concepts": ["concept1", "concept2"]
}`, query, userLocation)

	resp, err := s.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a query analysis assistant. Return only valid JSON.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: 0.3,
		},
	)

	if err != nil {
		fmt.Printf("LLM Error: %v\n", err)
		// Fallback to keyword analysis on error
		return s.fallbackAnalyzeQuery(query), nil
	}

	// Check if response has choices
	if len(resp.Choices) == 0 {
		return s.fallbackAnalyzeQuery(query), nil
	}

	var intent models.QueryIntent
	content := resp.Choices[0].Message.Content

	// Clean up response
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	if err := json.Unmarshal([]byte(content), &intent); err != nil {
		// Fallback on parse error
		return s.fallbackAnalyzeQuery(query), nil
	}

	return &intent, nil
}

func (s *LLMService) fallbackAnalyzeQuery(query string) *models.QueryIntent {
	lower := strings.ToLower(query)
	intent := &models.QueryIntent{
		Intent:   "search",
		Entities: []string{},
		Concepts: []string{},
	}

	// Simple keyword-based detection
	if strings.Contains(lower, "near") || strings.Contains(lower, "nearby") || strings.Contains(lower, "around") {
		intent.Intent = "nearby"
	} else if strings.Contains(lower, "category") || strings.Contains(lower, "technology") ||
		strings.Contains(lower, "business") || strings.Contains(lower, "sports") {
		intent.Intent = "category"
	} else if strings.Contains(lower, "from") && (strings.Contains(lower, "times") || strings.Contains(lower, "reuters")) {
		intent.Intent = "source"
	} else if strings.Contains(lower, "relevant") || strings.Contains(lower, "important") {
		intent.Intent = "score"
	}

	// Extract potential entities (simple approach)
	words := strings.Fields(query)
	for _, word := range words {
		if len(word) > 3 && strings.Title(word) == word {
			intent.Entities = append(intent.Entities, word)
		}
	}

	return intent
}

func (s *LLMService) GenerateSummary(title, description string) (string, error) {
	if s.client == nil {
		// Fallback: return truncated description
		if len(description) > 150 {
			return description[:150] + "...", nil
		}
		return description, nil
	}

	prompt := fmt.Sprintf(`Summarize this news article in 2-3 sentences:

Title: %s
Description: %s

Provide a concise, informative summary.`, title, description)

	resp, err := s.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: 0.5,
			MaxTokens:   150,
		},
	)

	if err != nil {
		// Fallback on error
		if len(description) > 150 {
			return description[:150] + "...", nil
		}
		return description, nil
	}

	// Check if response has choices
	if len(resp.Choices) == 0 {
		if len(description) > 150 {
			return description[:150] + "...", nil
		}
		return description, nil
	}

	summary := strings.TrimSpace(resp.Choices[0].Message.Content)
	if summary == "" {
		if len(description) > 150 {
			return description[:150] + "...", nil
		}
		return description, nil
	}

	return summary, nil
}

func (s *LLMService) BatchGenerateSummaries(articles []models.Article) (map[string]string, error) {
	summaries := make(map[string]string)

	for _, article := range articles {
		summary, err := s.GenerateSummary(article.Title, article.Description)
		if err != nil {
			// Use fallback on error
			summary = article.Description
			if len(summary) > 150 {
				summary = summary[:150] + "..."
			}
		}
		summaries[article.ID] = summary
	}

	return summaries, nil
}
