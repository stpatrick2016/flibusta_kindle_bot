package config

import (
	"os"
	"testing"
)

func TestLoad_Success(t *testing.T) {
	// Set required environment variables
	os.Setenv("TELEGRAM_BOT_TOKEN", "test_token_123")
	os.Setenv("DB_TYPE", "memory")
	defer os.Unsetenv("TELEGRAM_BOT_TOKEN")
	defer os.Unsetenv("DB_TYPE")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}

	if cfg.TelegramBotToken != "test_token_123" {
		t.Errorf("TelegramBotToken = %v, want %v", cfg.TelegramBotToken, "test_token_123")
	}

	if cfg.DBType != "memory" {
		t.Errorf("DBType = %v, want %v", cfg.DBType, "memory")
	}
}

func TestLoad_MissingBotToken(t *testing.T) {
	// Clear any existing token
	os.Unsetenv("TELEGRAM_BOT_TOKEN")

	_, err := Load()
	if err == nil {
		t.Error("Expected error for missing TELEGRAM_BOT_TOKEN, got nil")
	}

	expectedMsg := "TELEGRAM_BOT_TOKEN is required"
	if err.Error() != expectedMsg {
		t.Errorf("Error message = %v, want %v", err.Error(), expectedMsg)
	}
}

func TestLoad_DefaultValues(t *testing.T) {
	os.Setenv("TELEGRAM_BOT_TOKEN", "test_token")
	defer os.Unsetenv("TELEGRAM_BOT_TOKEN")

	// Clear optional variables
	os.Unsetenv("BOT_MODE")
	os.Unsetenv("DB_TYPE")
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("PORT")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_SSL_MODE")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Check default values
	if cfg.BotMode != "polling" {
		t.Errorf("BotMode = %v, want %v", cfg.BotMode, "polling")
	}

	if cfg.DBType != "memory" {
		t.Errorf("DBType = %v, want %v", cfg.DBType, "memory")
	}

	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %v, want %v", cfg.LogLevel, "info")
	}

	if cfg.Port != "8080" {
		t.Errorf("Port = %v, want %v", cfg.Port, "8080")
	}

	if cfg.DBPort != "5432" {
		t.Errorf("DBPort = %v, want %v", cfg.DBPort, "5432")
	}

	if cfg.DBSSLMode != "require" {
		t.Errorf("DBSSLMode = %v, want %v", cfg.DBSSLMode, "require")
	}
}

func TestLoad_WebhookMode_RequiresURL(t *testing.T) {
	os.Setenv("TELEGRAM_BOT_TOKEN", "test_token")
	os.Setenv("BOT_MODE", "webhook")
	defer os.Unsetenv("TELEGRAM_BOT_TOKEN")
	defer os.Unsetenv("BOT_MODE")
	defer os.Unsetenv("WEBHOOK_URL")

	// Without WEBHOOK_URL
	_, err := Load()
	if err == nil {
		t.Error("Expected error for missing WEBHOOK_URL in webhook mode, got nil")
	}

	// With WEBHOOK_URL
	os.Setenv("WEBHOOK_URL", "https://example.com/webhook")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.WebhookURL != "https://example.com/webhook" {
		t.Errorf("WebhookURL = %v, want %v", cfg.WebhookURL, "https://example.com/webhook")
	}
}

func TestLoad_PostgresValidation(t *testing.T) {
	os.Setenv("TELEGRAM_BOT_TOKEN", "test_token")
	os.Setenv("DB_TYPE", "postgres")
	defer os.Unsetenv("TELEGRAM_BOT_TOKEN")
	defer os.Unsetenv("DB_TYPE")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
	}()

	// Missing postgres configuration
	_, err := Load()
	if err == nil {
		t.Error("Expected error for incomplete postgres configuration, got nil")
	}

	// Complete postgres configuration
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.DBHost != "localhost" {
		t.Errorf("DBHost = %v, want %v", cfg.DBHost, "localhost")
	}
	if cfg.DBName != "testdb" {
		t.Errorf("DBName = %v, want %v", cfg.DBName, "testdb")
	}
	if cfg.DBUser != "testuser" {
		t.Errorf("DBUser = %v, want %v", cfg.DBUser, "testuser")
	}
	if cfg.DBPassword != "testpass" {
		t.Errorf("DBPassword = %v, want %v", cfg.DBPassword, "testpass")
	}
}

