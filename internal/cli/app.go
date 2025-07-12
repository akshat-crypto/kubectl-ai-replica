package cli

import (
	"fmt"

	"github.com/mcp-servers/cli/internal/commands"
	"github.com/mcp-servers/cli/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// App represents the main CLI application
type App struct {
	rootCmd *cobra.Command
	config  *config.Config
}

// NewApp creates a new CLI application
func NewApp(version, commit, date string) *App {
	app := &App{}
	app.setupRootCommand(version, commit, date)
	app.setupConfig()
	app.setupCommands()
	return app
}

// Execute runs the CLI application
func (a *App) Execute() error {
	return a.rootCmd.Execute()
}

// setupRootCommand configures the root command
func (a *App) setupRootCommand(version, commit, date string) {
	a.rootCmd = &cobra.Command{
		Use:     "mcp-cli",
		Short:   "MCP (Model Context Protocol) CLI tool for managing MCP servers",
		Long:    `A production-grade CLI tool for interacting with MCP servers, managing connections, and executing operations.`,
		Version: fmt.Sprintf("%s (commit: %s, date: %s)", version, commit, date),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return a.loadConfig()
		},
	}

	// Global flags
	a.rootCmd.PersistentFlags().StringP("config", "c", "", "config file (default is ./configs/config.yaml)")
	a.rootCmd.PersistentFlags().StringP("log-level", "l", "info", "log level (debug, info, warn, error)")
	a.rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")

	// Bind flags to viper
	viper.BindPFlag("config", a.rootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("log_level", a.rootCmd.PersistentFlags().Lookup("log-level"))
	viper.BindPFlag("verbose", a.rootCmd.PersistentFlags().Lookup("verbose"))
}

// setupConfig initializes configuration
func (a *App) setupConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
}

// setupCommands adds all subcommands
func (a *App) setupCommands() {
	// Initialize config if nil
	if a.config == nil {
		a.config = config.DefaultConfig()
	}
	
	// Server commands
	a.rootCmd.AddCommand(commands.NewServersCommand(a.config))
	
	// Connection commands
	a.rootCmd.AddCommand(commands.NewConnectCommand(a.config))
	
	// Config commands
	a.rootCmd.AddCommand(commands.NewConfigCommand(a.config))
	
	// Health commands
	a.rootCmd.AddCommand(commands.NewHealthCommand(a.config))
}

// loadConfig loads the configuration file
func (a *App) loadConfig() error {
	configFile := viper.GetString("config")
	if configFile != "" {
		viper.SetConfigFile(configFile)
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found, use defaults
		logrus.Warn("No config file found, using default configuration")
	}

	// Set log level
	logLevel := viper.GetString("log_level")
	if level, err := logrus.ParseLevel(logLevel); err == nil {
		logrus.SetLevel(level)
	}

	// Load configuration into struct
	a.config = &config.Config{}
	if err := viper.Unmarshal(a.config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}
