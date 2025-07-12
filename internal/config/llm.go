package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mcp-servers/cli/pkg/llm"
	"gopkg.in/yaml.v3"
)

// LLMConfig holds the complete LLM configuration
type LLMConfig struct {
	// LLM provider configuration
	Provider      string  `yaml:"provider" json:"provider"`
	Model         string  `yaml:"model" json:"model"`
	APIKey        string  `yaml:"api_key" json:"api_key"`
	MaxTokens     int     `yaml:"max_tokens" json:"max_tokens"`
	Temperature   float64 `yaml:"temperature" json:"temperature"`
	SkipVerifySSL bool    `yaml:"skip_verify_ssl" json:"skip_verify_ssl"`

	// Tool configuration
	CustomToolsConfig []string `yaml:"custom_tools_config" json:"custom_tools_config"`
	SkipPermissions   bool     `yaml:"skip_permissions" json:"skip_permissions"`
	EnableToolUseShim bool     `yaml:"enable_tool_use_shim" json:"enable_tool_use_shim"`

	// MCP configuration
	MCPServer     bool `yaml:"mcp_server" json:"mcp_server"`
	MCPClient     bool `yaml:"mcp_client" json:"mcp_client"`
	ExternalTools bool `yaml:"external_tools" json:"external_tools"`

	// Runtime settings
	MaxIterations int  `yaml:"max_iterations" json:"max_iterations"`
	Quiet         bool `yaml:"quiet" json:"quiet"`
	RemoveWorkdir bool `yaml:"remove_workdir" json:"remove_workdir"`

	// Kubernetes configuration
	Kubeconfig string `yaml:"kubeconfig" json:"kubeconfig"`

	// UI configuration
	UserInterface   string `yaml:"user_interface" json:"user_interface"`
	UIListenAddress string `yaml:"ui_listen_address" json:"ui_listen_address"`

	// Prompt configuration
	PromptTemplateFilePath string   `yaml:"prompt_template_file_path" json:"prompt_template_file_path"`
	ExtraPromptPaths       []string `yaml:"extra_prompt_paths" json:"extra_prompt_paths"`

	// Debug and trace settings
	TracePath string `yaml:"trace_path" json:"trace_path"`
}

// DefaultLLMConfig returns default configuration
func DefaultLLMConfig() *LLMConfig {
	return &LLMConfig{
		Provider:               "gemini",
		Model:                  "gemini-1.5-flash",
		MaxTokens:              2048,
		Temperature:            0.7,
		SkipVerifySSL:          false,
		CustomToolsConfig:      []string{"~/.config/mcp-servers/tools.yaml"},
		SkipPermissions:        false,
		EnableToolUseShim:      false,
		MCPServer:              false,
		MCPClient:              false,
		ExternalTools:          false,
		MaxIterations:          20,
		Quiet:                  false,
		RemoveWorkdir:          false,
		Kubeconfig:             "~/.kube/config",
		UserInterface:          "terminal",
		UIListenAddress:        "localhost:8888",
		PromptTemplateFilePath: "",
		ExtraPromptPaths:       []string{},
		TracePath:              "/tmp/mcp-servers-trace.txt",
	}
}

// LoadLLMConfig loads LLM configuration from file and environment
func LoadLLMConfig(configPath string) (*LLMConfig, error) {
	config := DefaultLLMConfig()

	// Load from config file if provided
	if configPath != "" {
		if err := loadConfigFromFile(config, configPath); err != nil {
			return nil, fmt.Errorf("failed to load config from file: %w", err)
		}
	}

	// Load from default config location
	defaultConfigPath := getDefaultConfigPath()
	if err := loadConfigFromFile(config, defaultConfigPath); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load default config: %w", err)
	}

	// Override with environment variables
	loadConfigFromEnv(config)

	// Validate configuration
	if err := validateLLMConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// loadConfigFromFile loads configuration from a YAML file
func loadConfigFromFile(config *LLMConfig, configPath string) error {
	// Expand home directory
	if strings.HasPrefix(configPath, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		configPath = filepath.Join(home, configPath[2:])
	}

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return err
	}

	// Read and parse YAML
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}

// loadConfigFromEnv loads configuration from environment variables
func loadConfigFromEnv(config *LLMConfig) {
	// LLM provider settings
	if provider := os.Getenv("LLM_PROVIDER"); provider != "" {
		config.Provider = provider
	}
	if model := os.Getenv("LLM_MODEL"); model != "" {
		config.Model = model
	}
	if apiKey := os.Getenv("LLM_API_KEY"); apiKey != "" {
		config.APIKey = apiKey
	}
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" && config.Provider == "openai" {
		config.APIKey = apiKey
	}
	if apiKey := os.Getenv("GEMINI_API_KEY"); apiKey != "" && config.Provider == "gemini" {
		config.APIKey = apiKey
	}
	if apiKey := os.Getenv("OPENROUTER_API_KEY"); apiKey != "" && config.Provider == "openrouter" {
		config.APIKey = apiKey
	}

	// MCP settings
	if mcpServer := os.Getenv("MCP_SERVER"); mcpServer != "" {
		config.MCPServer = mcpServer == "true"
	}
	if mcpClient := os.Getenv("MCP_CLIENT"); mcpClient != "" {
		config.MCPClient = mcpClient == "true"
	}

	// Kubernetes settings
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		config.Kubeconfig = kubeconfig
	}
}

// validateLLMConfig validates the configuration
func validateLLMConfig(config *LLMConfig) error {
	// Validate provider
	switch config.Provider {
	case "openai", "gemini", "openrouter":
		// Valid providers
	default:
		return fmt.Errorf("unsupported LLM provider: %s", config.Provider)
	}

	// Validate API key
	if config.APIKey == "" {
		return fmt.Errorf("API key is required for provider: %s", config.Provider)
	}

	// Validate model
	if config.Model == "" {
		return fmt.Errorf("model is required")
	}

	// Validate temperature
	if config.Temperature < 0 || config.Temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2")
	}

	// Validate max tokens
	if config.MaxTokens <= 0 {
		return fmt.Errorf("max_tokens must be positive")
	}

	return nil
}

// getDefaultConfigPath returns the default configuration file path
func getDefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "~/.config/mcp-servers/config.yaml"
	}
	return filepath.Join(home, ".config", "mcp-servers", "config.yaml")
}

// CreateLLMProvider creates an LLM provider from configuration
func (c *LLMConfig) CreateLLMProvider() (llm.Provider, error) {
	llmConfig := llm.Config{
		Provider:      c.Provider,
		Model:         c.Model,
		APIKey:        c.APIKey,
		MaxTokens:     c.MaxTokens,
		Temperature:   c.Temperature,
		SkipVerifySSL: c.SkipVerifySSL,
	}

	return llm.NewProvider(llmConfig)
}

// SaveConfig saves the configuration to file
func (c *LLMConfig) SaveConfig(configPath string) error {
	// Expand home directory
	if strings.HasPrefix(configPath, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		configPath = filepath.Join(home, configPath[2:])
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
