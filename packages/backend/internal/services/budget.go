package services

import (
	"context"
	"fmt"
	"time"

	"backend/internal/models"
	"backend/internal/repository"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// BudgetService handles budget-related business logic
type BudgetService struct {
	repo repository.Repository
}

// NewBudgetService creates a new budget service
func NewBudgetService(repo repository.Repository) *BudgetService {
	return &BudgetService{
		repo: repo,
	}
}

// CreateOrUpdateBudget creates or updates a budget for a specific month and category
func (s *BudgetService) CreateOrUpdateBudget(ctx context.Context, userID, month, category string, amount float64) error {
	if amount < 0 {
		return fmt.Errorf("budget amount cannot be negative")
	}

	// Validate month format (YYYY-MM)
	if _, err := time.Parse("2006-01", month); err != nil {
		return fmt.Errorf("invalid month format, expected YYYY-MM: %w", err)
	}

	budget := &models.Budget{
		PK:       fmt.Sprintf("USER#%s", userID),
		SK:       fmt.Sprintf("BUDGET#%s#%s", month, category),
		UserID:   userID,
		Month:    month,
		Category: category,
		Amount:   amount,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return s.repo.CreateOrUpdateBudget(ctx, budget)
}

// GetBudgetsByMonth retrieves all budgets for a user in a specific month
func (s *BudgetService) GetBudgetsByMonth(ctx context.Context, userID, month string) ([]models.Budget, error) {
	// Validate month format
	if _, err := time.Parse("2006-01", month); err != nil {
		return nil, fmt.Errorf("invalid month format, expected YYYY-MM: %w", err)
	}

	return s.repo.GetBudgetsByMonth(ctx, userID, month)
}

// GetBudget retrieves a specific budget
func (s *BudgetService) GetBudget(ctx context.Context, userID, month, category string) (*models.Budget, error) {
	// Validate month format
	if _, err := time.Parse("2006-01", month); err != nil {
		return nil, fmt.Errorf("invalid month format, expected YYYY-MM: %w", err)
	}

	return s.repo.GetBudget(ctx, userID, month, category)
}

// DeleteBudget removes a budget
func (s *BudgetService) DeleteBudget(ctx context.Context, userID, month, category string) error {
	return s.repo.DeleteBudget(ctx, userID, month, category)
}

// GetBudgetUtilization calculates budget utilization for a specific month
func (s *BudgetService) GetBudgetUtilization(ctx context.Context, userID, month string) (map[string]models.BudgetUtilization, error) {
	// Get budgets for the month
	budgets, err := s.GetBudgetsByMonth(ctx, userID, month)
	if err != nil {
		return nil, fmt.Errorf("failed to get budgets: %w", err)
	}

	// Get transactions for the month to calculate spending
	startDate, endDate, err := getMonthDateRange(month)
	if err != nil {
		return nil, fmt.Errorf("invalid month format: %w", err)
	}

	// Use the GetTransactionsByUser method with pagination to get all transactions
	var allTransactions []models.Transaction
	var lastKey map[string]types.AttributeValue
	limit := 1000 // Large limit to get most transactions

	for {
		transactions, nextKey, err := s.repo.GetTransactionsByUser(ctx, userID, limit, lastKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get transactions: %w", err)
		}

		allTransactions = append(allTransactions, transactions...)

		// If no more pages, break
		if nextKey == nil {
			break
		}
		lastKey = nextKey
	}

	// Filter transactions by date range and calculate spending by category (only expenses)
	categorySpending := make(map[string]float64)
	for _, tx := range allTransactions {
		// Convert transaction date to string for comparison
		txDateStr := tx.Date.Format("2006-01-02")
		if tx.Type == "expense" && txDateStr >= startDate && txDateStr <= endDate {
			// For expenses, amount is negative, so we use absolute value for spending
			if tx.Amount < 0 {
				categorySpending[tx.Category] += -tx.Amount
			} else {
				categorySpending[tx.Category] += tx.Amount
			}
		}
	}

	// Create budget utilization map
	utilization := make(map[string]models.BudgetUtilization)
	for _, budget := range budgets {
		spent := categorySpending[budget.Category]
		percentage := 0.0
		if budget.Amount > 0 {
			percentage = (spent / budget.Amount) * 100
		}

		utilization[budget.Category] = models.BudgetUtilization{
			Category:     budget.Category,
			BudgetAmount: budget.Amount,
			SpentAmount:  spent,
			Remaining:    budget.Amount - spent,
			Percentage:   percentage,
		}
	}

	return utilization, nil
}

// getMonthDateRange converts a month string (YYYY-MM) to start and end dates
func getMonthDateRange(month string) (string, string, error) {
	t, err := time.Parse("2006-01", month)
	if err != nil {
		return "", "", err
	}

	// Start of month
	startOfMonth := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	startDate := startOfMonth.Format("2006-01-02")

	// End of month
	endOfMonth := startOfMonth.AddDate(0, 1, -1)
	endDate := endOfMonth.Format("2006-01-02")

	return startDate, endDate, nil
}
