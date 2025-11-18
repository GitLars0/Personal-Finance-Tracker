package controllers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"Personal-Finance-Tracker-backend/db"
	"Personal-Finance-Tracker-backend/middleware"
	"Personal-Finance-Tracker-backend/models"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
)

// GetBudgets retrieves all budgets for the authenticated user
func GetBudgets(c *gin.Context) {
	// Step 1: Authenticate
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := uint(claims.(jwt.MapClaims)["sub"].(float64))

	var budgets []models.Budget

	// Step 2: Always filter by user_id (SECURITY CRITICAL!)
	query := db.DB.Where("user_id = ?", userID)

	// Step 3: Optional filters

	// Filter by active budgets (current date falls within period)
	if active := c.Query("active"); active == "true" {
		now := time.Now()
		query = query.Where("period_start <= ? AND period_end >= ?", now, now)
	}

	// Filter by date range
	if from := c.Query("from"); from != "" {
		fromDate, err := time.Parse("2006-01-02", from)
		if err == nil {
			query = query.Where("period_start >= ?", fromDate)
		}
	}

	if to := c.Query("to"); to != "" {
		toDate, err := time.Parse("2006-01-02", to)
		if err == nil {
			query = query.Where("period_end <= ?", toDate)
		}
	}

	// Filter by currency
	if currency := c.Query("currency"); currency != "" {
		query = query.Where("currency = ?", currency)
	}

	// Step 4: Execute query with preloads
	if err := query.
		Preload("Items.Category").
		Order("period_start DESC").
		Find(&budgets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch budgets"})
		return
	}

	// Add spending calculations to each budget
	type BudgetItemWithSpending struct {
		models.BudgetItem
		SpentCents int64   `json:"spent_cents"`
		Progress   float64 `json:"progress_percent"`
		Status     string  `json:"status"`
	}

	type BudgetWithSpending struct {
		models.Budget
		Items   []BudgetItemWithSpending `json:"items"`
		Summary struct {
			TotalPlannedCents   int64 `json:"total_planned_cents"`
			TotalSpentCents     int64 `json:"total_spent_cents"`
			TotalRemainingCents int64 `json:"total_remaining_cents"`
		} `json:"summary"`
	}

	var budgetsWithSpending []BudgetWithSpending

	for _, budget := range budgets {
		var itemsWithSpending []BudgetItemWithSpending
		var totalPlanned, totalSpent int64

		for _, item := range budget.Items {
			// Calculate actual spending for this category during budget period
			var spentCents int64
			db.DB.Model(&models.Transaction{}).
				Where("user_id = ? AND category_id = ? AND txn_date >= ? AND txn_date <= ? AND amount_cents < 0",
					userID, item.CategoryID, budget.PeriodStart, budget.PeriodEnd).
				Select("COALESCE(SUM(ABS(amount_cents)), 0)").
				Scan(&spentCents)

			// Also check transaction splits
			var splitSpent int64
			db.DB.Table("transaction_splits").
				Joins("JOIN transactions ON transactions.id = transaction_splits.parent_txn_id").
				Where("transactions.user_id = ? AND transaction_splits.category_id = ? AND transactions.txn_date >= ? AND transactions.txn_date <= ?",
					userID, item.CategoryID, budget.PeriodStart, budget.PeriodEnd).
				Select("COALESCE(SUM(ABS(transaction_splits.amount_cents)), 0)").
				Scan(&splitSpent)

			spentCents += splitSpent

			// Calculate progress and status
			progress := 0.0
			if item.PlannedCents > 0 {
				progress = (float64(spentCents) / float64(item.PlannedCents)) * 100
			}

			status := "under_budget"
			if progress > 100 {
				status = "over_budget"
			} else if progress >= 90 {
				status = "on_track"
			}

			totalPlanned += item.PlannedCents
			totalSpent += spentCents

			itemsWithSpending = append(itemsWithSpending, BudgetItemWithSpending{
				BudgetItem: item,
				SpentCents: spentCents,
				Progress:   progress,
				Status:     status,
			})
		}

		budgetWithSpending := BudgetWithSpending{
			Budget: budget,
			Items:  itemsWithSpending,
		}
		budgetWithSpending.Summary.TotalPlannedCents = totalPlanned
		budgetWithSpending.Summary.TotalSpentCents = totalSpent
		budgetWithSpending.Summary.TotalRemainingCents = totalPlanned - totalSpent

		budgetsWithSpending = append(budgetsWithSpending, budgetWithSpending)
	}

	c.JSON(http.StatusOK, budgetsWithSpending)
}

