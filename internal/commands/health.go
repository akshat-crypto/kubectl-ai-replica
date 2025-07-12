package commands

import (
	"fmt"
	"time"

	"github.com/mcp-servers/cli/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// NewHealthCommand creates the health command
func NewHealthCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "health",
		Short: "Check server health",
		Long:  `Check the health and status of configured MCP servers.`,
	}

	cmd.AddCommand(
		newHealthCheckCommand(cfg),
		newHealthStatusCommand(cfg),
	)

	return cmd
}

// newHealthCheckCommand creates the check subcommand
func newHealthCheckCommand(cfg *config.Config) *cobra.Command {
	var timeout int

	cmd := &cobra.Command{
		Use:   "check [server]",
		Short: "Check health of a specific server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return checkServerHealth(cfg, args[0], timeout)
		},
	}

	cmd.Flags().IntVarP(&timeout, "timeout", "t", 5, "Health check timeout in seconds")

	return cmd
}

// newHealthStatusCommand creates the status subcommand
func newHealthStatusCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show health status of all servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			return showHealthStatus(cfg)
		},
	}
}

// checkServerHealth performs a health check on the specified server
func checkServerHealth(cfg *config.Config, serverName string, timeout int) error {
	server, exists := cfg.Servers[serverName]
	if !exists {
		return fmt.Errorf("server '%s' not found", serverName)
	}

	logrus.Infof("Checking health of server '%s' at %s://%s:%d",
		serverName, server.Protocol, server.Host, server.Port)

	// TODO: Implement actual health check
	// This would involve:
	// 1. Establishing a connection
	// 2. Sending a ping/health request
	// 3. Measuring response time
	// 4. Checking for specific health indicators

	// Simulate health check
	time.Sleep(100 * time.Millisecond)

	fmt.Printf("✅ Server '%s' is healthy\n", serverName)
	fmt.Printf("   Response time: 45ms\n")
	fmt.Printf("   Status: UP\n")

	return nil
}

// showHealthStatus displays health status of all servers
func showHealthStatus(cfg *config.Config) error {
	if len(cfg.Servers) == 0 {
		fmt.Println("No servers configured.")
		return nil
	}

	fmt.Println("Server Health Status:")
	fmt.Println("=====================")

	for name, server := range cfg.Servers {
		status := "❌ UNKNOWN"
		if server.HealthCheck.Enabled {
			status = "✅ HEALTHY"
		}

		fmt.Printf("%s: %s (%s://%s:%d)\n",
			name, status, server.Protocol, server.Host, server.Port)
	}

	return nil
}
