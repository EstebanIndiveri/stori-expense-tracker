#!/bin/bash

# Test runner script for Stori Expense Tracker Backend

set -e

echo "🚀 Starting Stori Expense Tracker Test Suite"
echo "============================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
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

# Default values
RUN_UNIT=true
RUN_INTEGRATION=false
RUN_SMOKE=false
COVERAGE=false
VERBOSE=false
CLEANUP=true

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --unit)
            RUN_UNIT=true
            RUN_INTEGRATION=false
            RUN_SMOKE=false
            shift
            ;;
        --integration)
            RUN_INTEGRATION=true
            RUN_UNIT=false
            RUN_SMOKE=false
            shift
            ;;
        --smoke)
            RUN_SMOKE=true
            RUN_UNIT=false
            RUN_INTEGRATION=false
            shift
            ;;
        --all)
            RUN_UNIT=true
            RUN_INTEGRATION=true
            RUN_SMOKE=true
            shift
            ;;
        --coverage)
            COVERAGE=true
            shift
            ;;
        --verbose)
            VERBOSE=true
            shift
            ;;
        --no-cleanup)
            CLEANUP=false
            shift
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --unit         Run only unit tests (default)"
            echo "  --integration  Run only integration tests"
            echo "  --smoke        Run only smoke tests"
            echo "  --all          Run all test types"
            echo "  --coverage     Generate test coverage report"
            echo "  --verbose      Verbose output"
            echo "  --no-cleanup   Skip cleanup after tests"
            echo "  --help         Show this help message"
            echo ""
            echo "Environment Variables:"
            echo "  DYNAMODB_ENDPOINT  DynamoDB Local endpoint (default: http://localhost:8000)"
            echo "  API_BASE_URL       API base URL for smoke tests"
            echo "  OPENAI_API_KEY     OpenAI API key for AI service tests"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Change to backend directory
cd "$(dirname "$0")/.."

# Check if go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed or not in PATH"
    exit 1
fi

print_status "Go version: $(go version)"

# Install test dependencies
print_status "Installing test dependencies..."
go mod tidy
go mod download

# Create test output directory
mkdir -p test-results

# Coverage file
COVERAGE_FILE="test-results/coverage.out"
COVERAGE_HTML="test-results/coverage.html"

# Test flags
TEST_FLAGS=""
if [[ "$VERBOSE" == "true" ]]; then
    TEST_FLAGS="$TEST_FLAGS -v"
fi

if [[ "$COVERAGE" == "true" ]]; then
    TEST_FLAGS="$TEST_FLAGS -coverprofile=$COVERAGE_FILE"
fi

# Function to run unit tests
run_unit_tests() {
    print_status "Running unit tests..."
    
    if [[ "$COVERAGE" == "true" ]]; then
        go test $TEST_FLAGS ./tests/unit/... || {
            print_error "Unit tests failed"
            return 1
        }
    else
        go test $TEST_FLAGS ./tests/unit/... || {
            print_error "Unit tests failed"
            return 1
        }
    fi
    
    print_success "Unit tests completed successfully"
}

