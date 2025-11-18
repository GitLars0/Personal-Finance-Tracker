package controllers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"Personal-Finance-Tracker-backend/controllers"
	"Personal-Finance-Tracker-backend/db"
	"Personal-Finance-Tracker-backend/middleware"
	"Personal-Finance-Tracker-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// AdminControllerTestSuite defines the test suite for admin controller tests
type AdminControllerTestSuite struct {
	suite.Suite
	database   *gorm.DB
	adminUser  *models.User
	normalUser *models.User
	adminToken string
	userToken  string
	router     *gin.Engine
}

// SetupSuite is called once before all tests in the suite
func (suite *AdminControllerTestSuite) SetupSuite() {
	// Setup database
	suite.database = SetupTestDB()
	db.DB = suite.database

	// Create admin user
	adminHash, _ := controllers.HashPassword("admin123")
	suite.adminUser = &models.User{
		Username:     "admin",
		Email:        "admin@example.com",
		PasswordHash: adminHash,
		Role:         models.UserRoleAdmin,
		Name:         "Admin User",
	}
	suite.database.Create(suite.adminUser)

	// Create normal user
	userHash, _ := controllers.HashPassword("user123")
	suite.normalUser = &models.User{
		Username:     "normaluser",
		Email:        "user@example.com",
		PasswordHash: userHash,
		Role:         models.UserRoleUser,
		Name:         "Normal User",
	}
	suite.database.Create(suite.normalUser)

	// Generate tokens
	suite.adminToken, _ = controllers.GenerateToken(suite.adminUser.ID, suite.adminUser.Username, string(suite.adminUser.Role))
	suite.userToken, _ = controllers.GenerateToken(suite.normalUser.ID, suite.normalUser.Username, string(suite.normalUser.Role))

	// Setup router
	suite.router = SetupRouter()
	suite.setupAdminRoutes()
}

// setupAdminRoutes sets up admin routes for testing
func (suite *AdminControllerTestSuite) setupAdminRoutes() {
	api := suite.router.Group("/api")
	api.Use(controllers.AuthMiddleware())

	admin := api.Group("/admin")
	admin.Use(middleware.RequireAdmin())
	{
		// Dashboard
		admin.GET("/dashboard-stats", controllers.GetDashboardStats)

		// Users
		admin.GET("/users", controllers.GetAllUsers)
		admin.GET("/users/:id", controllers.GetUserDetails)
		admin.DELETE("/users/:id", controllers.DeleteUserAdmin)
		admin.PUT("/users/:id/role", controllers.UpdateUserRole)

		// Data overview
		admin.GET("/transactions", controllers.GetAllTransactions)
		admin.GET("/accounts", controllers.GetAllAccounts)
		admin.GET("/categories", controllers.GetAllCategories)
		admin.GET("/budgets", controllers.GetAllBudgets)
		admin.GET("/budgets/:id", controllers.GetBudgetDetails)

		// Data deletion
		admin.DELETE("/transactions/:id", controllers.DeleteTransactionAdmin)
		admin.DELETE("/accounts/:id", controllers.DeleteAccountAdmin)
		admin.DELETE("/categories/:id", controllers.DeleteCategoryAdmin)
		admin.DELETE("/budgets/:id", controllers.DeleteBudgetAdmin)
	}
}

// SetupTest is called before each test
func (suite *AdminControllerTestSuite) SetupTest() {
	// Clean up data before each test (except users)
	suite.database.Where("1=1").Delete(&models.TransactionSplit{})
	suite.database.Where("1=1").Delete(&models.Transaction{})
	suite.database.Where("1=1").Delete(&models.BudgetItem{})
	suite.database.Where("1=1").Delete(&models.Budget{})
	suite.database.Where("1=1").Delete(&models.Account{})
	suite.database.Where("1=1").Delete(&models.Category{})
}

