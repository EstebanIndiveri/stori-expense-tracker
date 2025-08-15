package mocks

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/mock"

	"backend/internal/models"
)

// MockRepository is a mock implementation of the Repository interface
type MockRepository struct {
	mock.Mock
}

// Transaction operations
func (m *MockRepository) CreateTransaction(ctx context.Context, transaction *models.Transaction) error {
	args := m.Called(ctx, transaction)
	return args.Error(0)
}

func (m *MockRepository) GetTransaction(ctx context.Context, userID, transactionID string) (*models.Transaction, error) {
	args := m.Called(ctx, userID, transactionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transaction), args.Error(1)
}

func (m *MockRepository) UpdateTransaction(ctx context.Context, transaction *models.Transaction) error {
	args := m.Called(ctx, transaction)
	return args.Error(0)
}

func (m *MockRepository) DeleteTransaction(ctx context.Context, userID, transactionID string) error {
	args := m.Called(ctx, userID, transactionID)
	return args.Error(0)
}

// Query operations with optimized access patterns
func (m *MockRepository) GetTransactionsByUser(ctx context.Context, userID string, limit int, lastKey map[string]types.AttributeValue) ([]models.Transaction, map[string]types.AttributeValue, error) {
	args := m.Called(ctx, userID, limit, lastKey)
	if args.Get(0) == nil {
		var nextKey map[string]types.AttributeValue
		if args.Get(1) != nil {
			nextKey = args.Get(1).(map[string]types.AttributeValue)
		}
		return nil, nextKey, args.Error(2)
	}
	var nextKey map[string]types.AttributeValue
	if args.Get(1) != nil {
		nextKey = args.Get(1).(map[string]types.AttributeValue)
	}
	return args.Get(0).([]models.Transaction), nextKey, args.Error(2)
}

func (m *MockRepository) GetTransactionsByMonth(ctx context.Context, userID string, month string, limit int, lastKey map[string]types.AttributeValue) ([]models.Transaction, map[string]types.AttributeValue, error) {
	args := m.Called(ctx, userID, month, limit, lastKey)
	if args.Get(0) == nil {
		var nextKey map[string]types.AttributeValue
		if args.Get(1) != nil {
			nextKey = args.Get(1).(map[string]types.AttributeValue)
		}
		return nil, nextKey, args.Error(2)
	}
	var nextKey map[string]types.AttributeValue
	if args.Get(1) != nil {
		nextKey = args.Get(1).(map[string]types.AttributeValue)
	}
	return args.Get(0).([]models.Transaction), nextKey, args.Error(2)
}

func (m *MockRepository) GetTransactionsByCategory(ctx context.Context, userID string, category string, limit int, lastKey map[string]types.AttributeValue) ([]models.Transaction, map[string]types.AttributeValue, error) {
	args := m.Called(ctx, userID, category, limit, lastKey)
	if args.Get(0) == nil {
		var nextKey map[string]types.AttributeValue
		if args.Get(1) != nil {
			nextKey = args.Get(1).(map[string]types.AttributeValue)
		}
		return nil, nextKey, args.Error(2)
	}
	var nextKey map[string]types.AttributeValue
	if args.Get(1) != nil {
		nextKey = args.Get(1).(map[string]types.AttributeValue)
	}
	return args.Get(0).([]models.Transaction), nextKey, args.Error(2)
}

// Batch operations
func (m *MockRepository) BatchCreateTransactions(ctx context.Context, transactions []models.Transaction) error {
	args := m.Called(ctx, transactions)
	return args.Error(0)
}

// Analytics
func (m *MockRepository) GetMonthlyAnalytics(ctx context.Context, userID string, month string) (*models.MonthlyAnalytics, error) {
	args := m.Called(ctx, userID, month)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.MonthlyAnalytics), args.Error(1)
}

// Budget operations
func (m *MockRepository) CreateOrUpdateBudget(ctx context.Context, budget *models.Budget) error {
	args := m.Called(ctx, budget)
	return args.Error(0)
}

func (m *MockRepository) GetBudgetsByMonth(ctx context.Context, userID, month string) ([]models.Budget, error) {
	args := m.Called(ctx, userID, month)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Budget), args.Error(1)
}

func (m *MockRepository) GetBudget(ctx context.Context, userID, month, category string) (*models.Budget, error) {
	args := m.Called(ctx, userID, month, category)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Budget), args.Error(1)
}

func (m *MockRepository) DeleteBudget(ctx context.Context, userID, month, category string) error {
	args := m.Called(ctx, userID, month, category)
	return args.Error(0)
}

// User operations
func (m *MockRepository) CreateUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRepository) GetUser(ctx context.Context, userID string) (*models.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// MockOpenAIClient is a mock implementation of the OpenAI client
type MockOpenAIClient struct {
	mock.Mock
}

func (m *MockOpenAIClient) CreateChatCompletion(ctx context.Context, prompt string) (string, error) {
	args := m.Called(ctx, prompt)
	return args.String(0), args.Error(1)
}

// NewMockRepository creates a new mock repository
func NewMockRepository() *MockRepository {
	return &MockRepository{}
}

// NewMockOpenAIClient creates a new mock OpenAI client
func NewMockOpenAIClient() *MockOpenAIClient {
	return &MockOpenAIClient{}
}
