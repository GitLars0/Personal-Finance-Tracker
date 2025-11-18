package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"Personal-Finance-Tracker-backend/controllers"
	"Personal-Finance-Tracker-backend/db"
	"Personal-Finance-Tracker-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type HealthCheckControllerTestSuite struct {
	suite.Suite
	database *gorm.DB
	router   *gin.Engine
	logger   *zap.Logger
}

// ============================================
// SETUP AND TEARDOWN
// ============================================
func (suite *HealthCheckControllerTestSuite) SetupSuite() {
	// Create in-memory SQLite database for testing
	testDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	// Set the database instance
	db.DB = testDB
	suite.database = testDB

	// Migrate basic models for database connectivity testing
	err = testDB.AutoMigrate(
		&models.User{},
		&models.Account{},
		&models.Category{},
		&models.Transaction{},
		&models.Budget{},
	)
	suite.Require().NoError(err)

	// Create test logger
	suite.logger = zap.NewNop() // No-op logger for testing

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()

	// Add health endpoints (no auth required - same as main.go)
	suite.router.GET("/health", controllers.HealthCheck)
	suite.router.GET("/health/detailed", func(c *gin.Context) {
		controllers.DetailedHealthCheck(c, suite.logger)
	})
	suite.router.GET("/health/ready", controllers.ReadinessCheck)
	suite.router.GET("/health/live", controllers.LivenessCheck)
}

func (suite *HealthCheckControllerTestSuite) TearDownSuite() {
	// Clean up database
	if suite.database != nil {
		sqlDB, _ := suite.database.DB()
		sqlDB.Close()
	}
}

// ============================================
// TEST 1: Basic Health Check
// ============================================
func (suite *HealthCheckControllerTestSuite) TestHealthCheck_Success() {
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)

	// Verify basic fields
	suite.Equal("ok", response["status"])
	suite.Contains(response, "timestamp")

	// Verify timestamp format (should be RFC3339)
	timestamp, ok := response["timestamp"].(string)
	suite.True(ok)
	_, err = time.Parse(time.RFC3339, timestamp)
	suite.NoError(err, "Timestamp should be in RFC3339 format")
}

func (suite *HealthCheckControllerTestSuite) TestHealthCheck_ResponseTime() {
	// Test that basic health check is fast
	start := time.Now()

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	elapsed := time.Since(start)

	suite.Equal(http.StatusOK, w.Code)
	suite.Less(elapsed, 100*time.Millisecond, "Basic health check should be very fast")
}

// ============================================
// TEST 2: Detailed Health Check
// ============================================
func (suite *HealthCheckControllerTestSuite) TestDetailedHealthCheck_Success() {
	req, _ := http.NewRequest("GET", "/health/detailed", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response controllers.HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)

	// Verify response structure
	suite.Equal("healthy", response.Status)
	suite.NotEmpty(response.Timestamp)
	suite.Equal("1.0.0", response.Version)
	suite.NotNil(response.Services)

	// Verify timestamp format
	_, err = time.Parse(time.RFC3339, response.Timestamp)
	suite.NoError(err, "Timestamp should be in RFC3339 format")

	// Verify database service check
	dbStatus, exists := response.Services["database"]
	suite.True(exists, "Database service should be checked")
	suite.Equal("healthy", dbStatus, "Database should be healthy in tests")
}