// ============================================
// TEST 1: Dashboard Stats
// ============================================
func (suite *AdminControllerTestSuite) TestGetDashboardStats() {
	// Create test data
	suite.database.Create(&models.Account{UserID: suite.normalUser.ID, Name: "Test Account", Type: "checking"})
	suite.database.Create(&models.Category{UserID: suite.normalUser.ID, Name: "Test Category", Kind: models.CategoryExpense})

	req, _ := http.NewRequest("GET", "/api/admin/dashboard-stats", nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	suite.Equal(float64(2), response["totalUsers"]) // admin + normal user
	suite.Equal(float64(1), response["totalAccounts"])
	suite.Equal(float64(1), response["totalCategories"])
	suite.Equal(float64(0), response["totalTransactions"])
	suite.Equal(float64(0), response["totalBudgets"])
}

func (suite *AdminControllerTestSuite) TestGetDashboardStats_Unauthorized() {
	req, _ := http.NewRequest("GET", "/api/admin/dashboard-stats", nil)
	req.Header.Set("Authorization", "Bearer "+suite.userToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusForbidden, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	suite.Contains(response["error"], "admin access required")
}

// ============================================
// TEST 2: User Management
// ============================================
func (suite *AdminControllerTestSuite) TestGetAllUsers() {
	req, _ := http.NewRequest("GET", "/api/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	users := response["users"].([]interface{})
	suite.Equal(2, len(users)) // admin + normal user

	// Check that password hashes are not included
	user := users[0].(map[string]interface{})
	suite.NotContains(user, "password_hash")
}

func (suite *AdminControllerTestSuite) TestGetUserDetails() {
	// Create additional test data for the user
	account := models.Account{
		UserID: suite.normalUser.ID,
		Name:   "User Account",
		Type:   "checking",
	}
	suite.database.Create(&account)

	category := models.Category{
		UserID: suite.normalUser.ID,
		Name:   "User Category",
		Kind:   models.CategoryExpense,
	}
	suite.database.Create(&category)

	transaction := models.Transaction{
		UserID:      suite.normalUser.ID,
		AccountID:   account.ID,
		CategoryID:  &category.ID,
		AmountCents: -1000,
		Description: "Test transaction",
		TxnDate:     time.Now(),
	}
	suite.database.Create(&transaction)

	req, _ := http.NewRequest("GET", "/api/admin/users/"+strconv.Itoa(int(suite.normalUser.ID)), nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	user := response["user"].(map[string]interface{})
	suite.Equal(suite.normalUser.Email, user["email"])

	stats := response["statistics"].(map[string]interface{})
	suite.Equal(float64(1), stats["accounts"])
	suite.Equal(float64(1), stats["categories"])
	suite.Equal(float64(1), stats["transactions"])
	suite.Equal(float64(0), stats["budgets"])
}

func (suite *AdminControllerTestSuite) TestGetUserDetails_NotFound() {
	req, _ := http.NewRequest("GET", "/api/admin/users/999", nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusNotFound, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	suite.Contains(response["error"], "user not found")
}

func (suite *AdminControllerTestSuite) TestUpdateUserRole() {
	updateData := map[string]interface{}{
		"role": "admin",
	}
	body, _ := json.Marshal(updateData)

	req, _ := http.NewRequest("PUT", "/api/admin/users/"+strconv.Itoa(int(suite.normalUser.ID))+"/role", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	// Verify role was updated
	var updatedUser models.User
	suite.database.First(&updatedUser, suite.normalUser.ID)
	suite.Equal(models.UserRoleAdmin, updatedUser.Role)
}

func (suite *AdminControllerTestSuite) TestUpdateUserRole_CannotDemoteSelf() {
	updateData := map[string]interface{}{
		"role": "user",
	}
	body, _ := json.Marshal(updateData)

	req, _ := http.NewRequest("PUT", "/api/admin/users/"+strconv.Itoa(int(suite.adminUser.ID))+"/role", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	suite.Contains(response["error"], "cannot change your own admin role")
}

func (suite *AdminControllerTestSuite) TestDeleteUserAdmin() {
	// Create a user to delete
	userToDelete := models.User{
		Username:     "deleteme",
		Email:        "delete@example.com",
		PasswordHash: "hash",
		Role:         models.UserRoleUser,
	}
	suite.database.Create(&userToDelete)

	// Create associated data
	account := models.Account{
		UserID: userToDelete.ID,
		Name:   "Account to delete",
		Type:   "checking",
	}
	suite.database.Create(&account)

	req, _ := http.NewRequest("DELETE", "/api/admin/users/"+strconv.Itoa(int(userToDelete.ID)), nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	// Verify user and associated data are deleted
	var count int64
	suite.database.Model(&models.User{}).Where("id = ?", userToDelete.ID).Count(&count)
	suite.Equal(int64(0), count)

	suite.database.Model(&models.Account{}).Where("user_id = ?", userToDelete.ID).Count(&count)
	suite.Equal(int64(0), count)
}

func (suite *AdminControllerTestSuite) TestDeleteUserAdmin_CannotDeleteSelf() {
	req, _ := http.NewRequest("DELETE", "/api/admin/users/"+strconv.Itoa(int(suite.adminUser.ID)), nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	suite.Contains(response["error"], "cannot delete your own admin account")
}

// ============================================
// TEST 3: Data Overview
// ============================================
func (suite *AdminControllerTestSuite) TestGetAllTransactions() {
	// Create test data
	account := models.Account{
		UserID: suite.normalUser.ID,
		Name:   "Test Account",
		Type:   "checking",
	}
	suite.database.Create(&account)

	category := models.Category{
		UserID: suite.normalUser.ID,
		Name:   "Test Category",
		Kind:   models.CategoryExpense,
	}
	suite.database.Create(&category)

	transaction := models.Transaction{
		UserID:      suite.normalUser.ID,
		AccountID:   account.ID,
		CategoryID:  &category.ID,
		AmountCents: -5000,
		Description: "Test transaction",
		TxnDate:     time.Now(),
	}
	suite.database.Create(&transaction)

	req, _ := http.NewRequest("GET", "/api/admin/transactions", nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	transactions := response["transactions"].([]interface{})
	suite.Equal(1, len(transactions))

	txn := transactions[0].(map[string]interface{})
	suite.Equal("Test transaction", txn["description"])
	suite.Equal(float64(-50), txn["amount"]) // Converted from cents
	suite.Equal("expense", txn["type"])
	suite.Equal(suite.normalUser.Username, txn["user_username"])
}

func (suite *AdminControllerTestSuite) TestGetAllAccounts() {
	account := models.Account{
		UserID:              suite.normalUser.ID,
		Name:                "Test Account",
		Type:                "checking",
		Currency:            "USD",
		InitialBalanceCents: 10000,
		CurrentBalanceCents: 8000,
	}
	suite.database.Create(&account)

	req, _ := http.NewRequest("GET", "/api/admin/accounts", nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	accounts := response["accounts"].([]interface{})
	suite.Equal(1, len(accounts))

	acc := accounts[0].(map[string]interface{})
	suite.Equal("Test Account", acc["name"])
	suite.Equal(float64(80), acc["balance"]) // Converted from cents
	suite.Equal(suite.normalUser.Username, acc["user_username"])
}

func (suite *AdminControllerTestSuite) TestGetAllCategories() {
	category := models.Category{
		UserID: suite.normalUser.ID,
		Name:   "Test Category",
		Kind:   models.CategoryExpense,
	}
	suite.database.Create(&category)

	req, _ := http.NewRequest("GET", "/api/admin/categories", nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	categories := response["categories"].([]interface{})
	suite.Equal(1, len(categories))

	cat := categories[0].(map[string]interface{})
	suite.Equal("Test Category", cat["name"])
	suite.Equal("expense", cat["kind"])
	suite.Equal(suite.normalUser.Username, cat["user_username"])
}

func (suite *AdminControllerTestSuite) TestGetAllBudgets() {
	budget := models.Budget{
		UserID:      suite.normalUser.ID,
		PeriodStart: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
		Currency:    "USD",
	}
	suite.database.Create(&budget)

	req, _ := http.NewRequest("GET", "/api/admin/budgets", nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	budgets := response["budgets"].([]interface{})
	suite.Equal(1, len(budgets))

	bud := budgets[0].(map[string]interface{})
	suite.Equal("Budget 2025-01", bud["name"])
	suite.Equal(suite.normalUser.Username, bud["user_username"])
}

// ============================================
// TEST 4: Admin Deletion Operations
// ============================================
func (suite *AdminControllerTestSuite) TestDeleteTransactionAdmin() {
	account := models.Account{
		UserID: suite.normalUser.ID,
		Name:   "Test Account",
		Type:   "checking",
	}
	suite.database.Create(&account)

	transaction := models.Transaction{
		UserID:      suite.normalUser.ID,
		AccountID:   account.ID,
		AmountCents: -1000,
		Description: "To delete",
		TxnDate:     time.Now(),
	}
	suite.database.Create(&transaction)

	req, _ := http.NewRequest("DELETE", "/api/admin/transactions/"+strconv.Itoa(int(transaction.ID)), nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	// Verify transaction is deleted
	var count int64
	suite.database.Model(&models.Transaction{}).Where("id = ?", transaction.ID).Count(&count)
	suite.Equal(int64(0), count)
}

func (suite *AdminControllerTestSuite) TestDeleteAccountAdmin() {
	account := models.Account{
		UserID: suite.normalUser.ID,
		Name:   "To delete",
		Type:   "checking",
	}
	suite.database.Create(&account)

	req, _ := http.NewRequest("DELETE", "/api/admin/accounts/"+strconv.Itoa(int(account.ID)), nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	// Verify account is deleted
	var count int64
	suite.database.Model(&models.Account{}).Where("id = ?", account.ID).Count(&count)
	suite.Equal(int64(0), count)
}

func (suite *AdminControllerTestSuite) TestDeleteCategoryAdmin() {
	category := models.Category{
		UserID: suite.normalUser.ID,
		Name:   "To delete",
		Kind:   models.CategoryExpense,
	}
	suite.database.Create(&category)

	req, _ := http.NewRequest("DELETE", "/api/admin/categories/"+strconv.Itoa(int(category.ID)), nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	// Verify category is deleted
	var count int64
	suite.database.Model(&models.Category{}).Where("id = ?", category.ID).Count(&count)
	suite.Equal(int64(0), count)
}

func (suite *AdminControllerTestSuite) TestDeleteBudgetAdmin() {
	budget := models.Budget{
		UserID:      suite.normalUser.ID,
		PeriodStart: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
		Currency:    "USD",
	}
	suite.database.Create(&budget)

	req, _ := http.NewRequest("DELETE", "/api/admin/budgets/"+strconv.Itoa(int(budget.ID)), nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	// Verify budget is deleted
	var count int64
	suite.database.Model(&models.Budget{}).Where("id = ?", budget.ID).Count(&count)
	suite.Equal(int64(0), count)
}

// ============================================
// TEST 5: Budget Details
// ============================================
func (suite *AdminControllerTestSuite) TestGetBudgetDetails() {
	// Create category
	category := models.Category{
		UserID: suite.normalUser.ID,
		Name:   "Food",
		Kind:   models.CategoryExpense,
	}
	suite.database.Create(&category)

	// Create account
	account := models.Account{
		UserID: suite.normalUser.ID,
		Name:   "Checking",
		Type:   "checking",
	}
	suite.database.Create(&account)

	// Create budget
	budget := models.Budget{
		UserID:      suite.normalUser.ID,
		PeriodStart: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
		Currency:    "USD",
	}
	suite.database.Create(&budget)

	// Create budget item
	budgetItem := models.BudgetItem{
		BudgetID:     budget.ID,
		CategoryID:   category.ID,
		PlannedCents: 50000, // $500
	}
	suite.database.Create(&budgetItem)

	// Create transaction within budget period
	transaction := models.Transaction{
		UserID:      suite.normalUser.ID,
		AccountID:   account.ID,
		CategoryID:  &category.ID,
		AmountCents: -20000, // $200 spent
		TxnDate:     time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
	}
	suite.database.Create(&transaction)

	req, _ := http.NewRequest("GET", "/api/admin/budgets/"+strconv.Itoa(int(budget.ID)), nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	budget_info := response["budget"].(map[string]interface{})
	suite.Equal("Budget 2025-01", budget_info["name"])
	suite.Equal(suite.normalUser.Username, budget_info["user_username"])

	summary := response["summary"].(map[string]interface{})
	suite.Equal(float64(500), summary["total_planned"])   // $500
	suite.Equal(float64(200), summary["total_spent"])     // $200
	suite.Equal(float64(300), summary["total_remaining"]) // $300
	suite.Equal(float64(40), summary["overall_progress"]) // 40%

	categories := response["categories"].([]interface{})
	suite.Equal(1, len(categories))

	cat := categories[0].(map[string]interface{})
	suite.Equal("Food", cat["name"])
	suite.Equal(float64(500), cat["planned_amount"])
	suite.Equal(float64(200), cat["spent_amount"])
	suite.Equal(float64(40), cat["progress_percentage"])
}

// ============================================
// TEST 6: Authorization Tests
// ============================================
func (suite *AdminControllerTestSuite) TestNonAdminCannotAccessAdminEndpoints() {
	endpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/api/admin/dashboard-stats"},
		{"GET", "/api/admin/users"},
		{"GET", "/api/admin/transactions"},
		{"GET", "/api/admin/accounts"},
		{"GET", "/api/admin/categories"},
		{"GET", "/api/admin/budgets"},
	}

	for _, endpoint := range endpoints {
		req, _ := http.NewRequest(endpoint.method, endpoint.path, nil)
		req.Header.Set("Authorization", "Bearer "+suite.userToken)
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		suite.Equal(http.StatusForbidden, w.Code, "Endpoint %s %s should be forbidden for normal users", endpoint.method, endpoint.path)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		suite.Contains(response["error"], "admin access required")
	}
}

func (suite *AdminControllerTestSuite) TestUnauthenticatedCannotAccessAdminEndpoints() {
	req, _ := http.NewRequest("GET", "/api/admin/dashboard-stats", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusUnauthorized, w.Code)
}

// ============================================
// TEST 7: Error Handling
// ============================================
func (suite *AdminControllerTestSuite) TestGetUserDetails_InvalidID() {
	req, _ := http.NewRequest("GET", "/api/admin/users/invalid", nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	suite.Contains(response["error"], "invalid user ID")
}

func (suite *AdminControllerTestSuite) TestUpdateUserRole_InvalidRole() {
	updateData := map[string]interface{}{
		"role": "invalid_role",
	}
	body, _ := json.Marshal(updateData)

	req, _ := http.NewRequest("PUT", "/api/admin/users/"+strconv.Itoa(int(suite.normalUser.ID))+"/role", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	suite.Contains(response["error"], "invalid role")
}

// TestAdminControllerTestSuite runs the admin controller test suite
func TestAdminControllerTestSuite(t *testing.T) {
	suite.Run(t, new(AdminControllerTestSuite))
}
