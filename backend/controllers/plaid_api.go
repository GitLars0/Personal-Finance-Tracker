package controllers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"Personal-Finance-Tracker-backend/db"
	"Personal-Finance-Tracker-backend/models"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/plaid/plaid-go/v29/plaid"
)

// PlaidClient manages Plaid API interactions
type PlaidClient struct {
	Client *plaid.APIClient
	Ctx    context.Context
}

var plaidClient *PlaidClient

// InitPlaidClient initializes the Plaid client
func InitPlaidClient(clientID, secret, environment string) error {
	var env plaid.Environment
	switch environment {
	case "sandbox":
		env = plaid.Sandbox
	case "production":
		env = plaid.Production
	default:
		env = plaid.Sandbox
	}

	configuration := plaid.NewConfiguration()
	configuration.AddDefaultHeader("PLAID-CLIENT-ID", clientID)
	configuration.AddDefaultHeader("PLAID-SECRET", secret)
	configuration.UseEnvironment(env)

	client := plaid.NewAPIClient(configuration)
	ctx := context.Background()

	plaidClient = &PlaidClient{
		Client: client,
		Ctx:    ctx,
	}

	return nil
}

// CreateLinkToken creates a Plaid Link token for the frontend
func CreateLinkToken(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}
	userID := uint(claims.(jwt.MapClaims)["sub"].(float64))

	if plaidClient == nil {
		c.JSON(500, gin.H{"error": "Plaid client not initialized"})
		return
	}

	// Create link token request
	user := plaid.LinkTokenCreateRequestUser{
		ClientUserId: fmt.Sprintf("user_%d", userID),
	}

	request := plaid.NewLinkTokenCreateRequest(
		"Personal Finance Tracker",
		"en",
		[]plaid.CountryCode{plaid.COUNTRYCODE_NO, plaid.COUNTRYCODE_GB, plaid.COUNTRYCODE_US},
		user,
	)

	// Set products to use
	request.SetProducts([]plaid.Products{
		plaid.PRODUCTS_AUTH,
		plaid.PRODUCTS_TRANSACTIONS,
	})

	// Optional: Set redirect URI for OAuth (only if needed)
	// request.SetRedirectUri("http://localhost:8080/banks")

	// Set webhook URL (optional)
	// request.SetWebhook("http://localhost:8080/api/plaid/webhook")

	// Create the link token
	resp, httpResp, err := plaidClient.Client.PlaidApi.LinkTokenCreate(plaidClient.Ctx).LinkTokenCreateRequest(*request).Execute()
	if err != nil {
		errMsg := fmt.Sprintf("Failed to create link token: %v", err)
		if httpResp != nil {
			errMsg = fmt.Sprintf("Plaid API error (status %d): %v", httpResp.StatusCode, err)
		}
		fmt.Printf("❌ Plaid Error: %s\n", errMsg)
		c.JSON(500, gin.H{"error": errMsg})
		return
	}

	c.JSON(200, gin.H{
		"link_token": resp.GetLinkToken(),
		"expiration": resp.GetExpiration(),
		"request_id": resp.GetRequestId(),
	})
}

