package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"

	"backend/internal/models"
)

type Repository interface {
	// Transaction operations
	CreateTransaction(ctx context.Context, transaction *models.Transaction) error
	GetTransaction(ctx context.Context, userID, transactionID string) (*models.Transaction, error)
	UpdateTransaction(ctx context.Context, transaction *models.Transaction) error
	DeleteTransaction(ctx context.Context, userID, transactionID string) error
	
	// Query operations with optimized access patterns
	GetTransactionsByUser(ctx context.Context, userID string, limit int, lastKey map[string]types.AttributeValue) ([]models.Transaction, map[string]types.AttributeValue, error)
	GetTransactionsByMonth(ctx context.Context, userID string, month string, limit int, lastKey map[string]types.AttributeValue) ([]models.Transaction, map[string]types.AttributeValue, error)
	GetTransactionsByCategory(ctx context.Context, userID string, category string, limit int, lastKey map[string]types.AttributeValue) ([]models.Transaction, map[string]types.AttributeValue, error)
	
	// Batch operations
	BatchCreateTransactions(ctx context.Context, transactions []models.Transaction) error
	
	// Analytics
	GetMonthlyAnalytics(ctx context.Context, userID string, month string) (*models.MonthlyAnalytics, error)
	
	// Budget operations
	CreateOrUpdateBudget(ctx context.Context, budget *models.Budget) error
	GetBudgetsByMonth(ctx context.Context, userID, month string) ([]models.Budget, error)
	GetBudget(ctx context.Context, userID, month, category string) (*models.Budget, error)
	DeleteBudget(ctx context.Context, userID, month, category string) error
	
	// User operations
	CreateUser(ctx context.Context, user *models.User) error
	GetUser(ctx context.Context, userID string) (*models.User, error)
}

type DynamoDBRepository struct {
	client    *dynamodb.Client
	tableName string
}

func NewDynamoDBRepository(client *dynamodb.Client, tableName string) Repository {
	return &DynamoDBRepository{
		client:    client,
		tableName: tableName,
	}
}

// CreateTransaction creates a new transaction
func (r *DynamoDBRepository) CreateTransaction(ctx context.Context, transaction *models.Transaction) error {
	item, err := transaction.ToDynamoDBItem()
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
		// Prevent overwriting existing transactions
		ConditionExpression: aws.String("attribute_not_exists(PK) AND attribute_not_exists(SK)"),
	}

	_, err = r.client.PutItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	log.Printf("Transaction created: %s for user %s", transaction.ID, transaction.UserID)
	return nil
}

// GetTransactionsByUser retrieves all transactions for a user with pagination
func (r *DynamoDBRepository) GetTransactionsByUser(ctx context.Context, userID string, limit int, lastKey map[string]types.AttributeValue) ([]models.Transaction, map[string]types.AttributeValue, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("USER#%s", userID)},
			":sk": &types.AttributeValueMemberS{Value: "TX#"},
		},
		ScanIndexForward: aws.Bool(false), // Most recent first
		Limit:           aws.Int32(int32(limit)),
	}

	if lastKey != nil {
		input.ExclusiveStartKey = lastKey
	}

	result, err := r.client.Query(ctx, input)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query transactions: %w", err)
	}

	var transactions []models.Transaction
	for _, item := range result.Items {
		var transaction models.Transaction
		if err := transaction.FromDynamoDBItem(item); err != nil {
			log.Printf("Failed to unmarshal transaction: %v", err)
			continue
		}
		transactions = append(transactions, transaction)
	}

	return transactions, result.LastEvaluatedKey, nil
}

// GetTransactionsByMonth retrieves transactions for a specific month using GSI1
func (r *DynamoDBRepository) GetTransactionsByMonth(ctx context.Context, userID string, month string, limit int, lastKey map[string]types.AttributeValue) ([]models.Transaction, map[string]types.AttributeValue, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("GSI1"),
		KeyConditionExpression: aws.String("GSI1PK = :gsi1pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":gsi1pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("MONTH#%s#%s", month, userID)},
		},
		ScanIndexForward: aws.Bool(false), // Most recent first
		Limit:           aws.Int32(int32(limit)),
	}

	if lastKey != nil {
		input.ExclusiveStartKey = lastKey
	}

	result, err := r.client.Query(ctx, input)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query transactions by month: %w", err)
	}

	var transactions []models.Transaction
	for _, item := range result.Items {
		var transaction models.Transaction
		if err := transaction.FromDynamoDBItem(item); err != nil {
			log.Printf("Failed to unmarshal transaction: %v", err)
			continue
		}
		transactions = append(transactions, transaction)
	}

	return transactions, result.LastEvaluatedKey, nil
}

