package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"backend/internal/config"
	"backend/internal/models"
	"backend/internal/repository"
	"backend/internal/services"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type AIApp struct {
	aiService services.AIService
}

func NewAIApp() (*AIApp, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	// Create DynamoDB client from config
	dynamoClient := dynamodb.NewFromConfig(cfg.AWSConfig)
	
	// Initialize repository with proper parameters
	repo := repository.NewDynamoDBRepository(dynamoClient, cfg.DynamoDBTableName)

	// Initialize AI service
	aiService, err := services.NewAIService(cfg, repo)
	if err != nil {
		return nil, err
	}

	return &AIApp{
		aiService: aiService,
	}, nil
}

func (app *AIApp) Handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("Processing AI request: %s %s", request.HTTPMethod, request.Path)
	
	// Handle different AI endpoints
	switch request.Path {
	case "/ai/advice":
		return app.handleAdviceRequest(ctx, request)
	case "/ai/advisor":
		return app.handleAdvisorRequest(ctx, request)
	default:
		return events.APIGatewayProxyResponse{
			StatusCode: 404,
			Body:       `{"error": "endpoint not found"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}, nil
	}
}

func (app *AIApp) handleAdviceRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if request.HTTPMethod != "POST" {
		return events.APIGatewayProxyResponse{
			StatusCode: 405,
			Body:       `{"error": "method not allowed"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}, nil
	}

	var adviceRequest models.AIAdviceRequest
	if err := json.Unmarshal([]byte(request.Body), &adviceRequest); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       `{"error": "invalid JSON"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}, nil
	}

	response, err := app.aiService.GetFinancialAdvice(ctx, &adviceRequest)
	if err != nil {
		log.Printf("Error getting AI advice: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       `{"error": "internal server error"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}, nil
	}

	responseBody, err := json.Marshal(response)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       `{"error": "failed to marshal response"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(responseBody),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}

func (app *AIApp) handleAdvisorRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if request.HTTPMethod != "GET" {
		return events.APIGatewayProxyResponse{
			StatusCode: 405,
			Body:       `{"error": "method not allowed"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}, nil
	}

	userID := request.QueryStringParameters["user_id"]
	if userID == "" {
		userID = "default-user"
	}

	// Build financial context for the user
	context, err := app.aiService.BuildFinancialContext(ctx, userID)
	if err != nil {
		log.Printf("Error building financial context: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       `{"error": "internal server error"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}, nil
	}

	// Generate personalized advice
	response, err := app.aiService.GeneratePersonalizedAdvice(ctx, context)
	if err != nil {
		log.Printf("Error getting personalized advice: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       `{"error": "internal server error"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}, nil
	}

	responseBody, err := json.Marshal(response)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       `{"error": "failed to marshal response"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(responseBody),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}

func main() {
	app, err := NewAIApp()
	if err != nil {
		log.Fatalf("Failed to initialize AI app: %v", err)
	}

	if os.Getenv("AWS_LAMBDA_RUNTIME_API") == "" {
		log.Fatal("AI advisor function should only run in Lambda environment")
	}

	lambda.Start(app.Handler)
}
