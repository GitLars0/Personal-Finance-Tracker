package controllers

import (
    "net/http"
    "strconv"
    "time"

    "Personal-Finance-Tracker-backend/db"
    "Personal-Finance-Tracker-backend/models"
    "github.com/gin-gonic/gin"
    jwt "github.com/golang-jwt/jwt/v5"
)

// GetSpendSummary provides spending breakdown by category
func GetSpendSummary(c *gin.Context) {
    claims, exists := c.Get("user")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }

    userID := uint(claims.(jwt.MapClaims)["sub"].(float64))

    from := c.Query("from")
    to := c.Query("to")
    
    var fromDate, toDate time.Time
    var err error

    if from == "" {
        now := time.Now()
        fromDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
    } else {
        fromDate, err = time.Parse("2006-01-02", from)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from date format, use YYYY-MM-DD"})
            return
        }
    }

    if to == "" {
        now := time.Now()
        toDate = time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 0, now.Location())
    } else {
        toDate, err = time.Parse("2006-01-02", to)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to date format, use YYYY-MM-DD"})
            return
        }
    }

    type CategorySpend struct {
        CategoryID   uint   `json:"category_id"`
        CategoryName string `json:"category_name"`
        CategoryKind string `json:"category_kind"`
        TotalCents   int64  `json:"total_cents"`
        Count        int64  `json:"transaction_count"`
    }

    var categorySpends []CategorySpend

    db.DB.Table("transactions").
        Select("categories.id as category_id, categories.name as category_name, categories.kind as category_kind, SUM(ABS(transactions.amount_cents)) as total_cents, COUNT(*) as count").
        Joins("JOIN categories ON categories.id = transactions.category_id").
        Where("transactions.user_id = ? AND transactions.txn_date >= ? AND transactions.txn_date <= ? AND transactions.amount_cents < 0", userID, fromDate, toDate).
        Group("categories.id, categories.name, categories.kind").
        Order("total_cents DESC").
        Scan(&categorySpends)

    var splitSpends []CategorySpend

    db.DB.Table("transaction_splits").
        Select("categories.id as category_id, categories.name as category_name, categories.kind as category_kind, SUM(ABS(transaction_splits.amount_cents)) as total_cents, COUNT(*) as count").
        Joins("JOIN categories ON categories.id = transaction_splits.category_id").
        Joins("JOIN transactions ON transactions.id = transaction_splits.parent_txn_id").
        Where("transactions.user_id = ? AND transactions.txn_date >= ? AND transactions.txn_date <= ?", userID, fromDate, toDate).
        Group("categories.id, categories.name, categories.kind").
        Scan(&splitSpends)

    categoryMap := make(map[uint]*CategorySpend)
    for i := range categorySpends {
        categoryMap[categorySpends[i].CategoryID] = &categorySpends[i]
    }

    for _, split := range splitSpends {
        if existing, exists := categoryMap[split.CategoryID]; exists {
            existing.TotalCents += split.TotalCents
            existing.Count += split.Count
        } else {
            categorySpends = append(categorySpends, split)
            categoryMap[split.CategoryID] = &split
        }
    }

    var totalSpending int64
    for _, spend := range categorySpends {
        totalSpending += spend.TotalCents
    }

    type CategorySpendWithPercent struct {
        CategorySpend
        Percentage float64 `json:"percentage"`
    }

    var result []CategorySpendWithPercent
    for _, spend := range categorySpends {
        percentage := 0.0
        if totalSpending > 0 {
            percentage = (float64(spend.TotalCents) / float64(totalSpending)) * 100
        }
        result = append(result, CategorySpendWithPercent{
            CategorySpend: spend,
            Percentage:    percentage,
        })
    }

    c.JSON(http.StatusOK, gin.H{
        "period": gin.H{
            "from": fromDate.Format("2006-01-02"),
            "to":   toDate.Format("2006-01-02"),
        },
        "total_spent_cents": totalSpending,
        "categories":        result,
    })
}

