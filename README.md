# Stori Expense Tracker

A modern, cloud-native expense tracking application built with Go backend, React frontend, and AWS infrastructure.

## 🏗️ Architecture

```
├── packages/
│   ├── backend/           # Go Lambda Functions & APIs
│   ├── frontend/          # React.js Application  
│   └── shared/            # Shared Types & Utils
├── infrastructure/        # Terraform Infrastructure
├── tools/                 # Data ingestion & Scripts
└── docs/                  # Documentation
```

## 🚀 Technologies

### Backend
- **Go 1.21**: High-performance Lambda functions
- **AWS Lambda**: Serverless compute
- **DynamoDB**: NoSQL database with optimized access patterns
- **API Gateway**: REST API management

### Frontend  
- **React 18**: Modern UI framework
- **TypeScript**: Type-safe development
- **Webpack**: Module bundling
- **Tailwind CSS**: Utility-first styling

### Infrastructure
- **Terraform**: Infrastructure as Code
- **AWS CloudFormation**: Resource orchestration
- **GitHub Actions**: CI/CD pipeline

## 🛠️ Development

### Prerequisites
- Go 1.21+
- Node.js 18+
- AWS CLI configured
- Terraform 1.6+

### Quick Start

```bash
# Install dependencies
go mod download
npm install

# Start local development
make dev-backend    # Starts local API server
make dev-frontend   # Starts React dev server

# Run tests
make test-backend
make test-frontend

# Deploy to AWS
make deploy-staging
make deploy-prod
```

## 📊 Features

- 📈 **Expense Categorization**: Automatic spending analysis
- 📅 **Timeline View**: Historical transaction tracking  
- 🤖 **AI Financial Advisor**: GPT-powered savings recommendations
- 📱 **Mobile-First**: Responsive design for all devices
- 🔐 **Secure**: Enterprise-grade security practices

## 🧪 Testing

- **Unit Tests**: Go & Jest
- **Integration Tests**: DynamoDB Local
- **E2E Tests**: Playwright
- **Load Tests**: Artillery.js

## 📈 Monitoring

- **CloudWatch**: Metrics & Logs
- **X-Ray**: Distributed tracing
- **Custom Dashboards**: Business KPIs

## 🚀 Deployment

Supports independent deployment of:
- Frontend (S3 + CloudFront)
- Backend (Lambda + API Gateway)  
- Infrastructure (Terraform)

## 📝 Documentation

- [API Documentation](./docs/api.md)
- [Architecture Guide](./docs/architecture.md)
- [Deployment Guide](./docs/deployment.md)
- [Contributing Guide](./CONTRIBUTING.md)

---

Built with ❤️ for the Stori Full Stack Challenge
