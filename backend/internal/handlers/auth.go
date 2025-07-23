package handlers

import (
	"net/http"
	"strconv"

	"fdip/internal/auth"
	"fdip/internal/database"
	"fdip/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRequest represents the registration request
type RegisterRequest struct {
	Username    string `json:"username" binding:"required,min=3,max=50"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=6"`
	DisplayName string `json:"display_name" binding:"required,min=1,max=100"`
	Bio         string `json:"bio"`
}

// LoginRequest represents the login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Register handles user registration
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if username already exists
	var existingUser models.User
	if err := database.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	// Check if email already exists
	if err := database.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		return
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create user
	user := models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		DisplayName:  req.DisplayName,
		Role:         models.RoleReader, // Default role
	}

	if req.Bio != "" {
		user.Bio = &req.Bio
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Create token balance for user
	balance := models.UserTokenBalance{
		UserID:      user.ID,
		Balance:     0,
		TotalEarned: 0,
		TotalSpent:  0,
	}

	if err := database.DB.Create(&balance).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token balance"})
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"token":   token,
		"user": gin.H{
			"id":           user.ID,
			"username":     user.Username,
			"email":        user.Email,
			"display_name": user.DisplayName,
			"role":         user.Role,
		},
	})
}

// Login handles user login
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user by username
	var user models.User
	if err := database.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Check password
	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
		"user": gin.H{
			"id":           user.ID,
			"username":     user.Username,
			"email":        user.Email,
			"display_name": user.DisplayName,
			"role":         user.Role,
		},
	})
}

// GetProfile returns the current user's profile
func GetProfile(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	currentUser := user.(*models.User)

	// Get token balance
	var balance models.UserTokenBalance
	if err := database.DB.Where("user_id = ?", currentUser.ID).First(&balance).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create balance if it doesn't exist
			balance = models.UserTokenBalance{
				UserID:      currentUser.ID,
				Balance:     0,
				TotalEarned: 0,
				TotalSpent:  0,
			}
			database.DB.Create(&balance)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get token balance"})
			return
		}
	}

	// Get follower and following counts
	followerCount, _ := models.GetFollowerCount(database.DB, currentUser.ID)
	followingCount, _ := models.GetFollowingCount(database.DB, currentUser.ID)

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":             currentUser.ID,
			"username":       currentUser.Username,
			"email":          currentUser.Email,
			"display_name":   currentUser.DisplayName,
			"bio":            currentUser.Bio,
			"avatar_url":     currentUser.AvatarURL,
			"role":           currentUser.Role,
			"created_at":     currentUser.CreatedAt,
			"token_balance":  balance.Balance,
			"total_earned":   balance.TotalEarned,
			"total_spent":    balance.TotalSpent,
			"follower_count": followerCount,
			"following_count": followingCount,
		},
	})
}

// UpdateProfile updates the current user's profile
func UpdateProfile(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	currentUser := user.(*models.User)

	var req struct {
		DisplayName string `json:"display_name" binding:"required,min=1,max=100"`
		Bio         string `json:"bio"`
		AvatarURL   string `json:"avatar_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update user
	updates := map[string]interface{}{
		"display_name": req.DisplayName,
	}

	if req.Bio != "" {
		updates["bio"] = req.Bio
	}

	if req.AvatarURL != "" {
		updates["avatar_url"] = req.AvatarURL
	}

	if err := database.DB.Model(&currentUser).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user": gin.H{
			"id":           currentUser.ID,
			"username":     currentUser.Username,
			"email":        currentUser.Email,
			"display_name": currentUser.DisplayName,
			"bio":          currentUser.Bio,
			"avatar_url":   currentUser.AvatarURL,
			"role":         currentUser.Role,
		},
	})
}

// PromoteToAuthor promotes a user to author role
func PromoteToAuthor(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if user.Role == models.RoleAuthor {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is already an author"})
		return
	}

	if err := database.DB.Model(&user).Update("role", models.RoleAuthor).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to promote user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User promoted to author successfully",
		"user": gin.H{
			"id":           user.ID,
			"username":     user.Username,
			"display_name": user.DisplayName,
			"role":         models.RoleAuthor,
		},
	})
} 