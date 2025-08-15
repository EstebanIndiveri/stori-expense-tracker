package smoke

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"backend/internal/models"
)

type SmokeTestSuite struct {
	suite.Suite
	baseURL   string
	client    *http.Client
	userID    string
	authToken string
}

func (suite *SmokeTestSuite) SetupSuite() {
	// Get API URL from environment or use default
	suite.baseURL = os.Getenv("API_BASE_URL")
	if suite.baseURL == "" {
		suite.baseURL = "http://localhost:8080" // Default for local development
	}

	suite.client = &http.Client{
		Timeout: 30 * time.Second,
	}

	// Test user ID for smoke tests
	suite.userID = "smoke-test-user-" + fmt.Sprintf("%d", time.Now().Unix())
	suite.authToken = "test-token" // In real scenarios, this would be a JWT
}

func (suite *SmokeTestSuite) TearDownSuite() {
	// Clean up any test data if needed
	// In a real scenario, you might want to delete test transactions
}

func (suite *SmokeTestSuite) TestHealthEndpoint() {
	resp, err := suite.client.Get(suite.baseURL + "/health")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var health map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&health)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "ok", health["status"])

	resp.Body.Close()
}

func (suite *SmokeTestSuite) TestCreateTransaction() {
	transaction := models.Transaction{
		Amount:      150.75,
		Description: "Smoke test transaction",
		Category:    "food",
		Type:        "expense",
		UserID:      suite.userID,
	}

	body, err := json.Marshal(transaction)
	assert.NoError(suite.T(), err)

	req, err := http.NewRequest("POST", suite.baseURL+"/api/v1/transactions", bytes.NewBuffer(body))
	assert.NoError(suite.T(), err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.authToken)
	req.Header.Set("X-User-ID", suite.userID)

	resp, err := suite.client.Do(req)
	assert.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)

	var createdTransaction models.Transaction
	err = json.NewDecoder(resp.Body).Decode(&createdTransaction)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), createdTransaction.ID)
	assert.Equal(suite.T(), transaction.Amount, createdTransaction.Amount)
	assert.Equal(suite.T(), transaction.Description, createdTransaction.Description)
}

func (suite *SmokeTestSuite) TestGetTransactions() {
	// First create a transaction to ensure we have data
	suite.TestCreateTransaction()

	req, err := http.NewRequest("GET", suite.baseURL+"/api/v1/transactions", nil)
	assert.NoError(suite.T(), err)

	req.Header.Set("Authorization", "Bearer "+suite.authToken)
	req.Header.Set("X-User-ID", suite.userID)

	// Add query parameters
	q := req.URL.Query()
	q.Add("limit", "10")
	req.URL.RawQuery = q.Encode()

	resp, err := suite.client.Do(req)
	assert.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var result []models.Transaction
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), len(result) > 0)
}

func (suite *SmokeTestSuite) TestGetTransactionsWithFilters() {
	// Test with category filter
	req, err := http.NewRequest("GET", suite.baseURL+"/api/v1/transactions", nil)
	assert.NoError(suite.T(), err)

	req.Header.Set("Authorization", "Bearer "+suite.authToken)
	req.Header.Set("X-User-ID", suite.userID)

	q := req.URL.Query()
	q.Add("category", "food")
	q.Add("month", "2024-01")
	q.Add("limit", "10")
	req.URL.RawQuery = q.Encode()

	resp, err := suite.client.Do(req)
	assert.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var transactions []models.Transaction
	err = json.NewDecoder(resp.Body).Decode(&transactions)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), transactions)
}

func (suite *SmokeTestSuite) TestGetAnalytics() {
	// First ensure we have some transactions
	suite.TestCreateTransaction()

	req, err := http.NewRequest("GET", suite.baseURL+"/api/v1/analytics", nil)
	assert.NoError(suite.T(), err)

	req.Header.Set("Authorization", "Bearer "+suite.authToken)
	req.Header.Set("X-User-ID", suite.userID)

	q := req.URL.Query()
	q.Add("months", "6")
	req.URL.RawQuery = q.Encode()

	resp, err := suite.client.Do(req)
	assert.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var analytics models.MonthlyAnalytics
	err = json.NewDecoder(resp.Body).Decode(&analytics)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), analytics.CategoryBreakdown)
}

func (suite *SmokeTestSuite) TestGetFinancialSummary() {
	// First ensure we have some transactions
	suite.TestCreateTransaction()

	req, err := http.NewRequest("GET", suite.baseURL+"/api/v1/analytics/financial-summary", nil)
	assert.NoError(suite.T(), err)

	req.Header.Set("Authorization", "Bearer "+suite.authToken)
	q := req.URL.Query()
	q.Add("user_id", suite.userID)
	req.URL.RawQuery = q.Encode()

	resp, err := suite.client.Do(req)
	assert.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var summary models.FinancialSummary
	err = json.NewDecoder(resp.Body).Decode(&summary)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), summary.CategoryBreakdown)
	assert.True(suite.T(), len(summary.CategoryBreakdown) > 0, "Category breakdown should not be empty")
}

