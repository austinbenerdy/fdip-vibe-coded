package models

import (
	"time"
)

type UserRole string

const (
	RoleReader UserRole = "reader"
	RoleAuthor UserRole = "author"
	RoleAdmin  UserRole = "admin"
)

type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Username     string    `json:"username" gorm:"uniqueIndex;size:50;not null"`
	Email        string    `json:"email" gorm:"uniqueIndex;size:255;not null"`
	PasswordHash string    `json:"-" gorm:"size:255;not null"`
	Role         UserRole  `json:"role" gorm:"type:ENUM('reader', 'author', 'admin');default:'reader'"`
	DisplayName  string    `json:"display_name" gorm:"size:100;not null"`
	Bio          *string   `json:"bio"`
	AvatarURL    *string   `json:"avatar_url" gorm:"size:500"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	Books           []Book              `json:"books,omitempty" gorm:"foreignKey:AuthorID"`
	TokenBalance    *UserTokenBalance   `json:"token_balance,omitempty"`
	TokenTransactions []TokenTransaction `json:"token_transactions,omitempty" gorm:"foreignKey:UserID"`
	Followers       []UserFollow        `json:"followers,omitempty" gorm:"foreignKey:FollowedID"`
	Following       []UserFollow        `json:"following,omitempty" gorm:"foreignKey:FollowerID"`
}

func (u *User) IsAuthor() bool {
	return u.Role == RoleAuthor || u.Role == RoleAdmin
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func (u *User) CanAccessAuthorFeatures() bool {
	return u.IsAuthor()
}

func (u *User) CanAccessAdminFeatures() bool {
	return u.IsAdmin()
} 