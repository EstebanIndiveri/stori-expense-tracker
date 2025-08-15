package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"backend/internal/config"
	"backend/internal/models"
	"backend/internal/repository"

	openai "github.com/sashabaranov/go-openai"
)

type AIService interface {
	GetFinancialAdvice(ctx context.Context, request *models.AIAdviceRequest) (*models.AIAdviceResponse, error)
	GeneratePersonalizedAdvice(ctx context.Context, userContext *models.FinancialContext) (*models.AIAdviceResponse, error)
	BuildFinancialContext(ctx context.Context, userID string) (*models.FinancialContext, error)
}

type aiService struct {
	client         *openai.Client
	repo           repository.Repository
	analyticsService AnalyticsService
	config         *config.Config
	isGroq         bool
}

func NewAIService(cfg *config.Config, repo repository.Repository) (AIService, error) {
	if cfg.OpenAIAPIKey == "" {
		return nil, fmt.Errorf("AI API key is required")
	}
	
	// Create OpenAI client with custom config for Groq
	clientConfig := openai.DefaultConfig(cfg.OpenAIAPIKey)
	
	// Configure for Groq if needed
	isGroq := cfg.AIProvider == "groq"
	if isGroq {
		clientConfig.BaseURL = cfg.AIBaseURL
	}
	
	client := openai.NewClientWithConfig(clientConfig)
	analyticsService := NewAnalyticsService(repo)
	
	return &aiService{
		client:           client,
		repo:             repo,
		analyticsService: analyticsService,
		config:           cfg,
		isGroq:           isGroq,
	}, nil
}

func (s *aiService) GetFinancialAdvice(ctx context.Context, request *models.AIAdviceRequest) (*models.AIAdviceResponse, error) {
	// Get user's financial context
	financialContext, err := s.buildFinancialContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build financial context: %w", err)
	}
	
	// Generate AI prompt
	prompt := s.buildAdvicePrompt(request.Question, financialContext)
	
	// Get model name
	model := s.config.AIModel
	if model == "" {
		if s.isGroq {
			model = "llama3-8b-8192" // Groq's fastest free model
		} else {
			model = "gpt-3.5-turbo"
		}
	}
	
	// Call AI API (OpenAI compatible)
	resp, err := s.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: s.getSystemPrompt(),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		MaxTokens:   500,
		Temperature: 0.7,
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to get AI response: %w", err)
	}
	
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI")
	}
	
	advice := resp.Choices[0].Message.Content
	suggestions := s.extractSuggestions(advice)
	
	return &models.AIAdviceResponse{
		Advice:      advice,
		Suggestions: suggestions,
		Context:     financialContext,
		Timestamp:   time.Now().UTC(),
		Provider:    s.config.AIProvider,
		Model:       model,
	}, nil
}

func (s *aiService) GeneratePersonalizedAdvice(ctx context.Context, userContext *models.FinancialContext) (*models.AIAdviceResponse, error) {
	prompt := s.buildPersonalizedPrompt(userContext)
	
	// Get model name
	model := s.config.AIModel
	if model == "" {
		if s.isGroq {
			model = "llama3-8b-8192"
		} else {
			model = "gpt-3.5-turbo"
		}
	}
	
	resp, err := s.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: s.getSystemPrompt(),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		MaxTokens:   500,
		Temperature: 0.7,
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to get AI response: %w", err)
	}
	
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI")
	}
	
	advice := resp.Choices[0].Message.Content
	suggestions := s.extractSuggestions(advice)
	
	return &models.AIAdviceResponse{
		Advice:      advice,
		Suggestions: suggestions,
		Context:     userContext,
		Timestamp:   time.Now().UTC(),
		Provider:    s.config.AIProvider,
		Model:       model,
	}, nil
}