func (suite *HealthCheckControllerTestSuite) TestDetailedHealthCheck_DatabaseConnected() {
	// Ensure we have a valid database connection
	req, _ := http.NewRequest("GET", "/health/detailed", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response controllers.HealthResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	// Database should be reported as healthy since we use in-memory SQLite
	suite.Equal("healthy", response.Status)
	suite.Equal("healthy", response.Services["database"])
}

// ============================================
// TEST 3: Readiness Check
// ============================================
func (suite *HealthCheckControllerTestSuite) TestReadinessCheck_Ready() {
	req, _ := http.NewRequest("GET", "/health/ready", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)

	suite.Equal("ready", response["status"])
}

func (suite *HealthCheckControllerTestSuite) TestReadinessCheck_DatabaseRequired() {
	// The readiness check should verify database connectivity
	// Since we have a valid SQLite connection, it should pass
	req, _ := http.NewRequest("GET", "/health/ready", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	suite.Equal("ready", response["status"])
	suite.NotContains(response, "reason", "Should not contain error reason when ready")
}

// ============================================
// TEST 4: Liveness Check
// ============================================
func (suite *HealthCheckControllerTestSuite) TestLivenessCheck_Alive() {
	req, _ := http.NewRequest("GET", "/health/live", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)

	suite.Equal("alive", response["status"])
}

func (suite *HealthCheckControllerTestSuite) TestLivenessCheck_AlwaysSucceeds() {
	// Liveness should always succeed as long as the process is running
	// Test multiple times to ensure consistency
	for i := 0; i < 5; i++ {
		req, _ := http.NewRequest("GET", "/health/live", nil)
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		suite.Equal(http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		suite.Equal("alive", response["status"])
	}
}

// ============================================
// TEST 5: Health Endpoints Integration
// ============================================
func (suite *HealthCheckControllerTestSuite) TestAllHealthEndpoints_NoAuthentication() {
	// All health endpoints should work without authentication
	endpoints := []string{"/health", "/health/detailed", "/health/ready", "/health/live"}

	for _, endpoint := range endpoints {
		suite.Run("endpoint_"+endpoint, func() {
			req, _ := http.NewRequest("GET", endpoint, nil)
			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			// All should return success (no 401 Unauthorized)
			suite.NotEqual(http.StatusUnauthorized, w.Code,
				"Health endpoint %s should not require authentication", endpoint)

			// Basic and liveness should always be 200
			if endpoint == "/health" || endpoint == "/health/live" {
				suite.Equal(http.StatusOK, w.Code,
					"Endpoint %s should always return 200", endpoint)
			}
		})
	}
}

func (suite *HealthCheckControllerTestSuite) TestHealthEndpoints_ContentType() {
	endpoints := []string{"/health", "/health/detailed", "/health/ready", "/health/live"}

	for _, endpoint := range endpoints {
		suite.Run("content_type_"+endpoint, func() {
			req, _ := http.NewRequest("GET", endpoint, nil)
			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			// Should return JSON content type
			suite.Contains(w.Header().Get("Content-Type"), "application/json",
				"Endpoint %s should return JSON", endpoint)
		})
	}
}

// ============================================
// TEST 6: Response Format Validation
// ============================================
func (suite *HealthCheckControllerTestSuite) TestDetailedHealthCheck_ResponseFormat() {
	req, _ := http.NewRequest("GET", "/health/detailed", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	var response controllers.HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)

	// Validate required fields are present and have correct types
	suite.NotEmpty(response.Status, "Status should not be empty")
	suite.NotEmpty(response.Timestamp, "Timestamp should not be empty")
	suite.NotEmpty(response.Version, "Version should not be empty")
	suite.NotNil(response.Services, "Services map should not be nil")

	// Validate status values
	validStatuses := []string{"healthy", "degraded", "unhealthy"}
	suite.Contains(validStatuses, response.Status, "Status should be one of the valid values")
}

func (suite *HealthCheckControllerTestSuite) TestHealthCheck_JSONStructure() {
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)

	// Should have exactly 2 fields for basic health check
	suite.Len(response, 2, "Basic health check should have exactly 2 fields")
	suite.Contains(response, "status")
	suite.Contains(response, "timestamp")

	// Status should be string
	_, ok := response["status"].(string)
	suite.True(ok, "Status should be a string")

	// Timestamp should be string
	_, ok = response["timestamp"].(string)
	suite.True(ok, "Timestamp should be a string")
}

// ============================================
// TEST 7: Error Scenarios and Edge Cases
// ============================================
func (suite *HealthCheckControllerTestSuite) TestHealthEndpoints_HTTPMethods() {
	endpoints := []string{"/health", "/health/detailed", "/health/ready", "/health/live"}
	methods := []string{"POST", "PUT", "DELETE", "PATCH"}

	for _, endpoint := range endpoints {
		for _, method := range methods {
			suite.Run("method_"+method+"_"+endpoint, func() {
				req, _ := http.NewRequest(method, endpoint, nil)
				w := httptest.NewRecorder()
				suite.router.ServeHTTP(w, req)

				// Should return 404 Method Not Allowed for non-GET methods
				suite.Equal(http.StatusNotFound, w.Code,
					"Endpoint %s should not accept %s method", endpoint, method)
			})
		}
	}
}

func (suite *HealthCheckControllerTestSuite) TestHealthCheck_ConcurrentRequests() {
	// Test that health checks can handle concurrent requests
	const numRequests = 10
	done := make(chan bool, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			defer func() { done <- true }()

			req, _ := http.NewRequest("GET", "/health", nil)
			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			suite.Equal(http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			suite.NoError(err)
			suite.Equal("ok", response["status"])
		}()
	}

	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		<-done
	}
}

func (suite *HealthCheckControllerTestSuite) TestDetailedHealthCheck_VersionInfo() {
	req, _ := http.NewRequest("GET", "/health/detailed", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	var response controllers.HealthResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	// Version should be present and follow semantic versioning pattern
	suite.NotEmpty(response.Version)
	suite.Regexp(`^\d+\.\d+\.\d+$`, response.Version,
		"Version should follow semantic versioning pattern (e.g., 1.0.0)")
}

// ============================================
// TEST 8: Performance and Reliability
// ============================================
func (suite *HealthCheckControllerTestSuite) TestHealthEndpoints_Performance() {
	endpoints := []string{"/health", "/health/ready", "/health/live"}

	for _, endpoint := range endpoints {
		suite.Run("performance_"+endpoint, func() {
			start := time.Now()

			req, _ := http.NewRequest("GET", endpoint, nil)
			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			elapsed := time.Since(start)

			suite.Equal(http.StatusOK, w.Code)
			// Health checks should be fast (under 1 second)
			suite.Less(elapsed, time.Second,
				"Health endpoint %s should respond quickly", endpoint)
		})
	}
}

func (suite *HealthCheckControllerTestSuite) TestDetailedHealthCheck_DatabaseStats() {
	// Detailed health check should not fail even when checking database stats
	req, _ := http.NewRequest("GET", "/health/detailed", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response controllers.HealthResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	// Should successfully get database status
	suite.Equal("healthy", response.Status)
	suite.Equal("healthy", response.Services["database"])
}

// ============================================
// TEST RUNNER
// ============================================
func TestHealthCheckControllerTestSuite(t *testing.T) {
	suite.Run(t, new(HealthCheckControllerTestSuite))
}
