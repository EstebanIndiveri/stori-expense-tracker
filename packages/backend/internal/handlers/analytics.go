package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"backend/internal/services"

	"github.com/gorilla/mux"
)

type AnalyticsHandler struct {
	service services.AnalyticsService
}

func NewAnalyticsHandler(service services.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		service: service,
	}
}

func (h *AnalyticsHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	month := r.URL.Query().Get("month")
	if month == "" {
		// Use current month if not specified
		month = time.Now().Format("2006-01")
	}

	// Check if budget information should be included
	includeBudgets := r.URL.Query().Get("include_budgets") == "true"
	
	if includeBudgets {
		// Return analytics with budget information
		analytics, err := h.service.GetFinancialSummaryWithBudgets(r.Context(), userID, month)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(analytics)
	} else {
		// Return regular monthly analytics
		analytics, err := h.service.GetMonthlyAnalytics(r.Context(), userID, month)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(analytics)
	}
}

func (h *AnalyticsHandler) GetMonthlyAnalytics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	month := vars["month"]
	
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	if month == "" {
		http.Error(w, "month is required", http.StatusBadRequest)
		return
	}

	analytics, err := h.service.GetMonthlyAnalytics(r.Context(), userID, month)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics)
}

func (h *AnalyticsHandler) GetCategoryBreakdown(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	// Check if they want just category names or full breakdown
	simple := r.URL.Query().Get("simple")
	if simple == "true" {
		// Return category options with label and value
		categoryOptions, err := h.service.GetCategoryOptions(r.Context(), userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    categoryOptions,
		})
		return
	}

	// Return full breakdown with amounts and percentages
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "month"
	}

	breakdown, err := h.service.GetCategoryBreakdown(r.Context(), userID, period)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    breakdown,
	})
}

// GetFinancialSummary returns comprehensive financial summary for dashboard
func (h *AnalyticsHandler) GetFinancialSummary(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	summary, err := h.service.GetFinancialSummary(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    summary,
	})
}

// GetMonthsWithTransactions returns list of months that have transactions
func (h *AnalyticsHandler) GetMonthsWithTransactions(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	months, err := h.service.GetMonthsWithTransactions(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    months,
	})
}