// GetTransactionsByCategory retrieves transactions for a specific category using GSI2
func (r *DynamoDBRepository) GetTransactionsByCategory(ctx context.Context, userID string, category string, limit int, lastKey map[string]types.AttributeValue) ([]models.Transaction, map[string]types.AttributeValue, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("GSI2"),
		KeyConditionExpression: aws.String("GSI2PK = :gsi2pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":gsi2pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("CATEGORY#%s#%s", strings.ToUpper(category), userID)},
		},
		ScanIndexForward: aws.Bool(false), // Most recent first
		Limit:           aws.Int32(int32(limit)),
	}

	if lastKey != nil {
		input.ExclusiveStartKey = lastKey
	}

	result, err := r.client.Query(ctx, input)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query transactions by category: %w", err)
	}

	var transactions []models.Transaction
	for _, item := range result.Items {
		var transaction models.Transaction
		if err := transaction.FromDynamoDBItem(item); err != nil {
			log.Printf("Failed to unmarshal transaction: %v", err)
			continue
		}
		transactions = append(transactions, transaction)
	}

	return transactions, result.LastEvaluatedKey, nil
}

// GetMonthlyAnalytics calculates analytics for a specific month
func (r *DynamoDBRepository) GetMonthlyAnalytics(ctx context.Context, userID string, month string) (*models.MonthlyAnalytics, error) {
	transactions, _, err := r.GetTransactionsByMonth(ctx, userID, month, 1000, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions for analytics: %w", err)
	}

	analytics := &models.MonthlyAnalytics{
		Month:             month,
		CategoryBreakdown: make(map[string]float64),
	}

	for _, tx := range transactions {
		analytics.TransactionCount++
		if tx.Type == models.TransactionTypeIncome {
			analytics.TotalIncome += tx.Amount
		} else {
			analytics.TotalExpense += tx.Amount
		}
		analytics.CategoryBreakdown[tx.Category] += tx.Amount
	}

	analytics.Balance = analytics.TotalIncome - analytics.TotalExpense
	return analytics, nil
}

// UpdateTransaction updates an existing transaction
func (r *DynamoDBRepository) UpdateTransaction(ctx context.Context, transaction *models.Transaction) error {
	if err := transaction.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// First, get the current transaction to preserve the original timestamps
	existing, err := r.GetTransaction(ctx, transaction.UserID, transaction.ID)
	if err != nil {
		return fmt.Errorf("transaction not found: %w", err)
	}

	// Preserve original CreatedAt and keys
	transaction.CreatedAt = existing.CreatedAt
	transaction.UpdatedAt = time.Now()
	transaction.Version = existing.Version + 1
	
	// Generate keys to ensure they're consistent
	transaction.GenerateKeys()

	item, err := transaction.ToDynamoDBItem()
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
		// Use optimistic locking but handle version mismatch gracefully
		ConditionExpression: aws.String("version = :oldVersion"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":oldVersion": &types.AttributeValueMemberN{Value: strconv.Itoa(existing.Version)},
		},
	}

	_, err = r.client.PutItem(ctx, input)
	if err != nil {
		// Check if it's a condition failed error (version mismatch)
		var condErr *types.ConditionalCheckFailedException
		if errors.As(err, &condErr) {
			return fmt.Errorf("transaction was modified by another process, please retry")
		}
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	return nil
}

// DeleteTransaction deletes a transaction by first finding it, then deleting
func (r *DynamoDBRepository) DeleteTransaction(ctx context.Context, userID, transactionID string) error {
	// First, find the transaction to get its full key
	transaction, err := r.GetTransaction(ctx, userID, transactionID)
	if err != nil {
		return fmt.Errorf("transaction not found: %w", err)
	}

	// Extract the SK from the found transaction
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: transaction.PK},
			"SK": &types.AttributeValueMemberS{Value: transaction.SK},
		},
		// Ensure transaction exists before deletion
		ConditionExpression: aws.String("attribute_exists(PK)"),
	}

	_, err = r.client.DeleteItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete transaction: %w", err)
	}

	return nil
}

// GetTransaction retrieves a single transaction by ID using Query and client-side filtering
func (r *DynamoDBRepository) GetTransaction(ctx context.Context, userID, transactionID string) (*models.Transaction, error) {
	// Query for all transactions of the user
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk_prefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":        &types.AttributeValueMemberS{Value: fmt.Sprintf("USER#%s", userID)},
			":sk_prefix": &types.AttributeValueMemberS{Value: "TX#"},
		},
	}

	result, err := r.client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}

	// Filter transactions by ID on the client side
	for _, item := range result.Items {
		var transaction models.Transaction
		if err := transaction.FromDynamoDBItem(item); err != nil {
			log.Printf("Failed to unmarshal transaction: %v", err)
			continue
		}
		
		if transaction.ID == transactionID {
			return &transaction, nil
		}
	}

	return nil, fmt.Errorf("transaction not found")
}

