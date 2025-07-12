package mcp

import (
	"encoding/json"
	"fmt"
	"time"
)

// MCP Protocol Version
const (
	ProtocolVersion = "2024-11-05"
)

// Message types for MCP protocol
const (
	MessageTypeInitialize     = "initialize"
	MessageTypeInitialization = "initialization"
	MessageTypePing           = "ping"
	MessageTypePong           = "pong"
	MessageTypeListResources  = "listResources"
	MessageTypeReadResource   = "readResource"
	MessageTypeListTools      = "listTools"
	MessageTypeCallTool       = "callTool"
	MessageTypeError          = "error"
)

// Message represents an MCP protocol message
type Message struct {
	Type      string          `json:"type"`
	ID        string          `json:"id,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data,omitempty"`
}

// InitializeRequest represents an initialization request from client
type InitializeRequest struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ClientCapabilities `json:"capabilities"`
	ClientInfo      ClientInfo         `json:"clientInfo"`
}

// ClientCapabilities represents client capabilities
type ClientCapabilities struct {
	Resources ResourceCapabilities `json:"resources"`
	Tools     ToolCapabilities     `json:"tools"`
}

// ResourceCapabilities represents resource-related capabilities
type ResourceCapabilities struct {
	Subscribe bool `json:"subscribe"`
}

// ToolCapabilities represents tool-related capabilities
type ToolCapabilities struct {
	Call bool `json:"call"`
}

// ClientInfo represents client information
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// InitializationResponse represents server initialization response
type InitializationResponse struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
}

// ServerCapabilities represents server capabilities
type ServerCapabilities struct {
	Resources ResourceCapabilities `json:"resources"`
	Tools     ToolCapabilities     `json:"tools"`
}

// ServerInfo represents server information
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Resource represents a resource that can be accessed
type Resource struct {
	URI         string            `json:"uri"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	MimeType    string            `json:"mimeType"`
	Content     json.RawMessage   `json:"content,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// Tool represents a tool that can be called
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// ToolCall represents a tool call request
type ToolCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolResult represents the result of a tool call
type ToolResult struct {
	Content []ToolResultContent `json:"content"`
}

// ToolResultContent represents content in a tool result
type ToolResultContent struct {
	Type     string           `json:"type"`
	Text     string           `json:"text,omitempty"`
	Image    *ImageContent    `json:"image,omitempty"`
	Embedded *EmbeddedContent `json:"embedded,omitempty"`
}

// ImageContent represents image content
type ImageContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType"`
	Width    int    `json:"width,omitempty"`
	Height   int    `json:"height,omitempty"`
}

// EmbeddedContent represents embedded content
type EmbeddedContent struct {
	MimeType string          `json:"mimeType"`
	Data     json.RawMessage `json:"data"`
}

// Error represents an MCP protocol error
type Error struct {
	Type    string      `json:"type"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// NewMessage creates a new MCP message
func NewMessage(msgType, id string, data interface{}) (*Message, error) {
	var rawData json.RawMessage
	var err error

	if data != nil {
		rawData, err = json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal message data: %w", err)
		}
	}

	return &Message{
		Type:      msgType,
		ID:        id,
		Timestamp: time.Now(),
		Data:      rawData,
	}, nil
}

// UnmarshalData unmarshals message data into the provided interface
func (m *Message) UnmarshalData(v interface{}) error {
	if m.Data == nil {
		return nil
	}
	return json.Unmarshal(m.Data, v)
}

// NewError creates a new error message
func NewError(errorType, message string, data interface{}) *Error {
	return &Error{
		Type:    errorType,
		Message: message,
		Data:    data,
	}
}