func (suite *SmokeTestSuite) TestGetCategoryOptions() {
	// First ensure we have some transactions
	suite.TestCreateTransaction()

	req, err := http.NewRequest("GET", suite.baseURL+"/api/v1/analytics/categories", nil)
	assert.NoError(suite.T(), err)

	req.Header.Set("Authorization", "Bearer "+suite.authToken)
	q := req.URL.Query()
	q.Add("user_id", suite.userID)
	req.URL.RawQuery = q.Encode()

	resp, err := suite.client.Do(req)
	assert.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var categories []models.CategoryOption
	err = json.NewDecoder(resp.Body).Decode(&categories)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), len(categories) > 0, "Categories should not be empty")
	
	// Check the structure of category options
	if len(categories) > 0 {
		assert.NotEmpty(suite.T(), categories[0].Label)
		assert.NotEmpty(suite.T(), categories[0].Value)
	}
}

func (suite *SmokeTestSuite) TestGetAIAdvice() {
	// Skip AI test if OpenAI API key is not available
	if os.Getenv("OPENAI_API_KEY") == "" {
		suite.T().Skip("Skipping AI test - OPENAI_API_KEY not set")
	}

	// First ensure we have some transactions for analysis
	suite.TestCreateTransaction()

	req, err := http.NewRequest("POST", suite.baseURL+"/api/v1/ai/advice", nil)
	assert.NoError(suite.T(), err)

	req.Header.Set("Authorization", "Bearer "+suite.authToken)
	req.Header.Set("X-User-ID", suite.userID)

	resp, err := suite.client.Do(req)
	assert.NoError(suite.T(), err)
	defer resp.Body.Close()

	// AI service might be slower, so we allow for longer response times
	assert.True(suite.T(), resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusServiceUnavailable)

	if resp.StatusCode == http.StatusOK {
		var advice models.AIAdviceResponse
		err = json.NewDecoder(resp.Body).Decode(&advice)
		assert.NoError(suite.T(), err)
		assert.NotEmpty(suite.T(), advice.Advice)
	}
}

func (suite *SmokeTestSuite) TestErrorHandling() {
	// Test invalid transaction creation
	invalidTransaction := map[string]interface{}{
		"amount":      -100.0, // Invalid negative amount
		"description": "",     // Empty description
		"category":    "invalid_category",
		"type":        "invalid_type",
		"userID":      suite.userID,
	}

	body, err := json.Marshal(invalidTransaction)
	assert.NoError(suite.T(), err)

	req, err := http.NewRequest("POST", suite.baseURL+"/api/v1/transactions", bytes.NewBuffer(body))
	assert.NoError(suite.T(), err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.authToken)
	req.Header.Set("X-User-ID", suite.userID)

	resp, err := suite.client.Do(req)
	assert.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode)

	var errorResponse models.APIResponse
	err = json.NewDecoder(resp.Body).Decode(&errorResponse)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), errorResponse.Success)
	assert.NotNil(suite.T(), errorResponse.Error)
}

func (suite *SmokeTestSuite) TestUnauthorizedAccess() {
	// Test without authorization header
	req, err := http.NewRequest("GET", suite.baseURL+"/api/v1/transactions", nil)
	assert.NoError(suite.T(), err)

	resp, err := suite.client.Do(req)
	assert.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusUnauthorized, resp.StatusCode)
}

func (suite *SmokeTestSuite) TestCORSHeaders() {
	req, err := http.NewRequest("OPTIONS", suite.baseURL+"/api/v1/transactions", nil)
	assert.NoError(suite.T(), err)

	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")

	resp, err := suite.client.Do(req)
	assert.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	assert.NotEmpty(suite.T(), resp.Header.Get("Access-Control-Allow-Origin"))
	assert.NotEmpty(suite.T(), resp.Header.Get("Access-Control-Allow-Methods"))
	assert.NotEmpty(suite.T(), resp.Header.Get("Access-Control-Allow-Headers"))
}

