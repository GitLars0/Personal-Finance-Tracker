package controllers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"Personal-Finance-Tracker-backend/controllers"
	"Personal-Finance-Tracker-backend/db"
	"Personal-Finance-Tracker-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// AIControllerTestSuite defines the test suite for AI controller tests
type AIControllerTestSuite struct {
	suite.Suite
	database     *gorm.DB
	user         *models.User
	userToken    string
	router       *gin.Engine
	mockAIServer *httptest.Server
	originalHost string
	originalPort string
}

// SetupSuite is called once before all tests in the suite
func (suite *AIControllerTestSuite) SetupSuite() {
	// Setup database
	suite.database = SetupTestDB()
	db.DB = suite.database

	// Create test user
	suite.user = CreateTestUser(suite.database)
	suite.userToken = GetTestToken(suite.user.ID, suite.user.Username)

	// Store original environment variables
	suite.originalHost = os.Getenv("AI_SERVICE_HOST")
	suite.originalPort = os.Getenv("AI_SERVICE_PORT")

	// Setup mock AI server
	suite.setupMockAIServer()

	// Setup router
	suite.router = SetupRouter()
	suite.setupAIRoutes()
}

// TearDownSuite is called once after all tests in the suite
func (suite *AIControllerTestSuite) TearDownSuite() {
	// Restore original environment variables
	if suite.originalHost != "" {
		os.Setenv("AI_SERVICE_HOST", suite.originalHost)
	} else {
		os.Unsetenv("AI_SERVICE_HOST")
	}

	if suite.originalPort != "" {
		os.Setenv("AI_SERVICE_PORT", suite.originalPort)
	} else {
		os.Unsetenv("AI_SERVICE_PORT")
	}

	// Close mock server
	if suite.mockAIServer != nil {
		suite.mockAIServer.Close()
	}
}

// setupMockAIServer creates a mock HTTP server for AI service responses
func (suite *AIControllerTestSuite) setupMockAIServer() {
	suite.mockAIServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/predict-budget":
			suite.handleBudgetPrediction(w, r)
		case "/analyze-patterns":
			suite.handleSpendingPatterns(w, r)
		default:
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "endpoint not found",
			})
		}
	}))

	// Extract host and port from mock server URL
	serverURL := strings.TrimPrefix(suite.mockAIServer.URL, "http://")
	parts := strings.Split(serverURL, ":")

	os.Setenv("AI_SERVICE_HOST", parts[0])
	os.Setenv("AI_SERVICE_PORT", parts[1])
}

// handleBudgetPrediction simulates AI service budget prediction response
func (suite *AIControllerTestSuite) handleBudgetPrediction(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var request map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "invalid request body",
		})
		return
	}

	// Check for required fields
	userID, hasUserID := request["user_id"]
	targetMonth, hasMonth := request["target_month"]
	targetYear, hasYear := request["target_year"]

	if !hasUserID || !hasMonth || !hasYear {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "missing required fields",
		})
		return
	}

	// Simulate different responses based on user_id for testing
	switch int(userID.(float64)) {
	case 999: // Test user that causes AI service error
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "AI service internal error",
		})
		return
	case 998: // Test user with no historical data
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"predictions":            []interface{}{},
			"target_month":           targetMonth,
			"target_year":            targetYear,
			"user_id":                userID,
			"historical_data_points": 0,
			"message":                "Insufficient historical data for predictions",
		})
		return
	default:
		// Normal successful response
		predictions := []map[string]interface{}{
			{
				"category_id":              1,
				"category_name":            "Groceries",
				"predicted_amount_cents":   45000,
				"predicted_amount_dollars": 450.0,
				"confidence_score":         0.85,
				"historical_avg_cents":     42000,
				"historical_avg_dollars":   420.0,
				"trend_direction":          "increasing",
				"reasoning":                "Based on historical spending patterns and seasonal trends",
			},
			{
				"category_id":              2,
				"category_name":            "Transportation",
				"predicted_amount_cents":   25000,
				"predicted_amount_dollars": 250.0,
				"confidence_score":         0.75,
				"historical_avg_cents":     28000,
				"historical_avg_dollars":   280.0,
				"trend_direction":          "decreasing",
				"reasoning":                "Recent reduction in commuting expenses",
			},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"predictions":            predictions,
			"target_month":           targetMonth,
			"target_year":            targetYear,
			"user_id":                userID,
			"historical_data_points": 12,
			"message":                "Predictions generated successfully",
		})
	}
}

