package controllers_test

import (
	"Personal-Finance-Tracker-backend/controllers"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/suite"
)

// AuthHelpersTestSuite defines the test suite for auth helper tests
type AuthHelpersTestSuite struct {
	suite.Suite
}

// SetupSuite is called once before all tests in the suite
func (suite *AuthHelpersTestSuite) SetupSuite() {
	// Set test environment
	os.Setenv("JWT_SECRET", "test_secret_key_for_testing")
	gin.SetMode(gin.TestMode)
}

func (suite *AuthHelpersTestSuite) TestHashPassword() {
	password := "testpassword123"

	hash, err := controllers.HashPassword(password)
	suite.NoError(err)
	suite.NotEmpty(hash)
	suite.Contains(hash, "$argon2id$")

	// Test that same password produces different hashes (due to salt)
	hash2, err := controllers.HashPassword(password)
	suite.NoError(err)
	suite.NotEqual(hash, hash2)
}

func (suite *AuthHelpersTestSuite) TestVerifyPassword() {
	password := "testpassword123"
	wrongPassword := "wrongpassword"

	hash, err := controllers.HashPassword(password)
	suite.NoError(err)

	// Test correct password
	valid := controllers.VerifyPassword(password, hash)
	suite.True(valid)

	// Test wrong password
	invalid := controllers.VerifyPassword(wrongPassword, hash)
	suite.False(invalid)

	// Test with malformed hash
	invalidHash := controllers.VerifyPassword(password, "invalid$hash$format")
	suite.False(invalidHash)
}

func (suite *AuthHelpersTestSuite) TestGenerateToken() {
	userID := uint(123)
	username := "testuser"
	role := "user"

	token, err := controllers.GenerateToken(userID, username, role)
	suite.NoError(err)
	suite.NotEmpty(token)

	// Verify token can be parsed
	parsedToken, err := controllers.ParseToken(token)
	suite.NoError(err)
	suite.True(parsedToken.Valid)

	// Check claims
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	suite.True(ok)
	suite.Equal(float64(userID), claims["sub"])
	suite.Equal(username, claims["name"])
	suite.Equal(role, claims["role"])

	// Check expiration
	exp, ok := claims["exp"].(float64)
	suite.True(ok)
	suite.True(time.Unix(int64(exp), 0).After(time.Now()))
}

func (suite *AuthHelpersTestSuite) TestParseToken() {
	userID := uint(123)
	username := "testuser"
	role := "role"

	// Generate valid token
	tokenStr, err := controllers.GenerateToken(userID, username, role)
	suite.NoError(err)

	// Parse valid token
	token, err := controllers.ParseToken(tokenStr)
	suite.NoError(err)
	suite.True(token.Valid)

	// Test invalid token
	_, err = controllers.ParseToken("invalid.token.here")
	suite.Error(err)
}

func (suite *AuthHelpersTestSuite) TestAuthMiddleware() {
	userID := uint(123)
	username := "testuser"
	role := "user"
	validToken, _ := controllers.GenerateToken(userID, username, role)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectNext     bool
	}{
		{
			name:           "Valid token",
			authHeader:     "Bearer " + validToken,
			expectedStatus: http.StatusOK,
			expectNext:     true,
		},
		{
			name:           "Missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectNext:     false,
		},
		{
			name:           "Invalid authorization format",
			authHeader:     "InvalidFormat",
			expectedStatus: http.StatusUnauthorized,
			expectNext:     false,
		},
		{
			name:           "Invalid token",
			authHeader:     "Bearer invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
			expectNext:     false,
		},
		{
			name:           "Wrong scheme",
			authHeader:     "Basic " + validToken,
			expectedStatus: http.StatusUnauthorized,
			expectNext:     false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			router := gin.New()

			nextCalled := false
			router.Use(controllers.AuthMiddleware())
			router.GET("/test", func(c *gin.Context) {
				nextCalled = true
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req, _ := http.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			suite.Equal(tt.expectedStatus, recorder.Code)
			suite.Equal(tt.expectNext, nextCalled)
		})
	}
}

// TestAuthHelpersTestSuite runs the auth helpers test suite
func TestAuthHelpersTestSuite(t *testing.T) {
	suite.Run(t, new(AuthHelpersTestSuite))
}
