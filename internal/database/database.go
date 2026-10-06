package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"backend-test/internal/entity"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config stores PostgreSQL connection configuration
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
	Timezone string
	URL      string
}

// LoadConfigFromEnv loads database configuration from environment variables
func LoadConfigFromEnv() Config {
	url := os.Getenv("DATABASE_URL")
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = getEnv("DB_PASS", "postgres")
	}
	dbname := getEnv("DB_NAME", "geospatial_db")
	sslmode := getEnv("DB_SSLMODE", "disable")
	timezone := getEnv("DB_TIMEZONE", "UTC")

	return Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbname,
		SSLMode:  sslmode,
		Timezone: timezone,
		URL:      url,
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// Connect initializes and returns a GORM database connection
func Connect(cfg Config) (*gorm.DB, error) {
	var dsn string
	if cfg.URL != "" {
		dsn = cfg.URL
	} else {
		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode, cfg.Timezone,
		)
	}

	gormLogLevel := logger.Warn
	if os.Getenv("ENV") == "development" || os.Getenv("GIN_MODE") != "release" {
		gormLogLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Auto-migrate tables
	if err := db.AutoMigrate(&entity.Entity{}); err != nil {
		return nil, fmt.Errorf("failed to auto-migrate schema: %w", err)
	}
	log.Println("Database migration completed successfully.")

	return db, nil
}
