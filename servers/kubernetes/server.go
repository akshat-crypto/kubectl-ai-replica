package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mcp-servers/cli/pkg/mcp"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Server represents a Kubernetes MCP server
type Server struct {
	clientset *kubernetes.Clientset
	config    *rest.Config
	server    *http.Server
	logger    *logrus.Logger
}

// NewServer creates a new Kubernetes MCP server
func NewServer(kubeconfig string) (*Server, error) {
	var config *rest.Config
	var err error

	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return &Server{
		clientset: clientset,
		config:    config,
		logger:    logrus.New(),
	}, nil
}

// Start starts the MCP server
func (s *Server) Start(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", s.handleMCP)

	s.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	s.logger.Infof("Starting Kubernetes MCP server on %s", addr)
	return s.server.ListenAndServe()
}

// Stop stops the MCP server
func (s *Server) Stop() error {
	if s.server != nil {
		return s.server.Shutdown(context.Background())
	}
	return nil
}

// handleMCP handles MCP protocol messages
func (s *Server) handleMCP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var msg mcp.Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	response, err := s.handleMessage(&msg)
	if err != nil {
		s.logger.Errorf("Error handling message: %v", err)
		response = &mcp.Message{
			Type: mcp.MessageTypeError,
			ID:   msg.ID,
			Data: json.RawMessage(fmt.Sprintf(`{"type":"error","message":"%s"}`, err.Error())),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleMessage processes MCP protocol messages
func (s *Server) handleMessage(msg *mcp.Message) (*mcp.Message, error) {
	switch msg.Type {
	case mcp.MessageTypeInitialize:
		return s.handleInitialize(msg)
	case mcp.MessageTypeListResources:
		return s.handleListResources(msg)
	case mcp.MessageTypeReadResource:
		return s.handleReadResource(msg)
	case mcp.MessageTypeListTools:
		return s.handleListTools(msg)
	case mcp.MessageTypeCallTool:
		return s.handleCallTool(msg)
	case mcp.MessageTypePing:
		return s.handlePing(msg)
	default:
		return nil, fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

// handleInitialize handles initialization requests
func (s *Server) handleInitialize(msg *mcp.Message) (*mcp.Message, error) {
	var req mcp.InitializeRequest
	if err := msg.UnmarshalData(&req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal initialize request: %w", err)
	}

	response := mcp.InitializationResponse{
		ProtocolVersion: mcp.ProtocolVersion,
		Capabilities: mcp.ServerCapabilities{
			Resources: mcp.ResourceCapabilities{Subscribe: true},
			Tools:     mcp.ToolCapabilities{Call: true},
		},
		ServerInfo: mcp.ServerInfo{
			Name:    "kubernetes-mcp-server",
			Version: "1.0.0",
		},
	}

	return mcp.NewMessage(mcp.MessageTypeInitialization, msg.ID, response)
}

// handleListResources handles resource listing requests
func (s *Server) handleListResources(msg *mcp.Message) (*mcp.Message, error) {
	resources := []mcp.Resource{
		{
			URI:         "kubernetes://pods",
			Name:        "Kubernetes Pods",
			Description: "List of all pods in the cluster",
			MimeType:    "application/json",
		},
		{
			URI:         "kubernetes://services",
			Name:        "Kubernetes Services",
			Description: "List of all services in the cluster",
			MimeType:    "application/json",
		},
		{
			URI:         "kubernetes://deployments",
			Name:        "Kubernetes Deployments",
			Description: "List of all deployments in the cluster",
			MimeType:    "application/json",
		},
		{
			URI:         "kubernetes://nodes",
			Name:        "Kubernetes Nodes",
			Description: "List of all nodes in the cluster",
			MimeType:    "application/json",
		},
	}

	return mcp.NewMessage("listResources", msg.ID, map[string]interface{}{
		"resources": resources,
	})
}

// handleReadResource handles resource reading requests
func (s *Server) handleReadResource(msg *mcp.Message) (*mcp.Message, error) {
	var req struct {
		URI string `json:"uri"`
	}
	if err := msg.UnmarshalData(&req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal read resource request: %w", err)
	}

	var content interface{}
	var err error

	switch req.URI {
	case "kubernetes://pods":
		content, err = s.getPods()
	case "kubernetes://services":
		content, err = s.getServices()
	case "kubernetes://deployments":
		content, err = s.getDeployments()
	case "kubernetes://nodes":
		content, err = s.getNodes()
	default:
		return nil, fmt.Errorf("unknown resource URI: %s", req.URI)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get resource %s: %w", req.URI, err)
	}

	contentBytes, err := json.Marshal(content)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resource content: %w", err)
	}

	resource := mcp.Resource{
		URI:      req.URI,
		Content:  contentBytes,
		MimeType: "application/json",
	}

	return mcp.NewMessage("readResource", msg.ID, resource)
}

// handleListTools handles tool listing requests
func (s *Server) handleListTools(msg *mcp.Message) (*mcp.Message, error) {
	tools := []mcp.Tool{
		{
			Name:        "get_pods",
			Description: "Get all pods in the cluster or in a specific namespace",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"namespace": map[string]interface{}{
						"type":        "string",
						"description": "Namespace to get pods from (optional)",
					},
				},
			},
		},
		{
			Name:        "create_deployment",
			Description: "Create a new deployment",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the deployment",
					},
					"namespace": map[string]interface{}{
						"type":        "string",
						"description": "Namespace for the deployment",
					},
					"image": map[string]interface{}{
						"type":        "string",
						"description": "Container image to deploy",
					},
					"replicas": map[string]interface{}{
						"type":        "integer",
						"description": "Number of replicas",
					},
				},
				"required": []string{"name", "namespace", "image"},
			},
		},
		{
			Name:        "scale_deployment",
			Description: "Scale a deployment to a specific number of replicas",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the deployment",
					},
					"namespace": map[string]interface{}{
						"type":        "string",
						"description": "Namespace of the deployment",
					},
					"replicas": map[string]interface{}{
						"type":        "integer",
						"description": "Number of replicas to scale to",
					},
				},
				"required": []string{"name", "namespace", "replicas"},
			},
		},
		{
			Name:        "delete_pod",
			Description: "Delete a pod",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the pod to delete",
					},
					"namespace": map[string]interface{}{
						"type":        "string",
						"description": "Namespace of the pod",
					},
				},
				"required": []string{"name", "namespace"},
			},
		},
	}

	return mcp.NewMessage("listTools", msg.ID, map[string]interface{}{
		"tools": tools,
	})
}