func (s *aiService) BuildFinancialContext(ctx context.Context, userID string) (*models.FinancialContext, error) {
	// Obtener TODAS las transacciones históricas del usuario
	allTransactions, _, err := s.repo.GetTransactionsByUser(ctx, userID, 1000, nil) // Aumentar límite para obtener historial completo
	if err != nil {
		return nil, fmt.Errorf("failed to get user transactions: %w", err)
	}
	
	// Calcular métricas financieras totales
	totalIncome := 0.0
	totalExpenses := 0.0
	categoryTotals := make(map[string]float64)
	categoryCounts := make(map[string]int)
	
	// Procesar todas las transacciones históricas
	for _, transaction := range allTransactions {
		if transaction.Type == "income" {
			totalIncome += transaction.Amount
		} else if transaction.Type == "expense" {
			totalExpenses += math.Abs(transaction.Amount) // Convertir a positivo para cálculos
		}
		
		// Acumular por categoría (mantenemos los montos originales para el contexto)
		categoryTotals[transaction.Category] += transaction.Amount
		categoryCounts[transaction.Category]++
	}
	
	// Calcular saldo actual total (ingresos - gastos)
	currentBalance := totalIncome - totalExpenses
	
	// Calcular savings rate: (saldo actual / ingresos totales) * 100
	savingsRate := 0.0
	if totalIncome > 0 {
		savingsRate = (currentBalance / totalIncome) * 100
	}
	
	// Crear categorías principales ordenadas por monto (solo gastos para el breakdown)
	var topCategories []*models.CategorySummary
	for category, amount := range categoryTotals {
		if amount < 0 { // Solo gastos para el breakdown de categorías
			percentage := 0.0
			if totalExpenses > 0 {
				percentage = (math.Abs(amount) / totalExpenses) * 100
			}
			
			topCategories = append(topCategories, &models.CategorySummary{
				Category:    category,
				TotalAmount: amount, // Mantener negativo para mostrar como gasto
				Percentage:  percentage,
				Count:       categoryCounts[category],
			})
		}
	}
	
	// Ordenar categorías por monto total (mayor gasto primero)
	sort.Slice(topCategories, func(i, j int) bool {
		return math.Abs(topCategories[i].TotalAmount) > math.Abs(topCategories[j].TotalAmount)
	})
	
	// Generar insights basados en datos históricos
	spendingTrends := s.generateHistoricalInsights(totalIncome, totalExpenses, currentBalance, topCategories)
	
	return &models.FinancialContext{
		MonthlyIncome:  totalIncome,  // Ahora es ingresos totales históricos
		MonthlyExpense: -totalExpenses, // Mantener negativo para consistencia
		SavingsRate:    savingsRate,
		TopCategories:  topCategories,
		SpendingTrends: spendingTrends,
	}, nil
}

func (s *aiService) buildFinancialContext(ctx context.Context) (*models.FinancialContext, error) {
	// Default user for backward compatibility
	return s.BuildFinancialContext(ctx, "user-123")
}

func (s *aiService) getSystemPrompt() string {
	if s.isGroq {
		return `Eres un asesor financiero experto de Stori, una fintech mexicana. 

Tu tarea es proporcionar consejos financieros prácticos y accionables basados en los datos de transacciones del usuario.

Instrucciones:
1. Responde en español claro y directo
2. Enfócate en recomendaciones específicas y realizables
3. Usa los datos financieros proporcionados para personalizar tu respuesta
4. Mantén un tono profesional pero amigable
5. Prioriza el ahorro, presupuesto y optimización de gastos
6. Limita tu respuesta a 3-4 puntos clave máximo

Formato de respuesta: Párrafos claros con recomendaciones numeradas al final.`
	}
	
	return `You are a professional financial advisor for Stori, a Mexican fintech company. 
Your role is to provide helpful, practical financial advice based on the user's spending data.

Guidelines:
1. Provide specific, actionable advice
2. Focus on realistic budget adjustments
3. Highlight spending patterns and opportunities for savings
4. Consider the Mexican financial context
5. Be encouraging but honest about financial habits
6. Suggest concrete steps the user can take
7. Keep responses concise and clear
8. Use a friendly, professional tone

Always base your advice on the provided financial data and avoid generic responses.`
}

func (s *aiService) buildAdvicePrompt(question string, context *models.FinancialContext) string {
	var prompt strings.Builder
	
	prompt.WriteString(fmt.Sprintf("Pregunta del Usuario: %s\n\n", question))
	prompt.WriteString("Contexto Financiero Histórico Completo:\n")
	prompt.WriteString(fmt.Sprintf("- Ingresos Totales Históricos: $%.2f\n", context.MonthlyIncome))
	prompt.WriteString(fmt.Sprintf("- Gastos Totales Históricos: $%.2f\n", math.Abs(context.MonthlyExpense)))
	
	// Calcular y mostrar balance actual
	currentBalance := context.MonthlyIncome + context.MonthlyExpense // MonthlyExpense ya es negativo
	prompt.WriteString(fmt.Sprintf("- Balance Actual Total: $%.2f\n", currentBalance))
	prompt.WriteString(fmt.Sprintf("- Tasa de Ahorro: %.1f%%\n", context.SavingsRate))
	
	if len(context.TopCategories) > 0 {
		prompt.WriteString("\nDesglose Histórico de Gastos por Categoría:\n")
		for i, category := range context.TopCategories {
			if i >= 10 { // Limitar a top 10 para no sobrecargar el prompt
				break
			}
			prompt.WriteString(fmt.Sprintf("- %s: $%.2f (%.1f%% del total, %d transacciones)\n", 
				category.Category, math.Abs(category.TotalAmount), category.Percentage, category.Count))
		}
	}
	
	if len(context.SpendingTrends) > 0 {
		prompt.WriteString("\nInsights Financieros:\n")
		for _, trend := range context.SpendingTrends {
			prompt.WriteString(fmt.Sprintf("- %s\n", trend))
		}
	}
	
	prompt.WriteString("\nBasándote en este historial financiero completo, proporciona consejos específicos y accionables.")
	
	return prompt.String()
}

