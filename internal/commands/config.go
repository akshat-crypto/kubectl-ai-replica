package commands

import (
	"fmt"
	"os"

	"github.com/mcp-servers/cli/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// NewConfigCommand creates the config command
func NewConfigCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Long:  `View, edit, and manage MCP CLI configuration settings.`,
	}

	cmd.AddCommand(
		newConfigShowCommand(cfg),
		newConfigInitCommand(),
		newConfigValidateCommand(cfg),
	)

	return cmd
}

// newConfigShowCommand creates the show subcommand
func newConfigShowCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return showConfig(cfg)
		},
	}
}

// newConfigInitCommand creates the init subcommand
func newConfigInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize default configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return initConfig()
		},
	}
}

// newConfigValidateCommand creates the validate subcommand
func newConfigValidateCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return validateConfig(cfg)
		},
	}
}

// showConfig displays the current configuration
func showConfig(cfg *config.Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

// initConfig creates a default configuration file
func initConfig() error {
	cfg := config.DefaultConfig()

	// Ensure configs directory exists
	if err := os.MkdirAll("configs", 0755); err != nil {
		return fmt.Errorf("failed to create configs directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile("configs/config.yaml", data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	logrus.Info("Configuration initialized at configs/config.yaml")
	return nil
}

// validateConfig validates the current configuration
func validateConfig(cfg *config.Config) error {
	errors := []string{}

	// Validate app settings
	if cfg.App.Name == "" {
		errors = append(errors, "app.name is required")
	}
	if cfg.App.Timeout <= 0 {
		errors = append(errors, "app.timeout must be positive")
	}

	// Validate server configurations
	for name, server := range cfg.Servers {
		if server.Host == "" {
			errors = append(errors, fmt.Sprintf("server %s: host is required", name))
		}
		if server.Port <= 0 || server.Port > 65535 {
			errors = append(errors, fmt.Sprintf("server %s: port must be between 1 and 65535", name))
		}
		if server.Protocol == "" {
			errors = append(errors, fmt.Sprintf("server %s: protocol is required", name))
		}
	}

	// Validate security settings
	if cfg.Security.JWT.Secret == "" {
		errors = append(errors, "security.jwt.secret is required")
	}

	// Validate logging settings
	if cfg.Logging.Level == "" {
		errors = append(errors, "logging.level is required")
	}

	if len(errors) > 0 {
		fmt.Println("Configuration validation failed:")
		for _, err := range errors {
			fmt.Printf("  - %s\n", err)
		}
		return fmt.Errorf("configuration validation failed")
	}

	logrus.Info("Configuration is valid")
	return nil
}
