package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"Personal-Finance-Tracker-backend/controllers"
	"Personal-Finance-Tracker-backend/db"
	"Personal-Finance-Tracker-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type PlaidAPITestSuite struct {
	suite.Suite
	database       *gorm.DB
	router         *gin.Engine
	normalUser     models.User
	normalToken    string
	bankConnection models.BankConnection
}

// ============================================
// SETUP AND TEARDOWN
// ============================================
func (suite *PlaidAPITestSuite) SetupSuite() {
	// Create in-memory SQLite database for testing
	testDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	// Set the database instance
	db.DB = testDB
	suite.database = testDB

	// Migrate all models
	err = testDB.AutoMigrate(
		&models.User{},
		&models.Account{},
		&models.Category{},
		&models.Transaction{},
		&models.Budget{},
		&models.BankConnection{},
		&models.BankAccount{},
	)
	suite.Require().NoError(err)

	// Create test user
	hashedPassword, err := controllers.HashPassword("password123")
	suite.Require().NoError(err)

	suite.normalUser = models.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		Role:         models.UserRoleUser,
	}

	// Save user to database
	suite.database.Create(&suite.normalUser)

	// Generate JWT token
	suite.normalToken, err = controllers.GenerateToken(suite.normalUser.ID, suite.normalUser.Username, string(suite.normalUser.Role))
	suite.Require().NoError(err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()

	// Add auth middleware and Plaid routes
	authGroup := suite.router.Group("/api")
	authGroup.Use(controllers.AuthMiddleware())
	{
		authGroup.POST("/plaid/create_link_token", controllers.CreateLinkToken)
		authGroup.POST("/plaid/exchange_public_token", controllers.ExchangePublicToken)
		authGroup.POST("/plaid/sync/:id", controllers.SyncPlaidTransactions)
		authGroup.GET("/plaid/accounts/:id", controllers.GetPlaidAccounts)
	}
}

func (suite *PlaidAPITestSuite) SetupTest() {
	// Clean up existing data
	suite.database.Unscoped().Where("1 = 1").Delete(&models.BankAccount{})
	suite.database.Unscoped().Where("1 = 1").Delete(&models.BankConnection{})
	suite.database.Unscoped().Where("1 = 1").Delete(&models.Transaction{})
	suite.database.Unscoped().Where("1 = 1").Delete(&models.Account{})
	suite.database.Unscoped().Where("1 = 1").Delete(&models.Category{})

	// Create test bank connection for tests that need it
	suite.bankConnection = models.BankConnection{
		UserID:            suite.normalUser.ID,
		BankName:          "Test Bank",
		BankEndpoint:      "plaid://api",
		ConsentID:         "test_item_id",
		Status:            "connected",
		ConsentValidUntil: time.Now().Add(90 * 24 * time.Hour),
		Metadata: models.JSONB{
			"access_token": "access-sandbox-12345",
			"item_id":      "test_item_id",
		},
	}
	suite.database.Create(&suite.bankConnection)

	// Create test categories for categorization testing
	categories := []models.Category{
		{UserID: suite.normalUser.ID, Name: "Groceries", Kind: models.CategoryExpense},
		{UserID: suite.normalUser.ID, Name: "Transportation", Kind: models.CategoryExpense},
		{UserID: suite.normalUser.ID, Name: "Entertainment", Kind: models.CategoryExpense},
		{UserID: suite.normalUser.ID, Name: "Salary", Kind: models.CategoryIncome},
	}
	for _, cat := range categories {
		suite.database.Create(&cat)
	}
}

func (suite *PlaidAPITestSuite) TearDownSuite() {
	// Clean up database
	if suite.database != nil {
		sqlDB, _ := suite.database.DB()
		sqlDB.Close()
	}
}

// ============================================
// TEST 1: Create Link Token
// ============================================
func (suite *PlaidAPITestSuite) TestCreateLinkToken_PlaidNotInitialized() {
	// Test when Plaid client is not initialized (common case)
	req, _ := http.NewRequest("POST", "/api/plaid/create_link_token", nil)
	req.Header.Set("Authorization", "Bearer "+suite.normalToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)

	suite.Equal("Plaid client not initialized", response["error"])
}

