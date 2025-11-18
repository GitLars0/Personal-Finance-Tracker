package controllers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"Personal-Finance-Tracker-backend/controllers"
	"Personal-Finance-Tracker-backend/db"
	"Personal-Finance-Tracker-backend/models"

	"github.com/stretchr/testify/assert"
)

// ============================================
// TEST 1: Create Budget
// ============================================
func TestCreateBudget(t *testing.T) {
	database := SetupTestDB()
	db.DB = database
	user := CreateTestUser(database)
	token := GetTestToken(user.ID, user.Username)

	// Create categories
	groceries := models.Category{
		UserID: user.ID,
		Name:   "Groceries",
		Kind:   models.CategoryExpense,
	}
	transport := models.Category{
		UserID: user.ID,
		Name:   "Transportation",
		Kind:   models.CategoryExpense,
	}
	database.Create(&groceries)
	database.Create(&transport)

	router := SetupRouter()
	router.POST("/api/budgets", controllers.AuthMiddleware(), controllers.CreateBudget)

	// Create budget for current month
	now := time.Now()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, -1)

	budgetData := map[string]interface{}{
		"period_start": periodStart.Format("2006-01-02"),
		"period_end":   periodEnd.Format("2006-01-02"),
		"currency":     "USD",
		"items": []map[string]interface{}{
			{
				"category_id":   groceries.ID,
				"planned_cents": 40000,
			},
			{
				"category_id":   transport.ID,
				"planned_cents": 60000,
			},
		},
	}
	body, _ := json.Marshal(budgetData)

	req, _ := http.NewRequest("POST", "/api/budgets", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code, "Should create budget successfully")

	var response models.Budget
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, user.ID, response.UserID)
	assert.Equal(t, "USD", response.Currency)
	assert.Equal(t, 2, len(response.Items))

	// Verify total planned amount
	var totalPlanned int64
	for _, item := range response.Items {
		totalPlanned += item.PlannedCents
	}
	assert.Equal(t, int64(100000), totalPlanned, "Total planned should equal 100000")
}

// ============================================
// TEST 2: Get Budgets
// ============================================
func TestGetBudgets(t *testing.T) {
	database := SetupTestDB()
	db.DB = database
	user := CreateTestUser(database)
	token := GetTestToken(user.ID, user.Username)

	// Create budgets
	jan2025 := models.Budget{
		UserID:      user.ID,
		PeriodStart: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
		Currency:    "USD",
	}
	feb2025 := models.Budget{
		UserID:      user.ID,
		PeriodStart: time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2025, 2, 28, 0, 0, 0, 0, time.UTC),
		Currency:    "USD",
	}
	database.Create(&jan2025)
	database.Create(&feb2025)

	router := SetupRouter()
	router.GET("/api/budgets", controllers.AuthMiddleware(), controllers.GetBudgets)

	// Test without filters
	req, _ := http.NewRequest("GET", "/api/budgets", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var budgets []interface{}
	json.Unmarshal(w.Body.Bytes(), &budgets)
	assert.Equal(t, 2, len(budgets))

	// Test with currency filter
	req, _ = http.NewRequest("GET", "/api/budgets?currency=USD", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	json.Unmarshal(w.Body.Bytes(), &budgets)
	assert.Equal(t, 2, len(budgets))
}

// ============================================
// TEST 3: Get Budget
// ============================================
func TestGetBudget(t *testing.T) {
	database := SetupTestDB()
	db.DB = database
	user := CreateTestUser(database)
	token := GetTestToken(user.ID, user.Username)

	// Create category
	category := models.Category{
		UserID: user.ID,
		Name:   "Groceries",
		Kind:   models.CategoryExpense,
	}
	database.Create(&category)

	// Create account
	account := models.Account{
		UserID:              user.ID,
		Name:                "Checking",
		Type:                "checking",
		InitialBalanceCents: 100000,
		CurrentBalanceCents: 100000,
	}
	database.Create(&account)

	// Create budget for current month
	now := time.Now()
	budget := models.Budget{
		UserID:      user.ID,
		PeriodStart: time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, -1),
		Currency:    "USD",
	}
	database.Create(&budget)

	budgetItem := models.BudgetItem{
		BudgetID:     budget.ID,
		CategoryID:   category.ID,
		PlannedCents: 40000,
	}
	database.Create(&budgetItem)

	// Create transactions
	txn1 := &models.Transaction{
		UserID:      user.ID,
		AccountID:   account.ID,
		CategoryID:  &category.ID,
		AmountCents: -15000,
		Description: "Groceries 1",
		TxnDate:     now,
	}
	result1 := database.Create(txn1)
	if result1.Error != nil {
		t.Fatalf("Failed to create transaction 1: %v", result1.Error)
	}

	txn2 := &models.Transaction{
		UserID:      user.ID,
		AccountID:   account.ID,
		CategoryID:  &category.ID,
		AmountCents: -10000,
		Description: "Groceries 2",
		TxnDate:     now,
	}
	result2 := database.Create(txn2)
	if result2.Error != nil {
		t.Fatalf("Failed to create transaction 2: %v", result2.Error)
	}

	router := SetupRouter()
	router.GET("/api/budgets/:id", controllers.AuthMiddleware(), controllers.GetBudget)

	req, _ := http.NewRequest("GET", "/api/budgets/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// Verify spending calculations
	assert.Equal(t, float64(40000), response["total_planned_cents"],
		"Total planned should be $400")
	assert.Equal(t, float64(25000), response["total_spent_cents"],
		"Total spent should be $250")

	// Check remaining cents with proper field name
	if remaining, ok := response["total_remaining_cents"]; ok {
		assert.Equal(t, float64(15000), remaining, "Remaining cents should match")
	} else {
		t.Logf("Response: %+v", response)
		t.Error("total_remaining_cents field not found")
	}
}

// ============================================
// TEST 4: Update Budget
// ============================================
func TestUpdateBudget(t *testing.T) {
	database := SetupTestDB()
	db.DB = database
	user := CreateTestUser(database)
	token := GetTestToken(user.ID, user.Username)

	// Create category FIRST
	category := models.Category{
		UserID: user.ID,
		Name:   "Food",
		Kind:   models.CategoryExpense,
	}
	database.Create(&category)

	// Create budget
	budget := models.Budget{
		UserID:      user.ID,
		PeriodStart: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
		Currency:    "USD",
	}
	database.Create(&budget)

	budgetItem := models.BudgetItem{
		BudgetID:     budget.ID,
		CategoryID:   category.ID,
		PlannedCents: 30000,
	}
	database.Create(&budgetItem)

	router := SetupRouter()
	router.PUT("/api/budgets/:id", controllers.AuthMiddleware(), controllers.UpdateBudget)

	// Update budget
	updateData := map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"category_id":   category.ID,
				"planned_cents": 50000,
			},
		},
	}
	body, _ := json.Marshal(updateData)

	req, _ := http.NewRequest("PUT", "/api/budgets/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Expected 200 OK status")

	var response models.Budget
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Should parse response")

	// Safety check before accessing array
	if len(response.Items) == 0 {
		var errorResponse map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &errorResponse)
		t.Fatalf("No items in response. Error: %v", errorResponse)
	}

	assert.Equal(t, 1, len(response.Items), "Budget should have 1 item after update")
	assert.Equal(t, int64(50000), response.Items[0].PlannedCents,
		"Planned amount should be updated to $500")
}

