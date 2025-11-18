package controllers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"Personal-Finance-Tracker-backend/controllers"
	"Personal-Finance-Tracker-backend/db"
	"Personal-Finance-Tracker-backend/models"

	"github.com/stretchr/testify/assert"
)

func TestCreateCategory(t *testing.T) {
	database := SetupTestDB()
	db.DB = database
	user := CreateTestUser(database)
	token := GetTestToken(user.ID, user.Username)

	router := SetupRouter()
	router.POST("/api/categories", controllers.AuthMiddleware(), controllers.CreateCategory)

	categoryData := map[string]interface{}{
		"name": "Category 1",
		"kind": "expense",
		"description": "Test category",
	}
	body, _ := json.Marshal(categoryData)
	req, _ := http.NewRequest("POST", "/api/categories", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code, "Expected 201 Created status")

	var response models.Category
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "Category 1", response.Name, "Category name should match")
	assert.Equal(t, models.CategoryExpense, response.Kind, "Category kind should match")
	assert.Equal(t, user.ID, response.UserID, "Category UserID should match the test user ID")
	assert.Equal(t, "Test category", *response.Description, "Category description should match")
	assert.Nil(t, response.ParentID, "Top-level category should have nil ParentID")
}

func TestCreateSubcategory(t *testing.T) {
	database := SetupTestDB()
	db.DB = database
	user := CreateTestUser(database)
	token := GetTestToken(user.ID, user.Username)

	parent := models.Category{
		UserID:    user.ID,
		Name:      "Parent Category",
		Kind:      models.CategoryIncome,
		CreatedAt: time.Now(),
	}
	database.Create(&parent)

	router := SetupRouter()
	router.POST("/api/categories", controllers.AuthMiddleware(), controllers.CreateCategory)

	subcategoryData := map[string]interface{}{
		"name":      "Subcategory 1",
		"kind":      "income",
		"parent_id": parent.ID,
	}
	body, _ := json.Marshal(subcategoryData)
	
	req, _ := http.NewRequest("POST", "/api/categories", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code, "Expected 201 Created status")

	var response models.Category
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "Subcategory 1", response.Name, "Subcategory name should match")
	assert.Equal(t, models.CategoryIncome, response.Kind, "Subcategory kind should match")
	assert.Equal(t, user.ID, response.UserID, "Subcategory UserID should match the test user ID")
	assert.Equal(t, parent.ID, *response.ParentID, "Subcategory ParentID should match the parent category ID")
}

func TestCreateCategory_ParentKindMismatch(t *testing.T) {
	database := SetupTestDB()
	db.DB = database
	user := CreateTestUser(database)
	token := GetTestToken(user.ID, user.Username)

	incomeParent := models.Category{
		UserID:    user.ID,
		Name:      "Parent Category",
		Kind:      models.CategoryIncome,
	}
	database.Create(&incomeParent)

	router := SetupRouter()
	router.POST("/api/categories", controllers.AuthMiddleware(), controllers.CreateCategory)

	subcategoryData := map[string]interface{}{
		"name":      "Subcategory 1",
		"kind":      "expense", // Diffenrent kind than parent
		"parent_id": incomeParent.ID,
	}
	body, _ := json.Marshal(subcategoryData)

	req, _ := http.NewRequest("POST", "/api/categories", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code, "Expected 400 Bad Request status")

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"].(string), "parent category must have the same kind", "Error message should indicate kind mismatch")
}

