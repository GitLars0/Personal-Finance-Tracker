package controllers_test

import (
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

func TestGetSpendSummary(t *testing.T) {
	database := SetupTestDB()
    db.DB = database
    user := CreateTestUser(database)
    token := GetTestToken(user.ID, user.Username)

	// Create test data
    account := models.Account{
        UserID:               user.ID,
        Name:                 "Checking",
        Type:                 "checking",
        InitialBalanceCents:  100000,
        CurrentBalanceCents:  100000,
    }
    database.Create(&account)

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

    // Create transactions in current month
    now := time.Now()
    database.Create(&models.Transaction{
        UserID:      user.ID,
        AccountID:   account.ID,
        CategoryID:  &groceries.ID,
        AmountCents: -5000, // $50 expense
        Description: "Whole Foods",
        TxnDate:     now,
    })
    database.Create(&models.Transaction{
        UserID:      user.ID,
        AccountID:   account.ID,
        CategoryID:  &groceries.ID,
        AmountCents: -3000, // $30 expense
        Description: "Safeway",
        TxnDate:     now,
    })
    database.Create(&models.Transaction{
        UserID:      user.ID,
        AccountID:   account.ID,
        CategoryID:  &transport.ID,
        AmountCents: -2000, // $20 expense
        Description: "Gas",
        TxnDate:     now,
    })

	router := SetupRouter()
	router.GET("/api/reports/spend-summary",controllers.AuthMiddleware(), controllers.GetSpendSummary)

	req, _ := http.NewRequest("GET", "/api/reports/spend-summary", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Expected status OK")

	var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)

	// Verify total spending
	assert.Equal(t, float64(10000), response["total_spent_cents"],
        "Total spent: $50 + $30 + $20 = $100")

	// Verify categories
	categories := response["categories"].([]interface{})
    assert.Equal(t, 2, len(categories), "Should have 2 categories")

    // First category should be Groceries (highest spending)
    firstCat := categories[0].(map[string]interface{})
    assert.Equal(t, "Groceries", firstCat["category_name"], "First category should be Groceries")
    assert.Equal(t, float64(8000), firstCat["total_cents"], "Groceries: $50 + $30 = $80")
    assert.Equal(t, float64(2), firstCat["transaction_count"], "2 transactions in Groceries")
    assert.Equal(t, float64(80), firstCat["percentage"], "80% of total")

    // Second category should be Transportation
    secondCat := categories[1].(map[string]interface{})
    assert.Equal(t, "Transportation", secondCat["category_name"], "Second category should be Transportation")
    assert.Equal(t, float64(2000), secondCat["total_cents"], "Transportation: $20 = $20")
    assert.Equal(t, float64(20), secondCat["percentage"], "20% of total")
}

func TestGetSpendSummary_DateRange(t *testing.T) {
	database := SetupTestDB()
	db.DB = database
	user := CreateTestUser(database)
	token := GetTestToken(user.ID, user.Username)

	account := models.Account{
        UserID:               user.ID,
        Name:                 "Checking",
        Type:                 "checking",
        InitialBalanceCents:  100000,
        CurrentBalanceCents:  100000,
    }
    database.Create(&account)

	category := models.Category{
        UserID: user.ID,
        Name:   "Food",
        Kind:   models.CategoryExpense,
    }
    database.Create(&category)

	// Create transactions in January 2025
	jan1 := time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC)
    jan2 := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
    feb1 := time.Date(2025, 2, 5, 0, 0, 0, 0, time.UTC)

	database.Create(&models.Transaction{
        UserID:      user.ID,
        AccountID:   account.ID,
        CategoryID:  &category.ID,
        AmountCents: -3000,
        TxnDate:     jan1,
    })
    database.Create(&models.Transaction{
        UserID:      user.ID,
        AccountID:   account.ID,
        CategoryID:  &category.ID,
        AmountCents: -2000,
        TxnDate:     jan2,
    })
    database.Create(&models.Transaction{
        UserID:      user.ID,
        AccountID:   account.ID,
        CategoryID:  &category.ID,
        AmountCents: -5000, // Should be excluded (February)
        TxnDate:     feb1,
    })

	router := SetupRouter()
    router.GET("/api/reports/spend-summary", controllers.AuthMiddleware(), controllers.GetSpendSummary)

	// Query for January only
	req, _ := http.NewRequest("GET", "/api/reports/spend-summary?from=2025-01-01&to=2025-01-31", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Expected status OK")

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// Should only include January transactions
	assert.Equal(t, float64(5000), response["total_spent_cents"], "Should only count January: $30 + $20 = $50")
}

func TestGetSpendSummary_WithSplits(t *testing.T) {
	database := SetupTestDB()
	db.DB = database
	user := CreateTestUser(database)
	token := GetTestToken(user.ID, user.Username)

	account := models.Account{
        UserID:               user.ID,
        Name:                 "Checking",
        Type:                 "checking",
        InitialBalanceCents:  100000,
        CurrentBalanceCents:  100000,
    }
    database.Create(&account)

    groceries := models.Category{
        UserID: user.ID,
        Name:   "Groceries",
        Kind:   models.CategoryExpense,
    }
    household := models.Category{
        UserID: user.ID,
        Name:   "Household",
        Kind:   models.CategoryExpense,
    }
    database.Create(&groceries)
    database.Create(&household)

    // Create split transaction
    splitTxn := models.Transaction{
        UserID:      user.ID,
        AccountID:   account.ID,
        AmountCents: -10000, // $100 total
        Description: "Costco",
        TxnDate:     time.Now(),
    }
    database.Create(&splitTxn)

    // Split: $60 groceries, $40 household
    database.Create(&models.TransactionSplit{
        ParentTxnID: splitTxn.ID,
        CategoryID:  groceries.ID,
        AmountCents: -6000,
    })
    database.Create(&models.TransactionSplit{
        ParentTxnID: splitTxn.ID,
        CategoryID:  household.ID,
        AmountCents: -4000,
    })

	router := SetupRouter()
	router.GET("/api/reports/spend-summary", controllers.AuthMiddleware(), controllers.GetSpendSummary)

	req, _ := http.NewRequest("GET", "/api/reports/spend-summary", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Expected status OK")

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// Should count split amounts
	assert.Equal(t, float64(10000), response["total_spent_cents"], "Total spent should be $100 from split")

	categories := response["categories"].([]interface{})
	assert.Equal(t, 2, len(categories), "Should have 2 categories from split")

    // Find groceries category
	var groceriesCat map[string]interface{}
    for _, cat := range categories {
        c := cat.(map[string]interface{})
        if c["category_name"] == "Groceries" {
            groceriesCat = c
            break
        }
    }
    assert.NotNil(t, groceriesCat)
    assert.Equal(t, float64(6000), groceriesCat["total_cents"], "Groceries from split: $60")
}

func TestGetCashflow(t *testing.T) {
	database := SetupTestDB()
    db.DB = database
    user := CreateTestUser(database)
    token := GetTestToken(user.ID, user.Username)

    account := models.Account{
        UserID:               user.ID,
        Name:                 "Checking",
        Type:                 "checking",
        InitialBalanceCents:  100000,
        CurrentBalanceCents:  100000,
    }
    database.Create(&account)

	// Create transactions across multiple months
    jan := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
    feb := time.Date(2025, 2, 15, 0, 0, 0, 0, time.UTC)

    // January: Income $1000, Expense $600
    database.Create(&models.Transaction{
        UserID:      user.ID,
        AccountID:   account.ID,
        AmountCents: 100000, // Income
        Description: "Salary",
        TxnDate:     jan,
    })
    database.Create(&models.Transaction{
        UserID:      user.ID,
        AccountID:   account.ID,
        AmountCents: -60000, // Expense
        Description: "Rent",
        TxnDate:     jan,
    })

    // February: Income $1000, Expense $500
    database.Create(&models.Transaction{
        UserID:      user.ID,
        AccountID:   account.ID,
        AmountCents: 100000,
        Description: "Salary",
        TxnDate:     feb,
    })
    database.Create(&models.Transaction{
        UserID:      user.ID,
        AccountID:   account.ID,
        AmountCents: -50000,
        Description: "Bills",
        TxnDate:     feb,
    })

	router := SetupRouter()
	router.GET("/api/reports/cashflow", controllers.AuthMiddleware(), controllers.GetCashflow)

	req, _ := http.NewRequest("GET", "/api/reports/cashflow?from=2025-01-01&to=2025-02-28&group_by=month", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)


	assert.Equal(t, http.StatusOK, w.Code, "Expected status OK")

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

    // Verify summary
	summary := response["summary"].(map[string]interface{})
    assert.Equal(t, float64(200000), summary["total_income_cents"], "Total income: $2000")
    assert.Equal(t, float64(110000), summary["total_expense_cents"], "Total expenses: $1100")
    assert.Equal(t, float64(90000), summary["net_cents"], "Net: $900")

    // Verify periods
    periods := response["periods"].([]interface{})
    assert.Equal(t, 2, len(periods), "Should have 2 months")

	// January
	janPeriod := periods[0].(map[string]interface{})
	assert.Equal(t, "2025-01", janPeriod["period"], "First period should be January 2025")
	assert.Equal(t, float64(100000), janPeriod["income_cents"], "January income: $1000")
	assert.Equal(t, float64(60000), janPeriod["expense_cents"], "January expenses: $600")
	assert.Equal(t, float64(40000), janPeriod["net_cents"], "January net: $400")
	assert.Equal(t, float64(40000), janPeriod["running_balance_cents"], "January running balance: $400")

	// February
	febPeriod := periods[1].(map[string]interface{})
	assert.Equal(t, "2025-02", febPeriod["period"], "Second period should be February 2025")
	assert.Equal(t, float64(100000), febPeriod["income_cents"], "February income: $1000")
	assert.Equal(t, float64(50000), febPeriod["expense_cents"], "February expenses: $500")
	assert.Equal(t, float64(50000), febPeriod["net_cents"], "February net: $500")
	assert.Equal(t, float64(90000), febPeriod["running_balance_cents"], "February running balance: $400 + $500 = $900")
}

func TestGetAccountBalances(t *testing.T) {
    database := SetupTestDB()
    db.DB = database
    user := CreateTestUser(database)
    token := GetTestToken(user.ID, user.Username)

    // Create accounts with different balances
    checking := models.Account{
        UserID:               user.ID,
        Name:                 "Checking",
        Type:                 "checking",
        InitialBalanceCents:  50000,
        CurrentBalanceCents:  60000, // $600
    }
    savings := models.Account{
        UserID:               user.ID,
        Name:                 "Savings",
        Type:                 "savings",
        InitialBalanceCents:  100000,
        CurrentBalanceCents:  105000, // $1050
    }
    database.Create(&checking)
    database.Create(&savings)

    // Create transactions for checking account
    database.Create(&models.Transaction{
        UserID:      user.ID,
        AccountID:   checking.ID,
        AmountCents: 1000,
        TxnDate:     time.Now(),
    })
    database.Create(&models.Transaction{
        UserID:      user.ID,
        AccountID:   checking.ID,
        AmountCents: -500,
        TxnDate:     time.Now(),
    })

    router := SetupRouter()
    router.GET("/api/reports/account-balances", controllers.AuthMiddleware(), controllers.GetAccountBalances)

    req, _ := http.NewRequest("GET", "/api/reports/account-balances", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)

    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)

    // Verify total balance
    assert.Equal(t, float64(165000), response["total_balance_cents"],
        "Total: $600 + $1050 = $1650")

    // Verify accounts
    accounts := response["accounts"].([]interface{})
    assert.Equal(t, 2, len(accounts))

    // Checking account
    checkingAccount := accounts[0].(map[string]interface{})
    assert.Equal(t, "Checking", checkingAccount["account_name"])
    assert.Equal(t, float64(60000), checkingAccount["balance_cents"])
    assert.Equal(t, float64(2), checkingAccount["transaction_count"])

    // Savings account
    savingsAccount := accounts[1].(map[string]interface{})
    assert.Equal(t, "Savings", savingsAccount["account_name"])
    assert.Equal(t, float64(105000), savingsAccount["balance_cents"])
    assert.Equal(t, float64(0), savingsAccount["transaction_count"])
}

func TestGetBudgetProgress(t *testing.T) {
    database := SetupTestDB()
    db.DB = database
    user := CreateTestUser(database)
    token := GetTestToken(user.ID, user.Username)

    account := models.Account{
        UserID:               user.ID,
        Name:                 "Checking",
        Type:                 "checking",
        InitialBalanceCents:  100000,
        CurrentBalanceCents:  100000,
    }
    database.Create(&account)

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

    // Create budget for current month
    now := time.Now()
    budget := models.Budget{
        UserID:      user.ID,
        PeriodStart: time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
        PeriodEnd:   time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, time.UTC),
        Currency:    "USD",
    }
    database.Create(&budget)

    // Budget items
    database.Create(&models.BudgetItem{
        BudgetID:     budget.ID,
        CategoryID:   groceries.ID,
        PlannedCents: 40000, // Planned: $400
    })
    database.Create(&models.BudgetItem{
        BudgetID:     budget.ID,
        CategoryID:   transport.ID,
        PlannedCents: 20000, // Planned: $200
    })

    // Create transactions
    database.Create(&models.Transaction{
        UserID:      user.ID,
        AccountID:   account.ID,
        CategoryID:  &groceries.ID,
        AmountCents: -15000, // Spent $150
        TxnDate:     now,
    })
    database.Create(&models.Transaction{
        UserID:      user.ID,
        AccountID:   account.ID,
        CategoryID:  &groceries.ID,
        AmountCents: -10000, // Spent $100
        TxnDate:     now,
    })
    database.Create(&models.Transaction{
        UserID:      user.ID,
        AccountID:   account.ID,
        CategoryID:  &transport.ID,
        AmountCents: -5000, // Spent $50
        TxnDate:     now,
    })

    router := SetupRouter()
    router.GET("/api/reports/budget-progress", controllers.AuthMiddleware(), controllers.GetBudgetProgress)

    req, _ := http.NewRequest("GET", "/api/reports/budget-progress?budget_id=1", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)

    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)

    // Verify summary
    summary := response["summary"].(map[string]interface{})
    assert.Equal(t, float64(60000), summary["total_planned_cents"], "Total planned: $600")
    assert.Equal(t, float64(30000), summary["total_spent_cents"], "Total spent: $300")
    assert.Equal(t, float64(30000), summary["total_remaining_cents"], "Remaining: $300")
    assert.Equal(t, float64(50), summary["overall_progress"], "50% spent")

    // Verify categories
    categories := response["categories"].([]interface{})
    assert.Equal(t, 2, len(categories))

    // Groceries
    var groceriesCat map[string]interface{}
    for _, cat := range categories {
        c := cat.(map[string]interface{})
        if c["category_name"] == "Groceries" {
            groceriesCat = c
            break
        }
    }
    assert.NotNil(t, groceriesCat)
    assert.Equal(t, float64(40000), groceriesCat["planned_cents"])
    assert.Equal(t, float64(25000), groceriesCat["spent_cents"], "$150 + $100 = $250")
    assert.Equal(t, float64(15000), groceriesCat["remaining_cents"])
    assert.Equal(t, float64(62.5), groceriesCat["progress_percent"])
    assert.Equal(t, "under_budget", groceriesCat["status"])
}