// GetBudget retrieves a single budget by ID with spending details
func GetBudget(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := uint(claims.(jwt.MapClaims)["sub"].(float64))
	budgetID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid budget ID"})
		return
	}

	var budget models.Budget
	if err := db.DB.
		Preload("Items.Category").
		Where("id = ? AND user_id = ?", budgetID, userID).
		First(&budget).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "budget not found"})
		return
	}

	// Calculate actual spending for each budget item
	type BudgetItemWithSpending struct {
		models.BudgetItem
		SpentCents int64   `json:"spent_cents"`
		Remaining  int64   `json:"remaining_cents"`
		Progress   float64 `json:"progress_percent"`
	}

	type BudgetWithSpending struct {
		models.Budget
		ItemsWithSpending   []BudgetItemWithSpending `json:"items_with_spending"`
		TotalPlannedCents   int64                    `json:"total_planned_cents"`
		TotalSpentCents     int64                    `json:"total_spent_cents"`
		TotalRemainingCents int64                    `json:"total_remaining_cents"`
	}

	var itemsWithSpending []BudgetItemWithSpending
	var totalPlanned int64
	var totalSpent int64

	for _, item := range budget.Items {
		// Calculate actual spending for this category during budget period
		// For expense transactions (negative amounts), sum the absolute values
		var spentCents int64
		err := db.DB.Model(&models.Transaction{}).
			Where("user_id = ? AND category_id = ? AND txn_date >= ? AND txn_date <= ? AND amount_cents < 0",
				userID, item.CategoryID, budget.PeriodStart, budget.PeriodEnd).
			Select("COALESCE(SUM(ABS(amount_cents)), 0)").
			Scan(&spentCents).Error

		if err != nil {
			log.Printf("Error calculating spending: %v", err)
			spentCents = 0
		}

		// Also check transaction splits
		var splitSpent int64
		err = db.DB.Table("transaction_splits").
			Joins("JOIN transactions ON transactions.id = transaction_splits.parent_txn_id").
			Where("transactions.user_id = ? AND transaction_splits.category_id = ? AND transactions.txn_date >= ? AND transactions.txn_date <= ? AND transaction_splits.amount_cents < 0",
				userID, item.CategoryID, budget.PeriodStart, budget.PeriodEnd).
			Select("COALESCE(SUM(ABS(transaction_splits.amount_cents)), 0)").
			Scan(&splitSpent).Error

		if err != nil {
			log.Printf("Error calculating split spending: %v", err)
			splitSpent = 0
		}

		spentCents += splitSpent

		remaining := item.PlannedCents - spentCents
		progress := 0.0
		if item.PlannedCents > 0 {
			progress = (float64(spentCents) / float64(item.PlannedCents)) * 100
		}

		itemsWithSpending = append(itemsWithSpending, BudgetItemWithSpending{
			BudgetItem: item,
			SpentCents: spentCents,
			Remaining:  remaining,
			Progress:   progress,
		})

		totalPlanned += item.PlannedCents
		totalSpent += spentCents
	}

	response := BudgetWithSpending{
		Budget:              budget,
		ItemsWithSpending:   itemsWithSpending,
		TotalPlannedCents:   totalPlanned,
		TotalSpentCents:     totalSpent,
		TotalRemainingCents: totalPlanned - totalSpent,
	}

	c.JSON(http.StatusOK, response)
}

