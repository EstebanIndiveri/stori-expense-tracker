package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	TransactionsTable = "transactions"
	UsersTable        = "users"
)

type DynamoDBClient struct {
	Client *dynamodb.Client
	Config Config
}

type Config struct {
	Environment  string // "local" or "aws"
	Region       string
	LocalPort    string
	TablePrefix  string
}

// NewDynamoDBClient creates a new DynamoDB client supporting both local and AWS environments
func NewDynamoDBClient(ctx context.Context) (*DynamoDBClient, error) {
	cfg := Config{
		Environment: getEnvOrDefault("ENVIRONMENT", "local"),
		Region:      getEnvOrDefault("AWS_REGION", "us-east-1"),
		LocalPort:   getEnvOrDefault("DYNAMODB_LOCAL_PORT", "8000"),
		TablePrefix: getEnvOrDefault("TABLE_PREFIX", "stori"),
	}

	var awsConfig aws.Config
	var err error

	if cfg.Environment == "local" {
		// Local DynamoDB configuration
		awsConfig, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(cfg.Region),
			config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
				Value: aws.Credentials{
					AccessKeyID:     "local",
					SecretAccessKey: "local",
				},
			}),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load local config: %w", err)
		}

		client := dynamodb.NewFromConfig(awsConfig, func(o *dynamodb.Options) {
			o.BaseEndpoint = aws.String(fmt.Sprintf("http://localhost:%s", cfg.LocalPort))
		})

		return &DynamoDBClient{
			Client: client,
			Config: cfg,
		}, nil
	}

	// AWS DynamoDB configuration
	awsConfig, err = config.LoadDefaultConfig(ctx, config.WithRegion(cfg.Region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := dynamodb.NewFromConfig(awsConfig)

	return &DynamoDBClient{
		Client: client,
		Config: cfg,
	}, nil
}

// CreateTablesIfNotExist creates the required tables with optimized schema
func (db *DynamoDBClient) CreateTablesIfNotExist(ctx context.Context) error {
	// Create transactions table with GSI for monthly and category queries
	err := db.createTransactionsTable(ctx)
	if err != nil {
		return fmt.Errorf("failed to create transactions table: %w", err)
	}

	// Create users table
	err = db.createUsersTable(ctx)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	log.Printf("Tables created successfully in %s environment", db.Config.Environment)
	return nil
}

func (db *DynamoDBClient) createTransactionsTable(ctx context.Context) error {
	tableName := db.getTableName(TransactionsTable)

	// Check if table exists
	_, err := db.Client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})
	if err == nil {
		log.Printf("Table %s already exists", tableName)
		return nil
	}

	input := &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("PK"),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: aws.String("SK"),
				KeyType:       types.KeyTypeRange,
			},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("PK"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("SK"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("GSI1PK"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("GSI1SK"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("GSI2PK"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("GSI2SK"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("GSI1"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("GSI1PK"),
						KeyType:       types.KeyTypeHash,
					},
					{
						AttributeName: aws.String("GSI1SK"),
						KeyType:       types.KeyTypeRange,
					},
				},
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll,
				},
				ProvisionedThroughput: &types.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(5),
					WriteCapacityUnits: aws.Int64(5),
				},
			},
			{
				IndexName: aws.String("GSI2"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("GSI2PK"),
						KeyType:       types.KeyTypeHash,
					},
					{
						AttributeName: aws.String("GSI2SK"),
						KeyType:       types.KeyTypeRange,
					},
				},
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll,
				},
				ProvisionedThroughput: &types.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(5),
					WriteCapacityUnits: aws.Int64(5),
				},
			},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
	}

	_, err = db.Client.CreateTable(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create table %s: %w", tableName, err)
	}

	log.Printf("Table %s created successfully", tableName)
	return nil
}

func (db *DynamoDBClient) createUsersTable(ctx context.Context) error {
	tableName := db.getTableName(UsersTable)

	// Check if table exists
	_, err := db.Client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})
	if err == nil {
		log.Printf("Table %s already exists", tableName)
		return nil
	}

	input := &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("PK"),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: aws.String("SK"),
				KeyType:       types.KeyTypeRange,
			},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("PK"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("SK"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
	}

	_, err = db.Client.CreateTable(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create table %s: %w", tableName, err)
	}

	log.Printf("Table %s created successfully", tableName)
	return nil
}

func (db *DynamoDBClient) getTableName(tableName string) string {
	if db.Config.TablePrefix != "" {
		return fmt.Sprintf("%s-%s", db.Config.TablePrefix, tableName)
	}
	return tableName
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
