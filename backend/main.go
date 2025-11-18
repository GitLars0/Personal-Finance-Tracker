package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"Personal-Finance-Tracker-backend/controllers"
	"Personal-Finance-Tracker-backend/db"
	"Personal-Finance-Tracker-backend/middleware"
	"Personal-Finance-Tracker-backend/redis"
	"Personal-Finance-Tracker-backend/routes"
	"Personal-Finance-Tracker-backend/seed"
	"Personal-Finance-Tracker-backend/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {
	// Initialize structured logger
	if err := utils.InitLogger(); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer utils.SyncLogger()

	logger := utils.Logger
	logger.Info("Starting Personal Finance Tracker API")

	// Initialize DB
	db.ConnectDatabase()
	if err := redis.InitRedis(); err != nil {
		log.Printf("⚠️ Redis init failed: %v (continuing without redis)", err)
	}

	logger.Info("Database connected successfully")

	// Initialize Plaid client
	plaidClientID := os.Getenv("PLAID_CLIENT_ID")
	plaidSecret := os.Getenv("PLAID_SECRET")
	plaidEnv := os.Getenv("PLAID_ENV")

	if plaidClientID != "" && plaidSecret != "" {
		if err := controllers.InitPlaidClient(plaidClientID, plaidSecret, plaidEnv); err != nil {
			logger.Warn("Failed to initialize Plaid client", zap.Error(err))
		} else {
			logger.Info("Plaid client initialized successfully", zap.String("environment", plaidEnv))
		}
	} else {
		logger.Info("Plaid credentials not configured, Plaid features disabled")
	}
	// Seed demo data
	seed.SeedDemoData(db.DB)

	// Set Gin to release mode for production
	gin.SetMode(gin.ReleaseMode)

	// Create Gin router without default middleware
	r := gin.New()

	// Add custom middleware
	r.Use(middleware.RecoveryMiddleware(logger))
	r.Use(middleware.LoggingMiddleware(logger))
	r.Use(middleware.MetricsMiddleware())

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Prometheus metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Health check endpoints (no auth required)
	r.GET("/health", controllers.HealthCheck)
	r.GET("/health/detailed", func(c *gin.Context) {
		controllers.DetailedHealthCheck(c, logger)
	})
	r.GET("/health/ready", controllers.ReadinessCheck)
	r.GET("/health/live", controllers.LivenessCheck)

	// Auth endpoints (no auth required)
	auth := r.Group("/auth")
	{
		auth.POST("/register", controllers.Register)
		auth.POST("/login", controllers.Login)
	}

	// API routes (protected)
	api := r.Group("/api")
	api.Use(controllers.AuthMiddleware())
	{
		routes.SetupRoutes(api)
	}

	// Serve static frontend files
	r.Static("/static", "./frontend/build/static")
	r.StaticFile("/favicon.ico", "./frontend/build/favicon.ico")
	r.StaticFile("/manifest.json", "./frontend/build/manifest.json")
	r.StaticFile("/robots.txt", "./frontend/build/robots.txt")

	// Simple Redis test endpoint for demo/report
	r.GET("/redistest", func(c *gin.Context) {
		if redis.RDB == nil {
			c.JSON(http.StatusOK, gin.H{
				"status": "redis not initialized",
			})
			return
		}

		ctx := context.Background()

		// Write a value into Redis
		if err := redis.RDB.Set(ctx, "hello", "from-redis-in-gcp", 0).Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  "failed to SET in redis",
				"detail": err.Error(),
			})
			return
		}

		// Read it back
		val, err := redis.RDB.Get(ctx, "hello").Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  "failed to GET from redis",
				"detail": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":     "Redis round-trip successful",
			"redis_value": val,
		})
	})

	// Optional: catch-all for any other frontend route
	r.NoRoute(func(c *gin.Context) {
		// If the path starts with /api, /auth, /health, /metrics, or /redistest, return 404 JSON
		path := c.Request.URL.Path

		// Check if it's an API route that should return JSON error
		isAPIRoute := false
		if len(path) >= 4 && path[:4] == "/api" {
			isAPIRoute = true
		} else if len(path) >= 5 && path[:5] == "/auth" {
			isAPIRoute = true
		} else if len(path) >= 7 && path[:7] == "/health" {
			isAPIRoute = true
		} else if len(path) >= 8 && path[:8] == "/metrics" {
			isAPIRoute = true
		} else if len(path) >= 10 && path[:10] == "/redistest" {
			isAPIRoute = true
		}

		if isAPIRoute {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Route not found",
				"path":  path,
			})
			return
		}

		// Otherwise, serve the React app for frontend routes
		c.File("./frontend/build/index.html")
	})

	logger.Info("Server starting on port 8080")
	if err := r.Run(":8080"); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