// CreateBudget creates a new budget with items
func CreateBudget(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := uint(claims.(jwt.MapClaims)["sub"].(float64))

	// Define input structure
	var input struct {
		PeriodStart string `json:"period_start" binding:"required"` // YYYY-MM-DD
		PeriodEnd   string `json:"period_end" binding:"required"`   // YYYY-MM-DD
		Currency    string `json:"currency"`                        // Default USD
		Items       []struct {
			CategoryID   uint  `json:"category_id" binding:"required"`
			PlannedCents int64 `json:"planned_cents" binding:"required,gt=0"`
		} `json:"items" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		log.Printf("❌ Budget creation bind error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("✅ Budget creation input: %+v", input)

	// Parse dates
	periodStart, err := time.Parse("2006-01-02", input.PeriodStart)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid period_start format, use YYYY-MM-DD"})
		return
	}

	periodEnd, err := time.Parse("2006-01-02", input.PeriodEnd)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid period_end format, use YYYY-MM-DD"})
		return
	}

	// Validate date range
	if periodEnd.Before(periodStart) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "period_end must be after period_start"})
		return
	}

	// Check for overlapping budgets (optional - depends on business rules)
	// You might want one budget per month, or allow overlapping budgets
	var overlappingCount int64
	db.DB.Model(&models.Budget{}).
		Where("user_id = ? AND ((period_start <= ? AND period_end >= ?) OR (period_start <= ? AND period_end >= ?))",
			userID, periodStart, periodStart, periodEnd, periodEnd).
		Count(&overlappingCount)

	if overlappingCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "budget period overlaps with existing budget"})
		return
	}

	// Set default currency
	currency := "USD"
	if input.Currency != "" {
		currency = input.Currency
	}

	// Verify all categories belong to user and are expense categories
	categoryMap := make(map[uint]bool)
	for _, item := range input.Items {
		// Check for duplicate categories in items
		if categoryMap[item.CategoryID] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "duplicate category in budget items"})
			return
		}
		categoryMap[item.CategoryID] = true

		var category models.Category
		if err := db.DB.Where("id = ? AND user_id = ?", item.CategoryID, userID).First(&category).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "category not found or does not belong to user"})
			return
		}

		// Allow budgets for both expense and income categories
		if category.Kind != models.CategoryExpense && category.Kind != models.CategoryIncome {
			c.JSON(http.StatusBadRequest, gin.H{"error": "budgets can only be created for expense or income categories"})
			return
		}
	}

	// Create budget
	budget := models.Budget{
		UserID:      userID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Currency:    currency,
	}

	// Start database transaction for atomicity
	tx := db.DB.Begin()

	if err := tx.Create(&budget).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create budget"})
		return
	}

	middleware.IncrementBudgetsCreated()

	// Create budget items
	for _, item := range input.Items {
		budgetItem := models.BudgetItem{
			BudgetID:     budget.ID,
			CategoryID:   item.CategoryID,
			PlannedCents: item.PlannedCents,
		}
		if err := tx.Create(&budgetItem).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create budget items"})
			return
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit budget"})
		return
	}

	// Reload with relationships
	db.DB.Preload("Items.Category").First(&budget, budget.ID)

	c.JSON(http.StatusCreated, budget)
}

// UpdateBudget updates an existing budget
func UpdateBudget(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := uint(claims.(jwt.MapClaims)["sub"].(float64))
	budgetID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid budget ID"})
		return
	}

	var budget models.Budget
	if err := db.DB.Where("id = ? AND user_id = ?", budgetID, userID).First(&budget).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "budget not found"})
		return
	}

	var input struct {
		PeriodStart string `json:"period_start"`
		PeriodEnd   string `json:"period_end"`
		Currency    string `json:"currency"`
		Items       []struct {
			ID           uint  `json:"id"` // If provided, update existing item
			CategoryID   uint  `json:"category_id" binding:"required"`
			PlannedCents int64 `json:"planned_cents" binding:"required,gt=0"`
		} `json:"items"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update period dates if provided
	if input.PeriodStart != "" {
		periodStart, err := time.Parse("2006-01-02", input.PeriodStart)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid period_start format, use YYYY-MM-DD"})
			return
		}
		budget.PeriodStart = periodStart
	}

	if input.PeriodEnd != "" {
		periodEnd, err := time.Parse("2006-01-02", input.PeriodEnd)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid period_end format, use YYYY-MM-DD"})
			return
		}
		budget.PeriodEnd = periodEnd
	}

	// Validate date range
	if budget.PeriodEnd.Before(budget.PeriodStart) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "period_end must be after period_start"})
		return
	}

	if input.Currency != "" {
		budget.Currency = input.Currency
	}

	// Start transaction
	tx := db.DB.Begin()

	// Update budget
	if err := tx.Save(&budget).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update budget"})
		return
	}

	// Update items if provided
	if len(input.Items) > 0 {
		// Delete old items (simpler than complex diff logic)
		if err := tx.Where("budget_id = ?", budgetID).Delete(&models.BudgetItem{}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update budget items"})
			return
		}

		// Create new items
		for _, item := range input.Items {
			// Verify category
			var category models.Category
			// ✅ Use tx for transaction consistency
			if err := tx.Where("id = ? AND user_id = ?", item.CategoryID, userID).First(&category).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{"error": "category not found or does not belong to user"})
				return
			}

			budgetItem := models.BudgetItem{
				BudgetID:     budget.ID,
				CategoryID:   item.CategoryID,
				PlannedCents: item.PlannedCents,
			}
			if err := tx.Create(&budgetItem).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create budget items"})
				return
			}
		}
	}

	// Commit
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit budget update"})
		return
	}

	// Reload with relationships
	db.DB.Preload("Items.Category").First(&budget, budget.ID)

	c.JSON(http.StatusOK, budget)
}

// DeleteBudget deletes a budget and its items
func DeleteBudget(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := uint(claims.(jwt.MapClaims)["sub"].(float64))
	budgetID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid budget ID"})
		return
	}

	var budget models.Budget
	if err := db.DB.Where("id = ? AND user_id = ?", budgetID, userID).First(&budget).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "budget not found"})
		return
	}

	// Start transaction
	tx := db.DB.Begin()

	// Delete budget items first (CASCADE should handle this, but be explicit)
	if err := tx.Where("budget_id = ?", budgetID).Delete(&models.BudgetItem{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete budget items"})
		return
	}

	// Delete budget
	if err := tx.Delete(&budget).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete budget"})
		return
	}

	// Commit
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit budget deletion"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "budget deleted successfully"})
}

// GetCurrentBudget gets the active budget for current month
func GetCurrentBudget(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := uint(claims.(jwt.MapClaims)["sub"].(float64))
	now := time.Now()

	var budget models.Budget
	if err := db.DB.
		Preload("Items.Category").
		Where("user_id = ? AND period_start <= ? AND period_end >= ?", userID, now, now).
		First(&budget).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no active budget found for current period"})
		return
	}

	c.JSON(http.StatusOK, budget)
}
