package services

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"backend/internal/models"
	"backend/internal/services"
	"backend/tests/mocks"
)

func TestTransactionService_CreateTransaction(t *testing.T) {
	mockTransaction := &models.Transaction{
		ID:          uuid.New().String(),
		UserID:      "user123",
		Amount:      100.50,
		Type:        "expense",
		Category:    "food",
		Description: "Lunch",
		Date:        time.Now(),
	}

	tests := []struct {
		name        string
		transaction *models.Transaction
		mockSetup   func(*mocks.MockRepository)
		expectError bool
	}{
		{
			name:        "successful transaction creation",
			transaction: mockTransaction,
			mockSetup: func(repo *mocks.MockRepository) {
				repo.On("CreateTransaction", mock.Anything, mock.AnythingOfType("*models.Transaction")).Return(nil)
			},
			expectError: false,
		},
		{
			name:        "nil transaction",
			transaction: nil,
			mockSetup:   func(repo *mocks.MockRepository) {},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockRepository()
			tt.mockSetup(mockRepo)

			service := services.NewTransactionService(mockRepo)
			ctx := context.Background()
			
			err := service.CreateTransaction(ctx, tt.transaction)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestTransactionService_GetTransactionsByUser(t *testing.T) {
	userID := "user123"
	mockTransactions := []models.Transaction{
		{
			ID:          uuid.New().String(),
			UserID:      userID,
			Amount:      100.50,
			Type:        "expense",
			Category:    "food",
			Description: "Lunch",
			Date:        time.Now(),
		},
	}

	tests := []struct {
		name        string
		userID      string
		limit       int
		mockSetup   func(*mocks.MockRepository)
		expectError bool
	}{
		{
			name:   "successful retrieval",
			userID: userID,
			limit:  10,
			mockSetup: func(repo *mocks.MockRepository) {
				repo.On("GetTransactionsByUser", mock.Anything, userID, 10, mock.Anything).Return(mockTransactions, map[string]types.AttributeValue{}, nil)
			},
			expectError: false,
		},
		{
			name:        "empty user ID",
			userID:      "",
			limit:       10,
			mockSetup:   func(repo *mocks.MockRepository) {},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockRepository()
			tt.mockSetup(mockRepo)

			service := services.NewTransactionService(mockRepo)
			ctx := context.Background()
			
			result, err := service.GetTransactionsByUser(ctx, tt.userID, tt.limit)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result, len(mockTransactions))
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestTransactionService_GetTransactionsByCategory(t *testing.T) {
	userID := "user123"
	category := "food"
	
	mockTransactions := []models.Transaction{
		{
			ID:          uuid.New().String(),
			UserID:      userID,
			Amount:      100.50,
			Type:        "expense",
			Category:    category,
			Description: "Lunch",
			Date:        time.Now(),
		},
	}

	tests := []struct {
		name        string
		userID      string
		category    string
		limit       int
		mockSetup   func(*mocks.MockRepository)
		expectError bool
	}{
		{
			name:     "successful retrieval",
			userID:   userID,
			category: category,
			limit:    10,
			mockSetup: func(repo *mocks.MockRepository) {
				repo.On("GetTransactionsByCategory", mock.Anything, userID, category, 10, mock.Anything).Return(mockTransactions, map[string]types.AttributeValue{}, nil)
			},
			expectError: false,
		},
		{
			name:        "empty user ID",
			userID:      "",
			category:    category,
			limit:       10,
			mockSetup:   func(repo *mocks.MockRepository) {},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockRepository()
			tt.mockSetup(mockRepo)

			service := services.NewTransactionService(mockRepo)
			ctx := context.Background()
			
			result, err := service.GetTransactionsByCategory(ctx, tt.userID, tt.category, tt.limit)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result, len(mockTransactions))
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
