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

// CreateBookRequest represents the book creation request
type CreateBookRequest struct {
	Title         string   `json:"title" binding:"required,min=1,max=255"`
	Description   string   `json:"description"`
	CoverImageURL string   `json:"cover_image_url"`
	Genres        []string `json:"genres"`
}

// UpdateBookRequest represents the book update request
type UpdateBookRequest struct {
	Title         string   `json:"title" binding:"required,min=1,max=255"`
	Description   string   `json:"description"`
	CoverImageURL string   `json:"cover_image_url"`
	Genres        []string `json:"genres"`
	IsPublished   *bool    `json:"is_published"`
}

// GetBooks returns all books for the authenticated user
func GetBooks(c *gin.Context) {
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	var books []models.Book
	query := database.DB.Preload("Author").Preload("Chapters")

	// If user is not admin, only show their own books
	if currentUser.Role != models.RoleAdmin {
		query = query.Where("author_id = ?", currentUser.ID)
	}

	if err := query.Find(&books).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch books"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"books": books})
}

// GetPublicBooks returns published books for public viewing
func GetPublicBooks(c *gin.Context) {
	var books []models.Book
	query := database.DB.Preload("Author").Preload("Chapters")

	// Only show published books
	query = query.Where("is_published = ?", true)

	// Apply filters
	if genre := c.Query("genre"); genre != "" {
		query = query.Where("JSON_CONTAINS(genres, ?)", genre)
	}

	if authorID := c.Query("author_id"); authorID != "" {
		query = query.Where("author_id = ?", authorID)
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	query = query.Offset(offset).Limit(limit).Order("created_at DESC")

	if err := query.Find(&books).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch books"})
		return
	}

	// Get total count for pagination
	var total int64
	database.DB.Model(&models.Book{}).Where("is_published = ?", true).Count(&total)

	c.JSON(http.StatusOK, gin.H{
		"books": books,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// CreateBook creates a new book
func CreateBook(c *gin.Context) {
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	var req CreateBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	book := models.Book{
		AuthorID: currentUser.ID,
		Title:    req.Title,
		Genres:   models.JSON(req.Genres),
	}

	if req.Description != "" {
		book.Description = &req.Description
	}

	if req.CoverImageURL != "" {
		book.CoverImageURL = &req.CoverImageURL
	}

	if err := database.DB.Create(&book).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create book"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Book created successfully",
		"book":    book,
	})
}

// GetBook returns a specific book
func GetBook(c *gin.Context) {
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

	var book models.Book
	query := database.DB.Preload("Author").Preload("Chapters")

	// If user is not admin, only allow access to their own books
	if currentUser.Role != models.RoleAdmin {
		query = query.Where("id = ? AND author_id = ?", bookID, currentUser.ID)
	} else {
		query = query.Where("id = ?", bookID)
	}

	if err := query.First(&book).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch book"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"book": book})
}

// GetPublicBook returns a published book for public viewing
func GetPublicBook(c *gin.Context) {
	bookID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID"})
		return
	}

	var book models.Book
	if err := database.DB.Preload("Author").Preload("Chapters").
		Where("id = ? AND is_published = ?", bookID, true).
		First(&book).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch book"})
		return
	}

	// Filter out private chapters
	publishedChapters := book.GetPublishedChapters()
	book.Chapters = publishedChapters

	c.JSON(http.StatusOK, gin.H{"book": book})
}

// UpdateBook updates a book
func UpdateBook(c *gin.Context) {
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

	var req UpdateBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the book
	var book models.Book
	query := database.DB.Where("id = ?", bookID)

	// If user is not admin, only allow updates to their own books
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

	// Update book
	updates := map[string]interface{}{
		"title": req.Title,
		"genres": models.JSON(req.Genres),
	}

	if req.Description != "" {
		updates["description"] = req.Description
	}

	if req.CoverImageURL != "" {
		updates["cover_image_url"] = req.CoverImageURL
	}

	if req.IsPublished != nil {
		updates["is_published"] = *req.IsPublished
	}

	if err := database.DB.Model(&book).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update book"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Book updated successfully",
		"book":    book,
	})
}

// DeleteBook deletes a book
func DeleteBook(c *gin.Context) {
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

	// Get the book
	var book models.Book
	query := database.DB.Where("id = ?", bookID)

	// If user is not admin, only allow deletion of their own books
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

	// Delete book (this will cascade to chapters due to foreign key constraint)
	if err := database.DB.Delete(&book).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete book"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Book deleted successfully"})
} 