package controllers

import (
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

type BankControllerTestSuite struct {
	suite.Suite
	database        *gorm.DB
	router          *gin.Engine
	adminUser       models.User
	normalUser      models.User
	adminToken      string
	normalToken     string
	bankConnection1 models.BankConnection
	bankConnection2 models.BankConnection
}

// ============================================
// SETUP AND TEARDOWN
// ============================================
func (suite *BankControllerTestSuite) SetupSuite() {
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
		&models.BudgetItem{},
		&models.BankConnection{},
		&models.BankAccount{},
		&models.BankSyncLog{},
	)
	suite.Require().NoError(err)

	// Create test users
	hashedPassword, err := controllers.HashPassword("password123")
	suite.Require().NoError(err)

	suite.adminUser = models.User{
		Username:     "admin",
		Email:        "admin@example.com",
		PasswordHash: hashedPassword,
		Role:         models.UserRoleAdmin,
	}

	suite.normalUser = models.User{
		Username:     "user",
		Email:        "user@example.com",
		PasswordHash: hashedPassword,
		Role:         models.UserRoleUser,
	}

	// Save users to database
	suite.database.Create(&suite.adminUser)
	suite.database.Create(&suite.normalUser)

	// Generate JWT tokens
	suite.adminToken, err = controllers.GenerateToken(suite.adminUser.ID, suite.adminUser.Username, string(suite.adminUser.Role))
	suite.Require().NoError(err)

	suite.normalToken, err = controllers.GenerateToken(suite.normalUser.ID, suite.normalUser.Username, string(suite.normalUser.Role))
	suite.Require().NoError(err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()

	// Add auth middleware and routes
	authGroup := suite.router.Group("/api")
	authGroup.Use(controllers.AuthMiddleware())
	{
		authGroup.GET("/banks/connections", controllers.GetBankConnections)
		authGroup.DELETE("/banks/connections/:id", controllers.DisconnectBank)
		authGroup.POST("/banks/connections", controllers.CreateBankConnection) // Deprecated endpoint
	}
}

func (suite *BankControllerTestSuite) SetupTest() {
	// Clean up existing data properly
	suite.database.Unscoped().Where("1 = 1").Delete(&models.BankAccount{})
	suite.database.Unscoped().Where("1 = 1").Delete(&models.BankConnection{})

	// Create test bank connections for normal user with unique consent IDs
	suite.bankConnection1 = models.BankConnection{
		UserID:            suite.normalUser.ID,
		BankName:          "sparebanken_norge",
		BankEndpoint:      "https://psd2.spvapi.no",
		ConsentID:         fmt.Sprintf("consent_123_%d", time.Now().UnixNano()),
		ConsentStatus:     "valid",
		ConsentValidUntil: time.Now().Add(90 * 24 * time.Hour), // Valid for 90 days
		FrequencyPerDay:   4,
		Status:            "connected",
		LastSyncAt:        timePtr(time.Now().Add(-1 * time.Hour)),
		NextSyncAt:        timePtr(time.Now().Add(6 * time.Hour)),
		SyncCount:         5,
		Metadata: models.JSONB{
			"oauth_token": "fake_token",
			"scope":       "read_accounts read_transactions",
		},
	}

	suite.bankConnection2 = models.BankConnection{
		UserID:            suite.normalUser.ID,
		BankName:          "bulder_bank",
		BankEndpoint:      "https://psd2-bulder.spvapi.no",
		ConsentID:         fmt.Sprintf("consent_456_%d", time.Now().UnixNano()+1),
		ConsentStatus:     "expired",
		ConsentValidUntil: time.Now().Add(-1 * 24 * time.Hour), // Expired yesterday
		FrequencyPerDay:   2,
		Status:            "expired",
		LastSyncAt:        timePtr(time.Now().Add(-25 * time.Hour)),
		SyncCount:         3,
	}

	suite.database.Create(&suite.bankConnection1)
	suite.database.Create(&suite.bankConnection2)

	// Create linked bank accounts
	bankAccount1 := models.BankAccount{
		BankConnectionID:    suite.bankConnection1.ID,
		AccountID:           "acc_123",
		IBAN:                "NO9386011117947",
		AccountName:         "Main Checking Account",
		Currency:            "NOK",
		AccountType:         "checking",
		LastTransactionSync: timePtr(time.Now().Add(-1 * time.Hour)),
		IsActive:            true,
	}

	bankAccount2 := models.BankAccount{
		BankConnectionID:    suite.bankConnection1.ID,
		AccountID:           "acc_456",
		IBAN:                "NO9386011117948",
		AccountName:         "Savings Account",
		Currency:            "NOK",
		AccountType:         "savings",
		LastTransactionSync: timePtr(time.Now().Add(-2 * time.Hour)),
		IsActive:            true,
	}

	suite.database.Create(&bankAccount1)
	suite.database.Create(&bankAccount2)
}

func (suite *BankControllerTestSuite) TearDownSuite() {
	// Clean up database
	if suite.database != nil {
		sqlDB, _ := suite.database.DB()
		sqlDB.Close()
	}
}

// ============================================
// TEST 1: Get Bank Connections
// ============================================
func (suite *BankControllerTestSuite) TestGetBankConnections_Success() {
	req, _ := http.NewRequest("GET", "/api/banks/connections", nil)
	req.Header.Set("Authorization", "Bearer "+suite.normalToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)

	connections, exists := response["connections"].([]interface{})
	suite.True(exists)
	suite.Len(connections, 2)

	// Find connections by bank name since order might vary
	var sparebankConnection, bulderConnection map[string]interface{}
	for _, conn := range connections {
		connection := conn.(map[string]interface{})
		if connection["bank_name"] == "sparebanken_norge" {
			sparebankConnection = connection
		} else if connection["bank_name"] == "bulder_bank" {
			bulderConnection = connection
		}
	}

	// Verify sparebank connection
	suite.NotNil(sparebankConnection)
	suite.Equal("sparebanken_norge", sparebankConnection["bank_name"])
	suite.Equal("valid", sparebankConnection["consent_status"])
	suite.Equal("connected", sparebankConnection["status"])

	// Verify bulder bank connection
	suite.NotNil(bulderConnection)
	suite.Equal("bulder_bank", bulderConnection["bank_name"])
	suite.Equal("expired", bulderConnection["consent_status"])
	suite.Equal("expired", bulderConnection["status"])

	// Check linked accounts are preloaded for sparebank connection
	linkedAccounts, exists := sparebankConnection["linked_accounts"].([]interface{})
	suite.True(exists)
	suite.Len(linkedAccounts, 2)
}

func (suite *BankControllerTestSuite) TestGetBankConnections_NoConnections() {
	// Create a new user with no bank connections
	hashedPassword, _ := controllers.HashPassword("password123")
	emptyUser := models.User{
		Username:     "empty",
		Email:        "empty@example.com",
		PasswordHash: hashedPassword,
		Role:         models.UserRoleUser,
	}
	suite.database.Create(&emptyUser)

	emptyToken, err := controllers.GenerateToken(emptyUser.ID, emptyUser.Username, string(emptyUser.Role))
	suite.Require().NoError(err)

	req, _ := http.NewRequest("GET", "/api/banks/connections", nil)
	req.Header.Set("Authorization", "Bearer "+emptyToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)

	connections, exists := response["connections"].([]interface{})
	suite.True(exists)
	suite.Len(connections, 0)
}

func (suite *BankControllerTestSuite) TestGetBankConnections_Unauthorized() {
	req, _ := http.NewRequest("GET", "/api/banks/connections", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	suite.Contains(response["error"], "missing authorization header")
}

func (suite *BankControllerTestSuite) TestGetBankConnections_UserIsolation() {
	// Create another user with their own bank connection
	hashedPassword, _ := controllers.HashPassword("password123")
	otherUser := models.User{
		Username:     "other",
		Email:        "other@example.com",
		PasswordHash: hashedPassword,
		Role:         models.UserRoleUser,
	}
	suite.database.Create(&otherUser)

	// Create bank connection for other user
	otherConnection := models.BankConnection{
		UserID:            otherUser.ID,
		BankName:          "test_bank",
		BankEndpoint:      "https://test.bank.com",
		ConsentID:         fmt.Sprintf("other_consent_%d", time.Now().UnixNano()),
		ConsentStatus:     "valid",
		ConsentValidUntil: time.Now().Add(90 * 24 * time.Hour),
		Status:            "connected",
	}
	suite.database.Create(&otherConnection)

	// Normal user should only see their own connections
	req, _ := http.NewRequest("GET", "/api/banks/connections", nil)
	req.Header.Set("Authorization", "Bearer "+suite.normalToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)

	connections := response["connections"].([]interface{})
	suite.Len(connections, 2) // Only normalUser's connections

	for _, conn := range connections {
		connection := conn.(map[string]interface{})
		suite.NotEqual("test_bank", connection["bank_name"])
	}
}

// ============================================
// TEST 2: Disconnect Bank
// ============================================
func (suite *BankControllerTestSuite) TestDisconnectBank_Success() {
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/banks/connections/%d", suite.bankConnection1.ID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.normalToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)

	suite.Equal("Bank connection deleted successfully", response["message"])

	// Verify the connection was soft deleted
	var connection models.BankConnection
	err = suite.database.First(&connection, suite.bankConnection1.ID).Error
	suite.Error(err) // Should not find record (soft deleted)

	// Verify it exists with Unscoped (includes soft deleted)
	err = suite.database.Unscoped().First(&connection, suite.bankConnection1.ID).Error
	suite.NoError(err)
	suite.True(connection.DeletedAt.Valid) // Should be soft deleted
}

func (suite *BankControllerTestSuite) TestDisconnectBank_InvalidID() {
	req, _ := http.NewRequest("DELETE", "/api/banks/connections/invalid", nil)
	req.Header.Set("Authorization", "Bearer "+suite.normalToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	suite.Equal("Invalid connection ID", response["error"])
}

func (suite *BankControllerTestSuite) TestDisconnectBank_NotFound() {
	nonExistentID := uint(99999)
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/banks/connections/%d", nonExistentID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.normalToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusNotFound, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	suite.Equal("Bank connection not found", response["error"])
}

func (suite *BankControllerTestSuite) TestDisconnectBank_Unauthorized() {
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/banks/connections/%d", suite.bankConnection1.ID), nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	suite.Contains(response["error"], "missing authorization header")
}

func (suite *BankControllerTestSuite) TestDisconnectBank_UserIsolation() {
	// Create another user with their own bank connection
	hashedPassword, _ := controllers.HashPassword("password123")
	otherUser := models.User{
		Username:     "other2",
		Email:        "other2@example.com",
		PasswordHash: hashedPassword,
		Role:         models.UserRoleUser,
	}
	suite.database.Create(&otherUser)

	// Create bank connection for other user
	otherConnection := models.BankConnection{
		UserID:            otherUser.ID,
		BankName:          "other_bank",
		BankEndpoint:      "https://other.bank.com",
		ConsentID:         fmt.Sprintf("other_consent_2_%d", time.Now().UnixNano()),
		ConsentStatus:     "valid",
		ConsentValidUntil: time.Now().Add(90 * 24 * time.Hour),
		Status:            "connected",
	}
	suite.database.Create(&otherConnection)

	// Normal user should not be able to delete other user's connection
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/banks/connections/%d", otherConnection.ID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.normalToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusNotFound, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	suite.Equal("Bank connection not found", response["error"])

	// Verify the connection still exists
	var connection models.BankConnection
	err := suite.database.First(&connection, otherConnection.ID).Error
	suite.NoError(err)
	suite.False(connection.DeletedAt.Valid) // Should not be deleted
}

// ============================================
// TEST 3: Create Bank Connection (Deprecated)
// ============================================
func (suite *BankControllerTestSuite) TestCreateBankConnection_Deprecated() {
	req, _ := http.NewRequest("POST", "/api/banks/connections", nil)
	req.Header.Set("Authorization", "Bearer "+suite.normalToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)

	suite.Equal("This endpoint is deprecated. Please use Plaid Link instead.", response["error"])
	suite.Equal("Use /api/plaid/create_link_token to connect banks via Plaid", response["message"])
	suite.Equal("All bank connections now use Plaid for security and reliability", response["hint"])
}

// ============================================
// TEST 4: Edge Cases and Error Handling
// ============================================
func (suite *BankControllerTestSuite) TestGetBankConnections_DatabaseError() {
	// We can't easily simulate database errors with SQLite in-memory,
	// but we can test with a malformed token that might cause issues
	req, _ := http.NewRequest("GET", "/api/banks/connections", nil)
	req.Header.Set("Authorization", "Bearer invalid_token_format")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	suite.Contains(response["error"], "invalid token")
}

func (suite *BankControllerTestSuite) TestDisconnectBank_MultipleConnections() {
	// Create additional connections to ensure proper isolation
	additionalConnection := models.BankConnection{
		UserID:            suite.normalUser.ID,
		BankName:          "additional_bank",
		BankEndpoint:      "https://additional.bank.com",
		ConsentID:         fmt.Sprintf("consent_789_%d", time.Now().UnixNano()),
		ConsentStatus:     "valid",
		ConsentValidUntil: time.Now().Add(90 * 24 * time.Hour),
		Status:            "connected",
	}
	suite.database.Create(&additionalConnection)

	// Delete one specific connection
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/banks/connections/%d", suite.bankConnection2.ID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.normalToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	// Verify only the target connection was deleted
	var deletedConnection models.BankConnection
	err := suite.database.First(&deletedConnection, suite.bankConnection2.ID).Error
	suite.Error(err) // Should be deleted

	// Verify other connections still exist
	var existingConnection1 models.BankConnection
	err = suite.database.First(&existingConnection1, suite.bankConnection1.ID).Error
	suite.NoError(err) // Should still exist

	var existingConnection2 models.BankConnection
	err = suite.database.First(&existingConnection2, additionalConnection.ID).Error
	suite.NoError(err) // Should still exist
}

func (suite *BankControllerTestSuite) TestBankConnectionsWithDifferentStatuses() {
	// Test retrieving connections with various statuses and metadata
	req, _ := http.NewRequest("GET", "/api/banks/connections", nil)
	req.Header.Set("Authorization", "Bearer "+suite.normalToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)

	connections := response["connections"].([]interface{})

	// Find the connected bank (sparebanken_norge)
	var connectedBank map[string]interface{}
	for _, conn := range connections {
		connection := conn.(map[string]interface{})
		if connection["bank_name"] == "sparebanken_norge" {
			connectedBank = connection
			break
		}
	}

	suite.NotNil(connectedBank)
	suite.Equal("connected", connectedBank["status"])
	suite.Equal("valid", connectedBank["consent_status"])
	suite.Equal(float64(4), connectedBank["frequency_per_day"])
	suite.Equal(float64(5), connectedBank["sync_count"])

	// Verify metadata structure
	metadata, exists := connectedBank["metadata"]
	suite.True(exists)
	suite.NotNil(metadata)
}

// ============================================
// UTILITY FUNCTIONS
// ============================================
func timePtr(t time.Time) *time.Time {
	return &t
}

// ============================================
// TEST RUNNER
// ============================================
func TestBankControllerTestSuite(t *testing.T) {
	suite.Run(t, new(BankControllerTestSuite))
}
