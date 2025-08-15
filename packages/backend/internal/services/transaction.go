package services

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"

	"backend/internal/models"
	"backend/internal/repository"
)

type TransactionService interface {
	GetTransactionsByUser(ctx context.Context, userID string, limit int) ([]models.Transaction, error)
	GetTransactionsByMonth(ctx context.Context, userID, month string, limit int) ([]models.Transaction, error)
	GetTransactionsByCategory(ctx context.Context, userID, category string, limit int) ([]models.Transaction, error)
	GetTransaction(ctx context.Context, userID, transactionID string) (*models.Transaction, error)
	CreateTransaction(ctx context.Context, transaction *models.Transaction) error
	UpdateTransaction(ctx context.Context, transaction *models.Transaction) error
	DeleteTransaction(ctx context.Context, userID, transactionID string) error
	ValidateTransaction(transaction *models.Transaction) error
}

type transactionService struct {
	repo repository.Repository
}

func NewTransactionService(repo repository.Repository) TransactionService {
	return &transactionService{
		repo: repo,
	}
}

func (s *transactionService) GetTransactionsByUser(ctx context.Context, userID string, limit int) ([]models.Transaction, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}
	
	if limit <= 0 {
		limit = 50
	}
	
	transactions, _, err := s.repo.GetTransactionsByUser(ctx, userID, limit, nil)
	if err != nil {
		return nil, err
	}
	
	// Sort transactions by date in descending order (most recent first)
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].Date.After(transactions[j].Date)
	})
	
	return transactions, nil
}

func (s *transactionService) GetTransactionsByMonth(ctx context.Context, userID, month string, limit int) ([]models.Transaction, error) {
	if userID == "" || month == "" {
		return nil, fmt.Errorf("userID and month are required")
	}
	
	if limit <= 0 {
		limit = 50
	}
	
	transactions, _, err := s.repo.GetTransactionsByMonth(ctx, userID, month, limit, nil)
	return transactions, err
}

func (s *transactionService) GetTransactionsByCategory(ctx context.Context, userID, category string, limit int) ([]models.Transaction, error) {
	if userID == "" || category == "" {
		return nil, fmt.Errorf("userID and category are required")
	}
	
	if limit <= 0 {
		limit = 50
	}
	
	transactions, _, err := s.repo.GetTransactionsByCategory(ctx, userID, category, limit, nil)
	return transactions, err
}

func (s *transactionService) GetTransaction(ctx context.Context, userID, transactionID string) (*models.Transaction, error) {
	if userID == "" || transactionID == "" {
		return nil, fmt.Errorf("userID and transactionID are required")
	}
	
	return s.repo.GetTransaction(ctx, userID, transactionID)
}

func (s *transactionService) CreateTransaction(ctx context.Context, transaction *models.Transaction) error {
	if err := s.ValidateTransaction(transaction); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	
	// Generate DynamoDB keys after validation and ID generation
	transaction.GenerateKeys()
	
	return s.repo.CreateTransaction(ctx, transaction)
}

func (s *transactionService) UpdateTransaction(ctx context.Context, transaction *models.Transaction) error {
	if err := s.ValidateTransaction(transaction); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	
	return s.repo.UpdateTransaction(ctx, transaction)
}

func (s *transactionService) DeleteTransaction(ctx context.Context, userID, transactionID string) error {
	if userID == "" || transactionID == "" {
		return fmt.Errorf("userID and transactionID are required")
	}
	
	return s.repo.DeleteTransaction(ctx, userID, transactionID)
}

func (s *transactionService) ValidateTransaction(transaction *models.Transaction) error {
	if transaction == nil {
		return fmt.Errorf("transaction cannot be nil")
	}
	
	// Generate ID if not provided
	if transaction.ID == "" {
		transaction.ID = uuid.New().String()
	}
	
	if transaction.UserID == "" {
		return fmt.Errorf("user ID is required")
	}
	
	if transaction.Amount == 0 {
		return fmt.Errorf("amount cannot be zero")
	}
	
	if transaction.Category == "" {
		return fmt.Errorf("category is required")
	}
	
	if transaction.Type != models.TransactionTypeIncome && transaction.Type != models.TransactionTypeExpense {
		return fmt.Errorf("type must be either income or expense")
	}
	
	if transaction.Description == "" {
		return fmt.Errorf("description is required")
	}
	
	if transaction.Date.IsZero() {
		transaction.Date = time.Now()
	}
	
	// Set timestamps and version for new transactions
	now := time.Now()
	if transaction.CreatedAt.IsZero() {
		transaction.CreatedAt = now
	}
	transaction.UpdatedAt = now
	
	if transaction.Version == 0 {
		transaction.Version = 1
	}
	
	return nil
}