// GetCashflow provides income vs expenses over time
func GetCashflow(c *gin.Context) {
    claims, exists := c.Get("user")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }

    userID := uint(claims.(jwt.MapClaims)["sub"].(float64))
    from := c.Query("from")
    to := c.Query("to")
    groupBy := c.DefaultQuery("group_by", "month")

    var fromDate, toDate time.Time
    var err error

    if from == "" {
        fromDate = time.Now().AddDate(0, -12, 0)
    } else {
        fromDate, err = time.Parse("2006-01-02", from)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from date format, use YYYY-MM-DD"})
            return
        }
    }

    if to == "" {
        toDate = time.Now()
    } else {
        toDate, err = time.Parse("2006-01-02", to)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to date format, use YYYY-MM-DD"})
            return
        }
    }

    // ✅ Detect database driver
    dbDriver := db.DB.Dialector.Name()
    var dateFormat string

    if dbDriver == "sqlite" {
        // SQLite syntax
        switch groupBy {
        case "day":
            dateFormat = "DATE(txn_date)"
        case "week":
            dateFormat = "DATE(txn_date, 'weekday 0', '-6 days')"
        case "year":
            dateFormat = "STRFTIME('%Y', txn_date)"
        default: // month
            dateFormat = "STRFTIME('%Y-%m', txn_date)"
        }
    } else {
        // PostgreSQL syntax
        switch groupBy {
        case "day":
            dateFormat = "DATE(txn_date)"
        case "week":
            dateFormat = "TO_CHAR(DATE_TRUNC('week', txn_date), 'YYYY-MM-DD')"
        case "year":
            dateFormat = "TO_CHAR(DATE_TRUNC('year', txn_date), 'YYYY')"
        default: // month
            dateFormat = "TO_CHAR(DATE_TRUNC('month', txn_date), 'YYYY-MM')"
        }
    }

    type CashflowPeriod struct {
        Period       string `json:"period"`
        IncomeCents  int64  `json:"income_cents"`
        ExpenseCents int64  `json:"expense_cents"`
        NetCents     int64  `json:"net_cents"`
    }

    var periods []CashflowPeriod

    db.DB.Raw(`
        SELECT 
            `+dateFormat+` as period,
            COALESCE(SUM(CASE WHEN amount_cents > 0 THEN amount_cents ELSE 0 END), 0) as income_cents,
            COALESCE(SUM(CASE WHEN amount_cents < 0 THEN ABS(amount_cents) ELSE 0 END), 0) as expense_cents,
            COALESCE(SUM(amount_cents), 0) as net_cents
        FROM transactions
        WHERE user_id = ? AND txn_date >= ? AND txn_date <= ?
        GROUP BY period
        ORDER BY period ASC
    `, userID, fromDate, toDate).Scan(&periods)

    type CashflowWithBalance struct {
        Period              string `json:"period"`
        IncomeCents         int64  `json:"income_cents"`
        ExpenseCents        int64  `json:"expense_cents"`
        NetCents            int64  `json:"net_cents"`
        RunningBalanceCents int64  `json:"running_balance_cents"`
    }

    var result []CashflowWithBalance
    var runningBalance int64

    for _, period := range periods {
        runningBalance += period.NetCents
        result = append(result, CashflowWithBalance{
            Period:              period.Period,
            IncomeCents:         period.IncomeCents,
            ExpenseCents:        period.ExpenseCents,
            NetCents:            period.NetCents,
            RunningBalanceCents: runningBalance,
        })
    }

    var totalIncome, totalExpenses int64
    for _, period := range periods {
        totalIncome += period.IncomeCents
        totalExpenses += period.ExpenseCents
    }

    c.JSON(http.StatusOK, gin.H{
        "period": gin.H{
            "from":     fromDate.Format("2006-01-02"),
            "to":       toDate.Format("2006-01-02"),
            "group_by": groupBy,
        },
        "summary": gin.H{
            "total_income_cents":  totalIncome,
            "total_expense_cents": totalExpenses,
            "net_cents":           totalIncome - totalExpenses,
        },
        "periods": result,
    })
}

