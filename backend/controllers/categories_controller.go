package controllers

import (
	"net/http"
	"strconv"

	"Personal-Finance-Tracker-backend/db"
	"Personal-Finance-Tracker-backend/models"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
)

// GetCategories retrieves all categories for the authenticated user.
func GetCategories(c *gin.Context) {
	// Extract JWT claims from context
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := uint(claims.(jwt.MapClaims)["sub"].(float64))

	var categories []models.Category

	// Always filter by user_id for multi-tenancy
	query := db.DB.Where("user_id = ?", userID)

	// Allow filtering by kind (income/expense)
	if kind := c.Query("kind"); kind != "" {
		query = query.Where("kind = ?", kind)
	}

	// Allow filtering by parent (get children or top-level)
	if parentID := c.Query("parent_id"); parentID != "" {
		if parentID == "null" {
			query = query.Where("parent_id IS NULL")
		} else {
			query = query.Where("parent_id = ?", parentID)
		}
	}

	if err := query.Order("kind, name").Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch categories"})
		return
	}

	c.JSON(http.StatusOK, categories)
}

func GetCategory(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := uint(claims.(jwt.MapClaims)["sub"].(float64))

	// Get ID from URL path /categories/:id
	categoryID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
		return
	}

	var category models.Category
	if err := db.DB.Where("id = ? AND user_id = ?", categoryID, userID).First(&category).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		return
	}

	c.JSON(http.StatusOK, category)
}

func CreateCategory(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := uint(claims.(jwt.MapClaims)["sub"].(float64))

	var input struct {
		Name        string              `json:"name" binding:"required"`
		Kind        models.CategoryKind `json:"kind" binding:"required,oneof=income expense"`
		ParentID    *uint               `json:"parent_id"`
		Description *string             `json:"description"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.ParentID != nil {
		var parent models.Category
		if err := db.DB.Where("id = ? AND user_id = ?", *input.ParentID, userID).First(&parent).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "parent category not found or does not belong to user"})
			return
		}

		// Income categories cant have expense parents
		if parent.Kind != input.Kind {
			c.JSON(http.StatusBadRequest, gin.H{"error": "parent category must have the same kind (income/expense)"})
			return
		}

		// Limit to 3 levels: Category -> Subcategory -> Sub-subcategory
		depth := 1
		currentParentID := parent.ParentID
		for currentParentID != nil && depth < 3 {
			var tempParent models.Category
			if err := db.DB.Where("id = ?", *currentParentID).First(&tempParent).Error; err != nil {
				break
			}
			currentParentID = tempParent.ParentID
			depth++
		}
		if depth >= 3 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "category nesting too deep (max 3 levels)"})
			return
		}
	}

	var existingCount int64
	query := db.DB.Model(&models.Category{}).Where("user_id = ? AND name = ? AND kind = ?", userID, input.Name, input.Kind)
	if input.ParentID != nil {
		query = query.Where("parent_id = ?", *input.ParentID)
	} else {
		query = query.Where("parent_id IS NULL")
	}
	query.Count(&existingCount)

	if existingCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "category with this name already exists"})
		return
	}

	category := models.Category{
		UserID:      userID,
		Name:        input.Name,
		Kind:        input.Kind,
		ParentID:    input.ParentID,
		Description: input.Description,
	}

	if err := db.DB.Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create category"})
		return
	}

	c.JSON(http.StatusCreated, category)
}

func UpdateCategory(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := uint(claims.(jwt.MapClaims)["sub"].(float64))

	categoryID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
		return
	}

	var category models.Category
	if err := db.DB.Where("id = ? AND user_id = ?", categoryID, userID).First(&category).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		return
	}

	var input struct {
		Name        string              `json:"name"`
		Kind        models.CategoryKind `json:"kind" binding:"omitempty,oneof=income expense"`
		ParentID    *uint               `json:"ParentID"`
		Description *string             `json:"description"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.ParentID != nil {
		if *input.ParentID == category.ID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "category cannot be its own parent"})
			return
		}

		// Prevent: A -> B -> C -> A (circular reference)
		if isDescendant(category.ID, *input.ParentID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "circular reference detected - parent cannot be a descendant"})
			return
		}
	}

	if input.Kind != "" && input.Kind != category.Kind {
		var childCount int64
		db.DB.Model(&models.Category{}).Where("parent_id = ?", categoryID).Count(&childCount)
		if childCount > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot change kind of category with subcategory"})
			return
		}
	}

	if input.Name != "" {
		category.Name = input.Name
	}
	if input.Kind != "" {
		category.Kind = input.Kind
	}
	if input.Description != nil {
		category.Description = input.Description
	}
	category.ParentID = input.ParentID // Always update

	if err := db.DB.Save(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed top update category"})
		return
	}

	c.JSON(http.StatusOK, category)
}

