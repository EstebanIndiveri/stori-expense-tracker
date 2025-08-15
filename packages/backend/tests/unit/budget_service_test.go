package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"backend/internal/models"
	"backend/internal/services"
	"backend/tests/mocks"
)

func TestBudgetService_CreateOrUpdateBudget(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		month       string
		category    string
		amount      float64
		mockSetup   func(*mocks.MockRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name:     "successful budget creation",
			userID:   "user-123",
			month:    "2025-08",
			category: "Food",
			amount:   500.0,
			mockSetup: func(repo *mocks.MockRepository) {
				repo.On("CreateOrUpdateBudget", mock.Anything, mock.AnythingOfType("*models.Budget")).Return(nil)
			},
			expectError: false,
		},
		{
			name:        "negative amount should fail",
			userID:      "user-123",
			month:       "2025-08",
			category:    "Food",
			amount:      -100.0,
			mockSetup:   func(repo *mocks.MockRepository) {},
			expectError: true,
			errorMsg:    "budget amount cannot be negative",
		},
		{
			name:        "invalid month format should fail",
			userID:      "user-123",
			month:       "invalid-month",
			category:    "Food",
			amount:      500.0,
			mockSetup:   func(repo *mocks.MockRepository) {},
			expectError: true,
			errorMsg:    "invalid month format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockRepository()
			tt.mockSetup(mockRepo)

			service := services.NewBudgetService(mockRepo)
			ctx := context.Background()

			err := service.CreateOrUpdateBudget(ctx, tt.userID, tt.month, tt.category, tt.amount)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestBudgetService_GetBudgetsByMonth(t *testing.T) {
	userID := "user-123"
	month := "2025-08"
	mockBudgets := []models.Budget{
		{
			UserID:   userID,
			Month:    month,
			Category: "Food",
			Amount:   500.0,
		},
		{
			UserID:   userID,
			Month:    month,
			Category: "Transportation",
			Amount:   200.0,
		},
	}

	tests := []struct {
		name        string
		userID      string
		month       string
		mockSetup   func(*mocks.MockRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name:   "successful retrieval",
			userID: userID,
			month:  month,
			mockSetup: func(repo *mocks.MockRepository) {
				repo.On("GetBudgetsByMonth", mock.Anything, userID, month).Return(mockBudgets, nil)
			},
			expectError: false,
		},
		{
			name:        "invalid month format",
			userID:      userID,
			month:       "invalid-month",
			mockSetup:   func(repo *mocks.MockRepository) {},
			expectError: true,
			errorMsg:    "invalid month format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockRepository()
			tt.mockSetup(mockRepo)

			service := services.NewBudgetService(mockRepo)
			ctx := context.Background()

			result, err := service.GetBudgetsByMonth(ctx, tt.userID, tt.month)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, len(mockBudgets), len(result))
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestBudgetService_GetBudgetUtilization(t *testing.T) {
	userID := "user-123"
	month := "2025-08"
	
	mockBudgets := []models.Budget{
		{
			UserID:   userID,
			Month:    month,
			Category: "Food",
			Amount:   500.0,
		},
	}

	mockTransactions := []models.Transaction{
		{
			ID:       "tx-1",
			UserID:   userID,
			Amount:   -50.0,
			Type:     "expense",
			Category: "Food",
			Date:     time.Date(2025, 8, 15, 0, 0, 0, 0, time.UTC),
		},
	}

	tests := []struct {
		name        string
		userID      string
		month       string
		mockSetup   func(*mocks.MockRepository)
		expectError bool
	}{
		{
			name:   "successful utilization calculation",
			userID: userID,
			month:  month,
			mockSetup: func(repo *mocks.MockRepository) {
				// Mock GetBudgetsByMonth
				repo.On("GetBudgetsByMonth", mock.Anything, userID, month).Return(mockBudgets, nil)
				// Mock GetTransactionsByUser for utilization calculation - return nil for nextKey to stop pagination
				repo.On("GetTransactionsByUser", mock.Anything, userID, 1000, mock.Anything).Return(mockTransactions, nil, nil)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockRepository()
			tt.mockSetup(mockRepo)

			service := services.NewBudgetService(mockRepo)
			ctx := context.Background()

			result, err := service.GetBudgetUtilization(ctx, tt.userID, tt.month)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				
				// Check that Food category exists in the result
				foodUtil, exists := result["Food"]
				assert.True(t, exists)
				assert.Equal(t, 500.0, foodUtil.BudgetAmount)
				assert.Equal(t, 50.0, foodUtil.SpentAmount) // Amount should be positive (absolute value)
				assert.Equal(t, 450.0, foodUtil.Remaining)
				assert.Equal(t, 10.0, foodUtil.Percentage) // 50/500 * 100
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
