package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

type RawTransaction struct {
	Date        string  `json:"date"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Type        string  `json:"type"`
}

// Transaction represents a financial transaction compatible with backend model
type Transaction struct {
	ID          string    `json:"id" dynamodbav:"id"`
	UserID      string    `json:"user_id" dynamodbav:"user_id"`
	Date        time.Time `json:"date" dynamodbav:"date"`
	Amount      float64   `json:"amount" dynamodbav:"amount"`
	Description string    `json:"description" dynamodbav:"description"`
	Category    string    `json:"category" dynamodbav:"category"`
	Type        string    `json:"type" dynamodbav:"type"`
	
	// DynamoDB keys - compatible with backend patterns
	PK     string `json:"-" dynamodbav:"PK"`     // USER#{userID}
	SK     string `json:"-" dynamodbav:"SK"`     // TX#{date}#{id}
	GSI1PK string `json:"-" dynamodbav:"GSI1PK"` // MONTH#{YYYY-MM}#{userID}
	GSI1SK string `json:"-" dynamodbav:"GSI1SK"` // TX#{date}#{id}
	GSI2PK string `json:"-" dynamodbav:"GSI2PK"` // CATEGORY#{category}#{userID}
	GSI2SK string `json:"-" dynamodbav:"GSI2SK"` // TX#{date}#{id}
	
	// Metadata
	CreatedAt time.Time `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt time.Time `json:"updated_at" dynamodbav:"updated_at"`
}

func main() {
	ctx := context.Background()

	// Get UserID from environment or use default
	userID := os.Getenv("USER_ID")
	if userID == "" {
		userID = "demo-user-123"
		fmt.Printf("Using default UserID: %s (set USER_ID env var to override)\n", userID)
	}

	// Load AWS config
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	// Override endpoint for local DynamoDB
	client := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		if endpoint := os.Getenv("DYNAMODB_ENDPOINT"); endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
	})

	tableName := os.Getenv("DYNAMODB_TABLE_NAME")
	if tableName == "" {
		tableName = "stori-transactions-dev"
	}

	// Read JSON file
	dataFile := "../../data/mock_expense_and_income.json"
	if len(os.Args) > 1 {
		dataFile = os.Args[1]
	}

	rawData, err := os.ReadFile(dataFile)
	if err != nil {
		log.Fatalf("Failed to read data file: %v", err)
	}

	var rawTransactions []RawTransaction
	if err := json.Unmarshal(rawData, &rawTransactions); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	// Transform and batch write
	fmt.Printf("Processing %d transactions for user %s...\n", len(rawTransactions), userID)

	for i := 0; i < len(rawTransactions); i += 25 {
		end := i + 25
		if end > len(rawTransactions) {
			end = len(rawTransactions)
		}

		batch := rawTransactions[i:end]
		if err := writeBatch(ctx, client, tableName, batch, userID); err != nil {
			log.Fatalf("Failed to write batch: %v", err)
		}

		fmt.Printf("Processed batch %d-%d\n", i+1, end)
		time.Sleep(100 * time.Millisecond) // Avoid throttling
	}

	fmt.Println("Data import completed successfully!")
}

func writeBatch(ctx context.Context, client *dynamodb.Client, tableName string, rawTransactions []RawTransaction, userID string) error {
	var writeRequests []types.WriteRequest

	for _, raw := range rawTransactions {
		transaction := transformTransaction(raw, userID)

		item, err := attributevalue.MarshalMap(transaction)
		if err != nil {
			return fmt.Errorf("failed to marshal transaction: %w", err)
		}

		writeRequests = append(writeRequests, types.WriteRequest{
			PutRequest: &types.PutRequest{
				Item: item,
			},
		})
	}

	input := &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			tableName: writeRequests,
		},
	}

	result, err := client.BatchWriteItem(ctx, input)
	if err != nil {
		return err
	}

	// Handle unprocessed items
	if len(result.UnprocessedItems) > 0 {
		fmt.Printf("Retrying %d unprocessed items...\n", len(result.UnprocessedItems[tableName]))
		time.Sleep(1 * time.Second)

		retryInput := &dynamodb.BatchWriteItemInput{
			RequestItems: result.UnprocessedItems,
		}

		_, err = client.BatchWriteItem(ctx, retryInput)
		if err != nil {
			return fmt.Errorf("failed to process retry batch: %w", err)
		}
	}

	return nil
}

func transformTransaction(raw RawTransaction, userID string) Transaction {
	txID := uuid.New().String()
	
	// Parse date string to time.Time
	parsedDate, err := time.Parse("2006-01-02", raw.Date)
	if err != nil {
		// If parsing fails, use current time
		log.Printf("Warning: Failed to parse date %s, using current time", raw.Date)
		parsedDate = time.Now().UTC()
	}

	now := time.Now().UTC()
	yearMonth := parsedDate.Format("2006-01")

	transaction := Transaction{
		ID:          txID,
		UserID:      userID,
		Date:        parsedDate,
		Amount:      raw.Amount,
		Category:    raw.Category,
		Type:        raw.Type,
		Description: raw.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Generate DynamoDB keys using backend patterns
	transaction.PK = fmt.Sprintf("USER#%s", userID)
	transaction.SK = fmt.Sprintf("TX#%s#%s", raw.Date, txID)
	transaction.GSI1PK = fmt.Sprintf("MONTH#%s#%s", yearMonth, userID)
	transaction.GSI1SK = fmt.Sprintf("TX#%s#%s", raw.Date, txID)
	transaction.GSI2PK = fmt.Sprintf("CATEGORY#%s#%s", strings.ToUpper(raw.Category), userID)
	transaction.GSI2SK = fmt.Sprintf("TX#%s#%s", raw.Date, txID)

	return transaction
}