// GetAccountBalances provides current balance for each account
func GetAccountBalances(c *gin.Context) {
    claims, exists := c.Get("user")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }

    userID := uint(claims.(jwt.MapClaims)["sub"].(float64))

    type AccountBalance struct {
        AccountID        uint   `json:"account_id"`
        AccountName      string `json:"account_name"`
        AccountType      string `json:"account_type"`
        BalanceCents     int64  `json:"balance_cents"`
        TransactionCount int64  `json:"transaction_count"`
    }

    var balances []AccountBalance

    db.DB.Table("accounts").
        Select(`
            accounts.id as account_id, 
            accounts.name as account_name, 
            accounts.type as account_type, 
            accounts.current_balance_cents as balance_cents,
            COUNT(transactions.id) as transaction_count
        `).
        Joins("LEFT JOIN transactions ON transactions.account_id = accounts.id").
        Where("accounts.user_id = ?", userID).
        Group("accounts.id, accounts.name, accounts.type, accounts.current_balance_cents").
        Order("account_type, account_name").
        Scan(&balances)

    var totalBalance int64
    for _, balance := range balances {
        totalBalance += balance.BalanceCents
    }

    c.JSON(http.StatusOK, gin.H{
        "accounts":            balances,
        "total_balance_cents": totalBalance,
    })
}

// GetBudgetProgress shows budget vs actual spending
func GetBudgetProgress(c *gin.Context) {
    claims, exists := c.Get("user")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }

    userID := uint(claims.(jwt.MapClaims)["sub"].(float64))
    budgetIDStr := c.Query("budget_id")
    
    var budget models.Budget

    if budgetIDStr != "" {
        if err := db.DB.
            Preload("Items.Category").
            Where("id = ? AND user_id = ?", budgetIDStr, userID).
            First(&budget).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "budget not found"})
            return
        }
    } else {
        now := time.Now()
        if err := db.DB.
            Preload("Items.Category").
            Where("user_id = ? AND period_start <= ? AND period_end >= ?", userID, now, now).
            First(&budget).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "no active budget found"})
            return
        }
    }

    type CategoryProgress struct {
        CategoryID     uint    `json:"category_id"`
        CategoryName   string  `json:"category_name"`
        PlannedCents   int64   `json:"planned_cents"`
        SpentCents     int64   `json:"spent_cents"`
        RemainingCents int64   `json:"remaining_cents"`
        Progress       float64 `json:"progress_percent"`
        Status         string  `json:"status"`
    }

    var categoryProgress []CategoryProgress
    var totalPlanned, totalSpent int64

    for _, item := range budget.Items {
        var spentCents int64
        db.DB.Model(&models.Transaction{}).
            Where("user_id = ? AND category_id = ? AND txn_date >= ? AND txn_date <= ? AND amount_cents < 0",
                userID, item.CategoryID, budget.PeriodStart, budget.PeriodEnd).
            Select("COALESCE(SUM(ABS(amount_cents)), 0)").
            Scan(&spentCents)

        var splitSpent int64
        db.DB.Table("transaction_splits").
            Joins("JOIN transactions ON transactions.id = transaction_splits.parent_txn_id").
            Where("transactions.user_id = ? AND transaction_splits.category_id = ? AND transactions.txn_date >= ? AND transactions.txn_date <= ?",
                userID, item.CategoryID, budget.PeriodStart, budget.PeriodEnd).
            Select("COALESCE(SUM(ABS(transaction_splits.amount_cents)), 0)").
            Scan(&splitSpent)

        spentCents += splitSpent

        remaining := item.PlannedCents - spentCents
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

        categoryProgress = append(categoryProgress, CategoryProgress{
            CategoryID:     item.CategoryID,
            CategoryName:   item.Category.Name,
            PlannedCents:   item.PlannedCents,
            SpentCents:     spentCents,
            RemainingCents: remaining,
            Progress:       progress,
            Status:         status,
        })

        totalPlanned += item.PlannedCents
        totalSpent += spentCents
    }

    now := time.Now()
    daysRemaining := int(budget.PeriodEnd.Sub(now).Hours() / 24)
    if daysRemaining < 0 {
        daysRemaining = 0
    }

    c.JSON(http.StatusOK, gin.H{
        "budget": gin.H{
            "id":             budget.ID,
            "period_start":   budget.PeriodStart.Format("2006-01-02"),
            "period_end":     budget.PeriodEnd.Format("2006-01-02"),
            "days_remaining": daysRemaining,
        },
        "summary": gin.H{
            "total_planned_cents":   totalPlanned,
            "total_spent_cents":     totalSpent,
            "total_remaining_cents": totalPlanned - totalSpent,
            "overall_progress":      (float64(totalSpent) / float64(totalPlanned)) * 100,
        },
        "categories": categoryProgress,
    })
}

