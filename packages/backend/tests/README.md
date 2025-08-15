# Testing Guide

## Overview

This document describes the comprehensive testing strategy for the Stori Expense Tracker backend, including unit tests, integration tests, and smoke tests.

## Test Structure

```
tests/
├── unit/                 # Unit tests for individual components
│   ├── transaction_service_test.go
│   ├── analytics_service_test.go
│   └── ai_service_test.go
├── integration/          # Integration tests with external services
│   └── dynamodb_test.go
├── smoke/               # End-to-end smoke tests
│   └── api_test.go
└── mocks/               # Mock implementations
    └── mocks.go
```

## Test Types

### 1. Unit Tests

Test individual components in isolation using mocks for dependencies.

**Coverage:**

- Service layer business logic
- Input validation
- Error handling
- Data transformation

**Run unit tests:**

```bash
make test-unit
# or
./scripts/test.sh --unit
```

### 2. Integration Tests

Test interactions with external services like DynamoDB.

**Requirements:**

- DynamoDB Local running on port 8000
- Docker (for automatic DynamoDB Local setup)

**Run integration tests:**

```bash
make test-integration
# or
./scripts/test.sh --integration
```

### 3. Smoke Tests

End-to-end tests against a running API server.

**Requirements:**

- API server running (default: http://localhost:8080)
- Environment variable `API_BASE_URL` (optional)

**Run smoke tests:**

```bash
make test-smoke
# or
./scripts/test.sh --smoke
```

## Quick Start

### Setup Development Environment

```bash
# Install dependencies and start infrastructure
make setup

# Run unit tests
make test

# Run all tests with coverage
make test-coverage-all
```

### Test Script Usage

The `scripts/test.sh` script provides comprehensive testing capabilities:

```bash
# Run only unit tests (default)
./scripts/test.sh

# Run specific test types
./scripts/test.sh --unit
./scripts/test.sh --integration
./scripts/test.sh --smoke

# Run all test types
./scripts/test.sh --all

# Generate coverage report
./scripts/test.sh --unit --coverage

# Verbose output
./scripts/test.sh --all --verbose

# Skip cleanup (useful for debugging)
./scripts/test.sh --integration --no-cleanup
```

## Environment Variables

### Integration Tests

- `DYNAMODB_ENDPOINT` - DynamoDB Local endpoint (default: http://localhost:8000)

### Smoke Tests

- `API_BASE_URL` - API server URL (default: http://localhost:8080)
- `RUN_SMOKE_TESTS` - Enable smoke tests when API_BASE_URL is not set

### AI Service Tests

- `OPENAI_API_KEY` - OpenAI API key (required for AI service tests)

## Make Commands

### Basic Testing

```bash
make test              # Run unit tests
make test-unit         # Run unit tests with verbose output
make test-integration  # Run integration tests
make test-smoke        # Run smoke tests
make test-all          # Run all test types
```

### Coverage Reports

```bash
make test-coverage     # Unit tests with coverage
make test-coverage-all # All tests with coverage
```

### Quick Tests

```bash
make test-quick        # Fast unit tests only
make test-services     # Service layer tests only
make test-handlers     # Handler layer tests only
```

### Code Quality

```bash
make check             # Run fmt, vet, and lint
make lint              # Run linter
make fmt               # Format code
make vet               # Run go vet
```

## Infrastructure Setup

### DynamoDB Local

**Using Make:**

```bash
make infra-start       # Start DynamoDB Local
make infra-stop        # Stop DynamoDB Local
make infra-restart     # Restart DynamoDB Local
make infra-logs        # View logs
```

**Manual Setup:**

```bash
docker run -d --name dynamodb-local -p 8000:8000 amazon/dynamodb-local
```

### API Server

**For Development:**

```bash
make dev               # Start development server
```

**For Testing:**

```bash
make run-api           # Build and run production binary
```

## Continuous Integration

### GitHub Actions

The project includes CI workflows that run:

- Unit tests on every push/PR
- Integration tests with DynamoDB Local
- Code quality checks (linting, formatting)
- Security scans

### Local CI Simulation

```bash
make ci                # Run full CI pipeline locally
make pre-commit        # Run pre-commit checks
```

## Test Data and Mocking

### Mock Implementations

The `tests/mocks/` directory contains mock implementations for:

- Repository interface
- OpenAI client interface

### Test Data

Tests use:

- Generated UUIDs for unique identifiers
- Realistic financial data
- Edge cases (zero amounts, empty strings, etc.)

## Coverage Reports

Coverage reports are generated in HTML format:

```bash
# Generate coverage report
make test-coverage

# View report
open test-results/coverage.html
```

**Coverage Targets:**

- Service layer: > 90%
- Handler layer: > 80%
- Repository layer: > 85%

## Debugging Tests

### Verbose Output

```bash
./scripts/test.sh --all --verbose
```

### Individual Test Files

```bash
go test -v ./tests/unit/transaction_service_test.go
```

### Race Condition Detection

```bash
go test -race ./tests/unit/...
```

### Benchmarking

```bash
make bench
```

## Best Practices

### Writing Tests

1. **Test Naming:** Use descriptive test names that explain the scenario
2. **Test Structure:** Follow Arrange-Act-Assert pattern
3. **Mocking:** Mock external dependencies, test business logic
4. **Data:** Use realistic test data, cover edge cases
5. **Cleanup:** Clean up resources after tests

### Example Test Structure

```go
func TestServiceMethod_Scenario(t *testing.T) {
    // Arrange
    mockRepo := new(MockRepository)
    service := NewService(mockRepo)

    // Setup expectations
    mockRepo.On("Method", mock.Anything).Return(expectedResult, nil)

    // Act
    result, err := service.Method(input)

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, expectedResult, result)
    mockRepo.AssertExpectations(t)
}
```

### Performance Testing

For performance-critical code:

```go
func BenchmarkServiceMethod(b *testing.B) {
    service := NewService(mockRepo)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        service.Method(input)
    }
}
```

## Troubleshooting

### Common Issues

1. **DynamoDB Local Connection Failed**

   - Ensure Docker is running
   - Check port 8000 is available
   - Verify container is healthy: `docker ps`

2. **Integration Tests Timeout**

   - Increase timeout in test configuration
   - Check DynamoDB Local responsiveness

3. **Smoke Tests Fail**

   - Verify API server is running
   - Check API_BASE_URL environment variable
   - Ensure proper authentication headers

4. **Coverage Report Not Generated**
   - Run tests with `--coverage` flag
   - Check file permissions in test-results directory

### Debug Commands

```bash
# Check DynamoDB Local status
curl http://localhost:8000

# Check API health
curl http://localhost:8080/health

# View test logs
./scripts/test.sh --all --verbose --no-cleanup

# Manual test execution
go test -v -run TestSpecificTest ./tests/unit/
```

## Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Testify Framework](https://github.com/stretchr/testify)
- [DynamoDB Local](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/DynamoDBLocal.html)
- [Go Test Coverage](https://go.dev/blog/cover)
