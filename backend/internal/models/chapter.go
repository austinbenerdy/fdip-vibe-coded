package models

import (
	"time"
	"strings"
	"fmt"
)

type ContentType string

const (
	ContentTypeMarkdown ContentType = "markdown"
	ContentTypeHTML     ContentType = "html"
)

type Chapter struct {
	ID           uint        `json:"id" gorm:"primaryKey"`
	BookID       uint        `json:"book_id" gorm:"not null"`
	Title        string      `json:"title" gorm:"size:255;not null"`
	Content      string      `json:"content" gorm:"type:longtext;not null"`
	ContentType  ContentType `json:"content_type" gorm:"type:ENUM('markdown', 'html');default:'markdown'"`
	ImageURL     *string     `json:"image_url" gorm:"size:500"`
	ChapterNumber uint       `json:"chapter_number" gorm:"not null"`
	IsPublished  bool        `json:"is_published" gorm:"default:false"`
	IsPrivate    bool        `json:"is_private" gorm:"default:false"`
	WordCount    uint        `json:"word_count" gorm:"default:0"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`

	// Relationships
	Book     Book             `json:"book,omitempty" gorm:"foreignKey:BookID"`
	Versions []ChapterVersion `json:"versions,omitempty" gorm:"foreignKey:ChapterID"`
}

type ChapterVersion struct {
	ID           uint        `json:"id" gorm:"primaryKey"`
	ChapterID    uint        `json:"chapter_id" gorm:"not null"`
	Content      string      `json:"content" gorm:"type:longtext;not null"`
	ContentType  ContentType `json:"content_type" gorm:"type:ENUM('markdown', 'html');default:'markdown'"`
	VersionNumber uint       `json:"version_number" gorm:"not null"`
	CreatedAt    time.Time   `json:"created_at"`

	// Relationships
	Chapter Chapter `json:"chapter,omitempty" gorm:"foreignKey:ChapterID"`
}

func (c *Chapter) CalculateWordCount() {
	if c.ContentType == ContentTypeMarkdown {
		// Remove markdown syntax for word count
		content := c.Content
		// Remove headers
		content = strings.ReplaceAll(content, "#", "")
		// Remove bold/italic
		content = strings.ReplaceAll(content, "*", "")
		content = strings.ReplaceAll(content, "_", "")
		// Remove links
		content = strings.ReplaceAll(content, "[", "")
		content = strings.ReplaceAll(content, "]", "")
		// Remove code blocks
		content = strings.ReplaceAll(content, "```", "")
		// Remove inline code
		content = strings.ReplaceAll(content, "`", "")
		
		words := strings.Fields(content)
		c.WordCount = uint(len(words))
	} else {
		// For HTML, strip tags and count words
		content := c.Content
		// Simple HTML tag removal (in production, use proper HTML parser)
		content = strings.ReplaceAll(content, "<", " <")
		content = strings.ReplaceAll(content, ">", "> ")
		words := strings.Fields(content)
		c.WordCount = uint(len(words))
	}
}

func (c *Chapter) IsVisible() bool {
	return c.IsPublished && !c.IsPrivate
}

func (c *Chapter) GetDisplayTitle() string {
	if c.ChapterNumber > 0 {
		return fmt.Sprintf("Chapter %d: %s", c.ChapterNumber, c.Title)
	}
	return c.Title
} 