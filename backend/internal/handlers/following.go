package handlers

import (
	"net/http"
	"strconv"

	"fdip/internal/database"
	"fdip/internal/middleware"
	"fdip/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetFollowing returns the list of authors the current user is following
func GetFollowing(c *gin.Context) {
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Get pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	following, err := models.GetFollowing(database.DB, currentUser.ID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get following list"})
		return
	}

	// Get total count for pagination
	followingCount, err := models.GetFollowingCount(database.DB, currentUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get following count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"following": following,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": followingCount,
		},
	})
}

// FollowAuthor allows a user to follow an author
func FollowAuthor(c *gin.Context) {
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	authorID, err := strconv.ParseUint(c.Param("authorId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid author ID"})
		return
	}

	// Check if user is trying to follow themselves
	if currentUser.ID == uint(authorID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot follow yourself"})
		return
	}

	// Check if author exists and is actually an author
	var author models.User
	if err := database.DB.Where("id = ? AND (role = ? OR role = ?)", authorID, models.RoleAuthor, models.RoleAdmin).First(&author).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Author not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch author"})
		return
	}

	// Check if already following
	isFollowing, err := models.IsFollowing(database.DB, currentUser.ID, uint(authorID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check following status"})
		return
	}

	if isFollowing {
		c.JSON(http.StatusConflict, gin.H{"error": "Already following this author"})
		return
	}

	// Create follow relationship
	follow := models.UserFollow{
		FollowerID: currentUser.ID,
		FollowedID: uint(authorID),
	}

	if err := database.DB.Create(&follow).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to follow author"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Successfully followed author",
		"author": gin.H{
			"id":           author.ID,
			"username":     author.Username,
			"display_name": author.DisplayName,
		},
	})
}

// UnfollowAuthor allows a user to unfollow an author
func UnfollowAuthor(c *gin.Context) {
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	authorID, err := strconv.ParseUint(c.Param("authorId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid author ID"})
		return
	}

	// Check if following
	isFollowing, err := models.IsFollowing(database.DB, currentUser.ID, uint(authorID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check following status"})
		return
	}

	if !isFollowing {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not following this author"})
		return
	}

	// Delete follow relationship
	if err := database.DB.Where("follower_id = ? AND followed_id = ?", currentUser.ID, authorID).
		Delete(&models.UserFollow{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unfollow author"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully unfollowed author"})
}

// GetAuthors returns a list of authors for public viewing
func GetAuthors(c *gin.Context) {
	// Get pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	var authors []models.User
	query := database.DB.Where("role IN (?, ?)", models.RoleAuthor, models.RoleAdmin)

	// Apply filters
	if search := c.Query("search"); search != "" {
		query = query.Where("display_name LIKE ? OR username LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Get total count for pagination
	var total int64
	query.Model(&models.User{}).Count(&total)

	// Get authors with pagination
	if err := query.Offset(offset).Limit(limit).Order("display_name ASC").Find(&authors).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch authors"})
		return
	}

	// Get follower counts for each author
	var authorData []gin.H
	for _, author := range authors {
		followerCount, _ := models.GetFollowerCount(database.DB, author.ID)
		
		// Check if current user is following this author
		var isFollowing bool
		if currentUser, exists := middleware.GetCurrentUser(c); exists {
			isFollowing, _ = models.IsFollowing(database.DB, currentUser.ID, author.ID)
		}

		authorData = append(authorData, gin.H{
			"id":             author.ID,
			"username":       author.Username,
			"display_name":   author.DisplayName,
			"bio":            author.Bio,
			"avatar_url":     author.AvatarURL,
			"follower_count": followerCount,
			"is_following":   isFollowing,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"authors": authorData,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetAuthor returns a specific author's profile
func GetAuthor(c *gin.Context) {
	authorID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid author ID"})
		return
	}

	var author models.User
	if err := database.DB.Where("id = ? AND (role = ? OR role = ?)", authorID, models.RoleAuthor, models.RoleAdmin).First(&author).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Author not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch author"})
		return
	}

	// Get follower and following counts
	followerCount, _ := models.GetFollowerCount(database.DB, author.ID)
	followingCount, _ := models.GetFollowingCount(database.DB, author.ID)

	// Get published books count
	var bookCount int64
	database.DB.Model(&models.Book{}).Where("author_id = ? AND is_published = ?", author.ID, true).Count(&bookCount)

	// Check if current user is following this author
	var isFollowing bool
	if currentUser, exists := middleware.GetCurrentUser(c); exists {
		isFollowing, _ = models.IsFollowing(database.DB, currentUser.ID, author.ID)
	}

	// Get recent published books
	var recentBooks []models.Book
	database.DB.Where("author_id = ? AND is_published = ?", author.ID, true).
		Order("created_at DESC").
		Limit(5).
		Find(&recentBooks)

	c.JSON(http.StatusOK, gin.H{
		"author": gin.H{
			"id":             author.ID,
			"username":       author.Username,
			"display_name":   author.DisplayName,
			"bio":            author.Bio,
			"avatar_url":     author.AvatarURL,
			"created_at":     author.CreatedAt,
			"follower_count": followerCount,
			"following_count": followingCount,
			"book_count":     bookCount,
			"is_following":   isFollowing,
			"recent_books":   recentBooks,
		},
	})
} 