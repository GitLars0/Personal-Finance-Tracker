package controllers

import (
	"net/http"
	"strconv"
	"time"

	"Personal-Finance-Tracker-backend/db"
	"Personal-Finance-Tracker-backend/middleware"
	"Personal-Finance-Tracker-backend/models"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
)

// GetTransactions retrieves all transactions for the authenticated user
func GetTransactions(c *gin.Context) {
	// Step 1: Authenticate
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := uint(claims.(jwt.MapClaims)["sub"].(float64))

	var transactions []models.Transaction

	// Step 2: Always filter by user_id (SECURITY CRITICAL!)
	query := db.DB.Where("user_id = ?", userID)

	// Step 3: Optional filters
	if accountID := c.Query("account_id"); accountID != "" {
		query = query.Where("account_id = ?", accountID)
	}

	if categoryID := c.Query("category_id"); categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}

	// Date range filters
	if from := c.Query("from"); from != "" {
		fromDate, err := time.Parse("2006-01-02", from)
		if err == nil {
			query = query.Where("txn_date >= ?", fromDate)
		}
	}

	if to := c.Query("to"); to != "" {
		toDate, err := time.Parse("2006-01-02", to)
		if err == nil {
			query = query.Where("txn_date <= ?", toDate)
		}
	}

	// Amount range filters
	if minAmount := c.Query("min_amount"); minAmount != "" {
		if min, err := strconv.ParseInt(minAmount, 10, 64); err == nil {
			query = query.Where("amount_cents >= ?", min)
		}
	}

	if maxAmount := c.Query("max_amount"); maxAmount != "" {
		if max, err := strconv.ParseInt(maxAmount, 10, 64); err == nil {
			query = query.Where("amount_cents <= ?", max)
		}
	}

	// Search in description/notes
	if search := c.Query("search"); search != "" {
		query = query.Where("description ILIKE ? OR notes ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Step 4: Execute query with preloads
	if err := query.
		Preload("Account").
		Preload("Category").
		Preload("Splits.Category").
		Order("txn_date DESC, created_at DESC").
		Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch transactions"})
		return
	}

	c.JSON(http.StatusOK, transactions)
}

// GetTransaction retrieves a single transaction by ID
func GetTransaction(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := uint(claims.(jwt.MapClaims)["sub"].(float64))
	transactionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction ID"})
		return
	}

	var transaction models.Transaction
	if err := db.DB.
		Preload("Account").
		Preload("Category").
		Preload("Splits.Category").
		Where("id = ? AND user_id = ?", transactionID, userID).
		First(&transaction).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
		return
	}

	c.JSON(http.StatusOK, transaction)
}

