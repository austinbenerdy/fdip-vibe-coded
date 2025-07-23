package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"fdip/internal/database"
	"fdip/internal/middleware"
	"fdip/internal/models"

	"os"

	"log"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/webhook"
	"gorm.io/gorm"
)

// PurchaseTokensRequest represents the token purchase request
type PurchaseTokensRequest struct {
	Amount float64 `json:"amount" binding:"required,min=1"` // Amount in USD
}

// TipRequest represents the tip request
type TipRequest struct {
	ChapterID uint `json:"chapter_id" binding:"required"`
	Amount    int  `json:"amount" binding:"required,min=1"` // Amount in tokens
}

// CashoutRequest represents the cashout request
type CashoutRequest struct {
	Amount int `json:"amount" binding:"required,min=10"` // Minimum 10 tokens
}

// GetTokenBalance returns the current user's token balance
func GetTokenBalance(c *gin.Context) {
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

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

	c.JSON(http.StatusOK, gin.H{
		"balance":      balance.Balance,
		"total_earned": balance.TotalEarned,
		"total_spent":  balance.TotalSpent,
	})
}

// PurchaseTokens handles token purchases via Stripe
func PurchaseTokens(c *gin.Context) {
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	var req PurchaseTokensRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Calculate tokens to be awarded
	tokensToAward := models.CalculateTokenCost(req.Amount)

	// Create Stripe payment intent
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(int64(req.Amount * 100)), // Convert to cents
		Currency: stripe.String("usd"),
		Metadata: map[string]string{
			"user_id":         strconv.FormatUint(uint64(currentUser.ID), 10),
			"tokens_to_award": strconv.Itoa(tokensToAward),
		},
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment intent"})
		return
	}

	// Create transaction record
	transaction := models.TokenTransaction{
		UserID:                currentUser.ID,
		TransactionType:       models.TransactionTypePurchase,
		Amount:                tokensToAward,
		StripePaymentIntentID: &pi.ID,
		Status:                models.TransactionStatusPending,
	}

	if err := database.DB.Create(&transaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"client_secret":     pi.ClientSecret,
		"payment_intent_id": pi.ID,
		"tokens_to_award":   tokensToAward,
	})
}

// TipAuthor handles tipping authors for their chapters
func TipAuthor(c *gin.Context) {
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	var req TipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the chapter and its author
	var chapter models.Chapter
	if err := database.DB.Preload("Book").Preload("Book.Author").
		Where("id = ? AND is_published = ? AND is_private = ?", req.ChapterID, true, false).
		First(&chapter).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Chapter not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch chapter"})
		return
	}

	// Check if user is trying to tip themselves
	if chapter.Book.AuthorID == currentUser.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot tip yourself"})
		return
	}

	// Check if user has enough tokens
	var userBalance models.UserTokenBalance
	if err := database.DB.Where("user_id = ?", currentUser.ID).First(&userBalance).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get token balance"})
		return
	}

	if userBalance.Balance < req.Amount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient token balance"})
		return
	}

	// Start transaction
	tx := database.DB.Begin()

	// Deduct tokens from user
	if err := tx.Model(&userBalance).Update("balance", userBalance.Balance-req.Amount).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deduct tokens"})
		return
	}

	// Add tokens to author
	var authorBalance models.UserTokenBalance
	if err := tx.Where("user_id = ?", chapter.Book.AuthorID).First(&authorBalance).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create balance if it doesn't exist
			authorBalance = models.UserTokenBalance{
				UserID:      chapter.Book.AuthorID,
				Balance:     0,
				TotalEarned: 0,
				TotalSpent:  0,
			}
			tx.Create(&authorBalance)
		} else {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get author balance"})
			return
		}
	}

	if err := tx.Model(&authorBalance).Updates(map[string]interface{}{
		"balance":      authorBalance.Balance + req.Amount,
		"total_earned": authorBalance.TotalEarned + req.Amount,
	}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add tokens to author"})
		return
	}

	// Create transaction records
	userTransaction := models.TokenTransaction{
		UserID:          currentUser.ID,
		TransactionType: models.TransactionTypeTip,
		Amount:          -req.Amount, // Negative for debit
		RecipientID:     &chapter.Book.AuthorID,
		ChapterID:       &req.ChapterID,
		Status:          models.TransactionStatusCompleted,
	}

	authorTransaction := models.TokenTransaction{
		UserID:          chapter.Book.AuthorID,
		TransactionType: models.TransactionTypeTip,
		Amount:          req.Amount, // Positive for credit
		RecipientID:     &currentUser.ID,
		ChapterID:       &req.ChapterID,
		Status:          models.TransactionStatusCompleted,
	}

	if err := tx.Create(&userTransaction).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user transaction"})
		return
	}

	if err := tx.Create(&authorTransaction).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create author transaction"})
		return
	}

	// Update user's total spent
	if err := tx.Model(&userBalance).Update("total_spent", userBalance.TotalSpent+req.Amount).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update total spent"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tip sent successfully",
		"amount":  req.Amount,
		"author":  chapter.Book.Author.DisplayName,
	})
}