func DeleteCategory(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := uint(claims.(jwt.MapClaims)["sub"].(float64))

	categoryID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
		return
	}

	force := c.Query("force") == "true"

	var category models.Category
	if err := db.DB.Where("id = ? AND user_id = ?", categoryID, userID).First(&category).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		return
	}

	var childCount int64
	db.DB.Model(&models.Category{}).Where("parent_id = ?", categoryID).Count(&childCount)
	if childCount > 0 && !force {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete category with subcategories"})
		return
	}

	// Check transactions
	var txnCount int64
	db.DB.Model(&models.Transaction{}).Where("category_id = ?", categoryID).Count(&txnCount)
	if txnCount > 0 && !force {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete category with existing transactions. Reassign transactions first."})
		return
	}

	// Check budget items
	var budgetItemCount int64
	db.DB.Model(&models.BudgetItem{}).Where("category_id = ?", categoryID).Count(&budgetItemCount)
	if budgetItemCount > 0 && !force {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete category used in budget items"})
		return
	}

	// Check transaction splits
	var splitCount int64
	db.DB.Model(&models.TransactionSplit{}).Where("category_id = ?", categoryID).Count(&splitCount)
	if splitCount > 0 && !force {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete category used in transaction splits"})
		return
	}

	// If force delete, delete all associated data
	if force {
		// Delete transaction splits first
		db.DB.Where("category_id = ?", categoryID).Delete(&models.TransactionSplit{})

		// Delete budget items
		db.DB.Where("category_id = ?", categoryID).Delete(&models.BudgetItem{})

		// Delete transactions
		db.DB.Where("category_id = ?", categoryID).Delete(&models.Transaction{})

		// Delete subcategories
		db.DB.Where("parent_id = ?", categoryID).Delete(&models.Category{})
	}

	if err := db.DB.Delete(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete category"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "category deleted successfully"})
}

// GetCategoryUsage returns usage statistics for a category
func GetCategoryUsage(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := uint(claims.(jwt.MapClaims)["sub"].(float64))

	categoryID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
		return
	}

	// Verify category belongs to user
	var category models.Category
	if err := db.DB.Where("id = ? AND user_id = ?", categoryID, userID).First(&category).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		return
	}

	// Count usage
	var transactionCount int64
	var budgetCount int64
	var subcategoryCount int64

	db.DB.Model(&models.Transaction{}).Where("category_id = ?", categoryID).Count(&transactionCount)
	db.DB.Model(&models.BudgetItem{}).Where("category_id = ?", categoryID).Count(&budgetCount)
	db.DB.Model(&models.Category{}).Where("parent_id = ?", categoryID).Count(&subcategoryCount)

	usage := gin.H{
		"transactionCount": transactionCount,
		"budgetCount":      budgetCount,
		"subcategoryCount": subcategoryCount,
		"hasUsage":         transactionCount > 0 || budgetCount > 0 || subcategoryCount > 0,
	}

	c.JSON(http.StatusOK, usage)
}

func GetCategoryTree(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	}

	userID := uint(claims.(jwt.MapClaims)["sub"].(float64))

	var categories []models.Category
	db.DB.Where("used_id = ?", userID).Order("kind, name").Find(&categories)

	// Nested struct
	type CategoryNode struct {
		models.Category
		Children []CategoryNode `json:"children,omitempty"`
	}

	// Pass 1: Creat all nodes in a map
	categoryMap := make(map[uint]*CategoryNode)
	var rootCategories []CategoryNode

	for _, cat := range categories {
		node := categoryMap[cat.ID]
		if cat.ParentID == nil {
			rootCategories = append(rootCategories, *node)
		} else {
			if parent, exists := categoryMap[*cat.ParentID]; exists {
				parent.Children = append(parent.Children, *node)
			}
		}
	}

	c.JSON(http.StatusOK, rootCategories)
}

// Helper Functions
func isDescendant(categoryID, potentialDescendantID uint) bool {
	// Base case: fetch the potential descendant
	var category models.Category
	if err := db.DB.Where("id = ?", potentialDescendantID).First(&category).Error; err != nil {
		return false // Doesn't exist, cant be a descendant
	}

	// If no parent, its a root category
	if category.ParentID == nil {
		return false
	}

	// Direct child check
	if *category.ParentID == categoryID {
		return true // Found it
	}

	// Recursive check: Is the parent a descendant
	return isDescendant(categoryID, *category.ParentID)
}
