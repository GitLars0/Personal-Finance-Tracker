package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
)

// BudgetPrediction represents an AI-generated budget prediction
type BudgetPrediction struct {
	CategoryID             uint    `json:"category_id"`
	CategoryName           string  `json:"category_name"`
	PredictedAmountCents   int64   `json:"predicted_amount_cents"`
	PredictedAmountDollars float64 `json:"predicted_amount_dollars"`
	ConfidenceScore        float64 `json:"confidence_score"`
	HistoricalAvgCents     int64   `json:"historical_avg_cents"`
	HistoricalAvgDollars   float64 `json:"historical_avg_dollars"`
	TrendDirection         string  `json:"trend_direction"`
	Reasoning              string  `json:"reasoning"`
}

// PredictBudgetRequest represents the request payload for budget prediction
type PredictBudgetRequest struct {
	TargetMonth      int `json:"target_month"`
	TargetYear       int `json:"target_year"`
	HistoricalMonths int `json:"historical_months"`
}

// PredictBudgetResponse represents the AI service response
type PredictBudgetResponse struct {
	Predictions          []BudgetPrediction `json:"predictions"`
	TargetMonth          int                `json:"target_month"`
	TargetYear           int                `json:"target_year"`
	UserID               uint               `json:"user_id"`
	HistoricalDataPoints int                `json:"historical_data_points"`
	Message              string             `json:"message"`
}

// GetBudgetPrediction generates AI-powered budget predictions for the user
func GetBudgetPrediction(c *gin.Context) {
	// Step 1: Authenticate
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := uint(claims.(jwt.MapClaims)["sub"].(float64))

	// Step 2: Parse query parameters with defaults
	targetMonth := time.Now().Month()
	targetYear := time.Now().Year()
	historicalMonths := 12

	if monthStr := c.Query("target_month"); monthStr != "" {
		if month, err := strconv.Atoi(monthStr); err == nil && month >= 1 && month <= 12 {
			targetMonth = time.Month(month)
		}
	}

	if yearStr := c.Query("target_year"); yearStr != "" {
		if year, err := strconv.Atoi(yearStr); err == nil && year >= 2020 && year <= 2030 {
			targetYear = year
		}
	}

	if monthsStr := c.Query("historical_months"); monthsStr != "" {
		if months, err := strconv.Atoi(monthsStr); err == nil && months >= 1 && months <= 36 {
			historicalMonths = months
		}
	}

	// Step 3: Prepare request to AI service
	aiRequest := map[string]interface{}{
		"user_id":           userID,
		"target_month":      int(targetMonth),
		"target_year":       targetYear,
		"historical_months": historicalMonths,
	}

	jsonData, err := json.Marshal(aiRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to prepare AI request"})
		return
	}

	// Step 4: Call AI service
	aiServiceURL := getAIServiceURL() + "/predict-budget"

	resp, err := http.Post(aiServiceURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "AI service unavailable",
			"details": err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	// Step 5: Parse AI service response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read AI response"})
		return
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		if err := json.Unmarshal(body, &errorResp); err == nil {
			c.JSON(resp.StatusCode, errorResp)
		} else {
			c.JSON(resp.StatusCode, gin.H{"error": "AI service error"})
		}
		return
	}

	var aiResponse PredictBudgetResponse
	if err := json.Unmarshal(body, &aiResponse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse AI response"})
		return
	}

	// Step 6: Return predictions
	c.JSON(http.StatusOK, gin.H{
		"predictions":            aiResponse.Predictions,
		"target_month":           aiResponse.TargetMonth,
		"target_year":            aiResponse.TargetYear,
		"user_id":                aiResponse.UserID,
		"historical_data_points": aiResponse.HistoricalDataPoints,
		"message":                aiResponse.Message,
		"generated_at":           time.Now().UTC(),
	})
}

// GetSpendingPatterns analyzes user spending patterns without generating predictions
func GetSpendingPatterns(c *gin.Context) {
	// Step 1: Authenticate
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := uint(claims.(jwt.MapClaims)["sub"].(float64))

	// Step 2: Parse query parameters
	historicalMonths := 12
	if monthsStr := c.Query("historical_months"); monthsStr != "" {
		if months, err := strconv.Atoi(monthsStr); err == nil && months >= 1 && months <= 36 {
			historicalMonths = months
		}
	}

	// Step 3: Prepare request to AI service
	aiRequest := map[string]interface{}{
		"user_id":           userID,
		"historical_months": historicalMonths,
	}

	jsonData, err := json.Marshal(aiRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to prepare AI request"})
		return
	}

	// Step 4: Call AI service
	aiServiceURL := getAIServiceURL() + "/analyze-patterns"

	resp, err := http.Post(aiServiceURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "AI service unavailable",
			"details": err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	// Step 5: Parse and return response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read AI response"})
		return
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		if err := json.Unmarshal(body, &errorResp); err == nil {
			c.JSON(resp.StatusCode, errorResp)
		} else {
			c.JSON(resp.StatusCode, gin.H{"error": "AI service error"})
		}
		return
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse AI response"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// getAIServiceURL returns the AI service URL from environment variables
func getAIServiceURL() string {
	aiServiceHost := os.Getenv("AI_SERVICE_HOST")
	if aiServiceHost == "" {
		aiServiceHost = "ai-service" // Default Docker service name
	}

	aiServicePort := os.Getenv("AI_SERVICE_PORT")
	if aiServicePort == "" {
		aiServicePort = "5001" // Default port
	}

	return fmt.Sprintf("http://%s:%s", aiServiceHost, aiServicePort)
}
