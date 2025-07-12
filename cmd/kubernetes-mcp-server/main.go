package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/mcp-servers/cli/servers/kubernetes"
)

func main() {
	var (
		addr       = flag.String("addr", ":8080", "Server address to listen on")
		kubeconfig = flag.String("kubeconfig", "", "Path to kubeconfig file (optional)")
	)
	flag.Parse()

	// Create and start the Kubernetes MCP server
	server, err := kubernetes.NewServer(*kubeconfig)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	fmt.Printf("Starting Kubernetes MCP server on %s\n", *addr)
	if err := server.Start(*addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
