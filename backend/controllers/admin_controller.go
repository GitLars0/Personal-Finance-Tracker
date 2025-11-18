package controllers

import (
	"net/http"
	"strconv"

	"Personal-Finance-Tracker-backend/db"
	"Personal-Finance-Tracker-backend/models"

	"github.com/gin-gonic/gin"
)

// GetAllUsers returns all users (admin only)
func GetAllUsers(c *gin.Context) {
	var users []models.User

	// Get all users with basic info (password hash excluded by json:"-" tag)
	if err := db.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

// GetUserDetails returns detailed info about a specific user (admin only)
func GetUserDetails(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var user models.User
	if err := db.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Get user's accounts count
	var accountCount int64
	db.DB.Model(&models.Account{}).Where("user_id = ?", userID).Count(&accountCount)

	// Get user's transactions count
	var transactionCount int64
	db.DB.Model(&models.Transaction{}).Where("user_id = ?", userID).Count(&transactionCount)

	// Get user's categories count
	var categoryCount int64
	db.DB.Model(&models.Category{}).Where("user_id = ?", userID).Count(&categoryCount)

	// Get user's budgets count
	var budgetCount int64
	db.DB.Model(&models.Budget{}).Where("user_id = ?", userID).Count(&budgetCount)

	response := gin.H{
		"user": user,
		"statistics": gin.H{
			"accounts":     accountCount,
			"transactions": transactionCount,
			"categories":   categoryCount,
			"budgets":      budgetCount,
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetAllTransactions returns all transactions from all users (admin only)
func GetAllTransactions(c *gin.Context) {
	type TransactionWithUser struct {
		ID           uint    `json:"id"`
		Description  string  `json:"description"`
		AmountCents  int64   `json:"amount_cents"`
		Amount       float64 `json:"amount"`
		Type         string  `json:"type"`
		TxnDate      string  `json:"txn_date"`
		CreatedAt    string  `json:"created_at"`
		UserID       uint    `json:"user_id"`
		UserUsername string  `json:"user_username"`
		UserEmail    string  `json:"user_email"`
		AccountID    uint    `json:"account_id"`
		AccountName  string  `json:"account_name"`
		CategoryID   *uint   `json:"category_id"`
		CategoryName string  `json:"category_name"`
	}

	var results []TransactionWithUser

	query := `
		SELECT 
			t.id, t.description, t.amount_cents, t.txn_date, t.created_at,
			t.user_id, u.username as user_username, u.email as user_email,
			t.account_id, a.name as account_name,
			t.category_id, COALESCE(c.name, 'Uncategorized') as category_name
		FROM transactions t
		LEFT JOIN users u ON t.user_id = u.id
		LEFT JOIN accounts a ON t.account_id = a.id
		LEFT JOIN categories c ON t.category_id = c.id
		ORDER BY t.created_at DESC
	`

	if err := db.DB.Raw(query).Scan(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch transactions"})
		return
	}

	// Convert amount_cents to amount and determine type
	for i := range results {
		results[i].Amount = float64(results[i].AmountCents) / 100.0
		if results[i].AmountCents > 0 {
			results[i].Type = "income"
		} else {
			results[i].Type = "expense"
		}
	}

	c.JSON(http.StatusOK, gin.H{"transactions": results})
}

// GetAllAccounts returns all accounts from all users (admin only)
func GetAllAccounts(c *gin.Context) {
	type AccountWithUser struct {
		ID                  uint    `json:"id"`
		Name                string  `json:"name"`
		AccountType         string  `json:"account_type"`
		Currency            string  `json:"currency"`
		InitialBalanceCents int64   `json:"initial_balance_cents"`
		CurrentBalanceCents int64   `json:"current_balance_cents"`
		Balance             float64 `json:"balance"`
		CreatedAt           string  `json:"created_at"`
		UserID              uint    `json:"user_id"`
		UserUsername        string  `json:"user_username"`
		UserEmail           string  `json:"user_email"`
	}

	var results []AccountWithUser

	query := `
		SELECT 
			a.id, a.name, a.type as account_type, a.currency, 
			a.initial_balance_cents, a.current_balance_cents, a.created_at,
			a.user_id, u.username as user_username, u.email as user_email
		FROM accounts a
		LEFT JOIN users u ON a.user_id = u.id
		ORDER BY a.created_at DESC
	`

	if err := db.DB.Raw(query).Scan(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch accounts"})
		return
	}

	// Convert balance_cents to balance
	for i := range results {
		results[i].Balance = float64(results[i].CurrentBalanceCents) / 100.0
	}

	c.JSON(http.StatusOK, gin.H{"accounts": results})
}

// GetAllCategories returns all categories from all users (admin only)
func GetAllCategories(c *gin.Context) {
	type CategoryWithUser struct {
		ID           uint   `json:"id"`
		Name         string `json:"name"`
		Kind         string `json:"kind"`
		Type         string `json:"type"` // Alias for kind to match frontend
		ParentID     *uint  `json:"parent_id"`
		ParentName   string `json:"parent_name"`
		CreatedAt    string `json:"created_at"`
		UserID       uint   `json:"user_id"`
		UserUsername string `json:"user_username"`
		UserEmail    string `json:"user_email"`
	}

	var results []CategoryWithUser

	query := `
		SELECT 
			c.id, c.name, c.kind, c.kind as type, c.parent_id, c.created_at,
			c.user_id, u.username as user_username, u.email as user_email,
			COALESCE(pc.name, '') as parent_name
		FROM categories c
		LEFT JOIN users u ON c.user_id = u.id
		LEFT JOIN categories pc ON c.parent_id = pc.id
		ORDER BY c.created_at DESC
	`

	if err := db.DB.Raw(query).Scan(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch categories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"categories": results})
}

// GetAllBudgets returns all budgets from all users (admin only)
func GetAllBudgets(c *gin.Context) {
	type BudgetWithUser struct {
		ID           uint    `json:"id"`
		PeriodStart  string  `json:"period_start"`
		PeriodEnd    string  `json:"period_end"`
		StartDate    string  `json:"start_date"` // Alias for frontend
		EndDate      string  `json:"end_date"`   // Alias for frontend
		Currency     string  `json:"currency"`
		CreatedAt    string  `json:"created_at"`
		UserID       uint    `json:"user_id"`
		UserUsername string  `json:"user_username"`
		UserEmail    string  `json:"user_email"`
		Name         string  `json:"name"`   // Computed name
		Amount       float64 `json:"amount"` // Total planned amount
		Spent        float64 `json:"spent"`  // Total spent (placeholder)
	}

	var results []BudgetWithUser

	query := `
		SELECT 
			b.id, b.period_start, b.period_end, b.currency, b.created_at,
			b.user_id, u.username as user_username, u.email as user_email,
			b.period_start as start_date, b.period_end as end_date
		FROM budgets b
		LEFT JOIN users u ON b.user_id = u.id
		ORDER BY b.created_at DESC
	`

	if err := db.DB.Raw(query).Scan(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch budgets"})
		return
	}

	// Calculate totals and create names for each budget
	for i := range results {
		// Create a simple name based on the period
		results[i].Name = "Budget " + results[i].PeriodStart[:7] // "Budget 2024-01"

		// Get total planned amount from budget items
		var totalPlannedCents int64
		db.DB.Table("budget_items").
			Where("budget_id = ?", results[i].ID).
			Select("COALESCE(SUM(planned_cents), 0)").
			Scan(&totalPlannedCents)

		results[i].Amount = float64(totalPlannedCents) / 100.0

		// Calculate spent amount from transactions in the budget period
		var totalSpentCents int64
		db.DB.Table("transactions").
			Where("user_id = ? AND txn_date >= ? AND txn_date <= ? AND amount_cents < 0",
				results[i].UserID, results[i].PeriodStart, results[i].PeriodEnd).
			Select("COALESCE(SUM(ABS(amount_cents)), 0)").
			Scan(&totalSpentCents)

		results[i].Spent = float64(totalSpentCents) / 100.0
	}

	c.JSON(http.StatusOK, gin.H{"budgets": results})
}

// DeleteUserAdmin allows admin to delete any user and all their data
func DeleteUserAdmin(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Prevent admin from deleting themselves
	adminUser, _ := c.Get("adminUser")
	if adminUser.(models.User).ID == uint(userID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete your own admin account"})
		return
	}

	// Check if user exists
	var user models.User
	if err := db.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Start transaction for atomic deletion
	tx := db.DB.Begin()

	// Delete user's budget items first (foreign key constraint)
	if err := tx.Where("budget_id IN (SELECT id FROM budgets WHERE user_id = ?)", userID).Delete(&models.BudgetItem{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user budget items"})
		return
	}

	// Delete user's budgets
	if err := tx.Where("user_id = ?", userID).Delete(&models.Budget{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user budgets"})
		return
	}

	// Delete user's transaction splits
	if err := tx.Where("parent_txn_id IN (SELECT id FROM transactions WHERE user_id = ?)", userID).Delete(&models.TransactionSplit{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user transaction splits"})
		return
	}

	// Delete user's transactions
	if err := tx.Where("user_id = ?", userID).Delete(&models.Transaction{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user transactions"})
		return
	}

	// Delete user's categories
	if err := tx.Where("user_id = ?", userID).Delete(&models.Category{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user categories"})
		return
	}

	// Delete user's accounts
	if err := tx.Where("user_id = ?", userID).Delete(&models.Account{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user accounts"})
		return
	}

	// Finally delete the user
	if err := tx.Delete(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit user deletion"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
}

// DeleteTransactionAdmin allows admin to delete any transaction
func DeleteTransactionAdmin(c *gin.Context) {
	transactionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction ID"})
		return
	}

	// Check if transaction exists
	var transaction models.Transaction
	if err := db.DB.Where("id = ?", transactionID).First(&transaction).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
		return
	}

	// Delete transaction splits first
	if err := db.DB.Where("parent_txn_id = ?", transactionID).Delete(&models.TransactionSplit{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete transaction splits"})
		return
	}

	// Delete the transaction
	if err := db.DB.Delete(&transaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "transaction deleted successfully"})
}

// DeleteAccountAdmin allows admin to delete any account
func DeleteAccountAdmin(c *gin.Context) {
	accountID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account ID"})
		return
	}

	// Check if account exists
	var account models.Account
	if err := db.DB.Where("id = ?", accountID).First(&account).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}

	// Start transaction for atomic deletion
	tx := db.DB.Begin()

	// Delete transaction splits first
	if err := tx.Where("parent_txn_id IN (SELECT id FROM transactions WHERE account_id = ?)", accountID).Delete(&models.TransactionSplit{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete transaction splits"})
		return
	}

	// Delete transactions associated with this account
	if err := tx.Where("account_id = ?", accountID).Delete(&models.Transaction{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete account transactions"})
		return
	}

	// Delete the account
	if err := tx.Delete(&account).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete account"})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit account deletion"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "account deleted successfully"})
}

// DeleteCategoryAdmin allows admin to delete any category
func DeleteCategoryAdmin(c *gin.Context) {
	categoryID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
		return
	}

	// Check if category exists
	var category models.Category
	if err := db.DB.Where("id = ?", categoryID).First(&category).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		return
	}

	// Start transaction for atomic deletion
	tx := db.DB.Begin()

	// Update transactions to remove category reference (set to null)
	if err := tx.Model(&models.Transaction{}).Where("category_id = ?", categoryID).Update("category_id", nil).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update transactions"})
		return
	}

	// Update budget items to remove category reference
	if err := tx.Where("category_id = ?", categoryID).Delete(&models.BudgetItem{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete budget items"})
		return
	}

	// Update child categories to remove parent reference
	if err := tx.Model(&models.Category{}).Where("parent_id = ?", categoryID).Update("parent_id", nil).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update child categories"})
		return
	}

	// Delete the category
	if err := tx.Delete(&category).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete category"})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit category deletion"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "category deleted successfully"})
}

// DeleteBudgetAdmin allows admin to delete any budget
func DeleteBudgetAdmin(c *gin.Context) {
	budgetID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid budget ID"})
		return
	}

	// Check if budget exists
	var budget models.Budget
	if err := db.DB.Where("id = ?", budgetID).First(&budget).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "budget not found"})
		return
	}

	// Start transaction for atomic deletion
	tx := db.DB.Begin()

	// Delete budget items first
	if err := tx.Where("budget_id = ?", budgetID).Delete(&models.BudgetItem{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete budget items"})
		return
	}

	// Delete the budget
	if err := tx.Delete(&budget).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete budget"})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit budget deletion"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "budget deleted successfully"})
}

