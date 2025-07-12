package nlp

import (
	"context"
	"fmt"

	"github.com/mcp-servers/cli/pkg/llm"
)

// Processor handles natural language processing for Kubernetes queries
type Processor struct {
	llmProvider llm.Provider
	tools       []llm.Tool
	history     []llm.Message
}

// NewProcessor creates a new NLP processor
func NewProcessor(llmProvider llm.Provider) *Processor {
	return &Processor{
		llmProvider: llmProvider,
		tools:       getDefaultKubernetesTools(),
		history:     []llm.Message{},
	}
}

// ProcessQuery processes a natural language query and returns the response
func (p *Processor) ProcessQuery(ctx context.Context, query string) (*llm.Response, error) {
	// Create query with context
	llmQuery := llm.Query{
		Text:    query,
		Tools:   p.tools,
		History: p.history,
		Context: map[string]interface{}{
			"domain": "kubernetes",
			"task":   "command_generation",
		},
	}

	// Generate response with tools
	response, err := p.llmProvider.(interface {
		GenerateResponseWithTools(context.Context, llm.Query) (*llm.Response, error)
	}).GenerateResponseWithTools(ctx, llmQuery)

	if err != nil {
		return nil, fmt.Errorf("failed to process query: %w", err)
	}

	// Update conversation history
	p.history = append(p.history, llm.Message{
		Role:    "user",
		Content: query,
	})
	p.history = append(p.history, llm.Message{
		Role:    "assistant",
		Content: response.Content,
	})

	// Keep history manageable (last 10 messages)
	if len(p.history) > 10 {
		p.history = p.history[len(p.history)-10:]
	}

	return response, nil
}

// getDefaultKubernetesTools returns the default set of Kubernetes tools
func getDefaultKubernetesTools() []llm.Tool {
	return []llm.Tool{
		{
			Name:        "kubectl_get_pods",
			Description: "List pods in a namespace or across all namespaces",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"namespace": map[string]interface{}{
						"type":        "string",
						"description": "Namespace to list pods from (optional)",
					},
					"all_namespaces": map[string]interface{}{
						"type":        "boolean",
						"description": "List pods from all namespaces",
					},
				},
			},
		},
		{
			Name:        "kubectl_get_services",
			Description: "List services in a namespace or across all namespaces",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"namespace": map[string]interface{}{
						"type":        "string",
						"description": "Namespace to list services from (optional)",
					},
					"all_namespaces": map[string]interface{}{
						"type":        "boolean",
						"description": "List services from all namespaces",
					},
				},
			},
		},
		{
			Name:        "kubectl_get_deployments",
			Description: "List deployments in a namespace or across all namespaces",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"namespace": map[string]interface{}{
						"type":        "string",
						"description": "Namespace to list deployments from (optional)",
					},
					"all_namespaces": map[string]interface{}{
						"type":        "boolean",
						"description": "List deployments from all namespaces",
					},
				},
			},
		},
		{
			Name:        "kubectl_create_deployment",
			Description: "Create a new deployment",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the deployment",
					},
					"image": map[string]interface{}{
						"type":        "string",
						"description": "Container image to use",
					},
					"namespace": map[string]interface{}{
						"type":        "string",
						"description": "Namespace to create deployment in (optional)",
					},
					"replicas": map[string]interface{}{
						"type":        "integer",
						"description": "Number of replicas (optional)",
					},
				},
				"required": []string{"name", "image"},
			},
		},
		{
			Name:        "kubectl_scale_deployment",
			Description: "Scale a deployment to a specific number of replicas",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the deployment",
					},
					"replicas": map[string]interface{}{
						"type":        "integer",
						"description": "Number of replicas",
					},
					"namespace": map[string]interface{}{
						"type":        "string",
						"description": "Namespace of the deployment (optional)",
					},
				},
				"required": []string{"name", "replicas"},
			},
		},
		{
			Name:        "kubectl_delete_pod",
			Description: "Delete a pod",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the pod",
					},
					"namespace": map[string]interface{}{
						"type":        "string",
						"description": "Namespace of the pod (optional)",
					},
				},
				"required": []string{"name"},
			},
		},
		{
			Name:        "kubectl_describe_pod",
			Description: "Describe a pod in detail",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the pod",
					},
					"namespace": map[string]interface{}{
						"type":        "string",
						"description": "Namespace of the pod (optional)",
					},
				},
				"required": []string{"name"},
			},
		},
	}
}