func (s *aiService) buildPersonalizedPrompt(context *models.FinancialContext) string {
	var prompt strings.Builder
	
	prompt.WriteString("Genera consejos financieros personalizados basados en el perfil histórico completo de este usuario:\n\n")
	prompt.WriteString(fmt.Sprintf("Ingresos Totales Históricos: $%.2f\n", context.MonthlyIncome))
	prompt.WriteString(fmt.Sprintf("Gastos Totales Históricos: $%.2f\n", math.Abs(context.MonthlyExpense)))
	
	// Calcular balance actual
	currentBalance := context.MonthlyIncome + context.MonthlyExpense
	prompt.WriteString(fmt.Sprintf("Balance Actual: $%.2f\n", currentBalance))
	prompt.WriteString(fmt.Sprintf("Tasa de Ahorro: %.1f%%\n", context.SavingsRate))
	
	if len(context.TopCategories) > 0 {
		prompt.WriteString("\nDesglose de Gastos Históricos:\n")
		for _, category := range context.TopCategories {
			prompt.WriteString(fmt.Sprintf("- %s: $%.2f (%.1f%%, %d transacciones)\n", 
				category.Category, math.Abs(category.TotalAmount), category.Percentage, category.Count))
		}
	}
	
	prompt.WriteString("\nProporciona 3-4 recomendaciones específicas y accionables para mejorar su situación financiera basándote en su historial completo.")
	
	return prompt.String()
}

func (s *aiService) extractSuggestions(advice string) []string {
	var suggestions []string
	
	// Simple extraction based on numbered lists or bullet points
	lines := strings.Split(advice, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		
		// Check for numbered items (1., 2., etc.)
		if len(line) > 2 && line[1] == '.' && line[0] >= '1' && line[0] <= '9' {
			suggestions = append(suggestions, strings.TrimSpace(line[2:]))
			continue
		}
		
		// Check for bullet points
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "• ") {
			suggestions = append(suggestions, strings.TrimSpace(line[2:]))
			continue
		}
		
		// Check for asterisk bullets
		if strings.HasPrefix(line, "* ") {
			suggestions = append(suggestions, strings.TrimSpace(line[2:]))
			continue
		}
	}
	
	return suggestions
}

func (s *aiService) generateHistoricalInsights(totalIncome, totalExpenses, currentBalance float64, topCategories []*models.CategorySummary) []string {
	var insights []string
	
	// Insight sobre el balance general
	if currentBalance > 0 {
		insights = append(insights, fmt.Sprintf("¡Excelente! Has ahorrado $%.2f en total", currentBalance))
	} else {
		insights = append(insights, fmt.Sprintf("Tienes un balance negativo de $%.2f - necesitas revisar tus gastos", math.Abs(currentBalance)))
	}
	
	// Insight sobre la categoría de mayor gasto
	if len(topCategories) > 0 {
		topCategory := topCategories[0]
		insights = append(insights, fmt.Sprintf("Tu categoría de mayor gasto es %s con $%.2f (%.1f%% del total)", 
			topCategory.Category, math.Abs(topCategory.TotalAmount), topCategory.Percentage))
	}
	
	// Insight sobre diversificación de gastos
	if len(topCategories) >= 3 {
		insights = append(insights, fmt.Sprintf("Tienes gastos distribuidos en %d categorías diferentes", len(topCategories)))
	} else if len(topCategories) == 1 {
		insights = append(insights, "Tus gastos se concentran en una sola categoría - considera diversificar")
	}
	
	// Insight sobre el total de ingresos vs gastos
	if totalIncome > 0 {
		expenseRatio := (totalExpenses / totalIncome) * 100
		if expenseRatio > 80 {
			insights = append(insights, fmt.Sprintf("Estás gastando %.1f%% de tus ingresos - considera reducir gastos", expenseRatio))
		} else if expenseRatio < 50 {
			insights = append(insights, fmt.Sprintf("Solo gastas %.1f%% de tus ingresos - ¡excelente control financiero!", expenseRatio))
		}
	}
	
	return insights
}
