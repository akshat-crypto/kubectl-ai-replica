package llm

import (
	"context"
	"fmt"
)

// Provider represents an LLM provider interface
type Provider interface {
	// GenerateResponse generates a response based on the input prompt
	GenerateResponse(ctx context.Context, prompt string) (string, error)

	// GetModel returns the current model name
	GetModel() string

	// GetProvider returns the provider name
	GetProvider() string
}

// Config holds LLM configuration
type Config struct {
	Provider      string  `yaml:"provider" json:"provider"`
	Model         string  `yaml:"model" json:"model"`
	APIKey        string  `yaml:"api_key" json:"api_key"`
	MaxTokens     int     `yaml:"max_tokens" json:"max_tokens"`
	Temperature   float64 `yaml:"temperature" json:"temperature"`
	SkipVerifySSL bool    `yaml:"skip_verify_ssl" json:"skip_verify_ssl"`
}

// NewProvider creates a new LLM provider based on configuration
func NewProvider(config Config) (Provider, error) {
	switch config.Provider {
	case "openai":
		return NewOpenAIProvider(config)
	case "gemini":
		return NewGeminiProvider(config)
	case "openrouter":
		return NewOpenRouterProvider(config)
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", config.Provider)
	}
}

// Query represents a natural language query with context
type Query struct {
	Text    string                 `json:"text"`
	Context map[string]interface{} `json:"context,omitempty"`
	Tools   []Tool                 `json:"tools,omitempty"`
	History []Message              `json:"history,omitempty"`
}

// Tool represents an available tool for the LLM
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// Message represents a conversation message
type Message struct {
	Role    string `json:"role"` // "user", "assistant", "system"
	Content string `json:"content"`
}

// Response represents the LLM response
type Response struct {
	Content    string                 `json:"content"`
	ToolCalls  []ToolCall             `json:"tool_calls,omitempty"`
	Confidence float64                `json:"confidence,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ToolCall represents a tool invocation request
type ToolCall struct {
	ToolName  string                 `json:"tool_name"`
	Arguments map[string]interface{} `json:"arguments"`
}
