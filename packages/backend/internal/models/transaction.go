package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

// Transaction represents a financial transaction in the system
type Transaction struct {
	ID          string    `json:"id" dynamodbav:"id"`
	Date        time.Time `json:"date" dynamodbav:"date"`
	Amount      float64   `json:"amount" dynamodbav:"amount"`
	Description string    `json:"description" dynamodbav:"description"`
	Category    string    `json:"category" dynamodbav:"category"`
	Type        string    `json:"type" dynamodbav:"type"` // "income" or "expense"
	UserID      string    `json:"user_id" dynamodbav:"user_id"`
	
	// DynamoDB keys for single-table design
	PK     string `json:"-" dynamodbav:"PK"`     // USER#{userID}
	SK     string `json:"-" dynamodbav:"SK"`     // TRANSACTION#{timestamp}#{id}
	GSI1PK string `json:"-" dynamodbav:"GSI1PK"` // MONTH#{YYYY-MM}#{userID}
	GSI1SK string `json:"-" dynamodbav:"GSI1SK"` // TRANSACTION#{timestamp}
	GSI2PK string `json:"-" dynamodbav:"GSI2PK"` // CATEGORY#{category}#{userID}
	GSI2SK string `json:"-" dynamodbav:"GSI2SK"` // TRANSACTION#{timestamp}
	
	// Metadata
	CreatedAt time.Time `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt time.Time `json:"updated_at" dynamodbav:"updated_at"`
	Version   int       `json:"version" dynamodbav:"version"`
}

// NewTransaction creates a new transaction with generated ID
func NewTransaction(userID, transactionType, category, description string, amount float64, date time.Time) *Transaction {
	now := time.Now()
	t := &Transaction{
		ID:          uuid.New().String(),
		UserID:      userID,
		Type:        transactionType,
		Category:    category,
		Description: description,
		Amount:      amount,
		Date:        date,
		CreatedAt:   now,
		UpdatedAt:   now,
		Version:     1,
	}
	t.GenerateKeys()
	return t
}

// GenerateKeys generates optimized DynamoDB keys for access patterns
func (t *Transaction) GenerateKeys() {
	// Primary access pattern: All transactions for a user
	t.PK = fmt.Sprintf("USER#%s", t.UserID)
	t.SK = fmt.Sprintf("TRANSACTION#%d#%s", t.Date.Unix(), t.ID)
	
	// GSI1: Monthly access pattern - Query transactions by month
	t.GSI1PK = fmt.Sprintf("MONTH#%s#%s", t.Date.Format("2006-01"), t.UserID)
	t.GSI1SK = fmt.Sprintf("TRANSACTION#%d", t.Date.Unix())
	
	// GSI2: Category access pattern - Query transactions by category
	t.GSI2PK = fmt.Sprintf("CATEGORY#%s#%s", strings.ToUpper(t.Category), t.UserID)
	t.GSI2SK = fmt.Sprintf("TRANSACTION#%d", t.Date.Unix())
}

// ToDynamoDBItem converts transaction to DynamoDB item
func (t *Transaction) ToDynamoDBItem() (map[string]types.AttributeValue, error) {
	t.GenerateKeys()
	return attributevalue.MarshalMap(t)
}

// FromDynamoDBItem creates transaction from DynamoDB item
func (t *Transaction) FromDynamoDBItem(item map[string]types.AttributeValue) error {
	return attributevalue.UnmarshalMap(item, t)
}

// Validate validates transaction fields
func (t *Transaction) Validate() error {
	if t.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	if t.Amount == 0 {
		return fmt.Errorf("amount must be non-zero")
	}
	if t.Type != "income" && t.Type != "expense" {
		return fmt.Errorf("type must be 'income' or 'expense'")
	}
	if t.Category == "" {
		return fmt.Errorf("category is required")
	}
	if t.Date.IsZero() {
		return fmt.Errorf("date is required")
	}
	return nil
}

// User represents a user in the system
type User struct {
	ID        string    `json:"id" dynamodbav:"id"`
	Email     string    `json:"email" dynamodbav:"email"`
	Name      string    `json:"name" dynamodbav:"name"`
	CreatedAt time.Time `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt time.Time `json:"updated_at" dynamodbav:"updated_at"`
	
	// DynamoDB keys
	PK string `json:"-" dynamodbav:"PK"` // USER#{id}
	SK string `json:"-" dynamodbav:"SK"` // PROFILE
}