// CreateTransaction creates a new transaction
func CreateTransaction(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := uint(claims.(jwt.MapClaims)["sub"].(float64))

	// Define input structure
	var input struct {
		AccountID   uint   `json:"account_id" binding:"required"`
		CategoryID  *uint  `json:"category_id"`
		AmountCents int64  `json:"amount_cents" binding:"required"`
		Description string `json:"description"`
		TxnDate     string `json:"txn_date" binding:"required"` // YYYY-MM-DD format
		Notes       string `json:"notes"`
		Splits      []struct {
			CategoryID  uint  `json:"category_id" binding:"required"`
			AmountCents int64 `json:"amount_cents" binding:"required"`
		} `json:"splits"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate amount (non-zero)
	if input.AmountCents == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount cannot be zero"})
		return
	}

	// Parse date
	txnDate, err := time.Parse("2006-01-02", input.TxnDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
		return
	}

	// Verify account belongs to user
	var account models.Account
	if err := db.DB.Where("id = ? AND user_id = ?", input.AccountID, userID).First(&account).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account not found or does not belong to user"})
		return
	}

	// Verify category belongs to user (if provided) and get category info
	var category *models.Category
	if input.CategoryID != nil {
		cat := models.Category{}
		if err := db.DB.Where("id = ? AND user_id = ?", *input.CategoryID, userID).First(&cat).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "category not found or does not belong to user"})
			return
		}
		category = &cat
	}

	// Adjust amount sign based on category type
	// Positive = income, Negative = expense
	finalAmount := input.AmountCents
	if category != nil && category.Kind == "expense" {
		// Make sure expenses are negative
		if finalAmount > 0 {
			finalAmount = -finalAmount
		}
	} else if category != nil && category.Kind == "income" {
		// Make sure income is positive
		if finalAmount < 0 {
			finalAmount = -finalAmount
		}
	}
	// If no category, keep the amount as provided by user

	// Validate splits if provided
	if len(input.Splits) > 0 {
		var splitTotal int64
		for _, split := range input.Splits {
			// Verify each split category
			var category models.Category
			if err := db.DB.Where("id = ? AND user_id = ?", split.CategoryID, userID).First(&category).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "split category not found or does not belong to user"})
				return
			}
			splitTotal += split.AmountCents
		}

		// Splits must equal transaction amount (with same sign)
		if splitTotal != input.AmountCents {
			c.JSON(http.StatusBadRequest, gin.H{"error": "split amounts must equal transaction amount"})
			return
		}

		// If splits exist, category_id should be null
		input.CategoryID = nil
	}

	// Create transaction
	transaction := models.Transaction{
		UserID:      userID,
		AccountID:   input.AccountID,
		CategoryID:  input.CategoryID,
		AmountCents: finalAmount,
		Description: input.Description,
		TxnDate:     txnDate,
		Notes:       input.Notes,
	} // Start database transaction for atomicity
	tx := db.DB.Begin()

	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create transaction"})
		return
	}

	// Track metrics
	middleware.IncrementTransactionsCreated()

	// Create splits if provided
	if len(input.Splits) > 0 {
		for _, split := range input.Splits {
			transactionSplit := models.TransactionSplit{
				ParentTxnID: transaction.ID,
				CategoryID:  split.CategoryID,
				AmountCents: split.AmountCents,
			}
			if err := tx.Create(&transactionSplit).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create transaction splits"})
				return
			}
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit transaction"})
		return
	}

	// Update account balance
	if err := UpdateAccountBalance(input.AccountID); err != nil {
		// Log error but don't fail the request since transaction was created
		// In production, you might want to use a job queue for this
	}

	// Reload with relationships
	db.DB.Preload("Account").Preload("Category").Preload("Splits.Category").First(&transaction, transaction.ID)

	c.JSON(http.StatusCreated, transaction)
}

// UpdateTransaction updates an existing transaction
func UpdateTransaction(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := uint(claims.(jwt.MapClaims)["sub"].(float64))
	transactionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction ID"})
		return
	}

	var transaction models.Transaction
	if err := db.DB.Where("id = ? AND user_id = ?", transactionID, userID).First(&transaction).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
		return
	}

	var input struct {
		AccountID   uint   `json:"account_id"`
		CategoryID  *uint  `json:"category_id"`
		AmountCents int64  `json:"amount_cents"`
		Description string `json:"description"`
		TxnDate     string `json:"txn_date"`
		Notes       string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify account if provided
	if input.AccountID != 0 {
		var account models.Account
		if err := db.DB.Where("id = ? AND user_id = ?", input.AccountID, userID).First(&account).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "account not found or does not belong to user"})
			return
		}
		transaction.AccountID = input.AccountID
	}

	// Verify category if provided and get category info
	var category *models.Category
	if input.CategoryID != nil {
		cat := models.Category{}
		if err := db.DB.Where("id = ? AND user_id = ?", *input.CategoryID, userID).First(&cat).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "category not found or does not belong to user"})
			return
		}
		transaction.CategoryID = input.CategoryID
		category = &cat
	}

	if input.AmountCents != 0 {
		// Adjust amount sign based on category type if category is being set
		finalAmount := input.AmountCents
		if category != nil && category.Kind == "expense" {
			// Make sure expenses are negative
			if finalAmount > 0 {
				finalAmount = -finalAmount
			}
		} else if category != nil && category.Kind == "income" {
			// Make sure income is positive
			if finalAmount < 0 {
				finalAmount = -finalAmount
			}
		}
		transaction.AmountCents = finalAmount
	}

	if input.Description != "" {
		transaction.Description = input.Description
	}

	if input.TxnDate != "" {
		txnDate, err := time.Parse("2006-01-02", input.TxnDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
			return
		}
		transaction.TxnDate = txnDate
	}

	if input.Notes != "" {
		transaction.Notes = input.Notes
	}

	if err := db.DB.Save(&transaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update transaction"})
		return
	}

	// Update account balance
	if err := UpdateAccountBalance(transaction.AccountID); err != nil {
		// Log error but don't fail the request since transaction was updated
	}

	// Reload with relationships
	db.DB.Preload("Account").Preload("Category").Preload("Splits.Category").First(&transaction, transaction.ID)

	c.JSON(http.StatusOK, transaction)
}

// DeleteTransaction deletes a transaction
func DeleteTransaction(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := uint(claims.(jwt.MapClaims)["sub"].(float64))
	transactionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction ID"})
		return
	}

	var transaction models.Transaction
	if err := db.DB.Where("id = ? AND user_id = ?", transactionID, userID).First(&transaction).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
		return
	}

	// Store account ID before deletion for balance update
	accountID := transaction.AccountID

	// Delete splits first (if any)
	db.DB.Where("parent_txn_id = ?", transactionID).Delete(&models.TransactionSplit{})

	// Delete transaction
	if err := db.DB.Delete(&transaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete transaction"})
		return
	}

	// Update account balance
	if err := UpdateAccountBalance(accountID); err != nil {
		// Log error but don't fail the request since transaction was deleted
	}

	c.JSON(http.StatusOK, gin.H{"message": "transaction deleted successfully"})
}
