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

// CreateChapterRequest represents the chapter creation request
type CreateChapterRequest struct {
	Title         string                `json:"title" binding:"required,min=1,max=255"`
	Content       string                `json:"content" binding:"required"`
	ContentType   models.ContentType    `json:"content_type" binding:"required"`
	ImageURL      string                `json:"image_url"`
	ChapterNumber uint                  `json:"chapter_number" binding:"required"`
	IsPublished   bool                  `json:"is_published"`
	IsPrivate     bool                  `json:"is_private"`
}

// UpdateChapterRequest represents the chapter update request
type UpdateChapterRequest struct {
	Title         string                `json:"title" binding:"required,min=1,max=255"`
	Content       string                `json:"content" binding:"required"`
	ContentType   models.ContentType    `json:"content_type" binding:"required"`
	ImageURL      string                `json:"image_url"`
	ChapterNumber uint                  `json:"chapter_number" binding:"required"`
	IsPublished   bool                  `json:"is_published"`
	IsPrivate     bool                  `json:"is_private"`
}

// GetChapters returns all chapters for a book
func GetChapters(c *gin.Context) {
	bookID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID"})
		return
	}

	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Check if user has access to this book
	var book models.Book
	query := database.DB.Where("id = ?", bookID)
	if currentUser.Role != models.RoleAdmin {
		query = query.Where("author_id = ?", currentUser.ID)
	}

	if err := query.First(&book).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch book"})
		return
	}

	var chapters []models.Chapter
	if err := database.DB.Where("book_id = ?", bookID).
		Order("chapter_number ASC").
		Find(&chapters).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch chapters"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"chapters": chapters})
}

// CreateChapter creates a new chapter
func CreateChapter(c *gin.Context) {
	bookID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID"})
		return
	}

	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	var req CreateChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user has access to this book
	var book models.Book
	query := database.DB.Where("id = ?", bookID)
	if currentUser.Role != models.RoleAdmin {
		query = query.Where("author_id = ?", currentUser.ID)
	}

	if err := query.First(&book).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch book"})
		return
	}

	// Check if chapter number already exists
	var existingChapter models.Chapter
	if err := database.DB.Where("book_id = ? AND chapter_number = ?", bookID, req.ChapterNumber).
		First(&existingChapter).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Chapter number already exists"})
		return
	}

	chapter := models.Chapter{
		BookID:       uint(bookID),
		Title:        req.Title,
		Content:      req.Content,
		ContentType:  req.ContentType,
		ChapterNumber: req.ChapterNumber,
		IsPublished:  req.IsPublished,
		IsPrivate:    req.IsPrivate,
	}

	if req.ImageURL != "" {
		chapter.ImageURL = &req.ImageURL
	}

	// Calculate word count
	chapter.CalculateWordCount()

	if err := database.DB.Create(&chapter).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create chapter"})
		return
	}

	// Create initial version if chapter is published
	if chapter.IsPublished {
		version := models.ChapterVersion{
			ChapterID:     chapter.ID,
			Content:       chapter.Content,
			ContentType:   chapter.ContentType,
			VersionNumber: 1,
		}
		database.DB.Create(&version)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Chapter created successfully",
		"chapter": chapter,
	})
}

// GetChapter returns a specific chapter
func GetChapter(c *gin.Context) {
	chapterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chapter ID"})
		return
	}

	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	var chapter models.Chapter
	query := database.DB.Preload("Book").Preload("Book.Author").Where("id = ?", chapterID)

	// If user is not admin, check if they own the book
	if currentUser.Role != models.RoleAdmin {
		query = query.Joins("JOIN books ON chapters.book_id = books.id").
			Where("books.author_id = ?", currentUser.ID)
	}

	if err := query.First(&chapter).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Chapter not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch chapter"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"chapter": chapter})
}

// GetPublicChapter returns a published chapter for public viewing
func GetPublicChapter(c *gin.Context) {
	chapterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chapter ID"})
		return
	}

	var chapter models.Chapter
	if err := database.DB.Preload("Book").Preload("Book.Author").
		Where("id = ? AND is_published = ? AND is_private = ?", chapterID, true, false).
		First(&chapter).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Chapter not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch chapter"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"chapter": chapter})
}

// UpdateChapter updates a chapter
func UpdateChapter(c *gin.Context) {
	chapterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chapter ID"})
		return
	}

	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	var req UpdateChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the chapter
	var chapter models.Chapter
	query := database.DB.Preload("Book").Where("id = ?", chapterID)

	// If user is not admin, check if they own the book
	if currentUser.Role != models.RoleAdmin {
		query = query.Joins("JOIN books ON chapters.book_id = books.id").
			Where("books.author_id = ?", currentUser.ID)
	}

	if err := query.First(&chapter).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Chapter not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch chapter"})
		return
	}

	// Check if new chapter number conflicts with existing chapter
	if req.ChapterNumber != chapter.ChapterNumber {
		var existingChapter models.Chapter
		if err := database.DB.Where("book_id = ? AND chapter_number = ? AND id != ?", 
			chapter.BookID, req.ChapterNumber, chapterID).First(&existingChapter).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Chapter number already exists"})
			return
		}
	}

	// Create version if chapter is being published and wasn't published before
	wasPublished := chapter.IsPublished
	if req.IsPublished && !wasPublished {
		// Get the latest version number
		var maxVersion uint
		database.DB.Model(&models.ChapterVersion{}).
			Where("chapter_id = ?", chapter.ID).
			Select("COALESCE(MAX(version_number), 0)").
			Scan(&maxVersion)

		version := models.ChapterVersion{
			ChapterID:     chapter.ID,
			Content:       req.Content,
			ContentType:   req.ContentType,
			VersionNumber: maxVersion + 1,
		}
		database.DB.Create(&version)
	}

	// Update chapter
	updates := map[string]interface{}{
		"title":          req.Title,
		"content":        req.Content,
		"content_type":   req.ContentType,
		"chapter_number": req.ChapterNumber,
		"is_published":   req.IsPublished,
		"is_private":     req.IsPrivate,
	}

	if req.ImageURL != "" {
		updates["image_url"] = req.ImageURL
	}

	// Calculate word count
	tempChapter := models.Chapter{
		Content:     req.Content,
		ContentType: req.ContentType,
	}
	tempChapter.CalculateWordCount()
	updates["word_count"] = tempChapter.WordCount

	if err := database.DB.Model(&chapter).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update chapter"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Chapter updated successfully",
		"chapter": chapter,
	})
}

// DeleteChapter deletes a chapter
func DeleteChapter(c *gin.Context) {
	chapterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chapter ID"})
		return
	}

	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Get the chapter
	var chapter models.Chapter
	query := database.DB.Preload("Book").Where("id = ?", chapterID)

	// If user is not admin, check if they own the book
	if currentUser.Role != models.RoleAdmin {
		query = query.Joins("JOIN books ON chapters.book_id = books.id").
			Where("books.author_id = ?", currentUser.ID)
	}

	if err := query.First(&chapter).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Chapter not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch chapter"})
		return
	}

	// Delete chapter (this will cascade to versions due to foreign key constraint)
	if err := database.DB.Delete(&chapter).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete chapter"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Chapter deleted successfully"})
} 