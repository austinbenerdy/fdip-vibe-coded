package models

import (
	"time"
	"encoding/json"
)

type Book struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	AuthorID      uint      `json:"author_id" gorm:"not null"`
	Title         string    `json:"title" gorm:"size:255;not null"`
	Description   *string   `json:"description"`
	CoverImageURL *string   `json:"cover_image_url" gorm:"size:500"`
	Genres        JSON      `json:"genres" gorm:"type:json"`
	IsPublished   bool      `json:"is_published" gorm:"default:false"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Relationships
	Author   User      `json:"author,omitempty" gorm:"foreignKey:AuthorID"`
	Chapters []Chapter `json:"chapters,omitempty" gorm:"foreignKey:BookID;order:chapter_number"`
}

// JSON type for handling JSON fields in GORM
type JSON []string

func (j JSON) Value() (interface{}, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, j)
	case string:
		return json.Unmarshal([]byte(v), j)
	default:
		return nil
	}
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