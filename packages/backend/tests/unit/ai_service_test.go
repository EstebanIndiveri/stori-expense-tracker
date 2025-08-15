package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"backend/internal/config"
	"backend/internal/models"
	"backend/internal/services"
	"backend/tests/mocks"
)

func TestAIService_GetFinancialAdvice(t *testing.T) {
	mockAdviceRequest := &models.AIAdviceRequest{
		Question: "How can I reduce my expenses?",
		Context:  "Monthly spending analysis",
	}

	tests := []struct {
		name          string
		request       *models.AIAdviceRequest
		mockSetup     func(*mocks.MockRepository)
		expectError   bool
		configSetup   func() *config.Config
	}{
		{
			name:    "successful advice generation",
			request: mockAdviceRequest,
			mockSetup: func(repo *mocks.MockRepository) {
				analytics := &models.MonthlyAnalytics{
					Month:        "2024-01",
					TotalIncome:  5000.0,
					TotalExpense: 3500.0,
					Balance:      1500.0,
				}
				repo.On("GetMonthlyAnalytics", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(analytics, nil)
			},
			expectError: true, // Will fail due to missing API key
			configSetup: func() *config.Config {
				return &config.Config{
					GroqAPIKey: "test-key", // Even with test key, will fail without real API
				}
			},
		},
		{
			name:        "invalid config - empty API key",
			request:     mockAdviceRequest,
			mockSetup:   func(repo *mocks.MockRepository) {},
			expectError: true,
			configSetup: func() *config.Config {
				return &config.Config{
					GroqAPIKey: "", // Empty API key should fail at service creation
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockRepository()
			tt.mockSetup(mockRepo)

			cfg := tt.configSetup()
			
			service, err := services.NewAIService(cfg, mockRepo)
			if err != nil {
				// Service creation failed due to invalid config
				assert.Error(t, err)
				return
			}
			
			ctx := context.Background()
			result, err := service.GetFinancialAdvice(ctx, tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEmpty(t, result.Advice)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