// BatchCreateTransactions creates multiple transactions in batches
func (r *DynamoDBRepository) BatchCreateTransactions(ctx context.Context, transactions []models.Transaction) error {
	const batchSize = 25 // DynamoDB batch limit

	for i := 0; i < len(transactions); i += batchSize {
		end := i + batchSize
		if end > len(transactions) {
			end = len(transactions)
		}

		batch := transactions[i:end]
		if err := r.batchWriteTransactions(ctx, batch); err != nil {
			return fmt.Errorf("failed to write batch %d-%d: %w", i, end, err)
		}

		log.Printf("Successfully wrote batch %d-%d (%d transactions)", i, end-1, len(batch))
	}

	return nil
}

func (r *DynamoDBRepository) batchWriteTransactions(ctx context.Context, transactions []models.Transaction) error {
	var writeRequests []types.WriteRequest

	for _, tx := range transactions {
		tx.GenerateKeys()
		tx.CreatedAt = time.Now()
		tx.UpdatedAt = time.Now()

		item, err := tx.ToDynamoDBItem()
		if err != nil {
			return fmt.Errorf("failed to marshal transaction %s: %w", tx.ID, err)
		}

		writeRequests = append(writeRequests, types.WriteRequest{
			PutRequest: &types.PutRequest{Item: item},
		})
	}

	input := &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			r.tableName: writeRequests,
		},
	}

	// Handle unprocessed items with retry
	maxRetries := 3
	for retry := 0; retry < maxRetries; retry++ {
		result, err := r.client.BatchWriteItem(ctx, input)
		if err != nil {
			return fmt.Errorf("failed to batch write items: %w", err)
		}

		if len(result.UnprocessedItems) == 0 {
			break
		}

		if retry == maxRetries-1 {
			return fmt.Errorf("failed to process all items after %d retries", maxRetries)
		}

		// Retry unprocessed items
		input.RequestItems = result.UnprocessedItems
		time.Sleep(time.Duration(retry+1) * 100 * time.Millisecond)
	}

	return nil
}

// User management methods
func (r *DynamoDBRepository) CreateUser(ctx context.Context, user *models.User) error {
	user.GenerateKeys()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	}

	_, err = r.client.PutItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *DynamoDBRepository) GetUser(ctx context.Context, userID string) (*models.User, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("USER#%s", userID)},
			"SK": &types.AttributeValueMemberS{Value: "PROFILE"},
		},
	}

	result, err := r.client.GetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("user not found")
	}

	var user models.User
	if err := attributevalue.UnmarshalMap(result.Item, &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return &user, nil
}

// Budget operations

// CreateOrUpdateBudget creates or updates a budget for a specific category and month
func (r *DynamoDBRepository) CreateOrUpdateBudget(ctx context.Context, budget *models.Budget) error {
	if budget.ID == "" {
		budget.ID = uuid.New().String()
	}
	
	now := time.Now()
	if budget.CreatedAt.IsZero() {
		budget.CreatedAt = now
	}
	budget.UpdatedAt = now
	
	budget.GenerateKeys()
	
	item, err := budget.ToDynamoDBItem()
	if err != nil {
		return fmt.Errorf("failed to marshal budget: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	}

	_, err = r.client.PutItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create/update budget: %w", err)
	}

	return nil
}

// GetBudgetsByMonth retrieves all budgets for a user in a specific month
func (r *DynamoDBRepository) GetBudgetsByMonth(ctx context.Context, userID, month string) ([]models.Budget, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk_prefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":        &types.AttributeValueMemberS{Value: fmt.Sprintf("USER#%s", userID)},
			":sk_prefix": &types.AttributeValueMemberS{Value: fmt.Sprintf("BUDGET#%s#", month)},
		},
	}

	result, err := r.client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query budgets: %w", err)
	}

	var budgets []models.Budget
	for _, item := range result.Items {
		var budget models.Budget
		if err := budget.FromDynamoDBItem(item); err != nil {
			log.Printf("Failed to unmarshal budget: %v", err)
			continue
		}
		budgets = append(budgets, budget)
	}

	return budgets, nil
}

// GetBudget retrieves a specific budget by category and month
func (r *DynamoDBRepository) GetBudget(ctx context.Context, userID, month, category string) (*models.Budget, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("USER#%s", userID)},
			"SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("BUDGET#%s#%s", month, strings.ToUpper(category))},
		},
	}

	result, err := r.client.GetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get budget: %w", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("budget not found")
	}

	var budget models.Budget
	if err := budget.FromDynamoDBItem(result.Item); err != nil {
		return nil, fmt.Errorf("failed to unmarshal budget: %w", err)
	}

	return &budget, nil
}

// DeleteBudget deletes a budget for a specific category and month
func (r *DynamoDBRepository) DeleteBudget(ctx context.Context, userID, month, category string) error {
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("USER#%s", userID)},
			"SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("BUDGET#%s#%s", month, strings.ToUpper(category))},
		},
		ConditionExpression: aws.String("attribute_exists(PK)"),
	}

	_, err := r.client.DeleteItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete budget: %w", err)
	}

	return nil
}
