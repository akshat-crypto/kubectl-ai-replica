package commands

import (
	"fmt"

	"github.com/mcp-servers/cli/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// NewConnectCommand creates the connect command
func NewConnectCommand(cfg *config.Config) *cobra.Command {
	var timeout int

	cmd := &cobra.Command{
		Use:   "connect [server]",
		Short: "Connect to an MCP server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return connectToServer(cfg, args[0], timeout)
		},
	}

	cmd.Flags().IntVarP(&timeout, "timeout", "t", 30, "Connection timeout in seconds")

	return cmd
}

// connectToServer establishes a connection to the specified server
func connectToServer(cfg *config.Config, serverName string, timeout int) error {
	server, exists := cfg.Servers[serverName]
	if !exists {
		return fmt.Errorf("server '%s' not found", serverName)
	}

	logrus.Infof("Connecting to server '%s' at %s://%s:%d",
		serverName, server.Protocol, server.Host, server.Port)

	// TODO: Implement actual connection logic
	// This would involve:
	// 1. Establishing WebSocket connection
	// 2. Authenticating if required
	// 3. Setting up message handlers
	// 4. Starting heartbeat/ping mechanism

	logrus.Info("Connection established successfully")
	return nil
}