// handleSpendingPatterns simulates AI service spending patterns response
func (suite *AIControllerTestSuite) handleSpendingPatterns(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var request map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "invalid request body",
		})
		return
	}

	userID, hasUserID := request["user_id"]
	if !hasUserID {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "missing user_id",
		})
		return
	}

	// Simulate different responses based on user_id
	switch int(userID.(float64)) {
	case 999: // Test user that causes AI service error
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "AI service temporarily unavailable",
		})
		return
	default:
		// Normal successful response
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"user_id": userID,
			"patterns": map[string]interface{}{
				"spending_velocity":    "moderate",
				"category_consistency": 0.78,
				"seasonal_trends": map[string]interface{}{
					"highest_month": "December",
					"lowest_month":  "February",
				},
				"weekend_vs_weekday": map[string]interface{}{
					"weekend_ratio": 0.35,
					"weekday_ratio": 0.65,
				},
			},
			"insights": []string{
				"Your grocery spending is highly consistent month-to-month",
				"Entertainment expenses spike significantly on weekends",
				"Transportation costs are lower in winter months",
			},
			"recommendations": []string{
				"Consider budgeting 15% more for December expenses",
				"Set weekend spending alerts for entertainment category",
			},
			"analyzed_period":  "12 months",
			"confidence_score": 0.82,
		})
	}
}

// setupAIRoutes sets up AI routes for testing
func (suite *AIControllerTestSuite) setupAIRoutes() {
	api := suite.router.Group("/api")
	api.Use(controllers.AuthMiddleware())
	{
		api.GET("/ai/budget-predictions", controllers.GetBudgetPrediction)
		api.GET("/ai/spending-patterns", controllers.GetSpendingPatterns)
	}
}

// SetupTest is called before each test
func (suite *AIControllerTestSuite) SetupTest() {
	// Clean up data before each test if needed
	// For AI controller tests, we mainly test the HTTP proxy functionality
}