// NewUser creates a new user
func NewUser(email, name string) *User {
	now := time.Now()
	u := &User{
		ID:        uuid.New().String(),
		Email:     email,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}
	u.GenerateKeys()
	return u
}

// GenerateKeys generates DynamoDB keys for a user
func (u *User) GenerateKeys() {
	u.PK = fmt.Sprintf("USER#%s", u.ID)
	u.SK = "PROFILE"
}

// BudgetUtilization represents budget vs spending analysis
type BudgetUtilization struct {
	Category     string  `json:"category"`
	BudgetAmount float64 `json:"budget_amount"`
	SpentAmount  float64 `json:"spent_amount"`
	Remaining    float64 `json:"remaining"`
	Percentage   float64 `json:"percentage"`
}
type TransactionSummary struct {
	TotalIncome  float64 `json:"total_income"`
	TotalExpense float64 `json:"total_expense"`
	Balance      float64 `json:"balance"`
	Count        int     `json:"count"`
}

type MonthlyAnalytics struct {
	Month             string                     `json:"month"`
	TotalIncome       float64                    `json:"total_income"`
	TotalExpense      float64                    `json:"total_expense"`
	Balance           float64                    `json:"balance"`
	CategoryBreakdown map[string]float64         `json:"category_breakdown"`
	TransactionCount  int                        `json:"transaction_count"`
	Transactions      []Transaction              `json:"transactions,omitempty"`
}

type CategorySummary struct {
	Category     string  `json:"category"`
	TotalAmount  float64 `json:"total_amount"`
	Count        int     `json:"count"`
	Percentage   float64 `json:"percentage"`
	AvgAmount    float64 `json:"avg_amount"`
}

// Query parameters for DynamoDB operations
type QueryParams struct {
	UserID     string
	StartDate  *time.Time
	EndDate    *time.Time
	Category   string
	Type       string
	Limit      int
	LastKey    map[string]types.AttributeValue
}

// PaginationResponse represents paginated response
type PaginationResponse struct {
	Items   interface{}                     `json:"items"`
	LastKey map[string]types.AttributeValue `json:"last_key,omitempty"`
	Count   int                             `json:"count"`
}

// AI models
type AIAdviceRequest struct {
	Question string `json:"question" validate:"required,min=10,max=500"`
	Context  string `json:"context,omitempty"`
}

type AIAdviceResponse struct {
	Advice      string              `json:"advice"`
	Suggestions []string            `json:"suggestions,omitempty"`
	Context     *FinancialContext   `json:"context,omitempty"`
	Timestamp   time.Time           `json:"timestamp"`
	Provider    string              `json:"provider,omitempty"`
	Model       string              `json:"model,omitempty"`
}

type FinancialContext struct {
	MonthlyIncome    float64            `json:"monthly_income"`
	MonthlyExpense   float64            `json:"monthly_expense"`
	SavingsRate      float64            `json:"savings_rate"`
	TopCategories    []*CategorySummary `json:"top_categories"`
	SpendingTrends   []string           `json:"spending_trends"`
}

// Constants
const (
	TransactionTypeIncome  = "income"
	TransactionTypeExpense = "expense"
)

const (
	CategorySalary         = "salary"
	CategoryRent           = "rent"
	CategoryGroceries      = "groceries"
	CategoryUtilities      = "utilities"
	CategoryDining         = "dining"
	CategoryTransportation = "transportation"
	CategoryEntertainment  = "entertainment"
	CategoryHealthcare     = "healthcare"
	CategoryShopping       = "shopping"
	CategoryOther          = "other"
)

// Custom JSON marshaling to handle time formatting
func (t *Transaction) MarshalJSON() ([]byte, error) {
	type Alias Transaction
	return json.Marshal(&struct {
		Date string `json:"date"`
		*Alias
	}{
		Date:  t.Date.Format("2006-01-02"),
		Alias: (*Alias)(t),
	})
}

