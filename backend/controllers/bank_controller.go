package controllers

import (
	"strconv"

	"Personal-Finance-Tracker-backend/db"
	"Personal-Finance-Tracker-backend/models"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
)

// CreateBankConnection - Deprecated endpoint, use Plaid Link instead
func CreateBankConnection(c *gin.Context) {
	c.JSON(400, gin.H{
		"error":   "This endpoint is deprecated. Please use Plaid Link instead.",
		"message": "Use /api/plaid/create_link_token to connect banks via Plaid",
		"hint":    "All bank connections now use Plaid for security and reliability",
	})
}

// GetBankConnections returns all bank connections for a user
func GetBankConnections(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(400, gin.H{"error": "unauthorized"})
		return
	}
	userID := uint(claims.(jwt.MapClaims)["sub"].(float64))

	var connections []models.BankConnection
	if err := db.DB.Where("user_id = ?", userID).
		Preload("LinkedAccounts").
		Order("created_at DESC").
		Find(&connections).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch bank connections"})
		return
	}

	c.JSON(200, gin.H{"connections": connections})
}

// DisconnectBank removes a bank connection
func DisconnectBank(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(400, gin.H{"error": "unauthorized"})
		return
	}
	userID := uint(claims.(jwt.MapClaims)["sub"].(float64))
	connectionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid connection ID"})
		return
	}

	var connection models.BankConnection
	if err := db.DB.Where("user_id = ? AND id = ?", userID, connectionID).First(&connection).Error; err != nil {
		c.JSON(404, gin.H{"error": "Bank connection not found"})
		return
	}

	// Delete from database (soft delete)
	if err := db.DB.Delete(&connection).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to delete bank connection"})
		return
	}

	c.JSON(200, gin.H{"message": "Bank connection deleted successfully"})
}
