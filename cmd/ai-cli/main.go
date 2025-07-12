package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/mcp-servers/cli/internal/config"
	"github.com/mcp-servers/cli/pkg/nlp"
	"github.com/sirupsen/logrus"
)

func main() {
	var (
		configPath  = flag.String("config", "", "Path to configuration file")
		query       = flag.String("query", "", "Natural language query to process")
		interactive = flag.Bool("interactive", false, "Run in interactive mode")
		quiet       = flag.Bool("quiet", false, "Suppress verbose output")
		model       = flag.String("model", "", "Override LLM model")
		provider    = flag.String("provider", "", "Override LLM provider")
	)
	flag.Parse()

	// Load configuration
	llmConfig, err := config.LoadLLMConfig(*configPath)
	if err != nil {
		logrus.Fatalf("Failed to load configuration: %v", err)
	}

	// Override with command line flags
	if *model != "" {
		llmConfig.Model = *model
	}
	if *provider != "" {
		llmConfig.Provider = *provider
	}
	if *quiet {
		llmConfig.Quiet = true
	}

	// Create LLM provider
	llmProvider, err := llmConfig.CreateLLMProvider()
	if err != nil {
		logrus.Fatalf("Failed to create LLM provider: %v", err)
	}

	// Create NLP processor
	processor := nlp.NewProcessor(llmProvider)

	// Set up logging
	if llmConfig.Quiet {
		logrus.SetLevel(logrus.ErrorLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	// Display current configuration
	if !llmConfig.Quiet {
		fmt.Printf("🤖 AI CLI - Kubernetes Assistant\n")
		fmt.Printf("Provider: %s\n", llmProvider.GetProvider())
		fmt.Printf("Model: %s\n", llmProvider.GetModel())
		fmt.Printf("Configuration: %s\n\n", *configPath)
	}

	// Process single query or run interactively
	if *query != "" {
		if err := processQuery(processor, *query); err != nil {
			logrus.Fatalf("Failed to process query: %v", err)
		}
	} else if *interactive {
		runInteractive(processor)
	} else {
		fmt.Println("Usage:")
		fmt.Println("  ./ai-cli --query 'list all pods'")
		fmt.Println("  ./ai-cli --interactive")
		fmt.Println("  ./ai-cli --help")
	}
}

// processQuery processes a single query
func processQuery(processor *nlp.Processor, query string) error {
	fmt.Printf("🔍 Processing: %s\n", query)

	ctx := context.Background()
	response, err := processor.ProcessQuery(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to process query: %w", err)
	}

	// Display response
	fmt.Printf("🤖 AI Response: %s\n", response.Content)

	// Process tool calls
	if len(response.ToolCalls) > 0 {
		fmt.Println("\n🔧 Tool Calls:")
		for i, toolCall := range response.ToolCalls {
			command, err := nlp.TranslateToolCallToCommand(toolCall)
			if err != nil {
				fmt.Printf("  %d. ❌ Error: %v\n", i+1, err)
				continue
			}
			fmt.Printf("  %d. %s\n", i+1, command)
		}
	}

	return nil
}

// runInteractive runs the CLI in interactive mode
func runInteractive(processor *nlp.Processor) {
	fmt.Println("🚀 Interactive Mode - Type 'exit' to quit, 'clear' to clear history")
	fmt.Println("Example queries:")
	fmt.Println("  - list all pods in default namespace")
	fmt.Println("  - create a deployment called myapp using nginx:latest")
	fmt.Println("  - scale deployment myapp to 5 replicas")
	fmt.Println("  - delete pod nginx-deployment-abc123")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("🤖 > ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		// Handle special commands
		switch input {
		case "exit", "quit":
			fmt.Println("👋 Goodbye!")
			return
		case "clear":
			processor.ClearHistory()
			fmt.Println("🧹 History cleared")
			continue
		case "history":
			history := processor.GetHistory()
			if len(history) == 0 {
				fmt.Println("📝 No conversation history")
			} else {
				fmt.Println("📝 Conversation History:")
				for _, msg := range history {
					role := "👤"
					if msg.Role == "assistant" {
						role = "🤖"
					}
					fmt.Printf("  %s %s: %s\n", role, msg.Role, msg.Content)
				}
			}
			continue
		case "help":
			fmt.Println("Available commands:")
			fmt.Println("  exit/quit - Exit the application")
			fmt.Println("  clear - Clear conversation history")
			fmt.Println("  history - Show conversation history")
			fmt.Println("  help - Show this help")
			fmt.Println()
			fmt.Println("Or ask natural language questions about Kubernetes!")
			continue
		}

		// Process the query
		if err := processQuery(processor, input); err != nil {
			fmt.Printf("❌ Error: %v\n", err)
		}
		fmt.Println()
	}
}
