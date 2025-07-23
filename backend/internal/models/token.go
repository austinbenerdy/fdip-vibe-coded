package models

import (
	"time"
)

type TransactionType string

const (
	TransactionTypePurchase TransactionType = "purchase"
	TransactionTypeTip     TransactionType = "tip"
	TransactionTypeCashout TransactionType = "cashout"
	TransactionTypeRefund  TransactionType = "refund"
)

type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusFailed    TransactionStatus = "failed"
	TransactionStatusCancelled TransactionStatus = "cancelled"
)

type TokenTransaction struct {
	ID                    uint              `json:"id" gorm:"primaryKey"`
	UserID                uint              `json:"user_id" gorm:"not null"`
	TransactionType       TransactionType   `json:"transaction_type" gorm:"type:ENUM('purchase', 'tip', 'cashout', 'refund');not null"`
	Amount                int               `json:"amount" gorm:"not null"` // Positive for credits, negative for debits
	StripePaymentIntentID *string           `json:"stripe_payment_intent_id" gorm:"size:255"`
	StripeTransferID      *string           `json:"stripe_transfer_id" gorm:"size:255"`
	RecipientID           *uint             `json:"recipient_id"` // For tips
	ChapterID             *uint             `json:"chapter_id"`   // For tips
	Status                TransactionStatus `json:"status" gorm:"type:ENUM('pending', 'completed', 'failed', 'cancelled');default:'pending'"`
	CreatedAt             time.Time         `json:"created_at"`
	UpdatedAt             time.Time         `json:"updated_at"`

	// Relationships
	User      User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Recipient *User     `json:"recipient,omitempty" gorm:"foreignKey:RecipientID"`
	Chapter   *Chapter  `json:"chapter,omitempty" gorm:"foreignKey:ChapterID"`
}

type UserTokenBalance struct {
	UserID       uint      `json:"user_id" gorm:"primaryKey"`
	Balance      int       `json:"balance" gorm:"not null;default:0"`
	TotalEarned  int       `json:"total_earned" gorm:"not null;default:0"`
	TotalSpent   int       `json:"total_spent" gorm:"not null;default:0"`
	LastUpdated  time.Time `json:"last_updated" gorm:"autoUpdateTime"`

	// Relationships
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// Token economics constants
const (
	TokensPerDollar = 10
	MinPayoutRate   = 0.60 // $0.60 per 10 tokens
	MaxPayoutRate   = 0.90 // $0.90 per 10 tokens
)

// CalculatePayoutRate determines the payout rate based on author performance
// This is a simplified version - in production, you might want more sophisticated logic
func CalculatePayoutRate(totalEarnings int, followerCount int) float64 {
	// Base rate starts at minimum
	rate := MinPayoutRate
	
	// Increase rate based on earnings (up to a point)
	if totalEarnings > 1000 { // 1000 tokens = $100
		rate += 0.10
	}
	
	// Increase rate based on follower count
	if followerCount > 100 {
		rate += 0.10
	}
	
	// Cap at maximum rate
	if rate > MaxPayoutRate {
		rate = MaxPayoutRate
	}
	
	return rate
}

// CalculatePayoutAmount calculates the USD amount for a given token amount
func CalculatePayoutAmount(tokens int, payoutRate float64) float64 {
	return float64(tokens) * payoutRate / float64(TokensPerDollar)
}

// CalculateTokenCost calculates how many tokens a user gets for a given USD amount
func CalculateTokenCost(usdAmount float64) int {
	return int(usdAmount * float64(TokensPerDollar))
}

// CalculateUSDValue calculates the USD value of tokens at purchase rate
func CalculateUSDValue(tokens int) float64 {
	return float64(tokens) / float64(TokensPerDollar)
} 