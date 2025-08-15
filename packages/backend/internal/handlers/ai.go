package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"backend/internal/models"
	"backend/internal/services"
)

type AIHandler struct {
	service services.AIService
}

func NewAIHandler(service services.AIService) *AIHandler {
	return &AIHandler{
		service: service,
	}
}

func (h *AIHandler) GetAdvice(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		http.Error(w, "AI service not available", http.StatusServiceUnavailable)
		return
	}

	var adviceRequest models.AIAdviceRequest
	
	if err := json.NewDecoder(r.Body).Decode(&adviceRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	advice, err := h.service.GetFinancialAdvice(r.Context(), &adviceRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(advice)
}

// GetPersonalizedAdvice provides personalized financial advice based on user's transaction data
func (h *AIHandler) GetPersonalizedAdvice(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		response := map[string]interface{}{
			"advice": "AI service is not configured. Please set GROQ_API_KEY or OPENAI_API_KEY environment variable.",
			"suggestions": []string{
				"Track your expenses regularly",
				"Create a monthly budget",
				"Save at least 20% of your income",
				"Reduce unnecessary expenses",
			},
			"timestamp": time.Now().Format(time.RFC3339),
			"provider": "mock",
			"mock": true,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id parameter is required", http.StatusBadRequest)
		return
	}

	// Build financial context first
	financialContext, err := h.service.BuildFinancialContext(r.Context(), userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to build financial context: %v", err), http.StatusInternalServerError)
		return
	}

	advice, err := h.service.GeneratePersonalizedAdvice(r.Context(), financialContext)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(advice)
}
