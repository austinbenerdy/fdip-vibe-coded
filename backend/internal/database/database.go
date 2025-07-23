package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"fdip/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Config holds database configuration
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// NewConfig creates a new database config from environment variables
func NewConfig() *Config {
	return &Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "3306"),
		User:     getEnv("DB_USER", "fdip_user"),
		Password: getEnv("DB_PASSWORD", ""),
		DBName:   getEnv("DB_NAME", "fdip"),
	}
}

// Connect establishes a connection to the database
func Connect(config *Config) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.DBName,
	)

	// Configure GORM logger
	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	DB = db
	return nil
}

// AutoMigrate runs database migrations
func AutoMigrate() error {
	return DB.AutoMigrate(
		&models.User{},
		&models.Book{},
		&models.Chapter{},
		&models.ChapterVersion{},
		&models.TokenTransaction{},
		&models.UserTokenBalance{},
		&models.UserFollow{},
	)
}

// CreateIndexes creates additional indexes for performance
func CreateIndexes() error {
	// These indexes are already defined in the models, but we can add additional ones here
	// if needed for specific query patterns
	
	return nil
}

// SeedData populates the database with initial data
func SeedData() error {
	// Check if admin user already exists
	var adminCount int64
	DB.Model(&models.User{}).Where("role = ?", models.RoleAdmin).Count(&adminCount)
	
	if adminCount == 0 {
		// Create default admin user
		adminUser := models.User{
			Username:     "admin",
			Email:        "admin@fdip.com",
			PasswordHash: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // admin123
			Role:         models.RoleAdmin,
			DisplayName:  "System Administrator",
		}
		
		if err := DB.Create(&adminUser).Error; err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}
		
		// Create token balance for admin
		adminBalance := models.UserTokenBalance{
			UserID:      adminUser.ID,
			Balance:     0,
			TotalEarned: 0,
			TotalSpent:  0,
		}
		
		if err := DB.Create(&adminBalance).Error; err != nil {
			return fmt.Errorf("failed to create admin token balance: %w", err)
		}
		
		log.Println("Created default admin user (username: admin, password: admin123)")
	}
	
	return nil
}

// getEnv gets an environment variable with a fallback default value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
} 