// ExchangePublicToken exchanges a public token for an access token
func ExchangePublicToken(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}
	userID := uint(claims.(jwt.MapClaims)["sub"].(float64))

	var req struct {
		PublicToken string `json:"public_token" binding:"required"`
		BankName    string `json:"bank_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	if plaidClient == nil {
		c.JSON(500, gin.H{"error": "Plaid client not initialized"})
		return
	}

	// Exchange public token for access token
	exchangeRequest := plaid.NewItemPublicTokenExchangeRequest(req.PublicToken)
	exchangeResp, _, err := plaidClient.Client.PlaidApi.ItemPublicTokenExchange(plaidClient.Ctx).ItemPublicTokenExchangeRequest(*exchangeRequest).Execute()
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to exchange token: " + err.Error()})
		return
	}

	accessToken := exchangeResp.GetAccessToken()
	itemID := exchangeResp.GetItemId()

	// Get institution info
	institutionName := req.BankName
	if institutionName == "" {
		institutionName = "Plaid Bank"
	}

	// Create bank connection
	connection := models.BankConnection{
		UserID:            userID,
		BankName:          institutionName,
		BankEndpoint:      "plaid://api",
		Status:            "connected",
		ConsentID:         itemID,
		ConsentValidUntil: time.Now().Add(90 * 24 * time.Hour),
		Metadata: models.JSONB{
			"access_token": accessToken,
			"item_id":      itemID,
		},
	}

	if err := db.DB.Create(&connection).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to save bank connection: " + err.Error()})
		return
	}

	// Fetch accounts
	accountsRequest := plaid.NewAccountsGetRequest(accessToken)
	accountsResp, _, err := plaidClient.Client.PlaidApi.AccountsGet(plaidClient.Ctx).AccountsGetRequest(*accountsRequest).Execute()
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch accounts: " + err.Error()})
		return
	}

	// Save accounts
	for _, acc := range accountsResp.GetAccounts() {
		balances := acc.GetBalances()

		// Create BankAccount
		bankAccount := models.BankAccount{
			BankConnectionID: connection.ID,
			AccountID:        acc.GetAccountId(),
			AccountName:      acc.GetName(),
			Currency:         balances.GetIsoCurrencyCode(),
			AccountType:      string(acc.GetSubtype()),
			IsActive:         true,
		}

		if err := db.DB.Create(&bankAccount).Error; err != nil {
			continue // Skip if error
		}

		// Create internal Account
		balanceCents := int64(balances.GetCurrent() * 100)
		account := models.Account{
			UserID:              userID,
			Name:                acc.GetName(),
			Type:                models.AccountChecking,
			Currency:            balances.GetIsoCurrencyCode(),
			InitialBalanceCents: balanceCents,
			CurrentBalanceCents: balanceCents,
		}

		if err := db.DB.Create(&account).Error; err != nil {
			continue
		}

		// Link accounts
		bankAccount.InternalAccountID = &account.ID
		db.DB.Save(&bankAccount)
	}

	c.JSON(200, gin.H{
		"success":       true,
		"message":       "Bank connected successfully via Plaid",
		"connection_id": connection.ID,
	})
}

// SyncPlaidTransactions syncs transactions from Plaid
func SyncPlaidTransactions(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}
	userID := uint(claims.(jwt.MapClaims)["sub"].(float64))

	connectionID := c.Param("id")

	// Get connection
	var connection models.BankConnection
	if err := db.DB.Where("id = ? AND user_id = ?", connectionID, userID).First(&connection).Error; err != nil {
		c.JSON(404, gin.H{"error": "Connection not found"})
		return
	}

	// Get access token from metadata
	accessToken, ok := connection.Metadata["access_token"].(string)
	if !ok {
		c.JSON(400, gin.H{"error": "Access token not found"})
		return
	}

	// Sync transactions from last 30 days
	startDate := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	endDate := time.Now().Format("2006-01-02")

	transactionsRequest := plaid.NewTransactionsGetRequest(accessToken, startDate, endDate)
	transactionsResp, _, err := plaidClient.Client.PlaidApi.TransactionsGet(plaidClient.Ctx).TransactionsGetRequest(*transactionsRequest).Execute()
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to sync transactions: " + err.Error()})
		return
	}

	// Get bank accounts for this connection
	var bankAccounts []models.BankAccount
	db.DB.Where("bank_connection_id = ?", connection.ID).Find(&bankAccounts)

	accountMap := make(map[string]uint)
	for _, ba := range bankAccounts {
		if ba.InternalAccountID != nil {
			accountMap[ba.AccountID] = *ba.InternalAccountID
		}
	}

	// Load user's categories for auto-categorization
	var userCategories []models.Category
	db.DB.Where("user_id = ?", userID).Find(&userCategories)

	fmt.Printf("🔍 Found %d categories for user %d\n", len(userCategories), userID)
	for _, cat := range userCategories {
		fmt.Printf("  - Category: %s (ID: %d, Kind: %s)\n", cat.Name, cat.ID, cat.Kind)
	}

	categoryMap := buildCategoryMap(userCategories)

	transactionsAdded := 0
	categorizedCount := 0
	for _, txn := range transactionsResp.GetTransactions() {
		accountID, ok := accountMap[txn.GetAccountId()]
		if !ok {
			continue // Skip if account not found
		}

		// Check if transaction already exists
		txnID := txn.GetTransactionId()
		var existing models.Transaction
		if err := db.DB.Where("bank_transaction_id = ?", txnID).First(&existing).Error; err == nil {
			continue // Skip if already exists
		}

		// Create transaction
		amountCents := int64(-txn.GetAmount() * 100) // Plaid uses positive for expenses
		txnDate, _ := time.Parse("2006-01-02", txn.GetDate())

		// Auto-categorize based on Plaid's category
		var categoryID *uint
		plaidCategories := txn.GetCategory()
		merchantName := txn.GetName()

		fmt.Printf("📦 Transaction: %s | Amount: %.2f | Plaid Categories: %v\n",
			merchantName, txn.GetAmount(), plaidCategories)

		if len(plaidCategories) > 0 {
			categoryID = matchPlaidCategory(plaidCategories, categoryMap, amountCents < 0)
		}

		// Fallback: Try keyword-based matching if no Plaid category
		if categoryID == nil {
			categoryID = matchByMerchantName(merchantName, categoryMap, amountCents < 0)
		}

		if categoryID != nil {
			categorizedCount++
			fmt.Printf("  ✅ Matched to category ID: %d\n", *categoryID)
		} else {
			fmt.Printf("  ❌ No category match found\n")
		}

		transaction := models.Transaction{
			UserID:            userID,
			AccountID:         accountID,
			CategoryID:        categoryID,
			AmountCents:       amountCents,
			Description:       txn.GetName(),
			TxnDate:           txnDate,
			BankTransactionID: &txnID,
		}

		if err := db.DB.Create(&transaction).Error; err == nil {
			transactionsAdded++
		}
	}

	fmt.Printf("📊 Sync Summary: %d transactions added, %d categorized\n", transactionsAdded, categorizedCount)

	c.JSON(200, gin.H{
		"success":             true,
		"transactions_synced": transactionsAdded,
	})
}

// GetPlaidAccounts retrieves account balances from Plaid
func GetPlaidAccounts(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}
	userID := uint(claims.(jwt.MapClaims)["sub"].(float64))

	connectionID := c.Param("id")

	// Get connection
	var connection models.BankConnection
	if err := db.DB.Where("id = ? AND user_id = ?", connectionID, userID).First(&connection).Error; err != nil {
		c.JSON(404, gin.H{"error": "Connection not found"})
		return
	}

	// Get access token from metadata
	accessToken, ok := connection.Metadata["access_token"].(string)
	if !ok {
		c.JSON(400, gin.H{"error": "Access token not found"})
		return
	}

	// Get accounts
	accountsRequest := plaid.NewAccountsGetRequest(accessToken)
	accountsResp, _, err := plaidClient.Client.PlaidApi.AccountsGet(plaidClient.Ctx).AccountsGetRequest(*accountsRequest).Execute()
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch accounts: " + err.Error()})
		return
	}

	c.JSON(200, accountsResp)
}

// buildCategoryMap creates a map of category names (lowercase) to category IDs
func buildCategoryMap(categories []models.Category) map[string]uint {
	categoryMap := make(map[string]uint)
	for _, cat := range categories {
		// Store by exact name and also lowercase version
		categoryMap[cat.Name] = cat.ID
		// Also store lowercase for fuzzy matching
		categoryMap["lower_"+cat.Name] = cat.ID
	}
	return categoryMap
}

// matchPlaidCategory maps Plaid category array to user's category ID
func matchPlaidCategory(plaidCategories []string, categoryMap map[string]uint, isExpense bool) *uint {
	// Plaid categories come in array like: ["Food and Drink", "Restaurants", "Fast Food"]
	// We'll try to match from most specific (last) to least specific (first)

	// Common Plaid category mappings to user categories
	plaidMapping := map[string]string{
		// Food categories
		"Food and Drink": "Groceries",
		"Restaurants":    "Entertainment",
		"Groceries":      "Groceries",

		// Transport
		"Transportation":        "Transportation",
		"Gas":                   "Transportation",
		"Public Transportation": "Transportation",
		"Taxi":                  "Transportation",

		// Housing
		"Rent":     "Rent",
		"Mortgage": "Rent",

		// Income
		"Income":  "Salary",
		"Payroll": "Salary",
		"Deposit": "Salary",

		// Entertainment
		"Entertainment":          "Entertainment",
		"Recreation":             "Entertainment",
		"Arts and Entertainment": "Entertainment",
		"Music":                  "Entertainment",
		"Movies":                 "Entertainment",
	}

	// Try to find a match from most specific to least
	for i := len(plaidCategories) - 1; i >= 0; i-- {
		plaidCat := plaidCategories[i]

		// First try direct mapping
		if userCatName, ok := plaidMapping[plaidCat]; ok {
			if catID, found := categoryMap[userCatName]; found {
				return &catID
			}
		}

		// Then try exact match with user's categories
		if catID, found := categoryMap[plaidCat]; found {
			return &catID
		}
	}

	// If no match found, return nil (uncategorized)
	return nil
}

// matchByMerchantName attempts to categorize based on merchant name keywords
func matchByMerchantName(merchantName string, categoryMap map[string]uint, isExpense bool) *uint {
	merchantLower := strings.ToLower(merchantName)

	// Keyword mappings for common merchants/patterns
	// Each entry maps to potential category names (will fuzzy match against user's categories)
	// Supports both English and Norwegian
	keywordMapping := map[string][]string{
		"food": {
			// Fast food & restaurants
			"starbucks", "mcdonald", "mcdonalds", "burger", "pizza", "subway", "kfc", "taco", "chipotle",
			"restaurant", "cafe", "coffee", "bistro", "diner", "bakery", "patisserie",
			// Grocery stores
			"food", "dining", "grocery", "groceries", "supermarket", "market", "rema", "kiwi", "coop",
			"meny", "spar", "bunnpris", "joker", "mat", "matbutikk",
		},
		"transport": {
			// Ride sharing & taxis
			"uber", "lyft", "taxi", "cab", "drosje",
			// Gas & fuel
			"gas", "fuel", "petrol", "bensin", "shell", "chevron", "esso", "circle k", "statoil", "7-eleven",
			// Public transport
			"parking", "transit", "metro", "subway", "bus", "train", "tog", "buss", "ruter", "nsb", "vy",
			// Car related & airlines (fallback)
			"transport", "bil", "vehicle", "auto", "airline", "airlines", "united", "delta", "southwest", "norwegian", "sas",
		},
		"travel": {
			// Airlines
			"airline", "airlines", "united", "delta", "american airlines", "southwest", "ryanair", "norwegian",
			"sas", "klm", "lufthansa", "british airways", "airways", "air",
			// Accommodation
			"hotel", "hostel", "motel", "airbnb", "booking", "expedia", "hotels.com",
			// Travel services
			"travel", "reise", "vacation", "ferie", "cruise",
		},
		"shopping": {
			// Online shopping
			"amazon", "ebay", "etsy", "alibaba", "wish",
			// Department stores
			"walmart", "target", "costco", "ikea", "h&m", "zara", "uniqlo",
			// Electronics & tech
			"apple", "best buy", "sparkfun", "elektronikk", "elkjøp", "power", "komplett",
			// General
			"shop", "shopping", "store", "mall", "butikk", "kjøpesenter",
		},
		"entertainment": {
			// Streaming services
			"netflix", "spotify", "hulu", "disney", "disney+", "hbo", "prime video", "youtube",
			// Entertainment venues
			"cinema", "movie", "theater", "kino", "concert", "konsert", "show",
			// Gaming
			"steam", "playstation", "xbox", "nintendo", "game", "gaming", "spill",
			// General
			"entertainment", "underholdning",
		},
		"utilities": {
			// Utilities
			"electric", "electricity", "strøm", "power", "water", "vann", "gas", "gass",
			// Internet & phone
			"internet", "broadband", "fiber", "phone", "telefon", "mobile", "mobil",
			// Service providers
			"telenor", "telia", "ice", "verizon", "at&t", "comcast", "xfinity",
			// Bills
			"utility", "utilities", "bill", "bills", "regning",
		},
		"healthcare": {
			// Medical facilities
			"hospital", "sykehus", "clinic", "klinikk", "doctor", "lege", "dentist", "tannlege",
			// Pharmacies
			"pharmacy", "apotek", "apoteket", "cvs", "walgreens", "boots",
			// General
			"medical", "medisin", "health", "helse", "care", "omsorg",
		},
		"education": {
			// Educational institutions
			"school", "skole", "university", "universitet", "college", "høyskole",
			// Online learning
			"udemy", "coursera", "skillshare", "lynda", "pluralsight",
			// General
			"education", "utdanning", "tuition", "skolepenger", "course", "kurs", "textbook", "lærebok",
		},
		"salary": {
			// Income related
			"payroll", "lønn", "salary", "income", "inntekt", "wage", "lønnsinntekt",
			"deposit", "payment received", "betaling mottatt", "transfer", "overføring",
		},
		"housing": {
			// Housing payments
			"rent", "leie", "husleie", "mortgage", "lån", "boliglån",
			// General
			"lease", "apartment", "leilighet", "house", "hus", "bolig",
		},
		"insurance": {
			// Insurance types
			"insurance", "forsikring", "assurance", "if", "tryg", "gjensidige", "sparebank",
			"life insurance", "livsforsikring", "car insurance", "bilforsikring",
		},
		"subscription": {
			// Subscriptions
			"subscription", "abonnement", "membership", "medlemskap", "monthly", "månedlig",
		},
	}

	// Try to match keywords
	for categoryHint, keywords := range keywordMapping {
		for _, keyword := range keywords {
			if strings.Contains(merchantLower, keyword) {
				// Found a keyword match, now find a category that matches
				catID := findMatchingCategory(categoryHint, categoryMap)
				if catID != nil {
					return catID
				}
			}
		}
	}

	return nil
}

// findMatchingCategory finds a category ID by fuzzy matching the category name
func findMatchingCategory(hint string, categoryMap map[string]uint) *uint {
	hintLower := strings.ToLower(hint)

	// Additional mappings for hints to common category variations
	hintToCategoryMapping := map[string][]string{
		"food":          {"food", "dining", "groceries", "grocery", "mat", "restaurant"},
		"transport":     {"transport", "transportation", "travel", "bil", "car"},
		"travel":        {"travel", "reise", "vacation", "ferie", "trip"},
		"shopping":      {"shop", "shopping", "butikk", "store", "groceries", "grocery"},
		"entertainment": {"entertainment", "underholdning", "fun", "leisure"},
		"utilities":     {"utilities", "utility", "bills", "regning"},
		"healthcare":    {"health", "healthcare", "medical", "helse"},
		"education":     {"education", "school", "utdanning", "skole"},
		"salary":        {"salary", "income", "lønn", "inntekt", "wage"},
		"housing":       {"housing", "rent", "mortgage", "bolig", "leie"},
		"insurance":     {"insurance", "forsikring"},
		"subscription":  {"subscription", "abonnement", "membership"},
	}

	// Try to find a category that contains the hint or vice versa
	for catName, catID := range categoryMap {
		catNameLower := strings.ToLower(catName)

		// Direct match: category name contains the hint or hint contains category name
		if strings.Contains(catNameLower, hintLower) || strings.Contains(hintLower, catNameLower) {
			return &catID
		}

		// Extended match: check if category name matches any of the hint's variations
		if variations, ok := hintToCategoryMapping[hintLower]; ok {
			for _, variation := range variations {
				if strings.Contains(catNameLower, variation) || strings.Contains(variation, catNameLower) {
					return &catID
				}
			}
		}
	}

	return nil
}
