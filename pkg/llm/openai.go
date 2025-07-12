package llm

import (
	"context"
	"encoding/json"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

// OpenAIProvider implements the Provider interface for OpenAI
type OpenAIProvider struct {
	client *openai.Client
	config Config
	model  string
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(config Config) (Provider, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	clientConfig := openai.DefaultConfig(config.APIKey)
	if config.SkipVerifySSL {
		// Note: SkipVerifySSL is not supported in this version
		// In production, you'd want to handle this properly
	}

	client := openai.NewClientWithConfig(clientConfig)

	model := config.Model
	if model == "" {
		model = openai.GPT4
	}

	return &OpenAIProvider{
		client: client,
		config: config,
		model:  model,
	}, nil
}

// GenerateResponse generates a response using OpenAI
func (p *OpenAIProvider) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	resp, err := p.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: p.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			MaxTokens:   p.config.MaxTokens,
			Temperature: float32(p.config.Temperature),
		},
	)

	if err != nil {
		return "", fmt.Errorf("failed to generate response: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response generated")
	}

	return resp.Choices[0].Message.Content, nil
}

// GetModel returns the current model name
func (p *OpenAIProvider) GetModel() string {
	return p.model
}

// GetProvider returns the provider name
func (p *OpenAIProvider) GetProvider() string {
	return "openai"
}

// GenerateResponseWithTools generates a response with tool calls
func (p *OpenAIProvider) GenerateResponseWithTools(ctx context.Context, query Query) (*Response, error) {
	// Build system message with tools
	systemMessage := "You are a Kubernetes assistant. You can use the following tools to help users:"
	for _, tool := range query.Tools {
		systemMessage += fmt.Sprintf("\n- %s: %s", tool.Name, tool.Description)
	}

	// Build messages
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemMessage,
		},
	}

	// Add conversation history
	for _, msg := range query.History {
		role := openai.ChatMessageRoleUser
		if msg.Role == "assistant" {
			role = openai.ChatMessageRoleAssistant
		}
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    role,
			Content: msg.Content,
		})
	}

	// Add current query
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: query.Text,
	})

	// Create completion request
	req := openai.ChatCompletionRequest{
		Model:       p.model,
		Messages:    messages,
		MaxTokens:   p.config.MaxTokens,
		Temperature: float32(p.config.Temperature),
	}

	// Add tools if available
	if len(query.Tools) > 0 {
		tools := make([]openai.Tool, len(query.Tools))
		for i, tool := range query.Tools {
			tools[i] = openai.Tool{
				Type: openai.ToolTypeFunction,
				Function: &openai.FunctionDefinition{
					Name:        tool.Name,
					Description: tool.Description,
					Parameters:  tool.Parameters,
				},
			}
		}
		req.Tools = tools
		req.ToolChoice = "auto"
	}

	resp, err := p.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response generated")
	}

	choice := resp.Choices[0]
	response := &Response{
		Content: choice.Message.Content,
	}

	// Extract tool calls
	if choice.Message.ToolCalls != nil {
		response.ToolCalls = make([]ToolCall, len(choice.Message.ToolCalls))
		for i, toolCall := range choice.Message.ToolCalls {
			response.ToolCalls[i] = ToolCall{
				ToolName:  toolCall.Function.Name,
				Arguments: parseJSONArguments(toolCall.Function.Arguments),
			}
		}
	}

	return response, nil
}

// parseJSONArguments parses JSON arguments string to map
func parseJSONArguments(args string) map[string]interface{} {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(args), &result); err != nil {
		return map[string]interface{}{"raw": args}
	}
	return result
}