# Function to run integration tests
run_integration_tests() {
    print_status "Running integration tests..."
    
    # Check if DynamoDB Local is running
    DYNAMODB_ENDPOINT=${DYNAMODB_ENDPOINT:-http://localhost:8000}
    
    print_status "Checking DynamoDB Local at $DYNAMODB_ENDPOINT..."
    
    if ! curl -s "$DYNAMODB_ENDPOINT" > /dev/null 2>&1; then
        print_warning "DynamoDB Local not running. Starting with Docker..."
        
        # Check if docker is available
        if ! command -v docker &> /dev/null; then
            print_error "Docker is not installed. Please install Docker or start DynamoDB Local manually."
            print_error "To start DynamoDB Local: docker run -p 8000:8000 amazon/dynamodb-local"
            return 1
        fi
        
        # Start DynamoDB Local
        docker run -d --name dynamodb-local-test -p 8000:8000 amazon/dynamodb-local > /dev/null 2>&1 || {
            print_warning "Failed to start DynamoDB Local with Docker. Attempting to use existing container..."
            docker start dynamodb-local-test > /dev/null 2>&1 || {
                print_error "Could not start DynamoDB Local"
                return 1
            }
        }
        
        # Wait for DynamoDB Local to be ready
        print_status "Waiting for DynamoDB Local to be ready..."
        for i in {1..30}; do
            if curl -s "$DYNAMODB_ENDPOINT" > /dev/null 2>&1; then
                break
            fi
            sleep 1
        done
        
        if ! curl -s "$DYNAMODB_ENDPOINT" > /dev/null 2>&1; then
            print_error "DynamoDB Local failed to start"
            return 1
        fi
        
        print_success "DynamoDB Local is running"
    else
        print_success "DynamoDB Local is already running"
    fi
    
    # Set environment variable for tests
    export DYNAMODB_ENDPOINT="$DYNAMODB_ENDPOINT"
    
    # Run integration tests
    go test $TEST_FLAGS ./tests/integration/... || {
        print_error "Integration tests failed"
        return 1
    }
    
    print_success "Integration tests completed successfully"
    
    # Cleanup DynamoDB Local if we started it
    if [[ "$CLEANUP" == "true" ]]; then
        print_status "Cleaning up DynamoDB Local..."
        docker stop dynamodb-local-test > /dev/null 2>&1 || true
        docker rm dynamodb-local-test > /dev/null 2>&1 || true
    fi
}

# Function to run smoke tests
run_smoke_tests() {
    print_status "Running smoke tests..."
    
    # Check if API is running
    API_BASE_URL=${API_BASE_URL:-http://localhost:8080}
    
    print_status "Testing API availability at $API_BASE_URL..."
    
    if ! curl -s -f "$API_BASE_URL/health" > /dev/null 2>&1; then
        print_error "API is not running at $API_BASE_URL"
        print_error "Please start the API server or set API_BASE_URL environment variable"
        return 1
    fi
    
    print_success "API is available"
    
    # Set environment variables for tests
    export API_BASE_URL="$API_BASE_URL"
    export RUN_SMOKE_TESTS="true"
    
    # Run smoke tests
    go test $TEST_FLAGS ./tests/smoke/... || {
        print_error "Smoke tests failed"
        return 1
    }
    
    print_success "Smoke tests completed successfully"
}

# Function to generate coverage report
generate_coverage_report() {
    if [[ "$COVERAGE" == "true" && -f "$COVERAGE_FILE" ]]; then
        print_status "Generating coverage report..."
        
        # Generate HTML coverage report
        go tool cover -html="$COVERAGE_FILE" -o "$COVERAGE_HTML"
        
        # Display coverage summary
        echo ""
        echo "Coverage Summary:"
        echo "=================="
        go tool cover -func="$COVERAGE_FILE" | tail -1
        
        print_success "Coverage report generated: $COVERAGE_HTML"
    fi
}

# Main execution
main() {
    local exit_code=0
    
    if [[ "$RUN_UNIT" == "true" ]]; then
        run_unit_tests || exit_code=$?
    fi
    
    if [[ "$RUN_INTEGRATION" == "true" ]]; then
        run_integration_tests || exit_code=$?
    fi
    
    if [[ "$RUN_SMOKE" == "true" ]]; then
        run_smoke_tests || exit_code=$?
    fi
    
    # Generate coverage report if requested
    generate_coverage_report
    
    # Final status
    echo ""
    echo "============================================="
    if [[ $exit_code -eq 0 ]]; then
        print_success "All tests completed successfully! 🎉"
    else
        print_error "Some tests failed! ❌"
    fi
    echo "============================================="
    
    return $exit_code
}

# Trap to ensure cleanup happens
trap 'if [[ "$CLEANUP" == "true" ]]; then docker stop dynamodb-local-test > /dev/null 2>&1 || true; docker rm dynamodb-local-test > /dev/null 2>&1 || true; fi' EXIT

# Run main function
main
exit $?
