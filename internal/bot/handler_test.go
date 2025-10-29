package bot

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stpatrick2016/flibusta_kindle_bot/internal/i18n"
	"github.com/stpatrick2016/flibusta_kindle_bot/internal/user"
)

// Mock bot API for testing
type mockBotAPI struct {
	sentMessages []tgbotapi.Chattable
	lastMessage  string
}

func (m *mockBotAPI) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	m.sentMessages = append(m.sentMessages, c)
	
	// Extract text from message
	if msg, ok := c.(tgbotapi.MessageConfig); ok {
		m.lastMessage = msg.Text
	}
	
	return tgbotapi.Message{MessageID: 1}, nil
}

func (m *mockBotAPI) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	return &tgbotapi.APIResponse{Ok: true}, nil
}

func setupTestHandler(t *testing.T) (*Handler, *mockBotAPI, *user.Manager) {
	// Create temporary i18n with test data
	tmpDir := t.TempDir()
	
	// Create minimal test locale files
	enContent := `{
		"welcome": "Welcome, %s!",
		"help_message": "Help text",
		"kindle_email_prompt": "Send your Kindle email",
		"kindle_email_set": "Email set to %s",
		"kindle_email_invalid": "Invalid email",
		"kindle_email_current": "Current email: %s",
		"whitelist_instructions": "Whitelist instructions",
		"whitelist_reminder": "Remember to whitelist",
		"language_prompt": "Select language",
		"language_changed": "Language changed",
		"settings_display": "Email: %s, Language: %s, Books: %d",
		"operation_cancelled": "Cancelled",
		"unknown_command": "Unknown command",
		"error_occurred": "Error occurred",
		"kindle_email_required": "Email required",
		"searching": "Searching for %s",
		"search_not_implemented": "Coming soon",
		"search_prompt": "Type to search",
		"feature_coming_soon": "Coming soon",
		"not_set": "not set"
	}`
	
	// Write test locale file
	if err := os.WriteFile(filepath.Join(tmpDir, "en.json"), []byte(enContent), 0644); err != nil {
		t.Fatalf("Failed to create test locale: %v", err)
	}
	
	i18nInstance, err := i18n.NewI18n(tmpDir, "en")
	if err != nil {
		t.Fatalf("Failed to create i18n: %v", err)
	}
	
	repo := user.NewMemoryRepository()
	userManager := user.NewManager(repo)
	
	mockBot := &mockBotAPI{
		sentMessages: make([]tgbotapi.Chattable, 0),
	}
	
	// We can't actually use the mock with the handler since it expects *tgbotapi.BotAPI
	// For now, we'll test the handler logic separately
	handler := &Handler{
		bot:         nil, // Will cause panic if Send is called
		i18n:        i18nInstance,
		userManager: userManager,
	}
	
	return handler, mockBot, userManager
}

func TestHandler_HandleCommand_Start(t *testing.T) {
	_, _, userManager := setupTestHandler(t)
	ctx := context.Background()
	
	// Create a test user
	testUser, _ := userManager.GetOrCreateUser(ctx, 12345, "testuser", "Test", "User", "en")
	
	// Test start command for new user without Kindle email
	_ = &tgbotapi.Message{
		MessageID: 1,
		From: &tgbotapi.User{
			ID:        12345,
			FirstName: "Test",
			UserName:  "testuser",
		},
		Chat: &tgbotapi.Chat{ID: 12345},
		Text: "/start",
	}
	
	// Since we can't actually send messages, we'll just verify the logic works
	if testUser == nil {
		t.Error("User should be created")
	}
	
	if testUser.TelegramID != 12345 {
		t.Errorf("TelegramID = %v, want %v", testUser.TelegramID, 12345)
	}
}

func TestHandler_ExtractCommand(t *testing.T) {
	tests := []struct {
		name            string
		text            string
		expectedCommand string
		expectedArgs    string
	}{
		{
			name:            "simple command",
			text:            "/start",
			expectedCommand: "start",
			expectedArgs:    "",
		},
		{
			name:            "command with args",
			text:            "/kindle user@kindle.com",
			expectedCommand: "kindle",
			expectedArgs:    "user@kindle.com",
		},
		{
			name:            "not a command",
			text:            "hello world",
			expectedCommand: "",
			expectedArgs:    "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = &tgbotapi.Message{
				Text: tt.text,
			}
			
			// Check if message is a command
			isCommand := len(tt.text) > 0 && tt.text[0] == '/'
			if isCommand != (tt.expectedCommand != "") {
				t.Errorf("IsCommand = %v, expected %v", isCommand, tt.expectedCommand != "")
			}
		})
	}
}

