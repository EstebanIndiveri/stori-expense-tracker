# Backend Go API

This package contains the Go backend implementation for the Stori Expense Tracker.

## 🏗️ Architecture

The backend follows Clean Architecture principles with the following layers:

```
cmd/                 # Application entry points (Lambda handlers)
├── api/            # Main API Lambda function
└── ai-advisor/     # AI advisor Lambda function

internal/           # Private application code
├── config/         # Configuration management
├── handlers/       # HTTP handlers (API layer)
├── services/       # Business logic (service layer)
├── repository/     # Data access (repository layer)
└── models/         # Data models and DTOs

tests/              # Test files
├── integration/    # Integration tests
├── smoke/          # Smoke tests
└── mocks/          # Test mocks
```

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- AWS CLI configured
- DynamoDB Local (for development)

### Development Setup

```bash
# Install dependencies
make deps-install

# Copy environment variables
cp .env.example .env

# Start DynamoDB Local
docker run -p 8000:8000 amazon/dynamodb-local

# Setup local database
make setup-db

# Seed sample data
make seed-data

# Start development server
make dev
```

### Running Tests

```bash
# Unit tests
make test

# Integration tests
make test-integration

# All tests with coverage
make coverage
```

## 📊 API Endpoints

### Transactions API

- `GET /api/v1/transactions` - Get transactions with filtering
- `POST /api/v1/transactions` - Create new transaction

### Analytics API

- `GET /api/v1/analytics/summary` - Get financial summary
- `GET /api/v1/analytics/timeline` - Get timeline data
- `GET /api/v1/analytics/categories` - Get category breakdown

### AI Advisor API

- `POST /api/v1/ai/advice` - Get AI financial advice

## 🗄️ Database Design

Uses DynamoDB with optimized single-table design:

### Primary Table: `transactions`

```
PK: DS#v1#M#2024-01          (Dataset + Version + Month)
SK: D#2024-01-15#TX#uuid     (Date + Transaction + ID)

GSI1PK: CAT#dining#M#2024-01 (Category + Month)
GSI1SK: D#2024-01-15#TX#uuid (Date + Transaction + ID)

GSI2PK: T#expense#M#2024-01  (Type + Month)
GSI2SK: D#2024-01-15#TX#uuid (Date + Transaction + ID)
```

### Access Patterns

1. **Get transactions by date range**: Query PK by month
2. **Get transactions by category**: Query GSI1 by category+month
3. **Get transactions by type**: Query GSI2 by type+month
4. **Analytics aggregations**: Performed in application layer

## 🤖 AI Integration

Uses OpenAI GPT-3.5 Turbo for financial advice:

- Context-aware prompts based on user's financial data
- Structured responses with actionable suggestions
- Fallback mechanisms for API failures
- Rate limiting and cost optimization

## 🔧 Configuration

Environment variables:

```env
# AWS Configuration
AWS_REGION=us-east-1
DYNAMODB_TABLE_NAME=stori-transactions-dev
DYNAMODB_ENDPOINT=http://localhost:8000  # For local development

# OpenAI Configuration
OPENAI_API_KEY_SSM=/stori/dev/openai-api-key

# Application Configuration
ENVIRONMENT=dev
LOG_LEVEL=debug
PORT=8080
```

## 🧪 Testing Strategy

### Unit Tests
- Repository layer with DynamoDB mocks
- Service layer with repository mocks
- Handler layer with service mocks

### Integration Tests
- End-to-end API testing with DynamoDB Local
- AI service integration testing
- Error handling and edge cases

### Smoke Tests
- Health checks against deployed environments
- Critical path validation
- Performance benchmarks

## 🚀 Deployment

### Manual Deployment

```bash
# Build binaries
make build

# Deploy to staging
make deploy-staging

# Deploy to production
make deploy-prod
```

### CI/CD Pipeline

Automated deployment via GitHub Actions:

1. **Code Quality**: Linting, formatting, security scans
2. **Testing**: Unit, integration, and smoke tests
3. **Building**: Cross-compilation for Lambda
4. **Deployment**: Environment-specific deployments
5. **Monitoring**: Post-deployment health checks

## 📊 Monitoring

### CloudWatch Metrics

- Lambda execution duration and errors
- DynamoDB read/write capacity and throttles
- Custom business metrics

### Logging

- Structured JSON logging
- Request/response tracing
- Error tracking with context

### Alerting

- High error rates
- Performance degradation
- Cost anomalies

## 🔒 Security

### Best Practices

- IAM least privilege principles
- Secrets management via AWS SSM
- Input validation and sanitization
- CORS configuration
- Rate limiting

### Compliance

- Data encryption at rest (DynamoDB)
- Data encryption in transit (HTTPS)
- Audit logging
- Access controls

## 🏗️ Development

### Code Style

- Follow Go idioms and conventions
- Use `gofmt` and `goimports`
- Comprehensive error handling
- Meaningful variable and function names

### Adding New Features

1. Define models in `internal/models/`
2. Implement repository methods
3. Add business logic in services
4. Create HTTP handlers
5. Add comprehensive tests
6. Update API documentation

### Performance Optimization

- Use DynamoDB query patterns efficiently
- Implement caching where appropriate
- Optimize Lambda cold starts
- Monitor and optimize costs

## 📚 Additional Resources

- [AWS Lambda Go Runtime](https://docs.aws.amazon.com/lambda/latest/dg/lambda-golang.html)
- [DynamoDB Best Practices](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/best-practices.html)
- [OpenAI API Documentation](https://platform.openai.com/docs)
- [Go Testing Best Practices](https://go.dev/doc/tutorial/add-a-test)
