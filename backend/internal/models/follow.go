package models

import (
	"time"
	"gorm.io/gorm"
)

type UserFollow struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	FollowerID uint      `json:"follower_id" gorm:"not null"`
	FollowedID uint      `json:"followed_id" gorm:"not null"`
	CreatedAt  time.Time `json:"created_at"`

	// Relationships
	Follower User `json:"follower,omitempty" gorm:"foreignKey:FollowerID"`
	Followed User `json:"followed,omitempty" gorm:"foreignKey:FollowedID"`
}

// TableName specifies the table name for UserFollow
func (UserFollow) TableName() string {
	return "user_follows"
}

// IsFollowing checks if a user is following another user
func IsFollowing(db *gorm.DB, followerID, followedID uint) (bool, error) {
	var count int64
	err := db.Model(&UserFollow{}).
		Where("follower_id = ? AND followed_id = ?", followerID, followedID).
		Count(&count).Error
	
	return count > 0, err
}

// GetFollowerCount returns the number of followers for a user
func GetFollowerCount(db *gorm.DB, userID uint) (int64, error) {
	var count int64
	err := db.Model(&UserFollow{}).
		Where("followed_id = ?", userID).
		Count(&count).Error
	
	return count, err
}

// GetFollowingCount returns the number of users a user is following
func GetFollowingCount(db *gorm.DB, userID uint) (int64, error) {
	var count int64
	err := db.Model(&UserFollow{}).
		Where("follower_id = ?", userID).
		Count(&count).Error
	
	return count, err
}

// GetFollowers returns the list of users following a specific user
func GetFollowers(db *gorm.DB, userID uint, limit, offset int) ([]User, error) {
	var followers []User
	err := db.Joins("JOIN user_follows ON users.id = user_follows.follower_id").
		Where("user_follows.followed_id = ?", userID).
		Order("user_follows.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&followers).Error
	
	return followers, err
}

// GetFollowing returns the list of users a specific user is following
func GetFollowing(db *gorm.DB, userID uint, limit, offset int) ([]User, error) {
	var following []User
	err := db.Joins("JOIN user_follows ON users.id = user_follows.followed_id").
		Where("user_follows.follower_id = ?", userID).
		Order("user_follows.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&following).Error
	
	return following, err
} 