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

func TestCreateTransaction(t *testing.T) {
	database := SetupTestDB()
	db.DB = database
	user := CreateTestUser(database)
	token := GetTestToken(user.ID, user.Username)

	account := models.Account{
		UserID:              user.ID,
		Name:                "Test Account",
		Type:                "checking",
		InitialBalanceCents: 1000,
		CurrentBalanceCents: 1000,
	}
	database.Create(&account)

	category := models.Category{
		UserID: user.ID,
		Name:   "Groceries",
		Kind:   models.CategoryExpense,
	}
	database.Create(&category)

	router := SetupRouter()
	router.POST("/api/transactions", controllers.AuthMiddleware(), controllers.CreateTransaction)

	txnData := map[string]interface{}{
		"account_id":   account.ID,
		"category_id":  category.ID,
		"amount_cents": -200,
		"description":  "Whole Foods",
		"txn_date":     time.Now().Format("2006-01-02"),
	}
	body, _ := json.Marshal(txnData)

	req, _ := http.NewRequest("POST", "/api/transactions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code, "Expected 201 Created status")

	var response models.Transaction
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, int64(-200), response.AmountCents, "Transaction amount should match")
	assert.Equal(t, "Whole Foods", response.Description, "Transaction description should match")

	var updatedAccount models.Account
	database.First(&updatedAccount, account.ID)
	assert.Equal(t, int64(800), updatedAccount.CurrentBalanceCents, "Account balance should be updated")
}

func TestGetTransactions(t *testing.T) {
	database := SetupTestDB()
	db.DB = database
	user := CreateTestUser(database)
	token := GetTestToken(user.ID, user.Username)

	account := models.Account{
		UserID:              user.ID,
		Name:                "Test Account",
		Type:                "checking",
		InitialBalanceCents: 100000,
		CurrentBalanceCents: 100000,
	}
	database.Create(&account)

	database.Create(&models.Transaction{
		UserID:      user.ID,
		AccountID:   account.ID,
		AmountCents: -5000,
		Description: "Transaction 1",
		TxnDate:     time.Now(),
	})
	database.Create(&models.Transaction{
		UserID:      user.ID,
		AccountID:   account.ID,
		AmountCents: -3000,
		Description: "Transaction 2",
		TxnDate:     time.Now(),
	})

	router := SetupRouter()
	router.GET("/api/transactions", controllers.AuthMiddleware(), controllers.GetTransactions)

	req, _ := http.NewRequest("GET", "/api/transactions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Expected 200 OK status")

	var transactions []models.Transaction
	json.Unmarshal(w.Body.Bytes(), &transactions)
	assert.Len(t, transactions, 2, "Should return 2 transactions")
}

func TestUpdateTransaction(t *testing.T) {
	database := SetupTestDB()
	db.DB = database
	user := CreateTestUser(database)
	token := GetTestToken(user.ID, user.Username)

	account := models.Account{
		UserID:              user.ID,
		Name:                "Test Account",
		Type:                "checking",
		InitialBalanceCents: 100000,
		CurrentBalanceCents: 95000,
	}
	database.Create(&account)

	transaction := models.Transaction{
		UserID:      user.ID,
		AccountID:   account.ID,
		AmountCents: -5000,
		Description: "Old Description",
		TxnDate:     time.Now(),
	}
	database.Create(&transaction)

	router := SetupRouter()
	router.PUT("/api/transactions/:id", controllers.AuthMiddleware(), controllers.UpdateTransaction)

	updateData := map[string]interface{}{
		"description":  "New Description",
		"amount_cents": -6000,
	}
	body, _ := json.Marshal(updateData)

	req, _ := http.NewRequest("PUT", "/api/transactions/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Expected 200 OK status")

	var updatedTransaction models.Transaction
	database.First(&updatedTransaction, transaction.ID)
	assert.Equal(t, "New Description", updatedTransaction.Description, "Transaction description should be updated")
	assert.Equal(t, int64(-6000), updatedTransaction.AmountCents, "Transaction amount should be updated")

	var updatedAccount models.Account
	database.First(&updatedAccount, account.ID)
	assert.Equal(t, int64(94000), updatedAccount.CurrentBalanceCents, "Account balance should be updated accordingly")
}

func TestDeleteTransaction(t *testing.T) {
	database := SetupTestDB()
	db.DB = database
	user := CreateTestUser(database)
	token := GetTestToken(user.ID, user.Username)

	account := models.Account{
		UserID:              user.ID,
		Name:                "Test Account",
		Type:                "checking",
		InitialBalanceCents: 100000,
		CurrentBalanceCents: 95000,
	}
	database.Create(&account)

	transaction := models.Transaction{
		UserID:      user.ID,
		AccountID:   account.ID,
		AmountCents: -5000,
		Description: "To delete",
		TxnDate:     time.Now(),
	}
	database.Create(&transaction)

	router := SetupRouter()
	router.DELETE("/api/transactions/:id", controllers.AuthMiddleware(), controllers.DeleteTransaction)

	req, _ := http.NewRequest("DELETE", "/api/transactions/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Expected 200 OK status")

	var count int64
	database.Model(&models.Transaction{}).Where("id = ?", transaction.ID).Count(&count)
	assert.Equal(t, int64(0), count, "Transaction should be deleted from database")

	var updatedAccount models.Account
	database.First(&updatedAccount, account.ID)
	assert.Equal(t, int64(100000), updatedAccount.CurrentBalanceCents, "Account balance should be restored")
}
