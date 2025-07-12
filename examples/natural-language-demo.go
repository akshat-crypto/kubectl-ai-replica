package main

import (
	"fmt"
	"strings"
)

// NaturalLanguageProcessor demonstrates how AI agents process natural language
type NaturalLanguageProcessor struct {
	commands map[string]string
}

// NewNaturalLanguageProcessor creates a new processor
func NewNaturalLanguageProcessor() *NaturalLanguageProcessor {
	return &NaturalLanguageProcessor{
		commands: make(map[string]string),
	}
}

// ProcessQuery processes natural language queries and converts them to commands
func (p *NaturalLanguageProcessor) ProcessQuery(query string) string {
	query = strings.ToLower(query)

	fmt.Printf("🤖 AI Agent: Processing query: '%s'\n", query)

	switch {
	case strings.Contains(query, "list") && strings.Contains(query, "pod"):
		return "kubectl get pods --all-namespaces"
	case strings.Contains(query, "list") && strings.Contains(query, "service"):
		return "kubectl get services --all-namespaces"
	case strings.Contains(query, "list") && strings.Contains(query, "deployment"):
		return "kubectl get deployments --all-namespaces"
	case strings.Contains(query, "create") && strings.Contains(query, "deployment"):
		// Extract deployment name and image
		parts := strings.Fields(query)
		for i, part := range parts {
			if part == "deployment" && i+1 < len(parts) {
				name := parts[i+1]
				image := "nginx:latest"
				if i+2 < len(parts) {
					image = parts[i+2]
				}
				return fmt.Sprintf("kubectl create deployment %s --image=%s", name, image)
			}
		}
		return "kubectl create deployment <name> --image=<image>"
	case strings.Contains(query, "scale") && strings.Contains(query, "deployment"):
		parts := strings.Fields(query)
		for i, part := range parts {
			if part == "deployment" && i+1 < len(parts) {
				name := parts[i+1]
				replicas := "3"
				if i+2 < len(parts) {
					replicas = parts[i+2]
				}
				return fmt.Sprintf("kubectl scale deployment %s --replicas=%s", name, replicas)
			}
		}
		return "kubectl scale deployment <name> --replicas=<number>"
	case strings.Contains(query, "delete") && strings.Contains(query, "pod"):
		parts := strings.Fields(query)
		for i, part := range parts {
			if part == "pod" && i+1 < len(parts) {
				return fmt.Sprintf("kubectl delete pod %s", parts[i+1])
			}
		}
		return "kubectl delete pod <name>"
	default:
		return "Unknown command. Try: list pods, create deployment, scale deployment, delete pod"
	}
}

// ExecuteCommand simulates executing the command
func (p *NaturalLanguageProcessor) ExecuteCommand(command string) {
	fmt.Printf("🔗 MCP Server: Executing command: %s\n", command)

	// Simulate command execution
	switch {
	case strings.Contains(command, "get pods"):
		fmt.Println("📊 Result:")
		fmt.Println("  📦 default/nginx-deployment-abc123 (Running) - Age: 2h")
		fmt.Println("  📦 default/redis-deployment-def456 (Running) - Age: 1h")
		fmt.Println("  📦 kube-system/coredns-xyz789 (Running) - Age: 5h")
	case strings.Contains(command, "get services"):
		fmt.Println("📊 Result:")
		fmt.Println("  🔗 default/kubernetes (ClusterIP) - IP: 10.96.0.1")
		fmt.Println("  🔗 default/nginx-service (ClusterIP) - IP: 10.96.1.100")
		fmt.Println("  🔗 kube-system/kube-dns (ClusterIP) - IP: 10.96.0.10")
	case strings.Contains(command, "get deployments"):
		fmt.Println("📊 Result:")
		fmt.Println("  🚀 default/nginx-deployment - Replicas: 3/3")
		fmt.Println("  🚀 default/redis-deployment - Replicas: 1/1")
		fmt.Println("  🚀 kube-system/coredns - Replicas: 2/2")
	case strings.Contains(command, "create deployment"):
		fmt.Println("📊 Result: deployment.apps/myapp created")
	case strings.Contains(command, "scale deployment"):
		fmt.Println("📊 Result: deployment.apps/myapp scaled")
	case strings.Contains(command, "delete pod"):
		fmt.Println("📊 Result: pod \"nginx-deployment-abc123\" deleted")
	default:
		fmt.Println("📊 Result: Command executed successfully")
	}
}

func main() {
	fmt.Println("🚀 Natural Language MCP Demo")
	fmt.Println("=============================")
	fmt.Println("")

	processor := NewNaturalLanguageProcessor()

	// Demo queries
	queries := []string{
		"Show me all the pods in my cluster",
		"List all services",
		"Get deployments",
		"Create a new deployment called myapp using nginx:latest",
		"Scale the myapp deployment to 5 replicas",
		"Delete the pod nginx-deployment-abc123",
		"What's the weather like?", // Unknown query
	}

	for i, query := range queries {
		fmt.Printf("Query %d: %s\n", i+1, query)
		fmt.Println("---")

		command := processor.ProcessQuery(query)
		processor.ExecuteCommand(command)

		fmt.Println("")
	}

	fmt.Println("🎯 How This Works:")
	fmt.Println("==================")
	fmt.Println("1. 🤖 AI Agent receives natural language query")
	fmt.Println("2. 🔍 AI processes the query and identifies intent")
	fmt.Println("3. 🔗 MCP Server translates intent to specific commands")
	fmt.Println("4. ☸️  Kubernetes API executes the commands")
	fmt.Println("5. 📊 Results are formatted and returned to AI")
	fmt.Println("")

	fmt.Println("💡 Real-World Applications:")
	fmt.Println("===========================")
	fmt.Println("• kubectl AI in Kubernetes 1.33")
	fmt.Println("• AWS Bedrock natural language operations")
	fmt.Println("• DevOps automation with AI assistants")
	fmt.Println("• Security operations with natural language")
	fmt.Println("• Database management through AI")
}
