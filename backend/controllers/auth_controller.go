package controllers

import (
    "Personal-Finance-Tracker-backend/db"
    "Personal-Finance-Tracker-backend/models"
    "Personal-Finance-Tracker-backend/utils"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    jwt "github.com/golang-jwt/jwt/v5"
    "go.uber.org/zap"
)

func Register(c *gin.Context) {
    var input struct {
        Username string `json:"username" binding:"required"`
        Email    string `json:"email" binding:"required,email"`
        Password string `json:"password" binding:"required,min=6"`
    }
    if err := c.ShouldBindJSON(&input); err != nil {
        utils.Logger.Warn("Registration validation failed",
            zap.Error(err),
            zap.String("ip", c.ClientIP()),
        )
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    hash, err := HashPassword(input.Password)
    if err != nil {
        utils.Logger.Error("Failed to hash password during registration",
            zap.Error(err),
            zap.String("username", input.Username),
        )
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
        return
    }

    user := models.User{
        Username:     input.Username,
        Email:        input.Email,
        PasswordHash: hash,
        CreatedAt:    time.Now(),
    }

    if err := db.DB.Create(&user).Error; err != nil {
        utils.Logger.Warn("User registration failed - duplicate username or email",
            zap.Error(err),
            zap.String("username", input.Username),
            zap.String("email", input.Email),
            zap.String("ip", c.ClientIP()),
        )
        c.JSON(http.StatusBadRequest, gin.H{"error": "Username or Email already exists"})
        return
    }

    utils.Logger.Info("User registered successfully",
        zap.Uint("user_id", user.ID),
        zap.String("username", user.Username),
        zap.String("email", user.Email),
        zap.String("ip", c.ClientIP()),
    )

    c.JSON(http.StatusCreated, gin.H{"message": "Registration successful"})
}

func Login(c *gin.Context) {
    var input struct {
        Username string `json:"username" binding:"required"`
        Password string `json:"password" binding:"required"`
    }
    if err := c.ShouldBindJSON(&input); err != nil {
        utils.Logger.Warn("Login validation failed",
            zap.Error(err),
            zap.String("ip", c.ClientIP()),
        )
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    var user models.User
    // Support login with username OR email
    if err := db.DB.Where("username = ? OR email = ?", input.Username, input.Username).First(&user).Error; err != nil {
        utils.Logger.Warn("Login failed - user not found",
            zap.String("username", input.Username),
            zap.String("ip", c.ClientIP()),
        )
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
        return
    }

    if !VerifyPassword(input.Password, user.PasswordHash) {
        utils.Logger.Warn("Login failed - invalid password",
            zap.String("username", input.Username),
            zap.Uint("user_id", user.ID),
            zap.String("ip", c.ClientIP()),
        )
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
        return
    }

    // Generate JWT token
    token, err := GenerateToken(user.ID, user.Username, string(user.Role))
    if err != nil {
        utils.Logger.Error("Failed to generate JWT token",
            zap.Error(err),
            zap.Uint("user_id", user.ID),
            zap.String("username", user.Username),
        )
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
        return
    }

    utils.Logger.Info("User logged in successfully",
        zap.Uint("user_id", user.ID),
        zap.String("username", user.Username),
        zap.String("role", string(user.Role)),
        zap.String("ip", c.ClientIP()),
    )

    // Return token and user object (not just username string)
    c.JSON(http.StatusOK, gin.H{
        "message": "Login successful",
        "token":   token,
        "user": gin.H{
            "id":       user.ID,
            "username": user.Username,
            "email":    user.Email,
            "name":     user.Name,
            "role":     user.Role,
        },
    })
}

// GetUserProfile returns the current user's profile information
func GetUserProfile(c *gin.Context) {
    claims, exists := c.Get("user")
    if !exists {
        utils.Logger.Warn("Unauthorized access to user profile",
            zap.String("ip", c.ClientIP()),
        )
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }

    userID := uint(claims.(jwt.MapClaims)["sub"].(float64))

    var user models.User
    if err := db.DB.First(&user, userID).Error; err != nil {
        utils.Logger.Warn("User profile not found",
            zap.Error(err),
            zap.Uint("user_id", userID),
        )
        c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
        return
    }

    utils.Logger.Debug("User profile retrieved",
        zap.Uint("user_id", userID),
        zap.String("username", user.Username),
    )

    c.JSON(http.StatusOK, gin.H{
        "id":         user.ID,
        "username":   user.Username,
        "email":      user.Email,
        "name":       user.Name,
        "role":       user.Role,
        "created_at": user.CreatedAt,
        "updated_at": user.UpdatedAt,
    })
}

// UpdateUserProfile updates the current user's profile information
func UpdateUserProfile(c *gin.Context) {
    claims, exists := c.Get("user")
    if !exists {
        utils.Logger.Warn("Unauthorized access to update profile",
            zap.String("ip", c.ClientIP()),
        )
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }

    userID := uint(claims.(jwt.MapClaims)["sub"].(float64))

    var user models.User
    if err := db.DB.First(&user, userID).Error; err != nil {
        utils.Logger.Warn("User not found for profile update",
            zap.Error(err),
            zap.Uint("user_id", userID),
        )
        c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
        return
    }

    var input struct {
        Name     *string `json:"name"`
        Email    *string `json:"email" binding:"omitempty,email"`
        Username *string `json:"username"`
    }

    if err := c.ShouldBindJSON(&input); err != nil {
        utils.Logger.Warn("Profile update validation failed",
            zap.Error(err),
            zap.Uint("user_id", userID),
        )
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Track what fields are being updated
    updatedFields := make([]string, 0)
    
    // Update fields only if provided
    if input.Name != nil {
        user.Name = *input.Name
        updatedFields = append(updatedFields, "name")
    }
    if input.Email != nil {
        user.Email = *input.Email
        updatedFields = append(updatedFields, "email")
    }
    if input.Username != nil {
        user.Username = *input.Username
        updatedFields = append(updatedFields, "username")
    }

    user.UpdatedAt = time.Now()

    if err := db.DB.Save(&user).Error; err != nil {
        utils.Logger.Warn("Profile update failed - duplicate username or email",
            zap.Error(err),
            zap.Uint("user_id", userID),
            zap.Strings("updated_fields", updatedFields),
        )
        c.JSON(http.StatusBadRequest, gin.H{"error": "Username or Email already exists"})
        return
    }

    utils.Logger.Info("User profile updated successfully",
        zap.Uint("user_id", userID),
        zap.String("username", user.Username),
        zap.Strings("updated_fields", updatedFields),
    )

    c.JSON(http.StatusOK, gin.H{
        "id":         user.ID,
        "username":   user.Username,
        "email":      user.Email,
        "name":       user.Name,
        "created_at": user.CreatedAt,
        "updated_at": user.UpdatedAt,
    })
}

// ChangePassword allows the current user to change their password
func ChangePassword(c *gin.Context) {
    claims, exists := c.Get("user")
    if !exists {
        utils.Logger.Warn("Unauthorized password change attempt",
            zap.String("ip", c.ClientIP()),
        )
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }

    userID := uint(claims.(jwt.MapClaims)["sub"].(float64))

    var user models.User
    if err := db.DB.First(&user, userID).Error; err != nil {
        utils.Logger.Warn("User not found for password change",
            zap.Error(err),
            zap.Uint("user_id", userID),
        )
        c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
        return
    }

    var input struct {
        CurrentPassword string `json:"current_password" binding:"required"`
        NewPassword     string `json:"new_password" binding:"required,min=6"`
    }

    if err := c.ShouldBindJSON(&input); err != nil {
        utils.Logger.Warn("Password change validation failed",
            zap.Error(err),
            zap.Uint("user_id", userID),
        )
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Verify current password
    if !VerifyPassword(input.CurrentPassword, user.PasswordHash) {
        utils.Logger.Warn("Password change failed - incorrect current password",
            zap.Uint("user_id", userID),
            zap.String("username", user.Username),
            zap.String("ip", c.ClientIP()),
        )
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
        return
    }

    // Hash new password
    newHash, err := HashPassword(input.NewPassword)
    if err != nil {
        utils.Logger.Error("Failed to hash new password",
            zap.Error(err),
            zap.Uint("user_id", userID),
        )
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password"})
        return
    }

    user.PasswordHash = newHash
    user.UpdatedAt = time.Now()

    if err := db.DB.Save(&user).Error; err != nil {
        utils.Logger.Error("Failed to save new password",
            zap.Error(err),
            zap.Uint("user_id", userID),
        )
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
        return
    }

    utils.Logger.Info("Password changed successfully",
        zap.Uint("user_id", userID),
        zap.String("username", user.Username),
        zap.String("ip", c.ClientIP()),
    )

    c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// DeleteUserAccount deletes the current user and all associated data
func DeleteUserAccount(c *gin.Context) {
    claims, exists := c.Get("user")
    if !exists {
        utils.Logger.Warn("Unauthorized account deletion attempt",
            zap.String("ip", c.ClientIP()),
        )
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }

    userID := uint(claims.(jwt.MapClaims)["sub"].(float64))

    // Get user info before deletion for logging
    var user models.User
    if err := db.DB.First(&user, userID).Error; err != nil {
        utils.Logger.Warn("User not found for account deletion",
            zap.Error(err),
            zap.Uint("user_id", userID),
        )
        c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
        return
    }

    utils.Logger.Info("Starting account deletion",
        zap.Uint("user_id", userID),
        zap.String("username", user.Username),
        zap.String("email", user.Email),
        zap.String("ip", c.ClientIP()),
    )

    // Start database transaction for atomicity
    tx := db.DB.Begin()

    // Delete all user data in correct order (respecting foreign key constraints)

    // 1. Delete budget items first (they reference budgets)
    if err := tx.Exec("DELETE FROM budget_items WHERE budget_id IN (SELECT id FROM budgets WHERE user_id = ?)", userID).Error; err != nil {
        tx.Rollback()
        utils.Logger.Error("Failed to delete budget items",
            zap.Error(err),
            zap.Uint("user_id", userID),
        )
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete budget items"})
        return
    }

    // 2. Delete budgets
    if err := tx.Where("user_id = ?", userID).Delete(&models.Budget{}).Error; err != nil {
        tx.Rollback()
        utils.Logger.Error("Failed to delete budgets",
            zap.Error(err),
            zap.Uint("user_id", userID),
        )
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete budgets"})
        return
    }

    // 3. Delete transaction splits (they reference transactions)
    if err := tx.Exec("DELETE FROM transaction_splits WHERE parent_txn_id IN (SELECT id FROM transactions WHERE user_id = ?)", userID).Error; err != nil {
        tx.Rollback()
        utils.Logger.Error("Failed to delete transaction splits",
            zap.Error(err),
            zap.Uint("user_id", userID),
        )
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete transaction splits"})
        return
    }

    // 4. Delete transactions
    if err := tx.Where("user_id = ?", userID).Delete(&models.Transaction{}).Error; err != nil {
        tx.Rollback()
        utils.Logger.Error("Failed to delete transactions",
            zap.Error(err),
            zap.Uint("user_id", userID),
        )
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete transactions"})
        return
    }

    // 5. Delete categories
    if err := tx.Where("user_id = ?", userID).Delete(&models.Category{}).Error; err != nil {
        tx.Rollback()
        utils.Logger.Error("Failed to delete categories",
            zap.Error(err),
            zap.Uint("user_id", userID),
        )
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete categories"})
        return
    }

    // 6. Delete accounts
    if err := tx.Where("user_id = ?", userID).Delete(&models.Account{}).Error; err != nil {
        tx.Rollback()
        utils.Logger.Error("Failed to delete accounts",
            zap.Error(err),
            zap.Uint("user_id", userID),
        )
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete accounts"})
        return
    }

    // 7. Finally delete the user
    if err := tx.Delete(&models.User{}, userID).Error; err != nil {
        tx.Rollback()
        utils.Logger.Error("Failed to delete user account",
            zap.Error(err),
            zap.Uint("user_id", userID),
        )
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user account"})
        return
    }

    // Commit transaction
    if err := tx.Commit().Error; err != nil {
        utils.Logger.Error("Failed to commit account deletion transaction",
            zap.Error(err),
            zap.Uint("user_id", userID),
        )
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete account deletion"})
        return
    }

    utils.Logger.Info("Account deleted successfully",
        zap.Uint("user_id", userID),
        zap.String("username", user.Username),
        zap.String("email", user.Email),
        zap.String("ip", c.ClientIP()),
    )

    c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
}