func TestCreateCategory_MaxDepth(t *testing.T) {
	database := SetupTestDB()
	db.DB = database
	user := CreateTestUser(database)
	token := GetTestToken(user.ID, user.Username)

    // Create 3-level hierarchy: Food > Groceries > Vegetables
    food := models.Category{
        UserID: user.ID,
        Name:   "Food",
        Kind:   models.CategoryExpense,
    }
    database.Create(&food)

    groceries := models.Category{
        UserID:   user.ID,
        Name:     "Groceries",
        Kind:     models.CategoryExpense,
        ParentID: &food.ID,
    }
    database.Create(&groceries)

    vegetables := models.Category{
        UserID:   user.ID,
        Name:     "Vegetables",
        Kind:     models.CategoryExpense,
        ParentID: &groceries.ID,
    }
    database.Create(&vegetables)

	router := SetupRouter()
	router.POST("/api/categories", controllers.AuthMiddleware(), controllers.CreateCategory)

	subcategoryData := map[string]interface{}{
		"name":      "Leafy Greens",
		"kind":      "expense",
		"parent_id": vegetables.ID,
	}
	body, _ := json.Marshal(subcategoryData)

	req, _ := http.NewRequest("POST", "/api/categories", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code, "Expected 400 Bad Request status")

	var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.Contains(t, response["error"].(string), "nesting too deep",
        "Should reject 4th level nesting")
}

