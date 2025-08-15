package services

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"backend/internal/models"
	"backend/internal/repository"
)

type AnalyticsService interface {
	GetMonthlyAnalytics(ctx context.Context, userID, month string) (*models.MonthlyAnalytics, error)
	GetFinancialInsights(ctx context.Context, userID, month string) ([]string, error)
	GetFinancialSummary(ctx context.Context, userID string) (*models.FinancialSummary, error)
	GetFinancialSummaryWithBudgets(ctx context.Context, userID, month string) (*models.MonthlyAnalyticsWithBudget, error)
	GetCategoryBreakdown(ctx context.Context, userID string, period string) ([]models.CategoryBreakdown, error)
	GetUniqueCategories(ctx context.Context, userID string) ([]string, error)
	GetCategoryOptions(ctx context.Context, userID string) ([]models.CategoryOption, error)
	GetMonthsWithTransactions(ctx context.Context, userID string) ([]string, error)
}

type analyticsService struct {
	repo          repository.Repository
	budgetService *BudgetService
}

func NewAnalyticsService(repo repository.Repository) AnalyticsService {
	return &analyticsService{
		repo:          repo,
		budgetService: NewBudgetService(repo),
	}
}

func (s *analyticsService) GetMonthlyAnalytics(ctx context.Context, userID, month string) (*models.MonthlyAnalytics, error) {
	if userID == "" || month == "" {
		return nil, fmt.Errorf("userID and month are required")
	}
	
	analytics, err := s.repo.GetMonthlyAnalytics(ctx, userID, month)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly analytics: %w", err)
	}
	
	return analytics, nil
}

func (s *analyticsService) GetFinancialSummary(ctx context.Context, userID string) (*models.FinancialSummary, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}
	
	// Get ALL historical transactions for user (not just current month)
	transactions, _, err := s.repo.GetTransactionsByUser(ctx, userID, 1000, nil) 
	if err != nil {
		return nil, fmt.Errorf("failed to get user transactions: %w", err)
	}
	
	// Calculate historical totals (same logic as AI service)
	totalIncome := 0.0
	totalExpenses := 0.0
	categoryMap := make(map[string]models.CategoryBreakdown)
	
	for _, transaction := range transactions {
		if transaction.Type == models.TransactionTypeIncome {
			totalIncome += transaction.Amount
		} else if transaction.Type == models.TransactionTypeExpense {
			totalExpenses += -transaction.Amount // Convert to positive for calculations
			
			breakdown, exists := categoryMap[transaction.Category]
			if !exists {
				breakdown = models.CategoryBreakdown{
					Category: transaction.Category,
					Amount:   0,
					Count:    0,
				}
			}
			// Sum absolute values of all transactions in this category
			breakdown.Amount += -transaction.Amount  // Convert negative expense to positive
			breakdown.Count++
			categoryMap[transaction.Category] = breakdown
		}
	}
	
	// Calculate total balance: income - expenses
	totalBalance := totalIncome - totalExpenses
	
	// Calculate savings rate: (total balance / total income) * 100
	savingsRate := 0.0
	if totalIncome > 0 {
		savingsRate = (totalBalance / totalIncome) * 100
	}
	
	// Calculate percentages for categories and convert to slice
	var categoryBreakdown []models.CategoryBreakdown
	for _, breakdown := range categoryMap {
		if totalExpenses > 0 {
			breakdown.Percentage = (breakdown.Amount / totalExpenses) * 100
		}
		categoryBreakdown = append(categoryBreakdown, breakdown)
	}
	
	// Sort category breakdown by category name
	sort.Slice(categoryBreakdown, func(i, j int) bool {
		return categoryBreakdown[i].Category < categoryBreakdown[j].Category
	})
	
	return &models.FinancialSummary{
		TotalBalance:      totalBalance,
		MonthlyIncome:     totalIncome,        // Now historical total income
		MonthlyExpenses:   -totalExpenses,     // Keep negative for consistency
		AvailableMoney:    totalBalance,
		SavingsRate:       savingsRate,        // New field
		CategoryBreakdown: categoryBreakdown,
		MonthlyTrend:      0, // TODO: Calculate trend
		Timestamp:         time.Now(),
	}, nil
}