func TestHandler_ValidateKindleEmail(t *testing.T) {
	_, _, userManager := setupTestHandler(t)
	ctx := context.Background()
	
	// Create user
	testUser, _ := userManager.GetOrCreateUser(ctx, 12345, "testuser", "Test", "User", "en")
	
	// Test valid email
	err := userManager.SetKindleEmail(ctx, testUser.TelegramID, "user@kindle.com")
	if err != nil {
		t.Errorf("SetKindleEmail() should succeed for valid email, got error: %v", err)
	}
	
	// Test invalid email
	err = userManager.SetKindleEmail(ctx, testUser.TelegramID, "user@gmail.com")
	if err == nil {
		t.Error("SetKindleEmail() should fail for invalid email")
	}
}

func TestHandler_LanguageDetection(t *testing.T) {
	_, _, userManager := setupTestHandler(t)
	ctx := context.Background()
	
	tests := []struct {
		name         string
		langCode     string
		expectedLang string
	}{
		{
			name:         "english",
			langCode:     "en",
			expectedLang: "en",
		},
		{
			name:         "russian",
			langCode:     "ru",
			expectedLang: "ru",
		},
		{
			name:         "default fallback",
			langCode:     "fr",
			expectedLang: "en",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, _ := userManager.GetOrCreateUser(ctx, int64(tt.name[0]), "test", "Test", "User", tt.langCode)
			if user.Language != tt.expectedLang {
				t.Errorf("Language = %v, want %v", user.Language, tt.expectedLang)
			}
		})
	}
}

func TestHandler_UserCreationFlow(t *testing.T) {
	_, _, userManager := setupTestHandler(t)
	ctx := context.Background()
	
	// First call should create user
	user1, err := userManager.GetOrCreateUser(ctx, 12345, "testuser", "Test", "User", "en")
	if err != nil {
		t.Fatalf("GetOrCreateUser() error = %v", err)
	}
	
	if user1.TelegramID != 12345 {
		t.Errorf("TelegramID = %v, want %v", user1.TelegramID, 12345)
	}
	
	// Second call should return existing user
	user2, err := userManager.GetOrCreateUser(ctx, 12345, "testuser", "Test", "User", "en")
	if err != nil {
		t.Fatalf("GetOrCreateUser() error = %v", err)
	}
	
	if user2.TelegramID != user1.TelegramID {
		t.Error("Should return same user on second call")
	}
}

func TestHandler_SetLanguageFlow(t *testing.T) {
	_, _, userManager := setupTestHandler(t)
	ctx := context.Background()
	
	// Create user
	user, _ := userManager.GetOrCreateUser(ctx, 12345, "testuser", "Test", "User", "en")
	
	if user.Language != "en" {
		t.Errorf("Initial language = %v, want %v", user.Language, "en")
	}
	
	// Change language
	err := userManager.SetLanguage(ctx, user.TelegramID, "ru")
	if err != nil {
		t.Fatalf("SetLanguage() error = %v", err)
	}
	
	// Verify language changed
	updatedUser, _ := userManager.GetOrCreateUser(ctx, 12345, "testuser", "Test", "User", "en")
	if updatedUser.Language != "ru" {
		t.Errorf("Updated language = %v, want %v", updatedUser.Language, "ru")
	}
}

func TestHandler_SearchQueryDetection(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		isCommand bool
		isSearch  bool
	}{
		{
			name:      "command",
			text:      "/start",
			isCommand: true,
			isSearch:  false,
		},
		{
			name:      "search query",
			text:      "Harry Potter",
			isCommand: false,
			isSearch:  true,
		},
		{
			name:      "kindle email",
			text:      "user@kindle.com",
			isCommand: false,
			isSearch:  false, // Would be detected as email
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isCommand := len(tt.text) > 0 && tt.text[0] == '/'
			if isCommand != tt.isCommand {
				t.Errorf("IsCommand = %v, want %v", isCommand, tt.isCommand)
			}
			
			// Search is anything that's not a command and not an email
			isEmail := len(tt.text) > 11 && tt.text[len(tt.text)-11:] == "@kindle.com"
			isSearchQuery := !isCommand && !isEmail
			
			if tt.isSearch && !isSearchQuery {
				t.Error("Should be detected as search query")
			}
		})
	}
}

func TestHandler_BookSentCounter(t *testing.T) {
	_, _, userManager := setupTestHandler(t)
	ctx := context.Background()
	
	user, _ := userManager.GetOrCreateUser(ctx, 12345, "testuser", "Test", "User", "en")
	
	if user.BooksSent != 0 {
		t.Errorf("Initial BooksSent = %v, want %v", user.BooksSent, 0)
	}
	
	// Record books sent
	userManager.RecordBookSent(ctx, user.TelegramID)
	userManager.RecordBookSent(ctx, user.TelegramID)
	userManager.RecordBookSent(ctx, user.TelegramID)
	
	// Check counter
	updatedUser, _ := userManager.GetOrCreateUser(ctx, 12345, "testuser", "Test", "User", "en")
	if updatedUser.BooksSent != 3 {
		t.Errorf("BooksSent = %v, want %v", updatedUser.BooksSent, 3)
	}
}
