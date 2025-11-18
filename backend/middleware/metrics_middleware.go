package middleware

import (
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // HTTP request metrics
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )

    httpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint"},
    )

    httpRequestSize = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_size_bytes",
            Help:    "HTTP request size in bytes",
            Buckets: prometheus.ExponentialBuckets(100, 10, 8),
        },
        []string{"method", "endpoint"},
    )

    httpResponseSize = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_response_size_bytes",
            Help:    "HTTP response size in bytes",
            Buckets: prometheus.ExponentialBuckets(100, 10, 8),
        },
        []string{"method", "endpoint"},
    )

    // Business metrics
    activeUsers = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "active_users_total",
            Help: "Number of currently active users",
        },
    )

    transactionsCreated = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "transactions_created_total",
            Help: "Total number of transactions created",
        },
    )

    accountsCreated = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "accounts_created_total",
            Help: "Total number of accounts created",
        },
    )

    budgetsCreated = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "budgets_created_total",
            Help: "Total number of budgets created",
        },
    )

    // Database metrics
    dbQueriesTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "db_queries_total",
            Help: "Total number of database queries",
        },
        []string{"operation"},
    )

    dbQueryDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "db_query_duration_seconds",
            Help:    "Database query duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"operation"},
    )
)

// MetricsMiddleware collects HTTP metrics
func MetricsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.FullPath()
        
        // If path is empty (404), use the request path
        if path == "" {
            path = c.Request.URL.Path
        }

        // Request size
        reqSize := float64(c.Request.ContentLength)
        httpRequestSize.WithLabelValues(c.Request.Method, path).Observe(reqSize)

        c.Next()

        // Duration
        duration := time.Since(start).Seconds()
        httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)

        // Status
        status := strconv.Itoa(c.Writer.Status())
        httpRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()

        // Response size
        resSize := float64(c.Writer.Size())
        httpResponseSize.WithLabelValues(c.Request.Method, path).Observe(resSize)
    }
}

// Helper functions to track business metrics
func IncrementTransactionsCreated() {
    transactionsCreated.Inc()
}

func IncrementAccountsCreated() {
    accountsCreated.Inc()
}

func IncrementBudgetsCreated() {
    budgetsCreated.Inc()
}

func SetActiveUsers(count float64) {
    activeUsers.Set(count)
}

func TrackDBQuery(operation string, duration time.Duration) {
    dbQueriesTotal.WithLabelValues(operation).Inc()
    dbQueryDuration.WithLabelValues(operation).Observe(duration.Seconds())
}