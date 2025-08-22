package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/handlers"
	"backend/internal/repository"
	"backend/internal/services"
)

func main() {
	// Check for health check flag
	if len(os.Args) > 1 && os.Args[1] == "--health-check" {
		healthCheck()
		return
	}

	startServer()
}

func healthCheck() {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("http://localhost:8080/api/v1/health")
	if err != nil {
		log.Printf("Health check failed: %v", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		log.Printf("Health check failed: status %d", resp.StatusCode)
		os.Exit(1)
	}
	
	os.Exit(0)
}

func startServer() {
	log.Println("Starting Stori Expense Tracker API Server...")

	// Start basic server first for health checks
	router := mux.NewRouter()
	
	// Add immediate health check that doesn't depend on DB
	router.PathPrefix("/api/v1").Subrouter().HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"healthy","timestamp":"`, time.Now().Format(time.RFC3339), `"}`)
	}).Methods("GET")

	// Configure CORS immediately
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	handler := c.Handler(router)

	// Server configuration
	port := getEnvOrDefault("PORT", "8080")
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Server starting on port %s", port)
	
	// Initialize services in the background
	go initializeServices(router)

	// Start server (this blocks)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func initializeServices(router *mux.Router) {
	ctx := context.Background()

	// Initialize DynamoDB client
	dbClient, err := database.NewDynamoDBClient(ctx)
	if err != nil {
		log.Printf("Warning: Failed to create DynamoDB client: %v", err)
		return
	}

	// Create tables if not exist
	if err := dbClient.CreateTablesIfNotExist(ctx); err != nil {
		log.Printf("Warning: Failed to create tables: %v", err)
	}

	// Initialize repository - use the full table name from config
	tableName := config.GetDynamoDBTableName()
	transactionRepo := repository.NewDynamoDBRepository(
		dbClient.Client,
		tableName,
	)

	// Initialize services
	transactionService := services.NewTransactionService(transactionRepo)
	budgetService := services.NewBudgetService(transactionRepo)
	analyticsService := services.NewAnalyticsService(transactionRepo)
	
	// Load configuration for AI service
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Warning: Failed to load config: %v", err)
		cfg = &config.Config{
			Environment:  "development",
			OpenAIAPIKey: getEnvOrDefault("OPENAI_API_KEY", ""), // Read from env
		}
	}
	
	aiService, err := services.NewAIService(cfg, transactionRepo)
	if err != nil {
		log.Printf("Warning: Failed to create AI service: %v", err)
		aiService = nil
	}

	// Initialize handlers
	transactionHandler := handlers.NewTransactionHandler(transactionService)
	budgetHandler := handlers.NewBudgetHandler(budgetService)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService)
	aiHandler := handlers.NewAIHandler(aiService)

	// Setup full routes
	setupFullRoutes(router, transactionHandler, budgetHandler, analyticsHandler, aiHandler)

	log.Printf("Services initialized successfully")
	log.Printf("Environment: %s", dbClient.Config.Environment)
	log.Printf("DynamoDB Table: %s", tableName)
}

func setupFullRoutes(
	router *mux.Router,
	transactionHandler *handlers.TransactionHandler,
	budgetHandler *handlers.BudgetHandler,
	analyticsHandler *handlers.AnalyticsHandler,
	aiHandler *handlers.AIHandler,
) {

	// API version prefix
	api := router.PathPrefix("/api/v1").Subrouter()

	// Transaction routes
	api.HandleFunc("/transactions", transactionHandler.CreateTransaction).Methods("POST")
	api.HandleFunc("/transactions", transactionHandler.GetTransactionsByUser).Methods("GET")
	api.HandleFunc("/transactions/{id}", transactionHandler.GetTransaction).Methods("GET")
	api.HandleFunc("/transactions/{id}", transactionHandler.UpdateTransaction).Methods("PUT")
	api.HandleFunc("/transactions/{id}", transactionHandler.DeleteTransaction).Methods("DELETE")

	// Advanced transaction queries
	api.HandleFunc("/transactions/month/{month}", transactionHandler.GetTransactionsByMonth).Methods("GET")
	api.HandleFunc("/transactions/category/{category}", transactionHandler.GetTransactionsByCategory).Methods("GET")

	// Analytics routes
	api.HandleFunc("/analytics/summary", analyticsHandler.GetSummary).Methods("GET")
	api.HandleFunc("/analytics/categories", analyticsHandler.GetCategoryBreakdown).Methods("GET")
	api.HandleFunc("/analytics/financial-summary", analyticsHandler.GetFinancialSummary).Methods("GET")
	api.HandleFunc("/analytics/months", analyticsHandler.GetMonthsWithTransactions).Methods("GET")

	// Budget routes
	api.HandleFunc("/budgets", budgetHandler.CreateOrUpdateBudget).Methods("POST")
	api.HandleFunc("/budgets/{month}", budgetHandler.GetBudgetsByMonth).Methods("GET")

	// AI advice routes
	api.HandleFunc("/ai/advice", aiHandler.GetAdvice).Methods("POST")
	api.HandleFunc("/ai/advisor", aiHandler.GetPersonalizedAdvice).Methods("GET")
	
	// Legacy analytics AI endpoint (backward compatibility)
	api.HandleFunc("/analytics/ai-advisor", aiHandler.GetPersonalizedAdvice).Methods("GET")

	// Health check
	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"healthy","timestamp":"`, time.Now().Format(time.RFC3339), `"}`)
	}).Methods("GET")

	// Add logging middleware
	api.Use(loggingMiddleware)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Wrap the response writer to capture status code
		wrappedWriter := &statusWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		next.ServeHTTP(wrappedWriter, r)
		
		duration := time.Since(start)
		log.Printf("%s %s %d %v", r.Method, r.URL.Path, wrappedWriter.statusCode, duration)
	})
}

type statusWriter struct {
	http.ResponseWriter
	statusCode int
}

func (sw *statusWriter) WriteHeader(statusCode int) {
	sw.statusCode = statusCode
	sw.ResponseWriter.WriteHeader(statusCode)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