// ============================================
// TEST 1: Budget Prediction Success Cases
// ============================================
func (suite *AIControllerTestSuite) TestGetBudgetPrediction_Success() {
	req, _ := http.NewRequest("GET", "/api/ai/budget-predictions", nil)
	req.Header.Set("Authorization", "Bearer "+suite.userToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// Verify response structure
	suite.Contains(response, "predictions")
	suite.Contains(response, "target_month")
	suite.Contains(response, "target_year")
	suite.Contains(response, "user_id")
	suite.Contains(response, "historical_data_points")
	suite.Contains(response, "message")
	suite.Contains(response, "generated_at")

	// Verify predictions content
	predictions := response["predictions"].([]interface{})
	suite.Equal(2, len(predictions))

	// Check first prediction
	pred1 := predictions[0].(map[string]interface{})
	suite.Equal("Groceries", pred1["category_name"])
	suite.Equal(float64(450), pred1["predicted_amount_dollars"])
	suite.Equal(float64(0.85), pred1["confidence_score"])
	suite.Equal("increasing", pred1["trend_direction"])

	// Verify user ID matches
	suite.Equal(float64(suite.user.ID), response["user_id"])

	// Verify current month/year defaults
	now := time.Now()
	suite.Equal(float64(now.Month()), response["target_month"])
	suite.Equal(float64(now.Year()), response["target_year"])
}

func (suite *AIControllerTestSuite) TestGetBudgetPrediction_WithQueryParameters() {
	req, _ := http.NewRequest("GET", "/api/ai/budget-predictions?target_month=6&target_year=2025&historical_months=18", nil)
	req.Header.Set("Authorization", "Bearer "+suite.userToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// Verify query parameters were processed
	suite.Equal(float64(6), response["target_month"])
	suite.Equal(float64(2025), response["target_year"])
	// Historical months parameter is passed to AI service but not returned directly
}

func (suite *AIControllerTestSuite) TestGetBudgetPrediction_InvalidQueryParameters() {
	// Test with invalid parameters - should use defaults
	req, _ := http.NewRequest("GET", "/api/ai/budget-predictions?target_month=15&target_year=1900&historical_months=100", nil)
	req.Header.Set("Authorization", "Bearer "+suite.userToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// Should use current month/year as defaults for invalid values
	now := time.Now()
	suite.Equal(float64(now.Month()), response["target_month"])
	suite.Equal(float64(now.Year()), response["target_year"])
}

// ============================================
// TEST 2: Budget Prediction Error Cases
// ============================================
func (suite *AIControllerTestSuite) TestGetBudgetPrediction_Unauthorized() {
	req, _ := http.NewRequest("GET", "/api/ai/budget-predictions", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	suite.Contains(response["error"], "missing authorization header")
}

func (suite *AIControllerTestSuite) TestGetBudgetPrediction_NoHistoricalData() {
	// Create a special user with ID 998 that triggers "no historical data" response
	userWithNoData := models.User{
		ID:           998,
		Username:     "nodata",
		Email:        "nodata@example.com",
		PasswordHash: "hash",
		Role:         models.UserRoleUser,
	}
	suite.database.Create(&userWithNoData)

	token := GetTestToken(userWithNoData.ID, userWithNoData.Username)

	req, _ := http.NewRequest("GET", "/api/ai/budget-predictions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	predictions := response["predictions"].([]interface{})
	suite.Equal(0, len(predictions))
	suite.Equal(float64(0), response["historical_data_points"])
	suite.Contains(response["message"], "Insufficient historical data")

	// Cleanup
	suite.database.Delete(&userWithNoData)
}

// ============================================
// TEST 3: Spending Patterns Success Cases
// ============================================
func (suite *AIControllerTestSuite) TestGetSpendingPatterns_Success() {
	req, _ := http.NewRequest("GET", "/api/ai/spending-patterns", nil)
	req.Header.Set("Authorization", "Bearer "+suite.userToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// Verify response structure
	suite.Contains(response, "user_id")
	suite.Contains(response, "patterns")
	suite.Contains(response, "insights")
	suite.Contains(response, "recommendations")
	suite.Contains(response, "analyzed_period")
	suite.Contains(response, "confidence_score")

	// Check patterns structure
	patterns := response["patterns"].(map[string]interface{})
	suite.Contains(patterns, "spending_velocity")
	suite.Contains(patterns, "category_consistency")
	suite.Contains(patterns, "seasonal_trends")
	suite.Contains(patterns, "weekend_vs_weekday")

	// Check insights and recommendations are arrays
	insights := response["insights"].([]interface{})
	suite.Greater(len(insights), 0)

	recommendations := response["recommendations"].([]interface{})
	suite.Greater(len(recommendations), 0)

	// Verify user ID
	suite.Equal(float64(suite.user.ID), response["user_id"])
}

func (suite *AIControllerTestSuite) TestGetSpendingPatterns_WithHistoricalMonths() {
	req, _ := http.NewRequest("GET", "/api/ai/spending-patterns?historical_months=6", nil)
	req.Header.Set("Authorization", "Bearer "+suite.userToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// Should still get successful response
	suite.Contains(response, "patterns")
	suite.Equal(float64(suite.user.ID), response["user_id"])
}

// ============================================
// TEST 4: Spending Patterns Error Cases
// ============================================
func (suite *AIControllerTestSuite) TestGetSpendingPatterns_Unauthorized() {
	req, _ := http.NewRequest("GET", "/api/ai/spending-patterns", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	suite.Contains(response["error"], "missing authorization header")
}

// ============================================
// TEST 5: AI Service Error Handling
// ============================================
func (suite *AIControllerTestSuite) TestAIService_InternalError() {
	// Create a special user with ID 999 that triggers AI service errors
	userWithError := models.User{
		ID:           999,
		Username:     "erroruser",
		Email:        "error@example.com",
		PasswordHash: "hash",
		Role:         models.UserRoleUser,
	}
	suite.database.Create(&userWithError)

	token := GetTestToken(userWithError.ID, userWithError.Username)

	// Test budget prediction error
	req, _ := http.NewRequest("GET", "/api/ai/budget-predictions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	suite.Contains(response["error"], "AI service internal error")

	// Test spending patterns error
	req, _ = http.NewRequest("GET", "/api/ai/spending-patterns", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusServiceUnavailable, w.Code)

	json.Unmarshal(w.Body.Bytes(), &response)
	suite.Contains(response["error"], "AI service temporarily unavailable")

	// Cleanup
	suite.database.Delete(&userWithError)
}

func (suite *AIControllerTestSuite) TestAIService_Unavailable() {
	// Temporarily close the mock server to simulate service unavailable
	suite.mockAIServer.Close()

	req, _ := http.NewRequest("GET", "/api/ai/budget-predictions", nil)
	req.Header.Set("Authorization", "Bearer "+suite.userToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusServiceUnavailable, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	suite.Contains(response["error"], "AI service unavailable")

	// Restart the mock server for subsequent tests
	suite.setupMockAIServer()
}

// ============================================
// TEST 6: Query Parameter Validation
// ============================================
func (suite *AIControllerTestSuite) TestBudgetPrediction_QueryParameterValidation() {
	testCases := []struct {
		name          string
		queryParams   string
		expectedMonth int
		expectedYear  int
		description   string
	}{
		{
			name:          "Valid parameters",
			queryParams:   "target_month=3&target_year=2024&historical_months=6",
			expectedMonth: 3,
			expectedYear:  2024,
			description:   "Should accept valid parameters",
		},
		{
			name:          "Month out of range high",
			queryParams:   "target_month=13",
			expectedMonth: int(time.Now().Month()),
			expectedYear:  time.Now().Year(),
			description:   "Should use defaults for month > 12",
		},
		{
			name:          "Month out of range low",
			queryParams:   "target_month=0",
			expectedMonth: int(time.Now().Month()),
			expectedYear:  time.Now().Year(),
			description:   "Should use defaults for month < 1",
		},
		{
			name:          "Year out of range low",
			queryParams:   "target_year=2010",
			expectedMonth: int(time.Now().Month()),
			expectedYear:  time.Now().Year(),
			description:   "Should use defaults for year < 2020",
		},
		{
			name:          "Year out of range high",
			queryParams:   "target_year=2040",
			expectedMonth: int(time.Now().Month()),
			expectedYear:  time.Now().Year(),
			description:   "Should use defaults for year > 2030",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			url := "/api/ai/budget-predictions"
			if tc.queryParams != "" {
				url += "?" + tc.queryParams
			}

			req, _ := http.NewRequest("GET", url, nil)
			req.Header.Set("Authorization", "Bearer "+suite.userToken)
			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			suite.Equal(http.StatusOK, w.Code, tc.description)

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)

			suite.Equal(float64(tc.expectedMonth), response["target_month"], tc.description+" - month")
			suite.Equal(float64(tc.expectedYear), response["target_year"], tc.description+" - year")
		})
	}
}

func (suite *AIControllerTestSuite) TestSpendingPatterns_HistoricalMonthsValidation() {
	testCases := []struct {
		name        string
		queryParam  string
		description string
		shouldWork  bool
	}{
		{
			name:        "Valid historical months",
			queryParam:  "historical_months=12",
			description: "Should accept valid historical months",
			shouldWork:  true,
		},
		{
			name:        "Historical months too low",
			queryParam:  "historical_months=0",
			description: "Should use default for historical_months < 1",
			shouldWork:  true,
		},
		{
			name:        "Historical months too high",
			queryParam:  "historical_months=50",
			description: "Should use default for historical_months > 36",
			shouldWork:  true,
		},
		{
			name:        "Invalid historical months",
			queryParam:  "historical_months=invalid",
			description: "Should use default for non-numeric historical_months",
			shouldWork:  true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			url := "/api/ai/spending-patterns?" + tc.queryParam

			req, _ := http.NewRequest("GET", url, nil)
			req.Header.Set("Authorization", "Bearer "+suite.userToken)
			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			if tc.shouldWork {
				suite.Equal(http.StatusOK, w.Code, tc.description)
			} else {
				suite.NotEqual(http.StatusOK, w.Code, tc.description)
			}
		})
	}
}

// TestAIControllerTestSuite runs the AI controller test suite
func TestAIControllerTestSuite(t *testing.T) {
	suite.Run(t, new(AIControllerTestSuite))
}
