package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
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

type Transaction struct {
	PK          string    `dynamodbav:"pk"`
	SK          string    `dynamodbav:"sk"`
	TxID        string    `dynamodbav:"txId"`
	Date        string    `dynamodbav:"date"`
	YearMonth   string    `dynamodbav:"yyyymm"`
	Amount      float64   `dynamodbav:"amount"`
	Category    string    `dynamodbav:"category"`
	Type        string    `dynamodbav:"type"`
	Description string    `dynamodbav:"description"`
	GSI1PK      string    `dynamodbav:"gsi1pk"`
	GSI1SK      string    `dynamodbav:"gsi1sk"`
	GSI2PK      string    `dynamodbav:"gsi2pk"`
	GSI2SK      string    `dynamodbav:"gsi2sk"`
	CreatedAt   time.Time `dynamodbav:"createdAt"`
	UpdatedAt   time.Time `dynamodbav:"updatedAt"`
}

func main() {
	ctx := context.Background()

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
	fmt.Printf("Processing %d transactions...\n", len(rawTransactions))

	for i := 0; i < len(rawTransactions); i += 25 {
		end := i + 25
		if end > len(rawTransactions) {
			end = len(rawTransactions)
		}

		batch := rawTransactions[i:end]
		if err := writeBatch(ctx, client, tableName, batch); err != nil {
			log.Fatalf("Failed to write batch: %v", err)
		}

		fmt.Printf("Processed batch %d-%d\n", i+1, end)
		time.Sleep(100 * time.Millisecond) // Avoid throttling
	}

	fmt.Println("Data import completed successfully!")
}

func writeBatch(ctx context.Context, client *dynamodb.Client, tableName string, rawTransactions []RawTransaction) error {
	var writeRequests []types.WriteRequest

	for _, raw := range rawTransactions {
		transaction := transformTransaction(raw)

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

func transformTransaction(raw RawTransaction) Transaction {
	txID := uuid.New().String()
	date := raw.Date
	yearMonth := date[:7] // "2024-01"

	now := time.Now().UTC()

	return Transaction{
		// Primary key
		PK: fmt.Sprintf("DS#v1#M#%s", yearMonth),
		SK: fmt.Sprintf("D#%s#TX#%s", date, txID),

		// Core attributes
		TxID:        txID,
		Date:        date,
		YearMonth:   yearMonth,
		Amount:      raw.Amount,
		Category:    raw.Category,
		Type:        raw.Type,
		Description: raw.Description,

		// GSI keys
		GSI1PK: fmt.Sprintf("CAT#%s#M#%s", raw.Category, yearMonth),
		GSI1SK: fmt.Sprintf("D#%s#TX#%s", date, txID),
		GSI2PK: fmt.Sprintf("T#%s#M#%s", raw.Type, yearMonth),
		GSI2SK: fmt.Sprintf("D#%s#TX#%s", date, txID),

		// Metadata
		CreatedAt: now,
		UpdatedAt: now,
	}
}
