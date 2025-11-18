package controllers_test

import (
	"bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "Personal-Finance-Tracker-backend/controllers"
    "Personal-Finance-Tracker-backend/models"
	"Personal-Finance-Tracker-backend/db"

    "github.com/gin-gonic/gin"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"

	"github.com/stretchr/testify/assert"

)

// setupTestDB creates an in-memory SQLite database for testing
func SetupTestDB() *gorm.DB {
    database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Silent),
    })
    if err != nil {
        panic("Failed to connect to test database: " + err.Error())
    }

    // Migrate ALL tables
    err = database.AutoMigrate(
        &models.User{},
        &models.Account{},
        &models.Category{},
        &models.Transaction{},
        &models.TransactionSplit{},
        &models.Budget{},
        &models.BudgetItem{},
    )
    if err != nil {
        panic("Failed to migrate test database: " + err.Error())
    }

    return database
}

// setupRouter creates a Gin router for testing
func SetupRouter() *gin.Engine {
    gin.SetMode(gin.TestMode)
    return gin.New()
}

// createTestUser creates a test user in the database
func CreateTestUser(database *gorm.DB) *models.User {
    hash, _ := controllers.HashPassword("testpassword")
    user := models.User{
        Username:     "testuser",
        Email:        "testuser@example.com",
        PasswordHash: hash,
        Role:         models.UserRoleUser,
    }
    database.Create(&user)
    return &user
}

// getTestToken generates a JWT token for testing
func GetTestToken(userID uint, username string) string {
    token, _ := controllers.GenerateToken(userID, username, "user")
    return token
}

func TestCreateAccount(t *testing.T) {
	// Setup fresh database
	database := SetupTestDB()
	db.DB = database

	// Create test user (needed for authentication)
	user := CreateTestUser(database)

	// Generate JWT token for test user
	token := GetTestToken(user.ID, user.Username)

	// Setup router with endpoints we are testing
	router := SetupRouter()
	router.POST(
		"/api/accounts",
		controllers.AuthMiddleware(),
		controllers.CreateAccount,
	)

	// Prepare test data (what we will send in the request)
	accountData := map[string]interface{}{
		"name":                  "Test Checking",
		"account_type":          "checking",
		"initial_balance_cents": 1000,
		"currency":              "USD",
	}

	body, _ := json.Marshal(accountData)

	// Create HTTP request
	req, _ := http.NewRequest("POST", "/api/accounts", bytes.NewBuffer(body))

	// Set required headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Create response recorder (captuers the response)
	w := httptest.NewRecorder()

	// Execute request through router
	router.ServeHTTP(w, req)

	// Check HTTP status code
	assert.Equal(t, http.StatusCreated, w.Code, "Expected 201 Created status")

	// Parse response body
	var response models.Account
	json.Unmarshal(w.Body.Bytes(), &response)

	// Verify response data matches what we sent
	assert.Equal(t, "Test Checking", response.Name, "Account name should match")
	assert.Equal(t, "checking", string(response.Type), "Account type should match")
	assert.Equal(t, int64(1000), response.InitialBalanceCents, "Initial balance should match")
	assert.Equal(t, int64(1000), response.CurrentBalanceCents, "Current balance should equal initial balance")

	// Verify data was actually saved to database
	var savedAccount models.Account
	database.First(&savedAccount, response.ID)
	assert.Equal(t, "Test Checking", savedAccount.Name, "Account should be saved in database")
}

func TestGetAccounts(t *testing.T) {
	// Setup/Arrange
	database := SetupTestDB()
	db.DB = database
	user := CreateTestUser(database)
	token := GetTestToken(user.ID, user.Username)

	// Create test accounts
	database.Create(&models.Account{
		UserID:              user.ID,
		Name:                "Checking",
		Type:                "Checking",
		InitialBalanceCents: 5000,
		CurrentBalanceCents: 5000,
	})
	database.Create(&models.Account{
		UserID:              user.ID,
		Name:                "Savings",
		Type:                "Saving",
		InitialBalanceCents: 5000,
		CurrentBalanceCents: 5000,
	})

	router := SetupRouter()
	router.GET("/api/accounts", controllers.AuthMiddleware(), controllers.GetAccounts)

	// Make request
	req, _ := http.NewRequest("GET", "/api/accounts", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code, "Expected 200 OK status")

	var accounts []models.Account
	json.Unmarshal(w.Body.Bytes(), &accounts)
	assert.Equal(t, 2, len(accounts), "Expected 2 accounts in response")
	assert.Equal(t, "Checking", accounts[0].Name, "First account name should match")
	assert.Equal(t, "Savings", accounts[1].Name, "Second account name should match")

	// Verify in database
	var dbAccounts []models.Account
	database.Where("user_id = ?", user.ID).Find(&dbAccounts)
	assert.Equal(t, 2, len(dbAccounts), "Expected 2 accounts in database")
}

func TestUpdateAccount(t *testing.T) {
	database := SetupTestDB()
	db.DB = database
	user := CreateTestUser(database)
	token := GetTestToken(user.ID, user.Username)

	account := models.Account{
		UserID:              user.ID,
		Name:                "Old Name",
		Type:                "checking",
		InitialBalanceCents: 5000,
		CurrentBalanceCents: 5000,
	}
	database.Create(&account)

	router := SetupRouter()
	router.PUT("/api/accounts/:id", controllers.AuthMiddleware(), controllers.UpdateAccount)

	// Update data
	updateData := map[string]interface{}{
		"name": "New Name",
	}
	body, _ := json.Marshal(updateData)

	// Make request
	req, _ := http.NewRequest("PUT", "/api/accounts/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code, "Expected 200 OK status")

	var response models.Account
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "New Name", response.Name, "Account name should be updated")

	// Verify in database
	var updatedAccount models.Account
	database.First(&updatedAccount, account.ID)
	assert.Equal(t, "New Name", updatedAccount.Name, "Account name in database should be updated")
}

func TestDeleteAccount(t *testing.T) {
	// Setup/Arrange
	database := SetupTestDB()
	db.DB = database
	user := CreateTestUser(database)
	token := GetTestToken(user.ID, user.Username)

	account := models.Account{
		UserID:              user.ID,
		Name:                "To Delete",
		Type:                "checking",
		InitialBalanceCents: 5000,
		CurrentBalanceCents: 5000,
	}
	database.Create(&account)

	router := SetupRouter()
	router.DELETE("/api/accounts/:id", controllers.AuthMiddleware(), controllers.DeleteAccount)

	// Make request
	req, _ := http.NewRequest("DELETE", "/api/accounts/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code, "Expected 200 OK status")

	// Verify account is deleted
	var count int64
	database.Model(&models.Account{}).Where("id = ?", account.ID).Count(&count)
	assert.Equal(t, int64(0), count, "Account should be deleted from database")
}

func TestUnauthorizedAccess(t *testing.T) {
	database := SetupTestDB()
	db.DB = database

	router := SetupRouter()
	router.GET("/api/accounts", controllers.AuthMiddleware(), controllers.GetAccounts)

	req, _ := http.NewRequest("GET", "/api/accounts", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code, "Should return 401 Unauthorized without token")

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["error"] != nil {
		assert.NotEmpty(t, response["error"], "unauthorized", "Error field should not be empty")
	} else {
		t.Error("Expecter error field in response")
	}
}
