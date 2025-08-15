package config

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/joho/godotenv"
)

type Config struct {
	// Environment
	Environment string
	Region      string
	
	// DynamoDB
	DynamoDBTableName     string
	DynamoDBEndpoint      string // For local testing
	
	// AI Configuration
	AIProvider      string
	OpenAIAPIKey    string
	OpenAIAPIKeySSM string
	GroqAPIKey      string
	AIModel         string
	AIBaseURL       string
	
	// CORS
	CORSOrigins []string
	
	// AWS Config
	AWSConfig aws.Config
}

func Load() (*Config, error) {
	// Load environment variables from .env files
	if err := loadEnvFiles(); err != nil {
		// Don't fail if .env files are not found, just log the warning
		fmt.Printf("Warning: Could not load .env files: %v\n", err)
	}
	
	cfg := &Config{
		Environment:       getEnv("ENVIRONMENT", "dev"),
		Region:           getEnv("AWS_REGION", "us-east-1"),
		DynamoDBTableName: getEnv("DYNAMODB_TABLE_NAME", "stori-transactions-dev"),
		DynamoDBEndpoint:  getEnv("DYNAMODB_ENDPOINT", ""),
		AIProvider:        getEnv("AI_PROVIDER", "groq"),
		OpenAIAPIKeySSM:   getEnv("OPENAI_API_KEY_SSM", "/stori/dev/openai-api-key"),
		GroqAPIKey:        getEnv("GROQ_API_KEY", ""),
		CORSOrigins:       []string{
			getEnv("FRONTEND_URL", "http://localhost:3000"),
		},
	}
	
	// Set AI configuration based on provider
	if cfg.AIProvider == "groq" {
		cfg.OpenAIAPIKey = cfg.GroqAPIKey
		cfg.AIModel = getEnv("GROQ_MODEL", "llama3-8b-8192")
		cfg.AIBaseURL = "https://api.groq.com/openai/v1"
	} else {
		cfg.OpenAIAPIKey = getEnv("OPENAI_API_KEY", "")
		cfg.AIModel = getEnv("OPENAI_MODEL", "gpt-3.5-turbo")
		cfg.AIBaseURL = "https://api.openai.com/v1"
	}
	
	// Load AWS config
	awsConfig, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	cfg.AWSConfig = awsConfig
	
	// Load OpenAI/Groq API Key from environment first, then try SSM
	if cfg.OpenAIAPIKey == "" && cfg.OpenAIAPIKeySSM != "" && cfg.Environment == "prod" {
		apiKey, err := getSSMParameter(awsConfig, cfg.OpenAIAPIKeySSM)
		if err != nil {
			return nil, fmt.Errorf("failed to load AI API key from SSM: %w", err)
		}
		cfg.OpenAIAPIKey = apiKey
	}
	
	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getSSMParameter(cfg aws.Config, parameterName string) (string, error) {
	ssmClient := ssm.NewFromConfig(cfg)
	
	withDecryption := true
	result, err := ssmClient.GetParameter(context.TODO(), &ssm.GetParameterInput{
		Name:           &parameterName,
		WithDecryption: &withDecryption,
	})
	if err != nil {
		return "", err
	}
	
	if result.Parameter == nil || result.Parameter.Value == nil {
		return "", fmt.Errorf("parameter %s not found", parameterName)
	}
	
	return *result.Parameter.Value, nil
}

// loadEnvFiles loads environment variables from .env files in order of precedence:
// 1. .env.local (highest priority, for local development)
// 2. .env.development (for development environment)
// 3. .env (fallback, general configuration)
func loadEnvFiles() error {
	envFiles := []string{".env.local", ".env.development", ".env"}
	var lastErr error
	loaded := false
	
	for _, envFile := range envFiles {
		if err := godotenv.Load(envFile); err != nil {
			lastErr = err
			continue
		}
		fmt.Printf("Loaded environment variables from %s\n", envFile)
		loaded = true
		break // Load only the first found file (highest priority)
	}
	
	if !loaded {
		return fmt.Errorf("no .env files found: %v", lastErr)
	}
	
	return nil
}
