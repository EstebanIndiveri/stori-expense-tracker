package services

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"backend/internal/models"
	"backend/internal/services"
	"backend/tests/mocks"
)

func TestAnalyticsService_GetMonthlyAnalytics(t *testing.T) {
	userID := "user123"
	month := "2024-01"
	mockAnalytics := &models.MonthlyAnalytics{
		Month:             month,
		TotalIncome:       5000.0,
		TotalExpense:      3500.0,
		Balance:           1500.0,
		CategoryBreakdown: map[string]float64{
			"food":           800.0,
			"transportation": 400.0,
			"entertainment":  300.0,
			"salary":         5000.0,
		},
		TransactionCount: 10,
	}

	tests := []struct {
		name          string
		userID        string
		month         string
		mockSetup     func(*mocks.MockRepository)
		expectedError error
		validateResult func(*testing.T, *models.MonthlyAnalytics)
	}{
		{
			name:   "successful analytics retrieval",
			userID: userID,
			month:  month,
			mockSetup: func(repo *mocks.MockRepository) {
				repo.On("GetMonthlyAnalytics", mock.Anything, userID, month).Return(mockAnalytics, nil)
			},
			expectedError: nil,
			validateResult: func(t *testing.T, result *models.MonthlyAnalytics) {
				assert.Equal(t, month, result.Month)
				assert.Equal(t, 5000.0, result.TotalIncome)
				assert.Equal(t, 3500.0, result.TotalExpense)
				assert.Equal(t, 1500.0, result.Balance)
				assert.Equal(t, 10, result.TransactionCount)
			},
		},
		{
			name:   "repository error",
			userID: userID,
			month:  month,
			mockSetup: func(repo *mocks.MockRepository) {
				repo.On("GetMonthlyAnalytics", mock.Anything, userID, month).Return(nil, errors.New("database error"))
			},
			expectedError: errors.New("database error"),
		},
		{
			name:          "missing user ID",
			userID:        "",
			month:         month,
			mockSetup:     func(repo *mocks.MockRepository) {},
			expectedError: errors.New("userID and month are required"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockRepository()
			tt.mockSetup(mockRepo)

			service := services.NewAnalyticsService(mockRepo)
			ctx := context.Background()

			result, err := service.GetMonthlyAnalytics(ctx, tt.userID, tt.month)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAnalyticsService_GetFinancialSummary(t *testing.T) {
	userID := "user123"
	// mockSummary := &models.FinancialSummary{
	// 	TotalBalance:    1500.0,
	// 	MonthlyIncome:   5000.0,
	// 	MonthlyExpenses: 3500.0,
	// 	AvailableMoney:  1500.0,
	// 	CategoryBreakdown: []models.CategoryBreakdown{
	// 		{Category: "food", Amount: 800.0, Percentage: 22.86, Count: 5},
	// 		{Category: "transportation", Amount: 400.0, Percentage: 11.43, Count: 3},
	// 	},
	// 	MonthlyTrend: 5.2,
	// }

	tests := []struct {
		name          string
		userID        string
		mockSetup     func(*mocks.MockRepository)
		expectedError error
	}{
		{
			name:   "successful financial summary",
			userID: userID,
			mockSetup: func(repo *mocks.MockRepository) {
				// Mock GetTransactionsByUser (the only method called now)
				mockTransactions := []models.Transaction{
					{
						ID:       "tx1",
						UserID:   userID,
						Amount:   2800.0,
						Type:     "income",
						Category: "salary",
					},
					{
						ID:       "tx2",
						UserID:   userID,
						Amount:   -100.0,
						Type:     "expense",
						Category: "food",
					},
				}
				repo.On("GetTransactionsByUser", mock.Anything, userID, 1000, mock.Anything).Return(mockTransactions, map[string]types.AttributeValue{}, nil)
			},
			expectedError: nil,
		},
		{
			name:          "missing user ID",
			userID:        "",
			mockSetup:     func(repo *mocks.MockRepository) {},
			expectedError: errors.New("userID is required"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockRepository()
			tt.mockSetup(mockRepo)

			service := services.NewAnalyticsService(mockRepo)
			ctx := context.Background()

			result, err := service.GetFinancialSummary(ctx, tt.userID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