// AddTool adds a custom tool to the processor
func (p *Processor) AddTool(tool llm.Tool) {
	p.tools = append(p.tools, tool)
}

// ClearHistory clears the conversation history
func (p *Processor) ClearHistory() {
	p.history = []llm.Message{}
}

// GetHistory returns the conversation history
func (p *Processor) GetHistory() []llm.Message {
	return p.history
}

// TranslateToolCallToCommand translates a tool call to a kubectl command
func TranslateToolCallToCommand(toolCall llm.ToolCall) (string, error) {
	switch toolCall.ToolName {
	case "kubectl_get_pods":
		return translateGetPods(toolCall.Arguments)
	case "kubectl_get_services":
		return translateGetServices(toolCall.Arguments)
	case "kubectl_get_deployments":
		return translateGetDeployments(toolCall.Arguments)
	case "kubectl_create_deployment":
		return translateCreateDeployment(toolCall.Arguments)
	case "kubectl_scale_deployment":
		return translateScaleDeployment(toolCall.Arguments)
	case "kubectl_delete_pod":
		return translateDeletePod(toolCall.Arguments)
	case "kubectl_describe_pod":
		return translateDescribePod(toolCall.Arguments)
	default:
		return "", fmt.Errorf("unknown tool: %s", toolCall.ToolName)
	}
}

// Helper functions to translate tool calls to commands
func translateGetPods(args map[string]interface{}) (string, error) {
	cmd := "kubectl get pods"
	if namespace, ok := args["namespace"].(string); ok && namespace != "" {
		cmd += " -n " + namespace
	} else if allNamespaces, ok := args["all_namespaces"].(bool); ok && allNamespaces {
		cmd += " --all-namespaces"
	}
	return cmd, nil
}

func translateGetServices(args map[string]interface{}) (string, error) {
	cmd := "kubectl get services"
	if namespace, ok := args["namespace"].(string); ok && namespace != "" {
		cmd += " -n " + namespace
	} else if allNamespaces, ok := args["all_namespaces"].(bool); ok && allNamespaces {
		cmd += " --all-namespaces"
	}
	return cmd, nil
}

func translateGetDeployments(args map[string]interface{}) (string, error) {
	cmd := "kubectl get deployments"
	if namespace, ok := args["namespace"].(string); ok && namespace != "" {
		cmd += " -n " + namespace
	} else if allNamespaces, ok := args["all_namespaces"].(bool); ok && allNamespaces {
		cmd += " --all-namespaces"
	}
	return cmd, nil
}

func translateCreateDeployment(args map[string]interface{}) (string, error) {
	name, ok := args["name"].(string)
	if !ok {
		return "", fmt.Errorf("deployment name is required")
	}
	image, ok := args["image"].(string)
	if !ok {
		return "", fmt.Errorf("image is required")
	}

	cmd := fmt.Sprintf("kubectl create deployment %s --image=%s", name, image)

	if namespace, ok := args["namespace"].(string); ok && namespace != "" {
		cmd += " -n " + namespace
	}

	if replicas, ok := args["replicas"].(float64); ok && replicas > 0 {
		cmd += fmt.Sprintf(" --replicas=%d", int(replicas))
	}

	return cmd, nil
}

func translateScaleDeployment(args map[string]interface{}) (string, error) {
	name, ok := args["name"].(string)
	if !ok {
		return "", fmt.Errorf("deployment name is required")
	}
	replicas, ok := args["replicas"].(float64)
	if !ok {
		return "", fmt.Errorf("replicas count is required")
	}

	cmd := fmt.Sprintf("kubectl scale deployment %s --replicas=%d", name, int(replicas))

	if namespace, ok := args["namespace"].(string); ok && namespace != "" {
		cmd += " -n " + namespace
	}

	return cmd, nil
}

func translateDeletePod(args map[string]interface{}) (string, error) {
	name, ok := args["name"].(string)
	if !ok {
		return "", fmt.Errorf("pod name is required")
	}

	cmd := fmt.Sprintf("kubectl delete pod %s", name)

	if namespace, ok := args["namespace"].(string); ok && namespace != "" {
		cmd += " -n " + namespace
	}

	return cmd, nil
}

func translateDescribePod(args map[string]interface{}) (string, error) {
	name, ok := args["name"].(string)
	if !ok {
		return "", fmt.Errorf("pod name is required")
	}

	cmd := fmt.Sprintf("kubectl describe pod %s", name)

	if namespace, ok := args["namespace"].(string); ok && namespace != "" {
		cmd += " -n " + namespace
	}

	return cmd, nil
}
