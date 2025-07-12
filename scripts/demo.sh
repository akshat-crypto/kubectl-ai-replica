#!/bin/bash

# MCP Server Demo Script
# This demonstrates how AI agents interact with MCP servers using natural language

set -e

echo "🚀 MCP Server Demo - AI Agent Integration"
echo "=========================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    print_error "kubectl is not installed. Please install kubectl first."
    exit 1
fi

# Check if we have access to a Kubernetes cluster
if ! kubectl cluster-info &> /dev/null; then
    print_warning "No Kubernetes cluster access. Starting in demo mode..."
    DEMO_MODE=true
else
    DEMO_MODE=false
    print_success "Kubernetes cluster detected"
fi

# Build the MCP server and client
print_status "Building MCP server and client..."

if [ "$DEMO_MODE" = true ]; then
    # In demo mode, we'll just show the structure
    print_status "Demo mode: Skipping actual build"
else
    # Build the Kubernetes MCP server
    cd servers/kubernetes
    go build -o ../../bin/kubernetes-mcp-server .
    cd ../..
    
    # Build the MCP client
    cd cmd/mcp-client
    go build -o ../../bin/mcp-client .
    cd ../..
    
    print_success "Build completed"
fi

echo ""
echo "📋 Demo Overview"
echo "================"
echo "This demo shows how AI agents can interact with Kubernetes through MCP servers:"
echo ""
echo "1. 🤖 AI Agent sends natural language queries"
echo "2. 🔗 MCP Server translates queries to Kubernetes operations"
echo "3. ☸️  Kubernetes API executes the operations"
echo "4. 📊 Results are returned to the AI agent"
echo ""

if [ "$DEMO_MODE" = true ]; then
    echo "🎭 Demo Mode - Simulated Interactions"
    echo "====================================="
    echo ""
    
    echo "🤖 AI Agent: 'Show me all the pods in my cluster'"
    echo "🔗 MCP Server: Translating to Kubernetes API call..."
    echo "☸️  Kubernetes: Executing 'kubectl get pods'"
    echo "📊 Result:"
    echo "  📦 default/nginx-deployment-abc123 (Running) - Age: 2h"
    echo "  📦 default/redis-deployment-def456 (Running) - Age: 1h"
    echo ""
    
    echo "🤖 AI Agent: 'Create a new deployment called myapp using nginx:latest'"
    echo "🔗 MCP Server: Translating to Kubernetes API call..."
    echo "☸️  Kubernetes: Executing deployment creation"
    echo "📊 Result: Successfully created deployment 'myapp' in namespace 'default' with 1 replicas"
    echo ""
    
    echo "🤖 AI Agent: 'Scale the myapp deployment to 5 replicas'"
    echo "🔗 MCP Server: Translating to Kubernetes API call..."
    echo "☸️  Kubernetes: Executing scale operation"
    echo "📊 Result: Successfully scaled deployment 'myapp' in namespace 'default' to 5 replicas"
    echo ""
    
    echo "🤖 AI Agent: 'Delete the pod nginx-deployment-abc123'"
    echo "🔗 MCP Server: Translating to Kubernetes API call..."
    echo "☸️  Kubernetes: Executing pod deletion"
    echo "📊 Result: Successfully deleted pod 'nginx-deployment-abc123' from namespace 'default'"
    echo ""
    
else
    echo "🔧 Live Demo - Starting MCP Server"
    echo "=================================="
    echo ""
    
    # Start the MCP server in the background
    print_status "Starting Kubernetes MCP server..."
    ./bin/kubernetes-mcp-server :8080 &
    SERVER_PID=$!
    
    # Wait for server to start
    sleep 2
    
    print_success "MCP server started on :8080"
    echo ""
    
    echo "🤖 AI Agent Interactions"
    echo "========================"
    echo ""
    
    # Demo 1: List pods
    echo "1. Listing pods..."
    ./bin/mcp-client http://localhost:8080 natural-language "list pods"
    echo ""
    
    # Demo 2: List services
    echo "2. Listing services..."
    ./bin/mcp-client http://localhost:8080 natural-language "list services"
    echo ""
    
    # Demo 3: Create deployment
    echo "3. Creating deployment..."
    ./bin/mcp-client http://localhost:8080 natural-language "create deployment demo-app nginx:latest"
    echo ""
    
    # Demo 4: Scale deployment
    echo "4. Scaling deployment..."
    ./bin/mcp-client http://localhost:8080 natural-language "scale deployment demo-app 3"
    echo ""
    
    # Demo 5: List deployments
    echo "5. Listing deployments..."
    ./bin/mcp-client http://localhost:8080 natural-language "list deployments"
    echo ""
    
    # Stop the server
    print_status "Stopping MCP server..."
    kill $SERVER_PID
    print_success "Demo completed"
fi

echo ""
echo "🎯 Key Benefits of MCP Servers"
echo "=============================="
echo "✅ Natural Language Interface: AI agents can use human-like queries"
echo "✅ Standardized Protocol: Works with any AI model that supports MCP"
echo "✅ Secure Access: Controlled access to sensitive systems"
echo "✅ Extensible: Easy to add new tools and resources"
echo "✅ Production Ready: Enterprise-grade security and monitoring"
echo ""

echo "🔗 Real-World Use Cases"
echo "======================="
echo "• Kubernetes 1.33 kubectl AI: Natural language Kubernetes operations"
echo "• AWS Bedrock Integration: AI-powered cloud management"
echo "• DevOps Automation: Intelligent CI/CD and infrastructure management"
echo "• Security Operations: AI-powered threat detection and response"
echo "• Data Analytics: Real-time data processing with AI insights"
echo ""

echo "📚 Next Steps"
echo "============="
echo "1. Explore the MCP specification: https://modelcontextprotocol.io"
echo "2. Build your own MCP server for your systems"
echo "3. Integrate with AI models like Claude, GPT, or local models"
echo "4. Deploy in production with proper security and monitoring"
echo ""

print_success "Demo completed! 🎉" 