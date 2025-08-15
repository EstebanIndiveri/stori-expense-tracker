<!-- Use this file to provide workspace-specific custom instructions to Copilot. For more details, visit https://code.visualstudio.com/docs/copilot/copilot-customization#_use-a-githubcopilotinstructionsmd-file -->

# Stori Expense Tracker - Copilot Instructions

## Project Context
This is a full-stack expense tracking application for the Stori challenge with the following architecture:

- **Backend**: Go Lambda functions with DynamoDB
- **Frontend**: React with TypeScript and Webpack
- **Infrastructure**: Terraform for AWS resources
- **Deployment**: Independent CI/CD pipelines

## Code Standards

### Go Backend
- Use Go 1.21 features and idiomatic patterns
- Follow AWS Lambda best practices for handlers
- Implement proper error handling with structured logging
- Use dependency injection for testability
- Write comprehensive unit and integration tests
- Follow the repository pattern for data access

### React Frontend
- Use functional components with hooks
- Implement TypeScript strictly (no `any` types)
- Use Tailwind CSS for styling
- Follow atomic design principles
- Implement proper error boundaries
- Use React Query for API state management

### Infrastructure
- Use Terraform modules for reusability
- Follow AWS Well-Architected Framework
- Implement proper IAM least privilege
- Use tags consistently across resources
- Enable monitoring and logging for all services

### DynamoDB Design
- Use single-table design with optimized access patterns
- Design GSIs based on query requirements
- Use composite keys for efficient sorting
- Implement proper pagination with LastEvaluatedKey

## API Design
- Follow RESTful conventions
- Use proper HTTP status codes
- Implement consistent error response format
- Add request validation at API Gateway level
- Use OpenAPI documentation

## Security
- Never commit secrets or API keys
- Use AWS Systems Manager for configuration
- Implement CORS properly
- Add request rate limiting
- Use HTTPS everywhere

## Testing
- Aim for >80% code coverage
- Use table-driven tests in Go
- Mock external dependencies
- Use DynamoDB Local for integration tests
- Implement smoke tests for deployments

## AI Integration
- Use structured prompts for financial advice
- Implement context-aware responses
- Add fallback mechanisms for API failures
- Log AI interactions for analysis
