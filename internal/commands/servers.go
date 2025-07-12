package commands

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/mcp-servers/cli/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// NewServersCommand creates the servers command
func NewServersCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "servers",
		Short: "Manage MCP servers",
		Long:  `List, add, remove, and manage MCP server configurations.`,
	}

	cmd.AddCommand(
		newServersListCommand(cfg),
		newServersAddCommand(cfg),
		newServersRemoveCommand(cfg),
		newServersShowCommand(cfg),
	)

	return cmd
}

// newServersListCommand creates the list subcommand
func newServersListCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Short:   "List configured MCP servers",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return listServers(cfg)
		},
	}
}

// newServersAddCommand creates the add subcommand
func newServersAddCommand(cfg *config.Config) *cobra.Command {
	var (
		host     string
		port     int
		protocol string
		authType string
	)

	cmd := &cobra.Command{
		Use:   "add [name]",
		Short: "Add a new MCP server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return addServer(cfg, args[0], host, port, protocol, authType)
		},
	}

	cmd.Flags().StringVarP(&host, "host", "H", "localhost", "Server host")
	cmd.Flags().IntVarP(&port, "port", "p", 8080, "Server port")
	cmd.Flags().StringVarP(&protocol, "protocol", "P", "http", "Server protocol")
	cmd.Flags().StringVarP(&authType, "auth", "a", "none", "Authentication type")

	return cmd
}

// newServersRemoveCommand creates the remove subcommand
func newServersRemoveCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:     "remove [name]",
		Short:   "Remove an MCP server",
		Aliases: []string{"rm"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return removeServer(cfg, args[0])
		},
	}
}

// newServersShowCommand creates the show subcommand
func newServersShowCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "show [name]",
		Short: "Show server configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return showServer(cfg, args[0])
		},
	}
}

// listServers displays all configured servers
func listServers(cfg *config.Config) error {
	if len(cfg.Servers) == 0 {
		fmt.Println("No servers configured.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tHOST\tPORT\tPROTOCOL\tSTATUS\t")
	fmt.Fprintln(w, "----\t----\t----\t--------\t------\t")

	for name, server := range cfg.Servers {
		status := "unknown"
		// TODO: Implement actual health check
		if server.HealthCheck.Enabled {
			status = "healthy"
		}

		fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t\n",
			name,
			server.Host,
			server.Port,
			server.Protocol,
			status,
		)
	}

	return w.Flush()
}

// addServer adds a new server configuration
func addServer(cfg *config.Config, name, host string, port int, protocol, authType string) error {
	if _, exists := cfg.Servers[name]; exists {
		return fmt.Errorf("server '%s' already exists", name)
	}

	cfg.Servers[name] = config.ServerConfig{
		Host:     host,
		Port:     port,
		Protocol: protocol,
		Auth: config.AuthConfig{
			Type: authType,
		},
		HealthCheck: config.HealthCheckConfig{
			Enabled:  true,
			Interval: 30,
			Timeout:  5,
		},
	}

	logrus.Infof("Added server '%s' (%s://%s:%d)", name, protocol, host, port)
	return nil
}

// removeServer removes a server configuration
func removeServer(cfg *config.Config, name string) error {
	if _, exists := cfg.Servers[name]; !exists {
		return fmt.Errorf("server '%s' not found", name)
	}

	delete(cfg.Servers, name)
	logrus.Infof("Removed server '%s'", name)
	return nil
}

// showServer displays detailed server configuration
func showServer(cfg *config.Config, name string) error {
	server, exists := cfg.Servers[name]
	if !exists {
		return fmt.Errorf("server '%s' not found", name)
	}

	fmt.Printf("Server: %s\n", name)
	fmt.Printf("  Host: %s\n", server.Host)
	fmt.Printf("  Port: %d\n", server.Port)
	fmt.Printf("  Protocol: %s\n", server.Protocol)
	fmt.Printf("  Auth Type: %s\n", server.Auth.Type)
	fmt.Printf("  TLS Enabled: %t\n", server.TLS.Enabled)
	fmt.Printf("  Health Check: %t\n", server.HealthCheck.Enabled)

	return nil
}