func TestGetBudgetProgress_OverBudget(t *testing.T) {
    database := SetupTestDB()
    db.DB = database
    user := CreateTestUser(database)
    token := GetTestToken(user.ID, user.Username)

    account := models.Account{
        UserID:               user.ID,
        Name:                 "Checking",
        Type:                 "checking",
        InitialBalanceCents:  100000,
        CurrentBalanceCents:  100000,
    }
    database.Create(&account)

    category := models.Category{
        UserID: user.ID,
        Name:   "Food",
        Kind:   models.CategoryExpense,
    }
    database.Create(&category)

    now := time.Now()
    budget := models.Budget{
        UserID:      user.ID,
        PeriodStart: time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
        PeriodEnd:   time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, time.UTC),
        Currency:    "USD",
    }
    database.Create(&budget)

    database.Create(&models.BudgetItem{
        BudgetID:     budget.ID,
        CategoryID:   category.ID,
        PlannedCents: 30000, // Planned: $300
    })

    // Spend more than budget
    database.Create(&models.Transaction{
        UserID:      user.ID,
        AccountID:   account.ID,
        CategoryID:  &category.ID,
        AmountCents: -35000, // Spent $350 (over budget!)
        TxnDate:     now,
    })

    router := SetupRouter()
    router.GET("/api/reports/budget-progress", controllers.AuthMiddleware(), controllers.GetBudgetProgress)

    req, _ := http.NewRequest("GET", "/api/reports/budget-progress?budget_id=1", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)

    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)

    categories := response["categories"].([]interface{})
    cat := categories[0].(map[string]interface{})

    assert.Equal(t, float64(35000), cat["spent_cents"])
    assert.Equal(t, float64(-5000), cat["remaining_cents"], "Negative remaining")
    assert.InDelta(t, 116.67, cat["progress_percent"].(float64), 0.1, "Over 100%")
    assert.Equal(t, "over_budget", cat["status"])
}

