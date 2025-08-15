package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"backend/internal/services"

	"github.com/gorilla/mux"
)

type BudgetHandler struct {
	budgetService *services.BudgetService
}

func NewBudgetHandler(budgetService *services.BudgetService) *BudgetHandler {
	return &BudgetHandler{
		budgetService: budgetService,
	}
}

// CreateOrUpdateBudgetRequest represents the request body for budget creation/update
type CreateOrUpdateBudgetRequest struct {
	Month    string  `json:"month"`    // Format: YYYY-MM
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
}

// CreateOrUpdateBudget handles POST /budgets requests
func (h *BudgetHandler) CreateOrUpdateBudget(w http.ResponseWriter, r *http.Request) {
	// For this demo, we'll use a hardcoded user ID
	userID := "user-123"

	var req CreateOrUpdateBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Month == "" {
		http.Error(w, "Month is required (format: YYYY-MM)", http.StatusBadRequest)
		return
	}
	if req.Category == "" {
		http.Error(w, "Category is required", http.StatusBadRequest)
		return
	}
	if req.Amount < 0 {
		http.Error(w, "Amount cannot be negative", http.StatusBadRequest)
		return
	}

	// Create or update the budget
	err := h.budgetService.CreateOrUpdateBudget(r.Context(), userID, req.Month, req.Category, req.Amount)
	if err != nil {
		log.Printf("Error creating/updating budget: %v", err)
		http.Error(w, fmt.Sprintf("Failed to create/update budget: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success response
	response := map[string]interface{}{
		"message":  "Budget created/updated successfully",
		"user_id":  userID,
		"month":    req.Month,
		"category": req.Category,
		"amount":   req.Amount,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetBudgetsByMonth handles GET /budgets/{month} requests
func (h *BudgetHandler) GetBudgetsByMonth(w http.ResponseWriter, r *http.Request) {
	userID := "user-123"
	
	vars := mux.Vars(r)
	month := vars["month"]
	
	if month == "" {
		http.Error(w, "Month parameter is required", http.StatusBadRequest)
		return
	}

	budgets, err := h.budgetService.GetBudgetsByMonth(r.Context(), userID, month)
	if err != nil {
		log.Printf("Error getting budgets for month %s: %v", month, err)
		http.Error(w, fmt.Sprintf("Failed to get budgets: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(budgets)
}

// GetBudget handles GET /budgets/{month}/{category} requests
func (h *BudgetHandler) GetBudget(w http.ResponseWriter, r *http.Request) {
	userID := "user-123"
	
	vars := mux.Vars(r)
	month := vars["month"]
	category := vars["category"]
	
	if month == "" || category == "" {
		http.Error(w, "Month and category parameters are required", http.StatusBadRequest)
		return
	}

	budget, err := h.budgetService.GetBudget(r.Context(), userID, month, category)
	if err != nil {
		log.Printf("Error getting budget for month %s, category %s: %v", month, category, err)
		http.Error(w, fmt.Sprintf("Failed to get budget: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(budget)
}

// DeleteBudget handles DELETE /budgets/{month}/{category} requests
func (h *BudgetHandler) DeleteBudget(w http.ResponseWriter, r *http.Request) {
	userID := "user-123"
	
	vars := mux.Vars(r)
	month := vars["month"]
	category := vars["category"]
	
	if month == "" || category == "" {
		http.Error(w, "Month and category parameters are required", http.StatusBadRequest)
		return
	}

	err := h.budgetService.DeleteBudget(r.Context(), userID, month, category)
	if err != nil {
		log.Printf("Error deleting budget for month %s, category %s: %v", month, category, err)
		http.Error(w, fmt.Sprintf("Failed to delete budget: %v", err), http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"message":  "Budget deleted successfully",
		"month":    month,
		"category": category,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetBudgetUtilization handles GET /budgets/{month}/utilization requests
func (h *BudgetHandler) GetBudgetUtilization(w http.ResponseWriter, r *http.Request) {
	userID := "user-123"
	
	vars := mux.Vars(r)
	month := vars["month"]
	
	if month == "" {
		http.Error(w, "Month parameter is required", http.StatusBadRequest)
		return
	}

	utilization, err := h.budgetService.GetBudgetUtilization(r.Context(), userID, month)
	if err != nil {
		log.Printf("Error getting budget utilization for month %s: %v", month, err)
		http.Error(w, fmt.Sprintf("Failed to get budget utilization: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(utilization)
}