// GetBudgetDetails returns detailed info about a specific budget (admin only)
func GetBudgetDetails(c *gin.Context) {
	budgetID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid budget ID"})
		return
	}

	// Get budget info
	var budget models.Budget
	if err := db.DB.Preload("Items.Category").Where("id = ?", budgetID).First(&budget).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "budget not found"})
		return
	}

	// Get user info
	var user models.User
	db.DB.Where("id = ?", budget.UserID).First(&user)

	// Prepare detailed response
	type CategoryProgress struct {
		ID                 uint    `json:"id"`
		Name               string  `json:"name"`
		PlannedAmount      float64 `json:"planned_amount"`
		SpentAmount        float64 `json:"spent_amount"`
		TransactionCount   int64   `json:"transaction_count"`
		ProgressPercentage float64 `json:"progress_percentage"`
	}

	var categories []CategoryProgress
	totalPlanned := 0.0
	totalSpent := 0.0

	// Process each budget item (category)
	for _, item := range budget.Items {
		plannedAmount := float64(item.PlannedCents) / 100.0
		totalPlanned += plannedAmount

		// Calculate spent amount for this category in the budget period
		var spentCents int64
		var transactionCount int64

		db.DB.Model(&models.Transaction{}).
			Where("user_id = ? AND category_id = ? AND txn_date >= ? AND txn_date <= ? AND amount_cents < 0",
				budget.UserID, item.CategoryID, budget.PeriodStart, budget.PeriodEnd).
			Count(&transactionCount)

		db.DB.Table("transactions").
			Where("user_id = ? AND category_id = ? AND txn_date >= ? AND txn_date <= ? AND amount_cents < 0",
				budget.UserID, item.CategoryID, budget.PeriodStart, budget.PeriodEnd).
			Select("COALESCE(SUM(ABS(amount_cents)), 0)").
			Scan(&spentCents)

		spentAmount := float64(spentCents) / 100.0
		totalSpent += spentAmount

		progressPercentage := 0.0
		if plannedAmount > 0 {
			progressPercentage = (spentAmount / plannedAmount) * 100.0
		}

		categories = append(categories, CategoryProgress{
			ID:                 item.Category.ID,
			Name:               item.Category.Name,
			PlannedAmount:      plannedAmount,
			SpentAmount:        spentAmount,
			TransactionCount:   transactionCount,
			ProgressPercentage: progressPercentage,
		})
	}

	// Calculate summary
	overallProgress := 0.0
	if totalPlanned > 0 {
		overallProgress = (totalSpent / totalPlanned) * 100.0
	}

	categoriesOverBudget := 0
	categoriesOnTrack := 0
	for _, cat := range categories {
		if cat.ProgressPercentage > 100 {
			categoriesOverBudget++
		} else {
			categoriesOnTrack++
		}
	}

	response := gin.H{
		"budget": gin.H{
			"id":            budget.ID,
			"name":          "Budget " + budget.PeriodStart.Format("2006-01"),
			"start_date":    budget.PeriodStart.Format("2006-01-02"),
			"end_date":      budget.PeriodEnd.Format("2006-01-02"),
			"user_username": user.Username,
			"user_email":    user.Email,
		},
		"categories": categories,
		"summary": gin.H{
			"total_planned":          totalPlanned,
			"total_spent":            totalSpent,
			"total_remaining":        totalPlanned - totalSpent,
			"overall_progress":       overallProgress,
			"categories_over_budget": categoriesOverBudget,
			"categories_on_track":    categoriesOnTrack,
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetDashboardStats returns dashboard statistics for admin
func GetDashboardStats(c *gin.Context) {
	var stats struct {
		TotalUsers        int64 `json:"totalUsers"`
		TotalTransactions int64 `json:"totalTransactions"`
		TotalAccounts     int64 `json:"totalAccounts"`
		TotalCategories   int64 `json:"totalCategories"`
		TotalBudgets      int64 `json:"totalBudgets"`
	}

	// Count users
	db.DB.Model(&models.User{}).Count(&stats.TotalUsers)

	// Count transactions
	db.DB.Model(&models.Transaction{}).Count(&stats.TotalTransactions)

	// Count accounts
	db.DB.Model(&models.Account{}).Count(&stats.TotalAccounts)

	// Count categories
	db.DB.Model(&models.Category{}).Count(&stats.TotalCategories)

	// Count budgets
	db.DB.Model(&models.Budget{}).Count(&stats.TotalBudgets)

	c.JSON(http.StatusOK, stats)
}

// UpdateUserRole allows admin to change user roles
func UpdateUserRole(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var input struct {
		Role models.UserRole `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate role
	if input.Role != models.UserRoleUser && input.Role != models.UserRoleAdmin {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		return
	}

	// Prevent admin from demoting themselves
	adminUser, _ := c.Get("adminUser")
	if adminUser.(models.User).ID == uint(userID) && input.Role != models.UserRoleAdmin {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot change your own admin role"})
		return
	}

	// Update user role
	if err := db.DB.Model(&models.User{}).Where("id = ?", userID).Update("role", input.Role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user role updated successfully"})
}
