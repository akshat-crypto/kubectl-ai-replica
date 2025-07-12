package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenRouterProvider implements the Provider interface for OpenRouter
type OpenRouterProvider struct {
	client  *http.Client
	config  Config
	apiKey  string
	baseURL string
}

// OpenRouterRequest represents the request payload for OpenRouter API
type OpenRouterRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	Tools       []Tool    `json:"tools,omitempty"`
	ToolChoice  string    `json:"tool_choice,omitempty"`
}

// OpenRouterResponse represents the response from OpenRouter API
type OpenRouterResponse struct {
	Choices []struct {
		Message struct {
			Content   string     `json:"content"`
			ToolCalls []ToolCall `json:"tool_calls,omitempty"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

// NewOpenRouterProvider creates a new OpenRouter provider
func NewOpenRouterProvider(config Config) (Provider, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("OpenRouter API key is required")
	}

	model := config.Model
	if model == "" {
		model = "openai/gpt-3.5-turbo"
	}

	return &OpenRouterProvider{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		config:  config,
		apiKey:  config.APIKey,
		baseURL: "https://openrouter.ai/api/v1",
	}, nil
}

// GenerateResponse generates a response using OpenRouter
func (p *OpenRouterProvider) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	payload := OpenRouterRequest{
		Model: p.config.Model,
		Messages: []Message{
			{
				Role:    "system",
				Content: "You are a Kubernetes expert. Provide clear, actionable responses and commands.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   p.config.MaxTokens,
		Temperature: p.config.Temperature,
	}

	return p.makeRequest(ctx, payload)
}

// GetModel returns the current model name
func (p *OpenRouterProvider) GetModel() string {
	return p.config.Model
}

// GetProvider returns the provider name
func (p *OpenRouterProvider) GetProvider() string {
	return "openrouter"
}

// GenerateResponseWithTools generates a response with tool calls
func (p *OpenRouterProvider) GenerateResponseWithTools(ctx context.Context, query Query) (*Response, error) {
	// Build system message with tools
	systemMessage := "You are a Kubernetes assistant. You can use the following tools to help users:\n"
	for _, tool := range query.Tools {
		systemMessage += fmt.Sprintf("- %s: %s\n", tool.Name, tool.Description)
	}

	// Build messages
	messages := []Message{
		{
			Role:    "system",
			Content: systemMessage,
		},
	}

	// Add conversation history
	for _, msg := range query.History {
		messages = append(messages, Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Add current query
	messages = append(messages, Message{
		Role:    "user",
		Content: query.Text,
	})

	// Create request payload
	payload := OpenRouterRequest{
		Model:       p.config.Model,
		Messages:    messages,
		MaxTokens:   p.config.MaxTokens,
		Temperature: p.config.Temperature,
	}

	// Add tools if available
	if len(query.Tools) > 0 {
		payload.Tools = query.Tools
		payload.ToolChoice = "auto"
	}

	// Make the request
	content, err := p.makeRequest(ctx, payload)
	if err != nil {
		return nil, err
	}

	response := &Response{
		Content: content,
	}

	// Parse tool calls from response (OpenRouter supports function calling)
	// Note: This is a simplified implementation. In production, you'd want to
	// properly parse the tool calls from the response.
	response.ToolCalls = p.parseToolCallsFromResponse(content)

	return response, nil
}

// makeRequest makes an HTTP request to the OpenRouter API
func (p *OpenRouterProvider) makeRequest(ctx context.Context, payload OpenRouterRequest) (string, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://mcp-servers-cli")
	req.Header.Set("X-Title", "MCP Servers CLI")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result OpenRouterResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("OpenRouter API error: %s", result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from AI model")
	}

	content := strings.TrimSpace(result.Choices[0].Message.Content)
	if content == "" {
		return "", fmt.Errorf("empty response from AI model")
	}

	return content, nil
}

// parseToolCallsFromResponse parses tool calls from OpenRouter response
func (p *OpenRouterProvider) parseToolCallsFromResponse(content string) []ToolCall {
	// This is a simplified parser - in production, you'd want more sophisticated parsing
	var toolCalls []ToolCall

	// Look for patterns like "TOOL: kubectl get pods" or "EXECUTE: kubectl scale deployment"
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "TOOL:") || strings.HasPrefix(line, "EXECUTE:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				command := strings.TrimSpace(parts[1])
				toolCalls = append(toolCalls, ToolCall{
					ToolName: "kubectl",
					Arguments: map[string]interface{}{
						"command": command,
					},
				})
			}
		}
	}

	return toolCalls
}
