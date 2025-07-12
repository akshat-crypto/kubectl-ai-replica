package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/mcp-servers/cli/pkg/mcp"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: mcp-client <server-url> [command]")
		fmt.Println("Commands:")
		fmt.Println("  list-pods                    - List all pods")
		fmt.Println("  list-services                - List all services")
		fmt.Println("  list-deployments             - List all deployments")
		fmt.Println("  create-deployment <name> <image> - Create a deployment")
		fmt.Println("  scale-deployment <name> <replicas> - Scale a deployment")
		fmt.Println("  delete-pod <name>            - Delete a pod")
		fmt.Println("  natural-language <query>     - Natural language query")
		os.Exit(1)
	}

	serverURL := os.Args[1]
	client := NewMCPClient(serverURL)

	// Initialize connection
	if err := client.Initialize(); err != nil {
		fmt.Printf("Failed to initialize: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) < 3 {
		fmt.Println("No command specified. Use 'help' for available commands.")
		os.Exit(1)
	}

	command := os.Args[2]
	args := os.Args[3:]

	switch command {
	case "list-pods":
		if err := client.ListPods(); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	case "list-services":
		if err := client.ListServices(); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	case "list-deployments":
		if err := client.ListDeployments(); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	case "create-deployment":
		if len(args) < 2 {
			fmt.Println("Usage: create-deployment <name> <image>")
			os.Exit(1)
		}
		if err := client.CreateDeployment(args[0], args[1]); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	case "scale-deployment":
		if len(args) < 2 {
			fmt.Println("Usage: scale-deployment <name> <replicas>")
			os.Exit(1)
		}
		if err := client.ScaleDeployment(args[0], args[1]); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	case "delete-pod":
		if len(args) < 1 {
			fmt.Println("Usage: delete-pod <name>")
			os.Exit(1)
		}
		if err := client.DeletePod(args[0]); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	case "natural-language":
		if len(args) < 1 {
			fmt.Println("Usage: natural-language <query>")
			os.Exit(1)
		}
		query := strings.Join(args, " ")
		if err := client.NaturalLanguageQuery(query); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

// MCPClient represents an MCP client
type MCPClient struct {
	serverURL string
	client    *http.Client
}

// NewMCPClient creates a new MCP client
func NewMCPClient(serverURL string) *MCPClient {
	return &MCPClient{
		serverURL: serverURL,
		client:    &http.Client{},
	}
}

// Initialize initializes the connection to the MCP server
func (c *MCPClient) Initialize() error {
	req := mcp.InitializeRequest{
		ProtocolVersion: mcp.ProtocolVersion,
		Capabilities: mcp.ClientCapabilities{
			Resources: mcp.ResourceCapabilities{Subscribe: true},
			Tools:     mcp.ToolCapabilities{Call: true},
		},
		ClientInfo: mcp.ClientInfo{
			Name:    "mcp-client",
			Version: "1.0.0",
		},
	}

	msg, err := mcp.NewMessage(mcp.MessageTypeInitialize, "init-1", req)
	if err != nil {
		return err
	}

	_, err = c.sendMessage(msg)
	return err
}

// ListPods lists all pods in the cluster
func (c *MCPClient) ListPods() error {
	fmt.Println("🤖 AI Agent: I'll get the list of pods for you...")

	// First, list available resources
	msg, err := mcp.NewMessage(mcp.MessageTypeListResources, "list-resources-1", nil)
	if err != nil {
		return err
	}

	_, err = c.sendMessage(msg)
	if err != nil {
		return err
	}

	// Read the pods resource
	readReq := map[string]string{"uri": "kubernetes://pods"}
	readMsg, err := mcp.NewMessage(mcp.MessageTypeReadResource, "read-pods-1", readReq)
	if err != nil {
		return err
	}

	readResp, err := c.sendMessage(readMsg)
	if err != nil {
		return err
	}

	var resource mcp.Resource
	if err := readResp.UnmarshalData(&resource); err != nil {
		return err
	}

	var podsData map[string]interface{}
	if err := json.Unmarshal(resource.Content, &podsData); err != nil {
		return err
	}

	fmt.Println("✅ Here are the pods in your cluster:")
	if pods, ok := podsData["pods"].([]interface{}); ok {
		for _, pod := range pods {
			if podMap, ok := pod.(map[string]interface{}); ok {
				fmt.Printf("  📦 %s/%s (%s) - Age: %s\n",
					podMap["namespace"], podMap["name"], podMap["status"], podMap["age"])
			}
		}
	}

	return nil
}

// ListServices lists all services in the cluster
func (c *MCPClient) ListServices() error {
	fmt.Println("🤖 AI Agent: I'll get the list of services for you...")

	readReq := map[string]string{"uri": "kubernetes://services"}
	readMsg, err := mcp.NewMessage(mcp.MessageTypeReadResource, "read-services-1", readReq)
	if err != nil {
		return err
	}

	readResp, err := c.sendMessage(readMsg)
	if err != nil {
		return err
	}

	var resource mcp.Resource
	if err := readResp.UnmarshalData(&resource); err != nil {
		return err
	}

	var servicesData map[string]interface{}
	if err := json.Unmarshal(resource.Content, &servicesData); err != nil {
		return err
	}

	fmt.Println("✅ Here are the services in your cluster:")
	if services, ok := servicesData["services"].([]interface{}); ok {
		for _, service := range services {
			if serviceMap, ok := service.(map[string]interface{}); ok {
				fmt.Printf("  🔗 %s/%s (%s) - IP: %s\n",
					serviceMap["namespace"], serviceMap["name"], serviceMap["type"], serviceMap["clusterIP"])
			}
		}
	}

	return nil
}

// ListDeployments lists all deployments in the cluster
func (c *MCPClient) ListDeployments() error {
	fmt.Println("🤖 AI Agent: I'll get the list of deployments for you...")

	readReq := map[string]string{"uri": "kubernetes://deployments"}
	readMsg, err := mcp.NewMessage(mcp.MessageTypeReadResource, "read-deployments-1", readReq)
	if err != nil {
		return err
	}

	readResp, err := c.sendMessage(readMsg)
	if err != nil {
		return err
	}

	var resource mcp.Resource
	if err := readResp.UnmarshalData(&resource); err != nil {
		return err
	}

	var deploymentsData map[string]interface{}
	if err := json.Unmarshal(resource.Content, &deploymentsData); err != nil {
		return err
	}

	fmt.Println("✅ Here are the deployments in your cluster:")
	if deployments, ok := deploymentsData["deployments"].([]interface{}); ok {
		for _, deployment := range deployments {
			if deploymentMap, ok := deployment.(map[string]interface{}); ok {
				fmt.Printf("  🚀 %s/%s - Replicas: %v/%v\n",
					deploymentMap["namespace"], deploymentMap["name"],
					deploymentMap["available"], deploymentMap["replicas"])
			}
		}
	}

	return nil
}

// CreateDeployment creates a new deployment
func (c *MCPClient) CreateDeployment(name, image string) error {
	fmt.Printf("🤖 AI Agent: I'll create a deployment named '%s' with image '%s'...\n", name, image)

	// First, list available tools
	msg, err := mcp.NewMessage(mcp.MessageTypeListTools, "list-tools-1", nil)
	if err != nil {
		return err
	}

	_, err = c.sendMessage(msg)
	if err != nil {
		return err
	}

	// Call the create deployment tool
	toolCall := mcp.ToolCall{
		Name: "create_deployment",
		Arguments: map[string]interface{}{
			"name":      name,
			"namespace": "default",
			"image":     image,
			"replicas":  1,
		},
	}

	callMsg, err := mcp.NewMessage(mcp.MessageTypeCallTool, "call-tool-1", toolCall)
	if err != nil {
		return err
	}

	callResp, err := c.sendMessage(callMsg)
	if err != nil {
		return err
	}

	var result mcp.ToolResult
	if err := callResp.UnmarshalData(&result); err != nil {
		return err
	}

	for _, content := range result.Content {
		if content.Type == "text" {
			fmt.Printf("✅ %s\n", content.Text)
		}
	}

	return nil
}

// ScaleDeployment scales a deployment
func (c *MCPClient) ScaleDeployment(name, replicas string) error {
	fmt.Printf("🤖 AI Agent: I'll scale deployment '%s' to %s replicas...\n", name, replicas)

	toolCall := mcp.ToolCall{
		Name: "scale_deployment",
		Arguments: map[string]interface{}{
			"name":      name,
			"namespace": "default",
			"replicas":  replicas,
		},
	}

	callMsg, err := mcp.NewMessage(mcp.MessageTypeCallTool, "call-tool-1", toolCall)
	if err != nil {
		return err
	}

	callResp, err := c.sendMessage(callMsg)
	if err != nil {
		return err
	}

	var result mcp.ToolResult
	if err := callResp.UnmarshalData(&result); err != nil {
		return err
	}

	for _, content := range result.Content {
		if content.Type == "text" {
			fmt.Printf("✅ %s\n", content.Text)
		}
	}

	return nil
}

// DeletePod deletes a pod
func (c *MCPClient) DeletePod(name string) error {
	fmt.Printf("🤖 AI Agent: I'll delete pod '%s'...\n", name)

	toolCall := mcp.ToolCall{
		Name: "delete_pod",
		Arguments: map[string]interface{}{
			"name":      name,
			"namespace": "default",
		},
	}

	callMsg, err := mcp.NewMessage(mcp.MessageTypeCallTool, "call-tool-1", toolCall)
	if err != nil {
		return err
	}

	callResp, err := c.sendMessage(callMsg)
	if err != nil {
		return err
	}

	var result mcp.ToolResult
	if err := callResp.UnmarshalData(&result); err != nil {
		return err
	}

	for _, content := range result.Content {
		if content.Type == "text" {
			fmt.Printf("✅ %s\n", content.Text)
		}
	}

	return nil
}

// NaturalLanguageQuery handles natural language queries
func (c *MCPClient) NaturalLanguageQuery(query string) error {
	fmt.Printf("🤖 AI Agent: Processing your query: '%s'\n", query)

	// Simple natural language processing
	query = strings.ToLower(query)

	switch {
	case strings.Contains(query, "pod") && strings.Contains(query, "list"):
		return c.ListPods()
	case strings.Contains(query, "service") && strings.Contains(query, "list"):
		return c.ListServices()
	case strings.Contains(query, "deployment") && strings.Contains(query, "list"):
		return c.ListDeployments()
	case strings.Contains(query, "create") && strings.Contains(query, "deployment"):
		// Extract name and image from query
		parts := strings.Fields(query)
		for i, part := range parts {
			if part == "deployment" && i+1 < len(parts) {
				name := parts[i+1]
				image := "nginx:latest" // default image
				if i+2 < len(parts) {
					image = parts[i+2]
				}
				return c.CreateDeployment(name, image)
			}
		}
		fmt.Println("❌ Please specify deployment name and image")
	case strings.Contains(query, "scale") && strings.Contains(query, "deployment"):
		parts := strings.Fields(query)
		for i, part := range parts {
			if part == "deployment" && i+1 < len(parts) {
				name := parts[i+1]
				replicas := "3" // default replicas
				if i+2 < len(parts) {
					replicas = parts[i+2]
				}
				return c.ScaleDeployment(name, replicas)
			}
		}
		fmt.Println("❌ Please specify deployment name and replicas")
	case strings.Contains(query, "delete") && strings.Contains(query, "pod"):
		parts := strings.Fields(query)
		for i, part := range parts {
			if part == "pod" && i+1 < len(parts) {
				return c.DeletePod(parts[i+1])
			}
		}
		fmt.Println("❌ Please specify pod name to delete")
	default:
		fmt.Println("❌ I don't understand that query. Try:")
		fmt.Println("  - 'list pods'")
		fmt.Println("  - 'list services'")
		fmt.Println("  - 'list deployments'")
		fmt.Println("  - 'create deployment myapp nginx:latest'")
		fmt.Println("  - 'scale deployment myapp 5'")
		fmt.Println("  - 'delete pod mypod'")
	}

	return nil
}

// sendMessage sends a message to the MCP server
func (c *MCPClient) sendMessage(msg *mcp.Message) (*mcp.Message, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Post(c.serverURL+"/mcp", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response mcp.Message
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return &response, nil
}