func (t *Transaction) UnmarshalJSON(data []byte) error {
	type Alias Transaction
	aux := &struct {
		Date string `json:"date"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}
	
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	
	// If date is empty, use current time
	if aux.Date == "" {
		t.Date = time.Now()
		return nil
	}
	
	// Try parsing RFC3339 format first (ISO 8601)
	if parsedTime, err := time.Parse(time.RFC3339, aux.Date); err == nil {
		t.Date = parsedTime
		return nil
	}
	
	// Fallback to date-only format
	var err error
	t.Date, err = time.Parse("2006-01-02", aux.Date)
	return err
}

// FinancialSummary represents an overall financial summary for a user
type FinancialSummary struct {
	TotalBalance      float64             `json:"total_balance"`
	MonthlyIncome     float64             `json:"monthly_income"`
	MonthlyExpenses   float64             `json:"monthly_expenses"`
	AvailableMoney    float64             `json:"available_money"`
	SavingsRate       float64             `json:"savings_rate"` // percentage of income saved
	CategoryBreakdown []CategoryBreakdown `json:"category_breakdown"` // Changed from map to slice
	MonthlyTrend      float64             `json:"monthly_trend"` // percentage change from last month
	Timestamp         time.Time           `json:"timestamp"`
}

// CategoryBreakdown represents spending breakdown by category
type CategoryBreakdown struct {
	Category    string  `json:"category"`
	Amount      float64 `json:"amount"`
	Percentage  float64 `json:"percentage"`
	Count       int     `json:"transaction_count"`
	Color       string  `json:"color,omitempty"` // For chart visualization
}

// MonthSummary represents summary data for a specific month
type MonthSummary struct {
	Month        string  `json:"month"`
	Income       float64 `json:"income"`
	Expenses     float64 `json:"expenses"`
	Balance      float64 `json:"balance"`
	HasTransactions bool `json:"has_transactions"`
}

// CategoryOption represents a category option for UI dropdowns/selects
type CategoryOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// Budget represents a budget for a specific category and month
type Budget struct {
	ID       string    `json:"id" dynamodbav:"id"`
	UserID   string    `json:"user_id" dynamodbav:"user_id"`
	Category string    `json:"category" dynamodbav:"category"`
	Month    string    `json:"month" dynamodbav:"month"` // YYYY-MM format
	Amount   float64   `json:"amount" dynamodbav:"amount"`
	CreatedAt time.Time `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt time.Time `json:"updated_at" dynamodbav:"updated_at"`
	
	// DynamoDB keys for single-table design
	PK string `json:"-" dynamodbav:"PK"` // USER#{userID}
	SK string `json:"-" dynamodbav:"SK"` // BUDGET#{month}#{category}
}

// GenerateKeys generates optimized DynamoDB keys for budget
func (b *Budget) GenerateKeys() {
	b.PK = fmt.Sprintf("USER#%s", b.UserID)
	b.SK = fmt.Sprintf("BUDGET#%s#%s", b.Month, strings.ToUpper(b.Category))
}

// ToDynamoDBItem converts budget to DynamoDB item
func (b *Budget) ToDynamoDBItem() (map[string]types.AttributeValue, error) {
	b.GenerateKeys()
	return attributevalue.MarshalMap(b)
}

// FromDynamoDBItem creates budget from DynamoDB item
func (b *Budget) FromDynamoDBItem(item map[string]types.AttributeValue) error {
	return attributevalue.UnmarshalMap(item, b)
}

// CategoryBudgetBreakdown represents spending breakdown with budget info
type CategoryBudgetBreakdown struct {
	Category  string  `json:"category"`
	Amount    float64 `json:"amount"`
	Budget    float64 `json:"budget"`
	Remaining float64 `json:"remaining"`
}

// MonthlyAnalyticsWithBudget extends MonthlyAnalytics with budget information
type MonthlyAnalyticsWithBudget struct {
	Month             string                     `json:"month"`
	TotalIncome       float64                    `json:"total_income"`
	TotalExpense      float64                    `json:"total_expense"`
	Balance           float64                    `json:"balance"`
	CategoryBreakdown []CategoryBudgetBreakdown  `json:"category_breakdown"` // Changed from map
	TransactionCount  int                        `json:"transaction_count"`
}