func TestLoad_CosmosValidation(t *testing.T) {
	os.Setenv("TELEGRAM_BOT_TOKEN", "test_token")
	os.Setenv("DB_TYPE", "cosmos")
	defer os.Unsetenv("TELEGRAM_BOT_TOKEN")
	defer os.Unsetenv("DB_TYPE")
	defer func() {
		os.Unsetenv("COSMOS_ENDPOINT")
		os.Unsetenv("COSMOS_KEY")
		os.Unsetenv("COSMOS_DATABASE")
		os.Unsetenv("COSMOS_CONTAINER")
	}()

	// Missing cosmos configuration
	_, err := Load()
	if err == nil {
		t.Error("Expected error for incomplete cosmos configuration, got nil")
	}

	// Complete cosmos configuration
	os.Setenv("COSMOS_ENDPOINT", "https://test.documents.azure.com")
	os.Setenv("COSMOS_KEY", "test_key")
	os.Setenv("COSMOS_DATABASE", "testdb")
	os.Setenv("COSMOS_CONTAINER", "users")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.CosmosEndpoint != "https://test.documents.azure.com" {
		t.Errorf("CosmosEndpoint = %v, want %v", cfg.CosmosEndpoint, "https://test.documents.azure.com")
	}
	if cfg.CosmosKey != "test_key" {
		t.Errorf("CosmosKey = %v, want %v", cfg.CosmosKey, "test_key")
	}
}

func TestLoad_InvalidDBType(t *testing.T) {
	os.Setenv("TELEGRAM_BOT_TOKEN", "test_token")
	os.Setenv("DB_TYPE", "invalid_db")
	defer os.Unsetenv("TELEGRAM_BOT_TOKEN")
	defer os.Unsetenv("DB_TYPE")

	_, err := Load()
	if err == nil {
		t.Error("Expected error for invalid DB_TYPE, got nil")
	}
}

func TestGetEnvOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "use environment value",
			key:          "TEST_VAR",
			defaultValue: "default",
			envValue:     "from_env",
			expected:     "from_env",
		},
		{
			name:         "use default when empty",
			key:          "TEST_VAR_EMPTY",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}

			result := getEnvOrDefault(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvOrDefault() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestLoad_AllOptionalFields(t *testing.T) {
	os.Setenv("TELEGRAM_BOT_TOKEN", "test_token")
	os.Setenv("BOT_MODE", "webhook")
	os.Setenv("WEBHOOK_URL", "https://example.com/webhook")
	os.Setenv("WEBHOOK_SECRET", "secret123")
	os.Setenv("AZURE_COMMUNICATION_CONNECTION_STRING", "connection_string")
	os.Setenv("SENDER_EMAIL", "bot@example.com")
	os.Setenv("APPINSIGHTS_INSTRUMENTATIONKEY", "insights_key")

	defer func() {
		os.Unsetenv("TELEGRAM_BOT_TOKEN")
		os.Unsetenv("BOT_MODE")
		os.Unsetenv("WEBHOOK_URL")
		os.Unsetenv("WEBHOOK_SECRET")
		os.Unsetenv("AZURE_COMMUNICATION_CONNECTION_STRING")
		os.Unsetenv("SENDER_EMAIL")
		os.Unsetenv("APPINSIGHTS_INSTRUMENTATIONKEY")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.WebhookSecret != "secret123" {
		t.Errorf("WebhookSecret = %v, want %v", cfg.WebhookSecret, "secret123")
	}
	if cfg.AzureCommunicationConnectionString != "connection_string" {
		t.Errorf("AzureCommunicationConnectionString = %v", cfg.AzureCommunicationConnectionString)
	}
	if cfg.SenderEmail != "bot@example.com" {
		t.Errorf("SenderEmail = %v, want %v", cfg.SenderEmail, "bot@example.com")
	}
	if cfg.AppInsightsInstrumentationKey != "insights_key" {
		t.Errorf("AppInsightsInstrumentationKey = %v, want %v", cfg.AppInsightsInstrumentationKey, "insights_key")
	}
}
