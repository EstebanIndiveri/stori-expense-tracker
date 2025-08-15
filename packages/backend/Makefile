# Stori Expense Tracker - Enterprise Makefile
# ===========================================

.PHONY: help build run dev test test-unit test-integration test-coverage seed seed-dry clean deps fmt check quick-start db-start db-stop db-reset db-ensure demo

# Default target
.DEFAULT_GOAL := help

# Colors for output
GREEN := \033[32m
YELLOW := \033[33m
RED := \033[31m
BLUE := \033[34m
RESET := \033[0m

## Help: Show this help message
help:
	@echo "$(BLUE)Stori Expense Tracker - Development Commands$(RESET)"
	@echo ""
	@echo "$(GREEN)Quick Start:$(RESET)"
	@echo "  make quick-start  - Complete setup for first-time users"
	@echo "  make dev          - Start development server"
	@echo ""
	@echo "$(GREEN)Core Commands:$(RESET)"
	@echo "  make deps         - Install dependencies"
	@echo "  make build        - Build the application" 
	@echo "  make run          - Run the application"
	@echo "  make test         - Run all tests"
	@echo "  make test-unit    - Run unit tests only"
	@echo "  make test-coverage - Generate test coverage report"
	@echo ""
	@echo "$(GREEN)Data Management:$(RESET)"
	@echo "  make seed         - Seed database with sample data"
	@echo "  make seed-dry     - Preview data to be seeded"
	@echo "  make db-start     - Start DynamoDB local"
	@echo "  make db-stop      - Stop DynamoDB local"
	@echo "  make db-reset     - Reset database"
	@echo ""
	@echo "$(GREEN)Code Quality:$(RESET)"
	@echo "  make fmt          - Format code"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make check        - Check system requirements"

## Quick Start: Complete setup for new developers
quick-start: check deps db-start
	@echo "$(YELLOW)Setting up development environment...$(RESET)"
	@sleep 2
	@make seed
	@echo "$(GREEN)✓ Setup complete! Run 'make dev' to start development$(RESET)"

## Check: Verify system requirements
check:
	@echo "$(YELLOW)Checking system requirements...$(RESET)"
	@echo "Go version: $(shell go version)"
	@echo "DynamoDB local: $(shell curl -s http://localhost:8000 >/dev/null 2>&1 && echo '$(GREEN)Running$(RESET)' || echo '$(RED)Not running$(RESET)')"
	@echo "JSON data: $(shell test -f ../../data/mock_expense_and_income.json && echo '$(GREEN)Found$(RESET)' || echo '$(RED)Missing$(RESET)')"

## Dependencies: Install Go dependencies
deps:
	@echo "$(YELLOW)Installing dependencies...$(RESET)"
	go mod download
	go mod tidy
	@echo "$(GREEN)✓ Dependencies installed$(RESET)"

## Build: Compile the application
build:
	@echo "$(YELLOW)Building application...$(RESET)"
	mkdir -p bin
	go build -o bin/api ./cmd/api
	go build -o bin/ai-advisor ./cmd/ai-advisor
	@echo "$(GREEN)✓ Build completed$(RESET)"

## Run: Start the application
run: build
	@echo "$(YELLOW)Starting application...$(RESET)"
	ENVIRONMENT=local ./bin/api

## Dev: Start development server with live reload
dev:
	@echo "$(YELLOW)Starting development server...$(RESET)"
	@echo "$(BLUE)API will be available at http://localhost:8080$(RESET)"
	ENVIRONMENT=local go run cmd/api/main.go

## Test: Run all tests
test:
	@echo "$(YELLOW)Running tests...$(RESET)"
	go test -v ./...
	@echo "$(GREEN)✓ Tests completed$(RESET)"

## Test Unit: Run only unit tests
test-unit:
	@echo "$(YELLOW)Running unit tests...$(RESET)"
	go test -v ./tests/unit/...
	@echo "$(GREEN)✓ Unit tests completed$(RESET)"

## Test Integration: Run integration tests
test-integration:
	@echo "$(YELLOW)Running integration tests...$(RESET)"
	go test -v ./tests/integration/...
	@echo "$(GREEN)✓ Integration tests completed$(RESET)"

## Test Coverage: Generate test coverage report
test-coverage:
	@echo "$(YELLOW)Generating test coverage...$(RESET)"
	mkdir -p test-results
	go test -coverprofile=test-results/coverage.out ./...
	go tool cover -html=test-results/coverage.out -o test-results/coverage.html
	@echo "$(GREEN)✓ Coverage report generated at test-results/coverage.html$(RESET)"

## Seed: Load sample data into database
seed: db-ensure
	@echo "$(YELLOW)Seeding database with sample data...$(RESET)"
	@echo "$(BLUE)Using seed command from cmd/seed...$(RESET)"
	go run cmd/seed/main.go
	@echo "$(GREEN)✓ Seed completed$(RESET)"

## Seed Dry: Preview data to be seeded  
seed-dry:
	@echo "$(YELLOW)Previewing seed data...$(RESET)"
	go run cmd/seed/main.go --dry-run

## DB Start: Start DynamoDB local
db-start:
	@echo "$(YELLOW)Starting DynamoDB local...$(RESET)"
	@if command -v docker >/dev/null 2>&1; then \
		if ! docker ps | grep -q dynamodb-local 2>/dev/null; then \
			docker run -d --name dynamodb-local -p 8000:8000 amazon/dynamodb-local:latest -jar DynamoDBLocal.jar -sharedDb >/dev/null 2>&1; \
			echo "$(GREEN)✓ DynamoDB local started on port 8000$(RESET)"; \
		else \
			echo "$(BLUE)DynamoDB local already running$(RESET)"; \
		fi \
	else \
		echo "$(RED)Docker not found. Please install Docker to run DynamoDB local$(RESET)"; \
		exit 1; \
	fi

## DB Stop: Stop DynamoDB local
db-stop:
	@echo "$(YELLOW)Stopping DynamoDB local...$(RESET)"
	@if docker ps | grep -q dynamodb-local 2>/dev/null; then \
		docker stop dynamodb-local >/dev/null 2>&1; \
		docker rm dynamodb-local >/dev/null 2>&1; \
		echo "$(GREEN)✓ DynamoDB local stopped$(RESET)"; \
	else \
		echo "$(BLUE)DynamoDB local not running$(RESET)"; \
	fi

## DB Reset: Reset local database
db-reset: db-stop db-start
	@echo "$(YELLOW)Waiting for DynamoDB to be ready...$(RESET)"
	@sleep 3
	@echo "$(GREEN)✓ Database reset completed$(RESET)"

## DB Ensure: Ensure DynamoDB is running
db-ensure:
	@if ! curl -s http://localhost:8000 >/dev/null 2>&1; then \
		echo "$(YELLOW)DynamoDB not running, starting...$(RESET)"; \
		make db-start; \
		sleep 2; \
	fi

## Format: Format all Go code
fmt:
	@echo "$(YELLOW)Formatting code...$(RESET)"
	go fmt ./...
	@echo "$(GREEN)✓ Code formatted$(RESET)"

## Clean: Remove build artifacts
clean:
	@echo "$(YELLOW)Cleaning up...$(RESET)"
	rm -rf bin/
	go clean -cache
	@echo "$(GREEN)✓ Cleanup completed$(RESET)"

## Demo: Show project capabilities
demo: db-ensure
	@echo "$(BLUE)Stori Expense Tracker - Demo$(RESET)"
	@echo "$(YELLOW)Testing components...$(RESET)"
	@echo ""
	@echo "$(GREEN)Project Features:$(RESET)"
	@echo "  • Go backend with Lambda functions"
	@echo "  • DynamoDB single-table design with local/AWS support"
	@echo "  • Budget management system"
	@echo "  • AI financial advisor integration"
	@echo "  • Comprehensive test suite"
	@echo "  • Repository pattern with pagination support"
	@echo ""
	@echo "$(YELLOW)Next steps:$(RESET)"
	@echo "  1. make dev     - Start development server"
	@echo "  2. make seed    - Load sample data"
	@echo "  3. make test    - Run test suite"
	@echo ""
