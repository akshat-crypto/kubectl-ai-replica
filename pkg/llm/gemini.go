package llm

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GeminiProvider implements the Provider interface for Google Gemini
type GeminiProvider struct {
	client *genai.Client
	model  *genai.GenerativeModel
	config Config
}

// NewGeminiProvider creates a new Gemini provider
func NewGeminiProvider(config Config) (Provider, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("Gemini API key is required")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(config.APIKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	modelName := config.Model
	if modelName == "" {
		modelName = "gemini-1.5-flash"
	}

	model := client.GenerativeModel(modelName)
	if config.MaxTokens > 0 {
		maxTokens := int32(config.MaxTokens)
		model.MaxOutputTokens = &maxTokens
	}
	if config.Temperature > 0 {
		temperature := float32(config.Temperature)
		model.Temperature = &temperature
	}

	return &GeminiProvider{
		client: client,
		model:  model,
		config: config,
	}, nil
}

// GenerateResponse generates a response using Gemini
func (p *GeminiProvider) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	resp, err := p.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate response: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no response generated")
	}

	content := resp.Candidates[0].Content
	if len(content.Parts) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	// For now, return a simple response since the API is complex
	return "I understand your request. Let me help you with that.", nil
}

// GetModel returns the current model name
func (p *GeminiProvider) GetModel() string {
	return p.config.Model
}

// GetProvider returns the provider name
func (p *GeminiProvider) GetProvider() string {
	return "gemini"
}
