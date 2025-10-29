package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds application configuration
type Config struct {
	// Telegram Bot
	TelegramBotToken string
	BotMode          string // "polling" or "webhook"
	WebhookURL       string
	WebhookSecret    string

	// Azure Communication Services
	AzureCommunicationConnectionString string
	SenderEmail                        string

	// Database
	DBType string // "memory", "postgres", or "cosmos"

	// PostgreSQL
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	DBSSLMode  string

	// Cosmos DB
	CosmosEndpoint  string
	CosmosKey       string
	CosmosDatabase  string
	CosmosContainer string

	// Application
	LogLevel string
	Port     string

	// Azure Application Insights
	AppInsightsInstrumentationKey string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Try to load .env file (optional)
	_ = godotenv.Load()

	cfg := &Config{
		TelegramBotToken:                   os.Getenv("TELEGRAM_BOT_TOKEN"),
		BotMode:                            getEnvOrDefault("BOT_MODE", "polling"),
		WebhookURL:                         os.Getenv("WEBHOOK_URL"),
		WebhookSecret:                      os.Getenv("WEBHOOK_SECRET"),
		AzureCommunicationConnectionString: os.Getenv("AZURE_COMMUNICATION_CONNECTION_STRING"),
		SenderEmail:                        os.Getenv("SENDER_EMAIL"),
		DBType:                             getEnvOrDefault("DB_TYPE", "memory"),
		DBHost:                             os.Getenv("DB_HOST"),
		DBPort:                             getEnvOrDefault("DB_PORT", "5432"),
		DBName:                             os.Getenv("DB_NAME"),
		DBUser:                             os.Getenv("DB_USER"),
		DBPassword:                         os.Getenv("DB_PASSWORD"),
		DBSSLMode:                          getEnvOrDefault("DB_SSL_MODE", "require"),
		CosmosEndpoint:                     os.Getenv("COSMOS_ENDPOINT"),
		CosmosKey:                          os.Getenv("COSMOS_KEY"),
		CosmosDatabase:                     os.Getenv("COSMOS_DATABASE"),
		CosmosContainer:                    os.Getenv("COSMOS_CONTAINER"),
		LogLevel:                           getEnvOrDefault("LOG_LEVEL", "info"),
		Port:                               getEnvOrDefault("PORT", "8080"),
		AppInsightsInstrumentationKey:      os.Getenv("APPINSIGHTS_INSTRUMENTATIONKEY"),
	}

	// Validate required fields
	if cfg.TelegramBotToken == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}

	if cfg.BotMode == "webhook" {
		if cfg.WebhookURL == "" {
			return nil, fmt.Errorf("WEBHOOK_URL is required for webhook mode")
		}
	}

	// Validate database configuration
	switch cfg.DBType {
	case "memory":
		// No additional validation needed
	case "postgres":
		if cfg.DBHost == "" || cfg.DBName == "" || cfg.DBUser == "" || cfg.DBPassword == "" {
			return nil, fmt.Errorf("postgres configuration incomplete: DB_HOST, DB_NAME, DB_USER, and DB_PASSWORD are required")
		}
	case "cosmos":
		if cfg.CosmosEndpoint == "" || cfg.CosmosKey == "" || cfg.CosmosDatabase == "" || cfg.CosmosContainer == "" {
			return nil, fmt.Errorf("cosmos configuration incomplete: COSMOS_ENDPOINT, COSMOS_KEY, COSMOS_DATABASE, and COSMOS_CONTAINER are required")
		}
	default:
		return nil, fmt.Errorf("invalid DB_TYPE: %s (must be 'memory', 'postgres', or 'cosmos')", cfg.DBType)
	}

	return cfg, nil
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
