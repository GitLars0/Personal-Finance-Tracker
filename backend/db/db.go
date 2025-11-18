package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"Personal-Finance-Tracker-backend/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	// ✅ Read environment variables
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")

	// 🧩 Local fallback (e.g., Docker Compose)
	if host == "" {
		host = "db"
	}
	if user == "" {
		user = "postgres"
	}
	if name == "" {
		name = "finance_tracker"
	}

	// ✅ Build connection string dynamically
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable",
		host, user, password, name,
	)

	var err error
	for i := 0; i < 10; i++ {
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			log.Println("✅ Connected to database!")

			// 🧱 AutoMigrate all models
			if err := DB.AutoMigrate(
				&models.User{},
				&models.Account{},
				&models.Category{},
				&models.Transaction{},
				&models.TransactionSplit{},
				&models.Budget{},
				&models.BudgetItem{},
				&models.BankConnection{},
				&models.BankAccount{},
				&models.BankSyncLog{},
			); err != nil {
				log.Fatalf("❌ Failed to migrate database: %v", err)
			}

			log.Println("✅ Database migration completed!")
			return
		}

		log.Println("⏳ Waiting for database...", err)
		time.Sleep(3 * time.Second)
	}

	log.Fatalf("❌ Failed to connect to database after retries: %v", err)
}
