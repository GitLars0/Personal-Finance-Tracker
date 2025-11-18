package middleware

import (
    "time"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

// LoggingMiddleware logs HTTP requests with structured logging
func LoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        query := c.Request.URL.RawQuery

        // Process request
        c.Next()

        // Log after processing
        duration := time.Since(start)
        
        fields := []zap.Field{
            zap.Int("status", c.Writer.Status()),
            zap.String("method", c.Request.Method),
            zap.String("path", path),
            zap.String("query", query),
            zap.String("ip", c.ClientIP()),
            zap.Duration("duration", duration),
            zap.String("user_agent", c.Request.UserAgent()),
        }

        // Add user ID if authenticated
        if userClaims, exists := c.Get("user"); exists {
            fields = append(fields, zap.Any("user", userClaims))
        }

        // Log errors if any
        if len(c.Errors) > 0 {
            fields = append(fields, zap.String("errors", c.Errors.String()))
        }

        // Choose log level based on status code
        switch {
        case c.Writer.Status() >= 500:
            logger.Error("Server error", fields...)
        case c.Writer.Status() >= 400:
            logger.Warn("Client error", fields...)
        default:
            logger.Info("Request completed", fields...)
        }
    }
}

// RecoveryMiddleware recovers from panics and logs them
func RecoveryMiddleware(logger *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                logger.Error("Panic recovered",
                    zap.Any("error", err),
                    zap.String("path", c.Request.URL.Path),
                    zap.String("method", c.Request.Method),
                )
                c.AbortWithStatusJSON(500, gin.H{
                    "error": "Internal server error",
                })
            }
        }()
        c.Next()
    }
}