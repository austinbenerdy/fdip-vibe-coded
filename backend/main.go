package main

import (
	"log"
	"os"

	"fdip/internal/auth"
	"fdip/internal/database"
	"fdip/internal/handlers"
	"fdip/internal/middleware"
	"fdip/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v76"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Set Stripe API key
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	// Initialize JWT
	if err := auth.InitJWT(); err != nil {
		log.Fatal("Failed to initialize JWT:", err)
	}

	// Initialize database
	dbConfig := database.NewConfig()
	if err := database.Connect(dbConfig); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	// Run migrations
	if err := database.AutoMigrate(); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Seed initial data
	if err := database.SeedData(); err != nil {
		log.Fatal("Failed to seed data:", err)
	}

	// Set Gin mode
	if os.Getenv("ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// API routes
	api := r.Group("/api")
	{
		// Authentication routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			// User profile
			protected.GET("/profile", handlers.GetProfile)
			protected.PUT("/profile", handlers.UpdateProfile)

			// Books routes (protected - for authors/admins)
			books := protected.Group("/books")
			{
				books.POST("", middleware.RequireAuthorOrAdmin(), handlers.CreateBook)
				books.PUT("/:id", middleware.RequireAuthorOrAdmin(), handlers.UpdateBook)
				books.DELETE("/:id", middleware.RequireAuthorOrAdmin(), handlers.DeleteBook)

				// Chapters routes
				chapters := books.Group("/:id/chapters")
				{
					chapters.GET("", handlers.GetChapters)
					chapters.POST("", middleware.RequireAuthorOrAdmin(), handlers.CreateChapter)
				}
			}

			// Chapter management routes (protected - for authors/admins)
			chapters := protected.Group("/chapters")
			{
				chapters.PUT("/:id", middleware.RequireAuthorOrAdmin(), handlers.UpdateChapter)
				chapters.DELETE("/:id", middleware.RequireAuthorOrAdmin(), handlers.DeleteChapter)
			}

			// Token routes
			tokens := protected.Group("/tokens")
			{
				tokens.GET("/balance", handlers.GetTokenBalance)
				tokens.GET("/transactions", handlers.GetTokenTransactions)
				tokens.POST("/purchase", handlers.PurchaseTokens)
				tokens.POST("/tip", handlers.TipAuthor)
				tokens.POST("/cashout", middleware.RequireAuthorOrAdmin(), handlers.CashoutTokens)
			}

			// Following routes
			following := protected.Group("/following")
			{
				following.GET("", handlers.GetFollowing)
				following.POST("/:authorId", handlers.FollowAuthor)
				following.DELETE("/:authorId", handlers.UnfollowAuthor)
			}

			// Admin routes
			admin := protected.Group("/admin")
			admin.Use(middleware.RequireRole(models.RoleAdmin))
			{
				admin.POST("/users/:id/promote", handlers.PromoteToAuthor)
			}
		}

		// Public routes (with optional auth)
		public := api.Group("")
		public.Use(middleware.OptionalAuthMiddleware())
		{
			public.GET("/books", handlers.GetPublicBooks)
			public.GET("/books/:id", handlers.GetPublicBook)
			public.GET("/chapters/:id", handlers.GetPublicChapter)
			public.GET("/authors", handlers.GetAuthors)
			public.GET("/authors/:id", handlers.GetAuthor)
		}

		// Stripe webhook (no auth)
		api.POST("/stripe/webhook", handlers.HandleStripeWebhook)
	}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
