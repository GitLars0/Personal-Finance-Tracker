package controllers

import (
    "net/http"
    "time"

    "Personal-Finance-Tracker-backend/db"
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

// HealthResponse represents the health check response
type HealthResponse struct {
    Status    string            `json:"status"`
    Timestamp string            `json:"timestamp"`
    Services  map[string]string `json:"services"`
    Version   string            `json:"version"`
}

// HealthCheck provides a simple health check endpoint
func HealthCheck(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status":    "ok",
        "timestamp": time.Now().Format(time.RFC3339),
    })
}

// DetailedHealthCheck provides detailed health information
func DetailedHealthCheck(c *gin.Context, logger *zap.Logger) {
    response := HealthResponse{
        Status:    "healthy",
        Timestamp: time.Now().Format(time.RFC3339),
        Services:  make(map[string]string),
        Version:   "1.0.0",
    }

    // Check database connection
    sqlDB, err := db.DB.DB()
    if err != nil {
        logger.Error("Database connection error",
            zap.Error(err),
        )
        response.Status = "unhealthy"
        response.Services["database"] = "down"
    } else {
        // Ping database
        err = sqlDB.Ping()
        if err != nil {
            logger.Error("Database ping failed",
                zap.Error(err),
            )
            response.Status = "degraded"
            response.Services["database"] = "unreachable"
        } else {
            response.Services["database"] = "healthy"
            
            // Get database stats
            stats := sqlDB.Stats()
            logger.Info("Database stats",
                zap.Int("open_connections", stats.OpenConnections),
                zap.Int("in_use", stats.InUse),
                zap.Int("idle", stats.Idle),
            )
        }
    }

    // Determine HTTP status based on health
    statusCode := http.StatusOK
    if response.Status == "unhealthy" {
        statusCode = http.StatusServiceUnavailable
    } else if response.Status == "degraded" {
        statusCode = http.StatusOK // Still return 200 for degraded
    }

    c.JSON(statusCode, response)
}

// ReadinessCheck checks if the application is ready to serve traffic
func ReadinessCheck(c *gin.Context) {
    // Check if database is accessible
    sqlDB, err := db.DB.DB()
    if err != nil || sqlDB.Ping() != nil {
        c.JSON(http.StatusServiceUnavailable, gin.H{
            "status": "not_ready",
            "reason": "database_unavailable",
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "status": "ready",
    })
}

// LivenessCheck checks if the application is alive
func LivenessCheck(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status": "alive",
    })
}