// ============================================
// TEST 5: Delete Budget
// ============================================
func TestDeleteBudget(t *testing.T) {
	database := SetupTestDB()
	db.DB = database
	user := CreateTestUser(database)
	token := GetTestToken(user.ID, user.Username)

	category := models.Category{
		UserID: user.ID,
		Name:   "Food",
		Kind:   models.CategoryExpense,
	}
	database.Create(&category)

	budget := models.Budget{
		UserID:      user.ID,
		PeriodStart: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
		Currency:    "USD",
	}
	database.Create(&budget)

	budgetItem := models.BudgetItem{
		BudgetID:     budget.ID,
		CategoryID:   category.ID,
		PlannedCents: 30000,
	}
	database.Create(&budgetItem)

	router := SetupRouter()
	router.DELETE("/api/budgets/:id", controllers.AuthMiddleware(), controllers.DeleteBudget)

	req, _ := http.NewRequest("DELETE", "/api/budgets/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify budget is deleted
	var count int64
	database.Model(&models.Budget{}).Where("id = ?", budget.ID).Count(&count)
	assert.Equal(t, int64(0), count, "Budget should be deleted")

	// Verify budget items are cascade deleted
	database.Model(&models.BudgetItem{}).Where("budget_id = ?", budget.ID).Count(&count)
	assert.Equal(t, int64(0), count, "Budget items should be deleted")
}

// ============================================
// TEST 6: User Isolation
// ============================================
func TestBudget_UserIsolation(t *testing.T) {
	database := SetupTestDB()
	db.DB = database

	user1 := CreateTestUser(database)

	hash, _ := controllers.HashPassword("testpass123")
	user2 := models.User{
		Username:     "user2",
		Email:        "user2@example.com",
		PasswordHash: hash,
		Role:         models.UserRoleUser,
	}
	database.Create(&user2)

	category := models.Category{
		UserID: user1.ID,
		Name:   "Food",
		Kind:   models.CategoryExpense,
	}
	database.Create(&category)

	budget := models.Budget{
		UserID:      user1.ID,
		PeriodStart: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
		Currency:    "USD",
	}
	database.Create(&budget)

	// User2 tries to access User1's budget
	token2 := GetTestToken(user2.ID, user2.Username)

	router := SetupRouter()
	router.GET("/api/budgets/:id", controllers.AuthMiddleware(), controllers.GetBudget)

	req, _ := http.NewRequest("GET", "/api/budgets/1", nil)
	req.Header.Set("Authorization", "Bearer "+token2)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code,
		"User2 should not access User1's budget")
}

// ============================================
// TEST 7: Get Current Budget
// ============================================
func TestGetCurrentBudget(t *testing.T) {
	database := SetupTestDB()
	db.DB = database
	user := CreateTestUser(database)
	token := GetTestToken(user.ID, user.Username)

	category := models.Category{
		UserID: user.ID,
		Name:   "Food",
		Kind:   models.CategoryExpense,
	}
	database.Create(&category)

	// Create budget for current month
	now := time.Now()
	currentBudget := models.Budget{
		UserID:      user.ID,
		PeriodStart: time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, -1),
		Currency:    "USD",
	}
	database.Create(&currentBudget)

	// Create budget for last month
	lastMonth := models.Budget{
		UserID:      user.ID,
		PeriodStart: time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(now.Year(), now.Month(), 0, 0, 0, 0, 0, time.UTC),
		Currency:    "USD",
	}
	database.Create(&lastMonth)

	router := SetupRouter()
	router.GET("/api/budgets/current", controllers.AuthMiddleware(), controllers.GetCurrentBudget)

	req, _ := http.NewRequest("GET", "/api/budgets/current", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Budget
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, currentBudget.ID, response.ID,
		"Should return current month's budget")
}