func TestCreateCategory_DuplicateName(t *testing.T) {
	database := SetupTestDB()
	db.DB = database
	user := CreateTestUser(database)
	token := GetTestToken(user.ID, user.Username)

	existingCategory := models.Category{
		UserID:    user.ID,
		Name:      "Category 1",
		Kind:      models.CategoryExpense,
	}
	database.Create(&existingCategory)

	router := SetupRouter()
	router.POST("/api/categories", controllers.AuthMiddleware(), controllers.CreateCategory)

	categoryData := map[string]interface{}{
		"name": "Category 1", // Duplicate name
		"kind": "expense",
	}
	body, _ := json.Marshal(categoryData)

	req, _ := http.NewRequest("POST", "/api/categories", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Line 210
	assert.Equal(t, http.StatusConflict, w.Code, "Expected 409 Conflict status")


	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"].(string), "category with this name already exists")
}

func TestGetCategories(t *testing.T) {
	database := SetupTestDB()
	db.DB = database
	user := CreateTestUser(database)
	token := GetTestToken(user.ID, user.Username)

	category1 := models.Category{
		UserID:    user.ID,
		Name:      "Category 1",
		Kind:      models.CategoryExpense,
		CreatedAt: time.Now(),
	}
	category2 := models.Category{
		UserID:    user.ID,
		Name:      "Category 2",
		Kind:      models.CategoryIncome,
		CreatedAt: time.Now(),
	}
	category3 := models.Category{
		UserID:    user.ID,
		Name:      "Category 3",
		Kind:      models.CategoryExpense,
		CreatedAt: time.Now(),
	}
	database.Create(&category1)
	database.Create(&category2)
	database.Create(&category3)

	router := SetupRouter()
	router.GET("/api/categories", controllers.AuthMiddleware(), controllers.GetCategories)

	req, _ := http.NewRequest("GET", "/api/categories", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Expected 200 OK status")

	var response []models.Category
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Len(t, response, 3, "Should return 3 categories")
}

func TestUpdateCategory(t *testing.T) {
	database := SetupTestDB()
	db.DB = database
	user := CreateTestUser(database)
	token := GetTestToken(user.ID, user.Username)
	
	category := models.Category{
		UserID:    user.ID,
		Name:      "Old Name",
		Kind:      models.CategoryExpense,
		CreatedAt: time.Now(),
	}
	database.Create(&category)

	router := SetupRouter()
	router.PUT("/api/categories/:id", controllers.AuthMiddleware(), controllers.UpdateCategory)

	updateData := map[string]interface{}{
		"name":        "New Name",
		"description": "Updated description",
	}
	body, _ := json.Marshal(updateData)

    req, _ := http.NewRequest("PUT", "/api/categories/"+strconv.FormatUint(uint64(category.ID), 10), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Expected 200 OK status")

	var response models.Category
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "New Name", response.Name, "Category name should be updated")
	assert.Equal(t, "Updated description", *response.Description, "Category description should be updated")
}

func TestUpdateCategory_CircularReference(t *testing.T) {
	database := SetupTestDB()
	db.DB = database
	user := CreateTestUser(database)
	token := GetTestToken(user.ID, user.Username)

	food := models.Category{
		UserID:    user.ID,
		Name:      "Food",
		Kind:      models.CategoryExpense,
		CreatedAt: time.Now(),
	}
	database.Create(&food)

	groceries := models.Category{
		UserID:    user.ID,
		Name:      "Groceries",
		Kind:      models.CategoryExpense,
		ParentID:  &food.ID,
		CreatedAt: time.Now(),
	}
	database.Create(&groceries)

	router := SetupRouter()
	router.PUT("/api/categories/:id", controllers.AuthMiddleware(), controllers.UpdateCategory)

	// Try to make Food a child of Groceries
	updateData := map[string]interface{}{
		"ParentID": groceries.ID,
	}
	body, _ := json.Marshal(updateData)

    req, _ := http.NewRequest("PUT", "/api/categories/"+strconv.FormatUint(uint64(food.ID), 10), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code, "Expected 400 Bad Request status")

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"].(string), "circular reference", "Error message should indicate circular reference")
}

func TestUpdateCategory_SelfParent(t *testing.T) {
	database := SetupTestDB()
	db.DB = database
	user := CreateTestUser(database)
	token := GetTestToken(user.ID, user.Username)

	category := models.Category{
		UserID:    user.ID,
		Name:      "Category 1",
		Kind:      models.CategoryExpense,
		CreatedAt: time.Now(),
	}
	database.Create(&category)

	router := SetupRouter()
	router.PUT("/api/categories/:id", controllers.AuthMiddleware(), controllers.UpdateCategory)

	// Try to set the category's ParentID to itself
	updateData := map[string]interface{}{
		"ParentID": category.ID,
	}
	body, _ := json.Marshal(updateData)

    req, _ := http.NewRequest("PUT", "/api/categories/"+strconv.FormatUint(uint64(category.ID), 10), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code, "Expected 400 Bad Request status")

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"].(string), "category cannot be its own parent", "Error message should indicate self-parenting")
}

func TestDeleteCategory_WithSubcategories(t *testing.T) {
	database := SetupTestDB()
	db.DB = database
	user := CreateTestUser(database)
	token := GetTestToken(user.ID, user.Username)

	parent := models.Category{
		UserID:    user.ID,
		Name:      "Parent Category",
		Kind:      models.CategoryExpense,
		CreatedAt: time.Now(),
	}
	database.Create(&parent)

	subcategory := models.Category{
		UserID:    user.ID,
		Name:      "Subcategory",
		Kind:      models.CategoryExpense,
		ParentID:  &parent.ID,
		CreatedAt: time.Now(),
	}
	database.Create(&subcategory)

	router := SetupRouter()
	router.DELETE("/api/categories/:id", controllers.AuthMiddleware(), controllers.DeleteCategory)

    req, _ := http.NewRequest("DELETE", "/api/categories/"+strconv.FormatUint(uint64(parent.ID), 10), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code, "Expected 400 Bad Request status")

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"].(string), "cannot delete category with subcategories", "Error message should indicate existing subcategories")
}

func TestDeleteCategory_ForceDelete(t *testing.T) {
	database := SetupTestDB()
	db.DB = database
	user := CreateTestUser(database)
	token := GetTestToken(user.ID, user.Username)

	parent := models.Category{
		UserID:    user.ID,
		Name:      "Parent Category",
		Kind:      models.CategoryExpense,
		CreatedAt: time.Now(),
	}
	database.Create(&parent)

	subcategory := models.Category{
		UserID:    user.ID,
		Name:      "Subcategory",
		Kind:      models.CategoryExpense,
		ParentID:  &parent.ID,
		CreatedAt: time.Now(),
	}
	database.Create(&subcategory)

	router := SetupRouter()
	router.DELETE("/api/categories/:id", controllers.AuthMiddleware(), controllers.DeleteCategory)

    req, _ := http.NewRequest("DELETE", "/api/categories/"+strconv.FormatUint(uint64(parent.ID), 10)+"?force=true", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Expected 200 OK status")

	var count int64
	database.Model(&models.Category{}).Where("id IN ?", []uint{parent.ID, subcategory.ID}).Count(&count)
	assert.Equal(t, int64(0), count, "Both parent and child should be deleted")
}