// CashoutTokens handles token cashout to USD via Stripe
func CashoutTokens(c *gin.Context) {
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	var req CashoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user has enough tokens
	var userBalance models.UserTokenBalance
	if err := database.DB.Where("user_id = ?", currentUser.ID).First(&userBalance).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get token balance"})
		return
	}

	if userBalance.Balance < req.Amount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient token balance"})
		return
	}

	// Calculate payout rate based on author performance
	followerCount, _ := models.GetFollowerCount(database.DB, currentUser.ID)
	payoutRate := models.CalculatePayoutRate(userBalance.TotalEarned, int(followerCount))
	payoutAmount := models.CalculatePayoutAmount(req.Amount, payoutRate)

	// Create Stripe transfer (this would require the author to have a Stripe Connect account)
	// For now, we'll just create a transaction record
	transaction := models.TokenTransaction{
		UserID:          currentUser.ID,
		TransactionType: models.TransactionTypeCashout,
		Amount:          -req.Amount, // Negative for debit
		Status:          models.TransactionStatusPending,
	}

	if err := database.DB.Create(&transaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create cashout transaction"})
		return
	}

	// Deduct tokens from user balance
	if err := database.DB.Model(&userBalance).Update("balance", userBalance.Balance-req.Amount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update balance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":           "Cashout request submitted successfully",
		"tokens_cashed_out": req.Amount,
		"payout_amount_usd": payoutAmount,
		"payout_rate":       payoutRate,
		"status":            "pending",
	})
}

// GetTokenTransactions returns the current user's token transaction history
func GetTokenTransactions(c *gin.Context) {
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	var transactions []models.TokenTransaction
	if err := database.DB.Where("user_id = ?", currentUser.ID).
		Preload("Recipient").
		Preload("Chapter").
		Order("created_at DESC").
		Limit(50).
		Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}

	// Format transactions for frontend
	var formattedTransactions []gin.H
	for _, t := range transactions {
		transaction := gin.H{
			"id":               t.ID,
			"transaction_type": t.TransactionType,
			"amount":           t.Amount,
			"status":           t.Status,
			"created_at":       t.CreatedAt,
		}

		if t.Recipient != nil {
			transaction["recipient"] = gin.H{
				"display_name": t.Recipient.DisplayName,
			}
		}

		if t.Chapter != nil {
			transaction["chapter"] = gin.H{
				"title": t.Chapter.Title,
			}
		}

		formattedTransactions = append(formattedTransactions, transaction)
	}

	c.JSON(http.StatusOK, gin.H{
		"transactions": formattedTransactions,
	})
}

// HandleStripeWebhook handles Stripe webhook events
func HandleStripeWebhook(c *gin.Context) {
	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	payload, err := c.GetRawData()
	if err != nil {
		log.Printf("[WEBHOOK] Failed to read request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}
	sigHeader := c.GetHeader("Stripe-Signature")
	event, err := webhook.ConstructEventWithOptions(payload, sigHeader, webhookSecret, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})
	if err != nil {
		log.Printf("[WEBHOOK] Invalid webhook signature: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook signature"})
		return
	}

	if event.Type == "payment_intent.succeeded" {
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			log.Printf("[WEBHOOK] Failed to parse payment intent: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse payment intent"})
			return
		}
		userIDStr := pi.Metadata["user_id"]
		tokensStr := pi.Metadata["tokens_to_award"]
		userID, err1 := strconv.ParseUint(userIDStr, 10, 32)
		tokens, err2 := strconv.Atoi(tokensStr)
		if err1 != nil || err2 != nil {
			log.Printf("[WEBHOOK] Invalid metadata in payment intent: user_id=%s tokens_to_award=%s", userIDStr, tokensStr)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid metadata in payment intent"})
			return
		}
		// Credit tokens to user
		var balance models.UserTokenBalance
		if err := database.DB.Where("user_id = ?", userID).First(&balance).Error; err != nil {
			log.Printf("[WEBHOOK] Failed to get user balance: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user balance"})
			return
		}
		if err := database.DB.Model(&balance).Update("balance", balance.Balance+tokens).Error; err != nil {
			log.Printf("[WEBHOOK] Failed to credit tokens: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to credit tokens"})
			return
		}
		// Mark transaction as completed
		if err := database.DB.Model(&models.TokenTransaction{}).
			Where("stripe_payment_intent_id = ?", pi.ID).
			Update("status", models.TransactionStatusCompleted).Error; err != nil {
			log.Printf("[WEBHOOK] Failed to update transaction status: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update transaction status"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Webhook received"})
}
