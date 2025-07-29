package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type Book struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	AuthorID      uint      `json:"author_id" gorm:"not null"`
	Title         string    `json:"title" gorm:"size:255;not null"`
	Description   *string   `json:"description"`
	CoverImageURL *string   `json:"cover_image_url" gorm:"size:500"`
	Genres        string    `json:"genres" gorm:"type:text;default:'[]'"`
	IsPublished   bool      `json:"is_published" gorm:"default:false"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Relationships
	Author   User      `json:"author,omitempty" gorm:"foreignKey:AuthorID"`
	Chapters []Chapter `json:"chapters,omitempty" gorm:"foreignKey:BookID;order:chapter_number"`
}

// GetGenres returns the genres as a string slice
func (b *Book) GetGenres() []string {
	if b.Genres == "" {
		return []string{}
	}
	var genres []string
	json.Unmarshal([]byte(b.Genres), &genres)
	return genres
}

// SetGenres sets the genres from a string slice
func (b *Book) SetGenres(genres []string) {
	if len(genres) == 0 {
		b.Genres = "[]"
		return
	}
	bytes, _ := json.Marshal(genres)
	b.Genres = string(bytes)
}

func (b *Book) GetPublishedChapters() []Chapter {
	var published []Chapter
	for _, chapter := range b.Chapters {
		if chapter.IsPublished && !chapter.IsPrivate {
			published = append(published, chapter)
		}
	}
	return published
}

func (b *Book) GetChapterCount() int {
	return len(b.Chapters)
}

func (b *Book) GetPublishedChapterCount() int {
	return len(b.GetPublishedChapters())
}

// MarshalJSON customizes the JSON output for the Book struct
func (b *Book) MarshalJSON() ([]byte, error) {
	type Alias Book
	return json.Marshal(&struct {
		*Alias
		Genres   []string  `json:"genres"`
		Chapters []Chapter `json:"chapters"`
	}{
		Alias:    (*Alias)(b),
		Genres:   b.GetGenres(),
		Chapters: b.Chapters,
	})
}

// AfterFind ensures chapters is never nil
func (b *Book) AfterFind(tx *gorm.DB) error {
	if b.Chapters == nil {
		b.Chapters = []Chapter{}
	}
	return nil
}