// handleCallTool handles tool execution requests
func (s *Server) handleCallTool(msg *mcp.Message) (*mcp.Message, error) {
	var req mcp.ToolCall
	if err := msg.UnmarshalData(&req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tool call request: %w", err)
	}

	var result *mcp.ToolResult
	var err error

	switch req.Name {
	case "get_pods":
		result, err = s.getPodsTool(req.Arguments)
	case "create_deployment":
		result, err = s.createDeploymentTool(req.Arguments)
	case "scale_deployment":
		result, err = s.scaleDeploymentTool(req.Arguments)
	case "delete_pod":
		result, err = s.deletePodTool(req.Arguments)
	default:
		return nil, fmt.Errorf("unknown tool: %s", req.Name)
	}

	if err != nil {
		return nil, fmt.Errorf("tool execution failed: %w", err)
	}

	return mcp.NewMessage("callTool", msg.ID, result)
}

// handlePing handles ping requests
func (s *Server) handlePing(msg *mcp.Message) (*mcp.Message, error) {
	return mcp.NewMessage(mcp.MessageTypePong, msg.ID, nil)
}

// Kubernetes resource methods
func (s *Server) getPods() (interface{}, error) {
	pods, err := s.clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// Simplify pod data for JSON response
	var simplifiedPods []map[string]interface{}
	for _, pod := range pods.Items {
		simplifiedPods = append(simplifiedPods, map[string]interface{}{
			"name":      pod.Name,
			"namespace": pod.Namespace,
			"status":    pod.Status.Phase,
			"age":       time.Since(pod.CreationTimestamp.Time).String(),
		})
	}

	return map[string]interface{}{
		"pods":  simplifiedPods,
		"total": len(simplifiedPods),
	}, nil
}

