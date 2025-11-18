package routes

import (
	"Personal-Finance-Tracker-backend/controllers"
	"Personal-Finance-Tracker-backend/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(rg *gin.RouterGroup) {
	// Accounts
	rg.GET("/accounts", controllers.GetAccounts)
	rg.GET("/accounts/:id", controllers.GetAccount)
	rg.POST("/accounts", controllers.CreateAccount)
	rg.PUT("/accounts/:id", controllers.UpdateAccount)
	rg.DELETE("/accounts/:id", controllers.DeleteAccount)

	// Transactions
	rg.GET("/transactions", controllers.GetTransactions)
	rg.GET("/transactions/:id", controllers.GetTransaction)
	rg.POST("/transactions", controllers.CreateTransaction)
	rg.PUT("/transactions/:id", controllers.UpdateTransaction)
	rg.DELETE("/transactions/:id", controllers.DeleteTransaction)

	// Categories
	rg.GET("/categories", controllers.GetCategories)
	rg.GET("/categories/tree", controllers.GetCategoryTree)
	rg.GET("/categories/:id", controllers.GetCategory)
	rg.GET("/categories/:id/usage", controllers.GetCategoryUsage)
	rg.POST("/categories", controllers.CreateCategory)
	rg.PUT("/categories/:id", controllers.UpdateCategory)
	rg.DELETE("/categories/:id", controllers.DeleteCategory)

	// Budgets
	rg.GET("/budgets", controllers.GetBudgets)
	rg.GET("/budgets/current", controllers.GetCurrentBudget)
	rg.GET("/budgets/:id", controllers.GetBudget)
	rg.POST("/budgets", controllers.CreateBudget)
	rg.PUT("/budgets/:id", controllers.UpdateBudget)
	rg.DELETE("/budgets/:id", controllers.DeleteBudget)

	// User Profile Management
	rg.GET("/user/profile", controllers.GetUserProfile)
	rg.PUT("/user/profile", controllers.UpdateUserProfile)
	rg.PUT("/user/change-password", controllers.ChangePassword)
	rg.DELETE("/user/account", controllers.DeleteUserAccount)

	// Reports & Analytics
	rg.GET("/reports/spend-summary", controllers.GetSpendSummary)
	rg.GET("/reports/cashflow", controllers.GetCashflow)
	rg.GET("/reports/account-balances", controllers.GetAccountBalances)
	rg.GET("/reports/budget-progress", controllers.GetBudgetProgress)
	rg.GET("/reports/monthly-trends", controllers.GetMonthlyTrends)
	rg.GET("/reports/top-merchants", controllers.GetTopMerchants)

	// AI-powered budget predictions
	rg.GET("/ai/budget-predictions", controllers.GetBudgetPrediction)
	rg.GET("/ai/spending-patterns", controllers.GetSpendingPatterns)

	// Bank Integration - Plaid only
	rg.GET("/banks/connections", controllers.GetBankConnections)
	rg.DELETE("/banks/connections/:id", controllers.DisconnectBank)

	// Plaid - FREE Banking API (100 users/month)
	rg.POST("/plaid/create_link_token", controllers.CreateLinkToken)
	rg.POST("/plaid/exchange_public_token", controllers.ExchangePublicToken)
	rg.POST("/plaid/sync/:id", controllers.SyncPlaidTransactions)
	rg.GET("/plaid/accounts/:id", controllers.GetPlaidAccounts)

	// Admin routes (require admin role)
	admin := rg.Group("/admin")
	admin.Use(middleware.RequireAdmin())
	{
		// Dashboard stats
		admin.GET("/dashboard-stats", controllers.GetDashboardStats)

		// User management
		admin.GET("/users", controllers.GetAllUsers)
		admin.GET("/users/:id", controllers.GetUserDetails)
		admin.DELETE("/users/:id", controllers.DeleteUserAdmin)
		admin.PUT("/users/:id/role", controllers.UpdateUserRole)

		// Data oversight
		admin.GET("/transactions", controllers.GetAllTransactions)
		admin.GET("/accounts", controllers.GetAllAccounts)
		admin.GET("/categories", controllers.GetAllCategories)
		admin.GET("/budgets", controllers.GetAllBudgets)
		admin.GET("/budgets/:id", controllers.GetBudgetDetails)

		// Data deletion (admin override)
		admin.DELETE("/transactions/:id", controllers.DeleteTransactionAdmin)
		admin.DELETE("/accounts/:id", controllers.DeleteAccountAdmin)
		admin.DELETE("/categories/:id", controllers.DeleteCategoryAdmin)
		admin.DELETE("/budgets/:id", controllers.DeleteBudgetAdmin)
	}
}