func (suite *SmokeTestSuite) TestRateLimit() {
	// Test rate limiting by making multiple rapid requests
	const numRequests = 10
	successCount := 0
	rateLimitCount := 0

	for i := 0; i < numRequests; i++ {
		req, err := http.NewRequest("GET", suite.baseURL+"/health", nil)
		assert.NoError(suite.T(), err)

		resp, err := suite.client.Do(req)
		assert.NoError(suite.T(), err)

		switch resp.StatusCode {
		case http.StatusOK:
			successCount++
		case http.StatusTooManyRequests:
			rateLimitCount++
		}

		resp.Body.Close()
	}

	// At least some requests should succeed
	assert.True(suite.T(), successCount > 0)

	// If rate limiting is enabled, we might see some rate limit responses
	// This is environment dependent, so we don't assert on it
	suite.T().Logf("Successful requests: %d, Rate limited: %d", successCount, rateLimitCount)
}

func (suite *SmokeTestSuite) TestResponseTimes() {
	endpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/health"},
		{"GET", "/api/v1/transactions"},
		{"GET", "/api/v1/analytics"},
	}

	for _, endpoint := range endpoints {
		suite.T().Run(fmt.Sprintf("%s %s", endpoint.method, endpoint.path), func(t *testing.T) {
			start := time.Now()

			req, err := http.NewRequest(endpoint.method, suite.baseURL+endpoint.path, nil)
			assert.NoError(t, err)

			if endpoint.path != "/health" {
				req.Header.Set("Authorization", "Bearer "+suite.authToken)
				req.Header.Set("X-User-ID", suite.userID)
			}

			resp, err := suite.client.Do(req)
			assert.NoError(t, err)
			defer resp.Body.Close()

			duration := time.Since(start)

			// Response should be reasonably fast (under 5 seconds for smoke tests)
			assert.True(t, duration < 5*time.Second, "Response took too long: %v", duration)

			t.Logf("Response time for %s %s: %v", endpoint.method, endpoint.path, duration)
		})
	}
}

func TestSmokeTestSuite(t *testing.T) {
	// Skip smoke tests if API_BASE_URL is not set
	if os.Getenv("API_BASE_URL") == "" && os.Getenv("RUN_SMOKE_TESTS") == "" {
		t.Skip("Skipping smoke tests - set API_BASE_URL or RUN_SMOKE_TESTS to run")
	}

	suite.Run(t, new(SmokeTestSuite))
}

// TestEndToEndFlow tests a complete user flow
func (suite *SmokeTestSuite) TestEndToEndFlow() {
	// 1. Create multiple transactions
	transactions := []models.Transaction{
		{
			Amount:      100.50,
			Description: "Grocery shopping",
			Category:    "food",
			Type:        "expense",
			UserID:      suite.userID,
		},
		{
			Amount:      2500.00,
			Description: "Monthly salary",
			Category:    "salary",
			Type:        "income",
			UserID:      suite.userID,
		},
		{
			Amount:      50.00,
			Description: "Gas",
			Category:    "transportation",
			Type:        "expense",
			UserID:      suite.userID,
		},
	}

	createdIDs := make([]string, 0, len(transactions))

	// Create transactions
	for _, transaction := range transactions {
		body, err := json.Marshal(transaction)
		assert.NoError(suite.T(), err)

		req, err := http.NewRequest("POST", suite.baseURL+"/api/v1/transactions", bytes.NewBuffer(body))
		assert.NoError(suite.T(), err)

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+suite.authToken)
		req.Header.Set("X-User-ID", suite.userID)

		resp, err := suite.client.Do(req)
		assert.NoError(suite.T(), err)

		assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)

		var created models.Transaction
		err = json.NewDecoder(resp.Body).Decode(&created)
		assert.NoError(suite.T(), err)
		createdIDs = append(createdIDs, created.ID)

		resp.Body.Close()
	}

	// 2. Retrieve all transactions
	req, err := http.NewRequest("GET", suite.baseURL+"/api/v1/transactions", nil)
	assert.NoError(suite.T(), err)

	req.Header.Set("Authorization", "Bearer "+suite.authToken)
	req.Header.Set("X-User-ID", suite.userID)

	resp, err := suite.client.Do(req)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var result []models.Transaction
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), len(result) >= len(transactions))

	resp.Body.Close()

	// 3. Get analytics
	req, err = http.NewRequest("GET", suite.baseURL+"/api/v1/analytics", nil)
	assert.NoError(suite.T(), err)

	req.Header.Set("Authorization", "Bearer "+suite.authToken)
	req.Header.Set("X-User-ID", suite.userID)

	resp, err = suite.client.Do(req)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var analytics models.MonthlyAnalytics
	err = json.NewDecoder(resp.Body).Decode(&analytics)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), analytics.TotalIncome >= 2500.00)
	assert.True(suite.T(), analytics.TotalExpense >= 150.50)

	resp.Body.Close()

	suite.T().Logf("End-to-end flow completed successfully. Created %d transactions.", len(createdIDs))
}
