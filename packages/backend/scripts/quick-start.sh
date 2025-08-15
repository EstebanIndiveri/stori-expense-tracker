#!/bin/bash

# Quick Start Script for Stori Expense Tracker - Local Development
# This script sets up everything needed for local development

set -e

echo "🚀 Stori Expense Tracker - Quick Local Setup"
echo "============================================="

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[!]${NC} $1"
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

# Change to backend directory
cd "$(dirname "$0")/.."

# Check prerequisites
print_status "Checking prerequisites..."

# Check Go
if ! command -v go &> /dev/null; then
    print_error "Go is not installed. Please install Go 1.21 or later."
    exit 1
fi
print_success "Go found: $(go version)"

# Check Docker
if ! command -v docker &> /dev/null; then
    print_warning "Docker not found. DynamoDB Local will need to be run manually."
    DOCKER_AVAILABLE=false
else
    print_success "Docker found: $(docker --version)"
    DOCKER_AVAILABLE=true
fi

# Create .env.local if it doesn't exist
if [ ! -f .env.local ]; then
    print_status "Creating .env.local file..."
    cp .env.example .env.local
    print_warning "Please edit .env.local and add your OpenAI API key for AI features"
fi

# Install dependencies
print_status "Installing Go dependencies..."
go mod tidy
go mod download
print_success "Dependencies installed"

# Start DynamoDB Local if Docker is available
if [ "$DOCKER_AVAILABLE" = true ]; then
    print_status "Starting DynamoDB Local..."
    
    # Stop existing container if running
    docker stop dynamodb-local 2>/dev/null || true
    docker rm dynamodb-local 2>/dev/null || true
    
    # Start new container
    docker run -d --name dynamodb-local -p 8000:8000 amazon/dynamodb-local:latest
    
    # Wait for DynamoDB to be ready
    print_status "Waiting for DynamoDB Local to be ready..."
    for i in {1..30}; do
        if curl -s http://localhost:8000 >/dev/null 2>&1; then
            break
        fi
        sleep 1
    done
    
    if curl -s http://localhost:8000 >/dev/null 2>&1; then
        print_success "DynamoDB Local is running on port 8000"
    else
        print_error "Failed to start DynamoDB Local"
        exit 1
    fi
else
    print_warning "Docker not available. Please start DynamoDB Local manually:"
    print_warning "docker run -p 8000:8000 amazon/dynamodb-local"
fi

# Create DynamoDB table
print_status "Setting up DynamoDB table..."
go run ../../tools/data-ingestion/main.go --setup-table || {
    print_warning "Table setup failed. This is normal if table already exists."
}

# Load sample data
print_status "Loading sample data..."
go run ../../tools/data-ingestion/main.go || {
    print_warning "Sample data loading failed. You can load it manually later."
}

# Build the application
print_status "Building application..."
go build -o bin/api-local cmd/api/main.go
go build -o bin/ai-advisor-local cmd/ai-advisor/main.go
print_success "Application built successfully"

# Show next steps
echo ""
echo "🎉 Setup Complete!"
echo "=================="
print_success "Local development environment is ready!"
echo ""
echo "Next steps:"
echo "  1. Start the API server:"
echo "     ${BLUE}make dev${NC} or ${BLUE}./bin/api-local${NC}"
echo ""
echo "  2. Test the API:"
echo "     ${BLUE}curl http://localhost:8080/health${NC}"
echo ""
echo "  3. View DynamoDB Local UI (if installed):"
echo "     ${BLUE}http://localhost:8000${NC}"
echo ""
echo "  4. Run tests:"
echo "     ${BLUE}make test${NC}"
echo ""
echo "Available Make commands:"
echo "  ${BLUE}make help${NC}          - Show all available commands"
echo "  ${BLUE}make dev${NC}           - Start development server"
echo "  ${BLUE}make test${NC}          - Run tests"
echo "  ${BLUE}make test-integration${NC} - Run integration tests"
echo ""

# Check for .env.local configuration
if grep -q "your-openai-api-key-here" .env.local 2>/dev/null; then
    print_warning "Remember to configure your OpenAI API key in .env.local for AI features!"
fi

echo "Happy coding! 🚀"