// GetFinancialSummaryWithBudgets returns financial summary with budget information included
func (s *analyticsService) GetFinancialSummaryWithBudgets(ctx context.Context, userID, month string) (*models.MonthlyAnalyticsWithBudget, error) {
	if userID == "" || month == "" {
		return nil, fmt.Errorf("userID and month are required")
	}

	// Get transactions for the specific month
	transactions, _, err := s.repo.GetTransactionsByMonth(ctx, userID, month, 1000, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions for month: %w", err)
	}

	// Calculate monthly totals and category breakdown
	monthlyIncome := 0.0
	monthlyExpenses := 0.0
	categorySpending := make(map[string]float64)

	for _, transaction := range transactions {
		if transaction.Type == models.TransactionTypeIncome {
			monthlyIncome += transaction.Amount
		} else if transaction.Type == models.TransactionTypeExpense {
			monthlyExpenses += -transaction.Amount // Convert to positive
			categorySpending[transaction.Category] += -transaction.Amount
		}
	}

	// Get budgets for the month
	budgets, err := s.budgetService.GetBudgetsByMonth(ctx, userID, month)
	if err != nil {
		return nil, fmt.Errorf("failed to get budgets: %w", err)
	}

	// Create budget map for quick lookup
	budgetMap := make(map[string]models.Budget)
	for _, budget := range budgets {
		budgetMap[budget.Category] = budget
	}

	// Create category breakdown with budget information
	var categoryBreakdown []models.CategoryBudgetBreakdown
	allCategories := make(map[string]bool)

	// Add categories from spending
	for category := range categorySpending {
		allCategories[category] = true
	}

	// Add categories from budgets
	for _, budget := range budgets {
		allCategories[budget.Category] = true
	}

	// Build the breakdown for each category
	for category := range allCategories {
		spent := categorySpending[category]
		budgetAmount := 0.0

		if budget, hasBudget := budgetMap[category]; hasBudget {
			budgetAmount = budget.Amount
		}

		breakdown := models.CategoryBudgetBreakdown{
			Category:  category,
			Amount:    spent,
			Budget:    budgetAmount,
			Remaining: budgetAmount - spent,
		}

		categoryBreakdown = append(categoryBreakdown, breakdown)
	}

	// Sort category breakdown by category name
	sort.Slice(categoryBreakdown, func(i, j int) bool {
		return categoryBreakdown[i].Category < categoryBreakdown[j].Category
	})

	// Count total transactions
	transactionCount := len(transactions)

	return &models.MonthlyAnalyticsWithBudget{
		Month:             month,
		TotalIncome:       monthlyIncome,
		TotalExpense:      monthlyExpenses,
		Balance:           monthlyIncome - monthlyExpenses,
		CategoryBreakdown: categoryBreakdown,
		TransactionCount:  transactionCount,
	}, nil
}

func (s *analyticsService) GetCategoryBreakdown(ctx context.Context, userID string, period string) ([]models.CategoryBreakdown, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}
	
	// For now, get current month's data
	currentMonth := time.Now().Format("2006-01")
	
	transactions, _, err := s.repo.GetTransactionsByMonth(ctx, userID, currentMonth, 1000, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions for period: %w", err)
	}
	
	categoryMap := make(map[string]models.CategoryBreakdown)
	totalExpenses := 0.0
	
	for _, transaction := range transactions {
		if transaction.Type == models.TransactionTypeExpense {
			totalExpenses += -transaction.Amount
			
			breakdown, exists := categoryMap[transaction.Category]
			if !exists {
				breakdown = models.CategoryBreakdown{
					Category: transaction.Category,
					Amount:   0,
					Count:    0,
				}
			}
			breakdown.Amount += -transaction.Amount
			breakdown.Count++
			categoryMap[transaction.Category] = breakdown
		}
	}
	
	// Convert to slice and calculate percentages
	var result []models.CategoryBreakdown
	for _, breakdown := range categoryMap {
		if totalExpenses > 0 {
			breakdown.Percentage = (breakdown.Amount / totalExpenses) * 100
		}
		result = append(result, breakdown)
	}
	
	return result, nil
}

// GetUniqueCategories returns just the unique category names for a user
func (s *analyticsService) GetUniqueCategories(ctx context.Context, userID string) ([]string, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}
	
	// Get all transactions for user to find all categories they've used
	transactions, _, err := s.repo.GetTransactionsByUser(ctx, userID, 1000, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get user transactions: %w", err)
	}
	
	categorySet := make(map[string]bool)
	for _, transaction := range transactions {
		categorySet[transaction.Category] = true
	}
	
	var categories []string
	for category := range categorySet {
		categories = append(categories, category)
	}
	
	return categories, nil
}

// GetCategoryOptions returns category options with label and value for UI components
func (s *analyticsService) GetCategoryOptions(ctx context.Context, userID string) ([]models.CategoryOption, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}
	
	// Get all transactions for user to find all categories they've used
	transactions, _, err := s.repo.GetTransactionsByUser(ctx, userID, 1000, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get user transactions: %w", err)
	}
	
	categorySet := make(map[string]bool)
	for _, transaction := range transactions {
		categorySet[transaction.Category] = true
	}
	
	var categoryOptions []models.CategoryOption
	for category := range categorySet {
		categoryOptions = append(categoryOptions, models.CategoryOption{
			Label: strings.ToUpper(string(category[0])) + category[1:],
			Value: category,
		})
	}
	
	return categoryOptions, nil
}

func (s *analyticsService) GetMonthsWithTransactions(ctx context.Context, userID string) ([]string, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}
	
	// Get all transactions for user
	transactions, _, err := s.repo.GetTransactionsByUser(ctx, userID, 1000, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get user transactions: %w", err)
	}
	
	monthSet := make(map[string]bool)
	for _, transaction := range transactions {
		month := transaction.Date.Format("2006-01")
		monthSet[month] = true
	}
	
	var months []string
	for month := range monthSet {
		months = append(months, month)
	}
	
	return months, nil
}

func (s *analyticsService) GetFinancialInsights(ctx context.Context, userID, month string) ([]string, error) {
	analytics, err := s.GetMonthlyAnalytics(ctx, userID, month)
	if err != nil {
		return nil, err
	}
	
	var insights []string
	
	// Generate basic insights
	if analytics.Balance > 0 {
		insights = append(insights, fmt.Sprintf("Great! You saved $%.2f this month", analytics.Balance))
	} else {
		insights = append(insights, fmt.Sprintf("You spent $%.2f more than you earned this month", -analytics.Balance))
	}
	
	// Category insights
	var topCategory string
	var topAmount float64
	for category, amount := range analytics.CategoryBreakdown {
		if amount < 0 && -amount > topAmount { // Look for highest expense
			topCategory = category
			topAmount = -amount
		}
	}
	
	if topCategory != "" {
		insights = append(insights, fmt.Sprintf("Your highest spending category was %s with $%.2f", topCategory, topAmount))
	}
	
	return insights, nil
}