func (s *Server) getServices() (interface{}, error) {
	services, err := s.clientset.CoreV1().Services("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var simplifiedServices []map[string]interface{}
	for _, service := range services.Items {
		simplifiedServices = append(simplifiedServices, map[string]interface{}{
			"name":      service.Name,
			"namespace": service.Namespace,
			"type":      service.Spec.Type,
			"clusterIP": service.Spec.ClusterIP,
		})
	}

	return map[string]interface{}{
		"services": simplifiedServices,
		"total":    len(simplifiedServices),
	}, nil
}

func (s *Server) getDeployments() (interface{}, error) {
	deployments, err := s.clientset.AppsV1().Deployments("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var simplifiedDeployments []map[string]interface{}
	for _, deployment := range deployments.Items {
		simplifiedDeployments = append(simplifiedDeployments, map[string]interface{}{
			"name":      deployment.Name,
			"namespace": deployment.Namespace,
			"replicas":  deployment.Spec.Replicas,
			"available": deployment.Status.AvailableReplicas,
		})
	}

	return map[string]interface{}{
		"deployments": simplifiedDeployments,
		"total":       len(simplifiedDeployments),
	}, nil
}

func (s *Server) getNodes() (interface{}, error) {
	nodes, err := s.clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var simplifiedNodes []map[string]interface{}
	for _, node := range nodes.Items {
		simplifiedNodes = append(simplifiedNodes, map[string]interface{}{
			"name":   node.Name,
			"status": node.Status.Conditions[len(node.Status.Conditions)-1].Type,
			"age":    time.Since(node.CreationTimestamp.Time).String(),
		})
	}

	return map[string]interface{}{
		"nodes": simplifiedNodes,
		"total": len(simplifiedNodes),
	}, nil
}

// Tool execution methods
func (s *Server) getPodsTool(args map[string]interface{}) (*mcp.ToolResult, error) {
	namespace := ""
	if ns, ok := args["namespace"].(string); ok {
		namespace = ns
	}

	var pods *corev1.PodList
	var err error

	if namespace != "" {
		pods, err = s.clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	} else {
		pods, err = s.clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
	}

	if err != nil {
		return nil, err
	}

	var podNames []string
	for _, pod := range pods.Items {
		podNames = append(podNames, fmt.Sprintf("%s/%s", pod.Namespace, pod.Name))
	}

	return &mcp.ToolResult{
		Content: []mcp.ToolResultContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Found %d pods:\n%s", len(podNames), strings.Join(podNames, "\n")),
			},
		},
	}, nil
}

func (s *Server) createDeploymentTool(args map[string]interface{}) (*mcp.ToolResult, error) {
	name := args["name"].(string)
	namespace := args["namespace"].(string)
	image := args["image"].(string)
	replicas := int32(1)
	if r, ok := args["replicas"].(float64); ok {
		replicas = int32(r)
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  name,
							Image: image,
						},
					},
				},
			},
		},
	}

	_, err := s.clientset.AppsV1().Deployments(namespace).Create(context.Background(), deployment, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return &mcp.ToolResult{
		Content: []mcp.ToolResultContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Successfully created deployment '%s' in namespace '%s' with %d replicas", name, namespace, replicas),
			},
		},
	}, nil
}

func (s *Server) scaleDeploymentTool(args map[string]interface{}) (*mcp.ToolResult, error) {
	name := args["name"].(string)
	namespace := args["namespace"].(string)
	replicas := int32(args["replicas"].(float64))

	scale, err := s.clientset.AppsV1().Deployments(namespace).GetScale(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	scale.Spec.Replicas = replicas
	_, err = s.clientset.AppsV1().Deployments(namespace).UpdateScale(context.Background(), name, scale, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	return &mcp.ToolResult{
		Content: []mcp.ToolResultContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Successfully scaled deployment '%s' in namespace '%s' to %d replicas", name, namespace, replicas),
			},
		},
	}, nil
}

func (s *Server) deletePodTool(args map[string]interface{}) (*mcp.ToolResult, error) {
	name := args["name"].(string)
	namespace := args["namespace"].(string)

	err := s.clientset.CoreV1().Pods(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		return nil, err
	}

	return &mcp.ToolResult{
		Content: []mcp.ToolResultContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Successfully deleted pod '%s' from namespace '%s'", name, namespace),
			},
		},
	}, nil
}
