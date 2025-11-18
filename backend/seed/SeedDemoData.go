package seed

import (
	"log"
	"time"

	"Personal-Finance-Tracker-backend/controllers"
	"Personal-Finance-Tracker-backend/models"

	"gorm.io/gorm"
)

// SeedDemoData adds sample data for testing/presentation
func SeedDemoData(db *gorm.DB) {
	// First, create the default admin user if it doesn't exist
	createDefaultAdmin(db)

	// Check if demo user exists
	var count int64
	db.Model(&models.User{}).Where("username = ?", "demo").Count(&count)
	if count > 0 {
		log.Println("ℹ️  Demo data already exists, skipping seed")
		return
	}

	log.Println("🌱 Seeding demo data...")

	// Use Argon2 hash (same as auth system)
	hash, err := controllers.HashPassword("demo123")
	if err != nil {
		log.Fatalf("❌ Failed to hash demo password: %v", err)
	}

	// Create demo user
	demoUser := models.User{
		Username:     "demo",
		Email:        "demo@example.com",
		PasswordHash: hash,
		Name:         "Demo User",
		Role:         models.UserRoleUser, // Explicitly set as regular user
	}
	db.Create(&demoUser)

	// Create additional demo users for clustering analysis
	additionalUsers := []models.User{
		{
			Username:     "user_conservative",
			Email:        "conservative@example.com",
			PasswordHash: hash,
			Name:         "Conservative User",
			Role:         models.UserRoleUser,
		},
		{
			Username:     "user_spender",
			Email:        "spender@example.com",
			PasswordHash: hash,
			Name:         "Big Spender",
			Role:         models.UserRoleUser,
		},
		{
			Username:     "user_balanced",
			Email:        "balanced@example.com",
			PasswordHash: hash,
			Name:         "Balanced User",
			Role:         models.UserRoleUser,
		},
		{
			Username:     "user_student",
			Email:        "student@example.com",
			PasswordHash: hash,
			Name:         "Student User",
			Role:         models.UserRoleUser,
		},
	}
	db.Create(&additionalUsers)

	// Create accounts for additional users
	allUsers := append([]models.User{demoUser}, additionalUsers...)
	var allAccounts []models.Account

	for _, user := range allUsers {
		userAccounts := []models.Account{
			{UserID: user.ID, Name: "Checking", Type: "checking", Currency: "USD"},
			{UserID: user.ID, Name: "Savings", Type: "savings", Currency: "USD"},
		}
		db.Create(&userAccounts)
		allAccounts = append(allAccounts, userAccounts...)
	}

	// Create demo accounts for main demo user
	accounts := []models.Account{
		{UserID: demoUser.ID, Name: "Main Checking", Type: "checking", Currency: "USD"},
		{UserID: demoUser.ID, Name: "Savings", Type: "savings", Currency: "USD"},
		{UserID: demoUser.ID, Name: "Credit Card", Type: "credit", Currency: "USD"},
	}
	db.Create(&accounts)

	// Create categories for all users
	var allCategories []models.Category
	for _, user := range allUsers {
		userCategories := []models.Category{
			{UserID: user.ID, Name: "Groceries", Kind: models.CategoryExpense},
			{UserID: user.ID, Name: "Salary", Kind: models.CategoryIncome},
			{UserID: user.ID, Name: "Rent", Kind: models.CategoryExpense},
			{UserID: user.ID, Name: "Transportation", Kind: models.CategoryExpense},
			{UserID: user.ID, Name: "Entertainment", Kind: models.CategoryExpense},
		}
		db.Create(&userCategories)
		allCategories = append(allCategories, userCategories...)
	}

	// Create demo categories for main demo user
	categories := []models.Category{
		{UserID: demoUser.ID, Name: "Groceries", Kind: models.CategoryExpense},
		{UserID: demoUser.ID, Name: "Salary", Kind: models.CategoryIncome},
		{UserID: demoUser.ID, Name: "Rent", Kind: models.CategoryExpense},
		{UserID: demoUser.ID, Name: "Transportation", Kind: models.CategoryExpense},
		{UserID: demoUser.ID, Name: "Entertainment", Kind: models.CategoryExpense},
	}
	db.Create(&categories)

	// Create demo transactions
	transactions := []models.Transaction{
		{
			UserID:      demoUser.ID,
			AccountID:   accounts[0].ID,
			CategoryID:  &categories[1].ID,
			AmountCents: 300000,
			Description: "October Salary",
			TxnDate:     time.Now().AddDate(0, 0, -10),
		},
		{
			UserID:      demoUser.ID,
			AccountID:   accounts[0].ID,
			CategoryID:  &categories[0].ID,
			AmountCents: -5000,
			Description: "Whole Foods",
			TxnDate:     time.Now().AddDate(0, 0, -5),
		},
		{
			UserID:      demoUser.ID,
			AccountID:   accounts[0].ID,
			CategoryID:  &categories[2].ID,
			AmountCents: -150000,
			Description: "Monthly Rent",
			TxnDate:     time.Now().AddDate(0, 0, -1),
		},
		{
			UserID:      demoUser.ID,
			AccountID:   accounts[0].ID,
			CategoryID:  &categories[3].ID,
			AmountCents: -3000,
			Description: "Uber",
			TxnDate:     time.Now().AddDate(0, 0, -3),
		},
		{
			UserID:      demoUser.ID,
			AccountID:   accounts[0].ID,
			CategoryID:  &categories[4].ID,
			AmountCents: -8000,
			Description: "Cinema Tickets",
			TxnDate:     time.Now().AddDate(0, 0, -2),
		},
	}
	db.Create(&transactions)

	// Create budgets for additional users with different spending patterns
	userTypes := map[string]map[string]int64{
		"user_conservative": {
			"Groceries":      25000, // $250 - conservative
			"Rent":           80000, // $800 - lower rent
			"Transportation": 10000, // $100 - minimal transport
			"Entertainment":  5000,  // $50 - very little entertainment
		},
		"user_spender": {
			"Groceries":      80000,  // $800 - premium groceries
			"Rent":           250000, // $2500 - expensive housing
			"Transportation": 50000,  // $500 - car payments
			"Entertainment":  40000,  // $400 - lots of entertainment
		},
		"user_balanced": {
			"Groceries":      45000,  // $450 - moderate
			"Rent":           120000, // $1200 - average rent
			"Transportation": 25000,  // $250 - reasonable transport
			"Entertainment":  20000,  // $200 - balanced entertainment
		},
		"user_student": {
			"Groceries":      15000, // $150 - tight budget
			"Rent":           60000, // $600 - shared housing
			"Transportation": 5000,  // $50 - public transport
			"Entertainment":  10000, // $100 - limited entertainment
		},
	}

	for i, user := range additionalUsers {
		username := user.Username
		budgetAmounts := userTypes[username]

		// Find user's categories
		var userCategories []models.Category
		db.Where("user_id = ?", user.ID).Find(&userCategories)

		// Create budget for this user
		userBudget := models.Budget{
			UserID:      user.ID,
			PeriodStart: time.Now().AddDate(0, 0, -15),
			PeriodEnd:   time.Now().AddDate(0, 0, 15),
			Currency:    "USD",
		}
		db.Create(&userBudget)

		// Create budget items based on user type
		var userBudgetItems []models.BudgetItem
		for _, category := range userCategories {
			if category.Kind == models.CategoryExpense {
				amount := budgetAmounts[category.Name]
				if amount > 0 {
					userBudgetItems = append(userBudgetItems, models.BudgetItem{
						BudgetID:     userBudget.ID,
						CategoryID:   category.ID,
						PlannedCents: amount,
					})
				}
			}
		}

		if len(userBudgetItems) > 0 {
			db.Create(&userBudgetItems)
		}

		// Create some transactions for spending pattern diversity
		var userAccounts []models.Account
		db.Where("user_id = ?", user.ID).Find(&userAccounts)

		if len(userAccounts) > 0 {
			// Create varied transaction patterns based on user type
			spendingMultiplier := []float64{0.8, 1.2, 1.0, 0.6}[i] // Conservative, Spender, Balanced, Student

			for _, item := range userBudgetItems {
				actualSpent := int64(float64(item.PlannedCents) * spendingMultiplier)
				userTransactions := []models.Transaction{
					{
						UserID:      user.ID,
						AccountID:   userAccounts[0].ID,
						CategoryID:  &item.CategoryID,
						AmountCents: -actualSpent,
						Description: "Monthly expense",
						TxnDate:     time.Now().AddDate(0, 0, -5),
					},
				}
				db.Create(&userTransactions)
			}
		}
	}

	// Create demo budget
	budget := models.Budget{
		UserID:      demoUser.ID,
		PeriodStart: time.Now().AddDate(0, 0, -15),
		PeriodEnd:   time.Now().AddDate(0, 0, 15),
		Currency:    "USD",
	}
	db.Create(&budget)

	// Create budget items
	budgetItems := []models.BudgetItem{
		{BudgetID: budget.ID, CategoryID: categories[0].ID, PlannedCents: 40000},  // Groceries: $400
		{BudgetID: budget.ID, CategoryID: categories[2].ID, PlannedCents: 150000}, // Rent: $1500
		{BudgetID: budget.ID, CategoryID: categories[3].ID, PlannedCents: 20000},  // Transportation: $200
		{BudgetID: budget.ID, CategoryID: categories[4].ID, PlannedCents: 15000},  // Entertainment: $150
	}
	db.Create(&budgetItems)

	log.Println("✅ Demo data seeded successfully!")
	log.Println("📧 Demo login credentials:")
	log.Println("   Username: demo")
	log.Println("   Password: demo123")
}

// createDefaultAdmin creates the default admin user if it doesn't exist
func createDefaultAdmin(db *gorm.DB) {
	var adminCount int64
	db.Model(&models.User{}).Where("role = ?", models.UserRoleAdmin).Count(&adminCount)

	if adminCount > 0 {
		log.Println("ℹ️  Admin user already exists, skipping creation")
		return
	}

	log.Println("🔑 Creating default admin user...")

	// Use the proper HashPassword function from controllers
	hash, err := controllers.HashPassword("admin123")
	if err != nil {
		log.Printf("❌ Failed to hash admin password: %v", err)
		return
	}

	admin := models.User{
		Username:     "admin",
		Email:        "admin@financetracker.com",
		PasswordHash: hash,
		Name:         "System Administrator",
		Role:         models.UserRoleAdmin,
	}

	if err := db.Create(&admin).Error; err != nil {
		log.Printf("❌ Failed to create admin user: %v", err)
		return
	}

	log.Println("✅ Default admin user created successfully!")
	log.Println("👑 Admin login credentials:")
	log.Println("   Username: admin")
	log.Println("   Password: admin123")
	log.Println("   ⚠️  Please change the admin password after first login!")
}