func TestGetTopMerchants(t *testing.T) {
    database := SetupTestDB()
    db.DB = database
    user := CreateTestUser(database)
    token := GetTestToken(user.ID, user.Username)

    account := models.Account{
        UserID:               user.ID,
        Name:                 "Checking",
        Type:                 "checking",
        InitialBalanceCents:  100000,
        CurrentBalanceCents:  100000,
    }
    database.Create(&account)

    // Create transactions with same merchants
    merchants := []struct {
        name   string
        amount int64
    }{
        {"Starbucks", -500},
        {"Starbucks", -600},
        {"Starbucks", -550},
        {"Whole Foods", -8000},
        {"Whole Foods", -7000},
        {"Shell Gas", -4000},
    }

    for _, m := range merchants {
        database.Create(&models.Transaction{
            UserID:      user.ID,
            AccountID:   account.ID,
            AmountCents: m.amount,
            Description: m.name,
            TxnDate:     time.Now(),
        })
    }

    router := SetupRouter()
    router.GET("/api/reports/top-merchants", controllers.AuthMiddleware(), controllers.GetTopMerchants)

    req, _ := http.NewRequest("GET", "/api/reports/top-merchants?limit=3", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)

    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)

    top_merchants := response["top_merchants"].([]interface{})
    assert.Equal(t, 3, len(top_merchants))

    // Top merchant should be Whole Foods
    top := top_merchants[0].(map[string]interface{})
    assert.Equal(t, "Whole Foods", top["description"])
    assert.Equal(t, float64(15000), top["total_cents"], "$80 + $70 = $150")
    assert.Equal(t, float64(2), top["transaction_count"])
    assert.Equal(t, float64(7500), top["avg_cents"], "Average: $75")
}

func TestGetSpendSummary_NoTransactions(t *testing.T) {
    database := SetupTestDB()
    db.DB = database
    user := CreateTestUser(database)
    token := GetTestToken(user.ID, user.Username)

    router := SetupRouter()
    router.GET("/api/reports/spend-summary", controllers.AuthMiddleware(), controllers.GetSpendSummary)

    req, _ := http.NewRequest("GET", "/api/reports/spend-summary", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)

    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)

    assert.Equal(t, float64(0), response["total_spent_cents"])
    
    categories := response["categories"]
    if categories != nil {
        assert.Equal(t, 0, len(categories.([]interface{})))
    }
}