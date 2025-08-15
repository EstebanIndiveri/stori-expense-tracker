package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"backend/internal/database"
	"backend/internal/models"
	"backend/internal/repository"
)

// JSONTransaction represents the structure in the JSON file
type JSONTransaction struct {
	ID          string  `json:"id"`
	Date        string  `json:"date"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Type        string  `json:"type"`
}

// SeedData contains the structure from JSON file
type SeedData struct {
	Transactions []JSONTransaction `json:"transactions"`
}

var (
	filePath    = flag.String("file", "./data/mock_expense_and_income.json", "Path to JSON seed file")
	version     = flag.String("version", "v1", "Seed version for tracking")
	userID      = flag.String("user", "default-user", "User ID to assign transactions")
	dryRun      = flag.Bool("dry-run", false, "Preview transactions without inserting")
	batchSize   = flag.Int("batch-size", 25, "Number of transactions per batch")
	maxRetries  = flag.Int("max-retries", 3, "Maximum retry attempts")
	environment = flag.String("env", "local", "Environment: local or aws")
)

func main() {
	flag.Parse()

	log.Printf("Starting seed process...")
	log.Printf("File: %s", *filePath)
	log.Printf("Version: %s", *version)
	log.Printf("User ID: %s", *userID)
	log.Printf("Environment: %s", *environment)
	log.Printf("Dry Run: %t", *dryRun)

	ctx := context.Background()

	// Initialize DynamoDB client
	dbClient, err := database.NewDynamoDBClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create DynamoDB client: %v", err)
	}

	if !*dryRun {
		// Create tables if not exist
		if err := dbClient.CreateTablesIfNotExist(ctx); err != nil {
			log.Fatalf("Failed to create tables: %v", err)
		}
		log.Println("Tables verified/created successfully")
	}

	// Initialize repository
	repo := repository.NewDynamoDBRepository(dbClient.Client, dbClient.Config.TablePrefix+"-transactions")

	// Load and validate JSON data
	transactions, err := loadTransactionsFromJSON(*filePath, *userID)
	if err != nil {
		log.Fatalf("Failed to load transactions: %v", err)
	}

	log.Printf("Loaded %d transactions from JSON", len(transactions))

	if *dryRun {
		fmt.Println("\n=== DRY RUN - Preview of transactions to be seeded ===")
		previewTransactions(transactions)
		return
	}

	// Create default user first
	user := models.NewUser("default@stori.com", "Default User")
	user.ID = *userID
	if err := repo.CreateUser(ctx, user); err != nil {
		log.Printf("User may already exist: %v", err)
	} else {
		log.Printf("Created user: %s", user.ID)
	}

	// Seed transactions with retry logic
	if err := seedTransactionsWithRetry(ctx, repo, transactions); err != nil {
		log.Fatalf("Failed to seed transactions: %v", err)
	}

	log.Printf("Successfully seeded %d transactions", len(transactions))

	// Verify seeding by querying some data
	if err := verifySeedData(ctx, repo, *userID); err != nil {
		log.Printf("Warning: Verification failed: %v", err)
	}
}

func loadTransactionsFromJSON(filePath, userID string) ([]models.Transaction, error) {
	// Read JSON file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	var jsonTransactions []JSONTransaction
	if err := json.Unmarshal(data, &jsonTransactions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	var transactions []models.Transaction
	for i, jsonTx := range jsonTransactions {
		// Parse date
		date, err := time.Parse("2006-01-02", jsonTx.Date)
		if err != nil {
			log.Printf("Warning: Invalid date format for transaction %d: %s", i, jsonTx.Date)
			continue
		}

		// Generate ID if not provided
		id := jsonTx.ID
		if id == "" {
			id = uuid.New().String()
		}

		// Validate required fields
		if jsonTx.Amount == 0 {
			log.Printf("Warning: Skipping transaction %d with zero amount", i)
			continue
		}

		if jsonTx.Type != "income" && jsonTx.Type != "expense" {
			log.Printf("Warning: Invalid type '%s' for transaction %d, defaulting to 'expense'", jsonTx.Type, i)
			jsonTx.Type = "expense"
		}

		if jsonTx.Category == "" {
			jsonTx.Category = "other"
		}

		// Create transaction
		transaction := models.NewTransaction(
			userID,
			jsonTx.Type,
			jsonTx.Category,
			jsonTx.Description,
			jsonTx.Amount,
			date,
		)
		transaction.ID = id // Override with specific ID if provided

		transactions = append(transactions, *transaction)
	}

	return transactions, nil
}

func previewTransactions(transactions []models.Transaction) {
	fmt.Printf("\nTotal transactions: %d\n", len(transactions))
	
	// Show first 5 transactions
	fmt.Println("\nFirst 5 transactions:")
	for i, tx := range transactions {
		if i >= 5 {
			break
		}
		fmt.Printf("  %d. %s | %s | %s | %.2f | %s\n", 
			i+1, tx.Date.Format("2006-01-02"), tx.Type, tx.Category, tx.Amount, tx.Description)
	}

	// Summary statistics
	var totalIncome, totalExpense float64
	categoryCount := make(map[string]int)
	typeCount := make(map[string]int)

	for _, tx := range transactions {
		if tx.Type == "income" {
			totalIncome += tx.Amount
		} else {
			totalExpense += tx.Amount
		}
		categoryCount[tx.Category]++
		typeCount[tx.Type]++
	}

	fmt.Printf("\nSummary:")
	fmt.Printf("  Total Income: $%.2f\n", totalIncome)
	fmt.Printf("  Total Expense: $%.2f\n", totalExpense)
	fmt.Printf("  Net Balance: $%.2f\n", totalIncome-totalExpense)
	
	fmt.Printf("\nBy Type:")
	for txType, count := range typeCount {
		fmt.Printf("  %s: %d\n", txType, count)
	}

	fmt.Printf("\nBy Category (top 5):")
	type catCount struct {
		category string
		count    int
	}
	var cats []catCount
	for cat, count := range categoryCount {
		cats = append(cats, catCount{cat, count})
	}
	// Simple sort by count (descending)
	for i := 0; i < len(cats)-1; i++ {
		for j := i + 1; j < len(cats); j++ {
			if cats[j].count > cats[i].count {
				cats[i], cats[j] = cats[j], cats[i]
			}
		}
	}
	for i, cat := range cats {
		if i >= 5 {
			break
		}
		fmt.Printf("  %s: %d\n", cat.category, cat.count)
	}
}

func seedTransactionsWithRetry(ctx context.Context, repo repository.Repository, transactions []models.Transaction) error {
	// Split into batches
	for i := 0; i < len(transactions); i += *batchSize {
		end := i + *batchSize
		if end > len(transactions) {
			end = len(transactions)
		}

		batch := transactions[i:end]
		
		// Retry logic for each batch
		var err error
		for retry := 0; retry < *maxRetries; retry++ {
			err = repo.BatchCreateTransactions(ctx, batch)
			if err == nil {
				log.Printf("Successfully seeded batch %d-%d (%d transactions)", i+1, end, len(batch))
				break
			}

			if retry == *maxRetries-1 {
				return fmt.Errorf("failed to seed batch %d-%d after %d retries: %w", i+1, end, *maxRetries, err)
			}

			log.Printf("Retry %d/%d for batch %d-%d: %v", retry+1, *maxRetries, i+1, end, err)
			time.Sleep(time.Duration(retry+1) * 500 * time.Millisecond)
		}
	}

	return nil
}

func verifySeedData(ctx context.Context, repo repository.Repository, userID string) error {
	log.Println("Verifying seed data...")

	// Test basic query
	transactions, _, err := repo.GetTransactionsByUser(ctx, userID, 10, nil)
	if err != nil {
		return fmt.Errorf("failed to verify user transactions: %w", err)
	}

	if len(transactions) == 0 {
		return fmt.Errorf("no transactions found for user %s", userID)
	}

	log.Printf("✓ Found %d transactions for user %s", len(transactions), userID)

	// Test monthly query with a sample month
	if len(transactions) > 0 {
		sampleMonth := transactions[0].Date.Format("2006-01")
		monthlyTxs, _, err := repo.GetTransactionsByMonth(ctx, userID, sampleMonth, 10, nil)
		if err != nil {
			return fmt.Errorf("failed to verify monthly query: %w", err)
		}
		log.Printf("✓ Found %d transactions for month %s", len(monthlyTxs), sampleMonth)
	}

	// Test category query
	if len(transactions) > 0 {
		sampleCategory := transactions[0].Category
		categoryTxs, _, err := repo.GetTransactionsByCategory(ctx, userID, sampleCategory, 10, nil)
		if err != nil {
			return fmt.Errorf("failed to verify category query: %w", err)
		}
		log.Printf("✓ Found %d transactions for category %s", len(categoryTxs), sampleCategory)
	}

	log.Println("✓ Seed verification completed successfully")
	return nil
}
