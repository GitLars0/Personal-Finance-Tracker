package controllers

import (
	"net/http"

	"Personal-Finance-Tracker-backend/db"
	"Personal-Finance-Tracker-backend/models"
	"Personal-Finance-Tracker-backend/middleware"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
)

// GetAccounts retrieves all accounts for the authenticated user
func GetAccounts(c *gin.Context) {
	// Extract JWT claims from context
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := uint(claims.(jwt.MapClaims)["sub"].(float64))

	var accounts []models.Account
	result := db.DB.Where("user_id = ?", userID).Find(&accounts)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch accounts"})
		return
	}

	c.JSON(http.StatusOK, accounts)
}

// GetAccount retrieves a specific account by ID for the authenticated user
func GetAccount(c *gin.Context) {
	// Extract JWT claims from context
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := uint(claims.(jwt.MapClaims)["sub"].(float64))
	accountID := c.Param("id")

	var account models.Account
	result := db.DB.Where("id = ? AND user_id = ?", accountID, userID).First(&account)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	c.JSON(http.StatusOK, account)
}

// CreateAccount creates a new account for the authenticated user
func CreateAccount(c *gin.Context) {
	// Extract JWT claims from context
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := uint(claims.(jwt.MapClaims)["sub"].(float64))

	var request struct {
		Name                string             `json:"name" binding:"required"`
		AccountType         models.AccountType `json:"account_type" binding:"required"`
		InitialBalanceCents int64              `json:"initial_balance_cents"`
		Description         string             `json:"description"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate account type
	validTypes := map[models.AccountType]bool{
		models.AccountCash:       true,
		models.AccountChecking:   true,
		models.AccountSavings:    true,
		models.AccountCredit:     true,
		models.AccountInvestment: true,
		models.AccountOther:      true,
	}

	if !validTypes[request.AccountType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account type"})
		return
	}

	account := models.Account{
		UserID:              userID,
		Name:                request.Name,
		Type:                request.AccountType,
		InitialBalanceCents: request.InitialBalanceCents,
		CurrentBalanceCents: request.InitialBalanceCents, // Start with initial balance
		Description:         request.Description,
	}

	result := db.DB.Create(&account)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create account"})
		return
	}

	middleware.IncrementAccountsCreated()

	c.JSON(http.StatusCreated, account)
}

// UpdateAccount updates an existing account for the authenticated user
func UpdateAccount(c *gin.Context) {
	// Extract JWT claims from context
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := uint(claims.(jwt.MapClaims)["sub"].(float64))
	accountID := c.Param("id")

	var account models.Account
	result := db.DB.Where("id = ? AND user_id = ?", accountID, userID).First(&account)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	var request struct {
		Name                string             `json:"name"`
		AccountType         models.AccountType `json:"account_type"`
		InitialBalanceCents *int64             `json:"initial_balance_cents"`
		Description         string             `json:"description"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate account type if provided
	if request.AccountType != "" {
		validTypes := map[models.AccountType]bool{
			models.AccountCash:       true,
			models.AccountChecking:   true,
			models.AccountSavings:    true,
			models.AccountCredit:     true,
			models.AccountInvestment: true,
			models.AccountOther:      true,
		}

		if !validTypes[request.AccountType] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account type"})
			return
		}
		account.Type = request.AccountType
	}

	if request.Name != "" {
		account.Name = request.Name
	}

	account.Description = request.Description

	// Update initial balance if provided
	if request.InitialBalanceCents != nil {
		account.InitialBalanceCents = *request.InitialBalanceCents
	}

	result = db.DB.Save(&account)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update account"})
		return
	}

	// Recalculate current balance after initial balance change
	if request.InitialBalanceCents != nil {
		if err := UpdateAccountBalance(account.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update account balance"})
			return
		}
		// Reload the account to get the updated current balance
		db.DB.First(&account, account.ID)
	}

	c.JSON(http.StatusOK, account)
}

// DeleteAccount deletes an account for the authenticated user
func DeleteAccount(c *gin.Context) {
	// Extract JWT claims from context
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := uint(claims.(jwt.MapClaims)["sub"].(float64))
	accountID := c.Param("id")

	var account models.Account
	result := db.DB.Where("id = ? AND user_id = ?", accountID, userID).First(&account)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	// Check if account has transactions
	var transactionCount int64
	db.DB.Model(&models.Transaction{}).Where("account_id = ?", accountID).Count(&transactionCount)

	if transactionCount > 0 {
		// Delete all transactions for this account first
		if err := db.DB.Where("account_id = ?", accountID).Delete(&models.Transaction{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete account transactions"})
			return
		}
	}

	// Also unlink any bank accounts that reference this internal account
	db.DB.Model(&models.BankAccount{}).Where("internal_account_id = ?", accountID).Update("internal_account_id", nil)

	// Now delete the account
	result = db.DB.Delete(&account)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":              "Account deleted successfully",
		"transactions_deleted": transactionCount,
	})
}

// UpdateAccountBalance recalculates the current balance for an account based on transactions
func UpdateAccountBalance(accountID uint) error {
	var account models.Account
	if err := db.DB.First(&account, accountID).Error; err != nil {
		return err
	}

	// Calculate total transaction amount for this account
	var totalTransactions int64
	db.DB.Model(&models.Transaction{}).
		Where("account_id = ?", accountID).
		Select("COALESCE(SUM(amount_cents), 0)").
		Scan(&totalTransactions)

	// Current balance = initial balance + transactions
	account.CurrentBalanceCents = account.InitialBalanceCents + totalTransactions

	return db.DB.Save(&account).Error
}