// GetMonthlyTrends shows spending trends over multiple months
func GetMonthlyTrends(c *gin.Context) {
    claims, exists := c.Get("user")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }

    userID := uint(claims.(jwt.MapClaims)["sub"].(float64))
    monthsStr := c.DefaultQuery("months", "12")
    months := 12
    if m, _ := strconv.Atoi(monthsStr); m > 0 {
        months = m
    }

    type MonthlyData struct {
        Month        string  `json:"month"`
        IncomeCents  int64   `json:"income_cents"`
        ExpenseCents int64   `json:"expense_cents"`
        NetCents     int64   `json:"net_cents"`
        SavingsRate  float64 `json:"savings_rate_percent"`
    }

    var trends []MonthlyData

    // ✅ Detect database driver
    dbDriver := db.DB.Dialector.Name()
    cutoffDate := time.Now().AddDate(0, -months, 0)

    if dbDriver == "sqlite" {
        db.DB.Raw(`
            SELECT 
                STRFTIME('%Y-%m', txn_date) as month,
                COALESCE(SUM(CASE WHEN amount_cents > 0 THEN amount_cents ELSE 0 END), 0) as income_cents,
                COALESCE(SUM(CASE WHEN amount_cents < 0 THEN ABS(amount_cents) ELSE 0 END), 0) as expense_cents,
                COALESCE(SUM(amount_cents), 0) as net_cents
            FROM transactions
            WHERE user_id = ? AND txn_date >= ?
            GROUP BY STRFTIME('%Y-%m', txn_date)
            ORDER BY month ASC
        `, userID, cutoffDate.Format("2006-01-02")).Scan(&trends)
    } else {
        db.DB.Raw(`
            SELECT 
                TO_CHAR(DATE_TRUNC('month', txn_date), 'YYYY-MM') as month,
                COALESCE(SUM(CASE WHEN amount_cents > 0 THEN amount_cents ELSE 0 END), 0) as income_cents,
                COALESCE(SUM(CASE WHEN amount_cents < 0 THEN ABS(amount_cents) ELSE 0 END), 0) as expense_cents,
                COALESCE(SUM(amount_cents), 0) as net_cents
            FROM transactions
            WHERE user_id = ? AND txn_date >= ?
            GROUP BY DATE_TRUNC('month', txn_date)
            ORDER BY month ASC
        `, userID, cutoffDate).Scan(&trends)
    }

    for i := range trends {
        if trends[i].IncomeCents > 0 {
            trends[i].SavingsRate = (float64(trends[i].NetCents) / float64(trends[i].IncomeCents)) * 100
        }
    }

    c.JSON(http.StatusOK, gin.H{
        "months": months,
        "trends": trends,
    })
}

// GetTopMerchants shows most frequent transaction descriptions
func GetTopMerchants(c *gin.Context) {
    claims, exists := c.Get("user")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }

    userID := uint(claims.(jwt.MapClaims)["sub"].(float64))
    limit := c.DefaultQuery("limit", "10")

    type MerchantSpend struct {
        Description string `json:"description"`
        TotalCents  int64  `json:"total_cents"`
        Count       int64  `json:"transaction_count"`
        AvgCents    int64  `json:"avg_cents"`
    }

    var merchants []MerchantSpend

    db.DB.Raw(`
        SELECT 
            description,
            SUM(ABS(amount_cents)) as total_cents,
            COUNT(*) as count,
            AVG(ABS(amount_cents)) as avg_cents
        FROM transactions
        WHERE user_id = ? AND amount_cents < 0 AND description != ''
        GROUP BY description
        ORDER BY total_cents DESC
        LIMIT ?
    `, userID, limit).Scan(&merchants)

    c.JSON(http.StatusOK, gin.H{
        "top_merchants": merchants,
    })
}