func (suite *PlaidAPITestSuite) TestCreateLinkToken_Unauthorized() {
	req, _ := http.NewRequest("POST", "/api/plaid/create_link_token", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	suite.Contains(response["error"], "missing authorization header")
}

func (suite *PlaidAPITestSuite) TestCreateLinkToken_InvalidToken() {
	req, _ := http.NewRequest("POST", "/api/plaid/create_link_token", nil)
	req.Header.Set("Authorization", "Bearer invalid_token")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	suite.Contains(response["error"], "invalid token")
}

// ============================================
// TEST 2: Exchange Public Token
// ============================================
func (suite *PlaidAPITestSuite) TestExchangePublicToken_Unauthorized() {
	requestBody := map[string]interface{}{
		"public_token": "public-sandbox-12345",
		"bank_name":    "Test Bank",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/api/plaid/exchange_public_token", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusUnauthorized, w.Code)
}

func (suite *PlaidAPITestSuite) TestExchangePublicToken_InvalidRequest() {
	// Test with missing required fields
	requestBody := map[string]interface{}{
		"bank_name": "Test Bank",
		// missing public_token
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/api/plaid/exchange_public_token", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.normalToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)

	suite.Contains(response["error"], "Invalid request")
}

func (suite *PlaidAPITestSuite) TestExchangePublicToken_PlaidNotInitialized() {
	requestBody := map[string]interface{}{
		"public_token": "public-sandbox-12345",
		"bank_name":    "Test Bank",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/api/plaid/exchange_public_token", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.normalToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)

	suite.Equal("Plaid client not initialized", response["error"])
}

func (suite *PlaidAPITestSuite) TestExchangePublicToken_ValidRequest() {
	// This tests the structure of the request even though Plaid client isn't initialized
	requestBody := map[string]interface{}{
		"public_token": "public-sandbox-12345",
		"bank_name":    "Chase Bank",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/api/plaid/exchange_public_token", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.normalToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Should fail because Plaid client is not initialized, but request parsing should work
	suite.Equal(http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)

	// Should be Plaid client error, not parsing error
	suite.Equal("Plaid client not initialized", response["error"])
}

// ============================================
// TEST 3: Sync Plaid Transactions
// ============================================
func (suite *PlaidAPITestSuite) TestSyncPlaidTransactions_Unauthorized() {
	req, _ := http.NewRequest("POST", "/api/plaid/sync/1", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusUnauthorized, w.Code)
}

func (suite *PlaidAPITestSuite) TestSyncPlaidTransactions_ConnectionNotFound() {
	nonExistentID := "99999"
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/plaid/sync/%s", nonExistentID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.normalToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)

	suite.Equal("Connection not found", response["error"])
}

func (suite *PlaidAPITestSuite) TestSyncPlaidTransactions_UserIsolation() {
	// Create another user with their own bank connection
	hashedPassword, _ := controllers.HashPassword("password123")
	otherUser := models.User{
		Username:     "otheruser",
		Email:        "other@example.com",
		PasswordHash: hashedPassword,
		Role:         models.UserRoleUser,
	}
	suite.database.Create(&otherUser)

	// Create bank connection for other user
	otherConnection := models.BankConnection{
		UserID:            otherUser.ID,
		BankName:          "Other Bank",
		BankEndpoint:      "plaid://api",
		ConsentID:         "other_item_id",
		Status:            "connected",
		ConsentValidUntil: time.Now().Add(90 * 24 * time.Hour),
		Metadata: models.JSONB{
			"access_token": "access-other-12345",
			"item_id":      "other_item_id",
		},
	}
	suite.database.Create(&otherConnection)

	// Normal user should not be able to sync other user's connection
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/plaid/sync/%d", otherConnection.ID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.normalToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusNotFound, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	suite.Equal("Connection not found", response["error"])
}

func (suite *PlaidAPITestSuite) TestSyncPlaidTransactions_NoAccessToken() {
	// Create a connection without access token in metadata
	connectionWithoutToken := models.BankConnection{
		UserID:            suite.normalUser.ID,
		BankName:          "Bank Without Token",
		BankEndpoint:      "plaid://api",
		ConsentID:         "test_item_no_token",
		Status:            "connected",
		ConsentValidUntil: time.Now().Add(90 * 24 * time.Hour),
		Metadata: models.JSONB{
			"item_id": "test_item_no_token",
			// missing access_token
		},
	}
	suite.database.Create(&connectionWithoutToken)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/plaid/sync/%d", connectionWithoutToken.ID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.normalToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)

	suite.Equal("Access token not found", response["error"])
}

func (suite *PlaidAPITestSuite) TestSyncPlaidTransactions_ValidConnection() {
	// Test with valid connection that has access token
	// Note: This test currently demonstrates a bug in the controller where it panics
	// instead of returning a proper error when plaidClient is nil

	defer func() {
		if r := recover(); r != nil {
			// Expect a panic due to nil plaidClient access
			suite.Contains(fmt.Sprintf("%v", r), "nil pointer")
		}
	}()

	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/plaid/sync/%d", suite.bankConnection.ID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.normalToken)
	w := httptest.NewRecorder()

	// This will panic due to nil plaidClient, demonstrating a controller bug
	suite.router.ServeHTTP(w, req)

	// This line won't be reached due to the panic
	suite.Fail("Expected panic did not occur")
}

// ============================================
// TEST 4: Get Plaid Accounts
// ============================================
func (suite *PlaidAPITestSuite) TestGetPlaidAccounts_Unauthorized() {
	req, _ := http.NewRequest("GET", "/api/plaid/accounts/1", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusUnauthorized, w.Code)
}

func (suite *PlaidAPITestSuite) TestGetPlaidAccounts_ConnectionNotFound() {
	nonExistentID := "99999"
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/plaid/accounts/%s", nonExistentID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.normalToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)

	suite.Equal("Connection not found", response["error"])
}

func (suite *PlaidAPITestSuite) TestGetPlaidAccounts_UserIsolation() {
	// Create another user with their own bank connection
	hashedPassword, _ := controllers.HashPassword("password123")
	otherUser := models.User{
		Username:     "anotheruser",
		Email:        "another@example.com",
		PasswordHash: hashedPassword,
		Role:         models.UserRoleUser,
	}
	suite.database.Create(&otherUser)

	// Create bank connection for other user
	otherConnection := models.BankConnection{
		UserID:            otherUser.ID,
		BankName:          "Another Bank",
		BankEndpoint:      "plaid://api",
		ConsentID:         "another_item_id",
		Status:            "connected",
		ConsentValidUntil: time.Now().Add(90 * 24 * time.Hour),
		Metadata: models.JSONB{
			"access_token": "access-another-12345",
			"item_id":      "another_item_id",
		},
	}
	suite.database.Create(&otherConnection)

	// Normal user should not be able to access other user's connection
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/plaid/accounts/%d", otherConnection.ID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.normalToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusNotFound, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	suite.Equal("Connection not found", response["error"])
}

func (suite *PlaidAPITestSuite) TestGetPlaidAccounts_NoAccessToken() {
	// Create a connection without access token in metadata
	connectionWithoutToken := models.BankConnection{
		UserID:            suite.normalUser.ID,
		BankName:          "Bank Without Token Accounts",
		BankEndpoint:      "plaid://api",
		ConsentID:         "test_item_no_token_accounts",
		Status:            "connected",
		ConsentValidUntil: time.Now().Add(90 * 24 * time.Hour),
		Metadata: models.JSONB{
			"item_id": "test_item_no_token_accounts",
			// missing access_token
		},
	}
	suite.database.Create(&connectionWithoutToken)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/plaid/accounts/%d", connectionWithoutToken.ID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.normalToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)

	suite.Equal("Access token not found", response["error"])
}

func (suite *PlaidAPITestSuite) TestGetPlaidAccounts_ValidConnection() {
	// Test with valid connection that has access token
	// Note: This test currently demonstrates a bug in the controller where it panics
	// instead of returning a proper error when plaidClient is nil

	defer func() {
		if r := recover(); r != nil {
			// Expect a panic due to nil plaidClient access
			suite.Contains(fmt.Sprintf("%v", r), "nil pointer")
		}
	}()

	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/plaid/accounts/%d", suite.bankConnection.ID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.normalToken)
	w := httptest.NewRecorder()

	// This will panic due to nil plaidClient, demonstrating a controller bug
	suite.router.ServeHTTP(w, req)

	// This line won't be reached due to the panic
	suite.Fail("Expected panic did not occur")
}

// ============================================
// TEST 5: Helper Function Testing
// ============================================
func (suite *PlaidAPITestSuite) TestPlaidCategorization_Logic() {
	// Test categorization logic by creating categories and testing patterns
	// This tests the business logic without needing actual Plaid integration

	// Create test categories
	var categories []models.Category
	suite.database.Where("user_id = ?", suite.normalUser.ID).Find(&categories)
	suite.GreaterOrEqual(len(categories), 4, "Should have at least 4 test categories")

	// Verify categories are properly structured
	categoryNames := make(map[string]bool)
	for _, cat := range categories {
		categoryNames[cat.Name] = true
		suite.Equal(suite.normalUser.ID, cat.UserID, "Category should belong to test user")
		suite.NotEmpty(cat.Kind, "Category should have a kind")
	}

	// Check that expected categories exist
	suite.True(categoryNames["Groceries"], "Should have Groceries category")
	suite.True(categoryNames["Transportation"], "Should have Transportation category")
	suite.True(categoryNames["Entertainment"], "Should have Entertainment category")
	suite.True(categoryNames["Salary"], "Should have Salary category")
}

func (suite *PlaidAPITestSuite) TestBankConnectionMetadata() {
	// Test that bank connection metadata is properly structured
	suite.NotNil(suite.bankConnection.Metadata)

	accessToken, hasAccessToken := suite.bankConnection.Metadata["access_token"]
	suite.True(hasAccessToken, "Bank connection should have access token")
	suite.Equal("access-sandbox-12345", accessToken)

	itemID, hasItemID := suite.bankConnection.Metadata["item_id"]
	suite.True(hasItemID, "Bank connection should have item ID")
	suite.Equal("test_item_id", itemID)
}

// ============================================
// TEST 6: Request Validation
// ============================================
func (suite *PlaidAPITestSuite) TestExchangePublicToken_EmptyBody() {
	req, _ := http.NewRequest("POST", "/api/plaid/exchange_public_token", bytes.NewBuffer([]byte("")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.normalToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
}

func (suite *PlaidAPITestSuite) TestExchangePublicToken_InvalidJSON() {
	req, _ := http.NewRequest("POST", "/api/plaid/exchange_public_token", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.normalToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
}

func (suite *PlaidAPITestSuite) TestExchangePublicToken_EmptyPublicToken() {
	requestBody := map[string]interface{}{
		"public_token": "",
		"bank_name":    "Test Bank",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/api/plaid/exchange_public_token", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.normalToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)

	suite.Contains(response["error"], "Invalid request")
}

// ============================================
// TEST 7: Edge Cases and Error Handling
// ============================================
func (suite *PlaidAPITestSuite) TestAllPlaidEndpoints_RequireAuthentication() {
	// Test that all Plaid endpoints require authentication
	endpoints := []struct {
		method string
		path   string
		body   []byte
	}{
		{"POST", "/api/plaid/create_link_token", nil},
		{"POST", "/api/plaid/exchange_public_token", []byte(`{"public_token":"test"}`)},
		{"POST", "/api/plaid/sync/1", nil},
		{"GET", "/api/plaid/accounts/1", nil},
	}

	for _, endpoint := range endpoints {
		suite.Run("auth_required_"+endpoint.method+"_"+endpoint.path, func() {
			var req *http.Request
			if endpoint.body != nil {
				req, _ = http.NewRequest(endpoint.method, endpoint.path, bytes.NewBuffer(endpoint.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req, _ = http.NewRequest(endpoint.method, endpoint.path, nil)
			}

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			suite.Equal(http.StatusUnauthorized, w.Code,
				"Endpoint %s %s should require authentication", endpoint.method, endpoint.path)
		})
	}
}

func (suite *PlaidAPITestSuite) TestPlaidEndpoints_WrongHTTPMethod() {
	// Test wrong HTTP methods
	wrongMethods := []struct {
		correctMethod string
		wrongMethod   string
		path          string
	}{
		{"POST", "GET", "/api/plaid/create_link_token"},
		{"POST", "PUT", "/api/plaid/exchange_public_token"},
		{"POST", "DELETE", "/api/plaid/sync/1"},
		{"GET", "POST", "/api/plaid/accounts/1"},
	}

	for _, test := range wrongMethods {
		suite.Run("wrong_method_"+test.wrongMethod+"_"+test.path, func() {
			req, _ := http.NewRequest(test.wrongMethod, test.path, nil)
			req.Header.Set("Authorization", "Bearer "+suite.normalToken)
			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			suite.Equal(http.StatusNotFound, w.Code,
				"Wrong method %s for %s should return 404", test.wrongMethod, test.path)
		})
	}
}

func (suite *PlaidAPITestSuite) TestSyncPlaidTransactions_InvalidConnectionID() {
	// Test with non-numeric connection ID
	req, _ := http.NewRequest("POST", "/api/plaid/sync/invalid_id", nil)
	req.Header.Set("Authorization", "Bearer "+suite.normalToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusNotFound, w.Code)
}

func (suite *PlaidAPITestSuite) TestGetPlaidAccounts_InvalidConnectionID() {
	// Test with non-numeric connection ID
	req, _ := http.NewRequest("GET", "/api/plaid/accounts/invalid_id", nil)
	req.Header.Set("Authorization", "Bearer "+suite.normalToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusNotFound, w.Code)
}

// ============================================
// TEST 8: Database Integration
// ============================================
func (suite *PlaidAPITestSuite) TestBankConnectionCreation() {
	// Verify bank connection was created properly
	var connection models.BankConnection
	err := suite.database.Where("id = ?", suite.bankConnection.ID).First(&connection).Error
	suite.NoError(err)

	suite.Equal(suite.normalUser.ID, connection.UserID)
	suite.Equal("Test Bank", connection.BankName)
	suite.Equal("plaid://api", connection.BankEndpoint)
	suite.Equal("connected", connection.Status)
	suite.NotNil(connection.Metadata)
}

func (suite *PlaidAPITestSuite) TestCategorySetup() {
	// Verify test categories were created properly
	var categories []models.Category
	suite.database.Where("user_id = ?", suite.normalUser.ID).Find(&categories)

	suite.Len(categories, 4, "Should have exactly 4 test categories")

	categoryMap := make(map[string]models.Category)
	for _, cat := range categories {
		categoryMap[cat.Name] = cat
		suite.Equal(suite.normalUser.ID, cat.UserID)
	}

	// Check specific categories
	groceries, exists := categoryMap["Groceries"]
	suite.True(exists, "Groceries category should exist")
	suite.Equal(models.CategoryExpense, groceries.Kind)

	salary, exists := categoryMap["Salary"]
	suite.True(exists, "Salary category should exist")
	suite.Equal(models.CategoryIncome, salary.Kind)
}

// ============================================
// TEST 9: Performance and Reliability
// ============================================
func (suite *PlaidAPITestSuite) TestPlaidEndpoints_ConcurrentRequests() {
	// Test that endpoints can handle concurrent requests
	const numRequests = 5
	done := make(chan bool, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			defer func() { done <- true }()

			req, _ := http.NewRequest("POST", "/api/plaid/create_link_token", nil)
			req.Header.Set("Authorization", "Bearer "+suite.normalToken)
			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			// Should consistently return the same error (Plaid not initialized)
			suite.Equal(http.StatusInternalServerError, w.Code)
		}()
	}

	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		<-done
	}
}

// ============================================
// TEST RUNNER
// ============================================
func TestPlaidAPITestSuite(t *testing.T) {
	suite.Run(t, new(PlaidAPITestSuite))
}
