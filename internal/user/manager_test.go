package user

import (
	"context"
	"testing"

	"github.com/stpatrick2016/flibusta_kindle_bot/pkg/models"
)

func TestValidateKindleEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "valid kindle email",
			email:   "user@kindle.com",
			wantErr: false,
		},
		{
			name:    "valid kindle email with numbers",
			email:   "user123@kindle.com",
			wantErr: false,
		},
		{
			name:    "valid kindle email with dots",
			email:   "user.name@kindle.com",
			wantErr: false,
		},
		{
			name:    "invalid - wrong domain",
			email:   "user@gmail.com",
			wantErr: true,
		},
		{
			name:    "invalid - no @ symbol",
			email:   "userkindle.com",
			wantErr: true,
		},
		{
			name:    "invalid - empty email",
			email:   "",
			wantErr: true,
		},
		{
			name:    "invalid - only @kindle.com",
			email:   "@kindle.com",
			wantErr: true,
		},
		{
			name:    "invalid - spaces",
			email:   "user @kindle.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateKindleEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateKindleEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != ErrInvalidEmail {
				t.Errorf("Expected ErrInvalidEmail, got %v", err)
			}
		})
	}
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name     string
		langCode string
		expected string
	}{
		{
			name:     "russian",
			langCode: "ru",
			expected: "ru",
		},
		{
			name:     "russian with region",
			langCode: "ru-RU",
			expected: "ru",
		},
		{
			name:     "english",
			langCode: "en",
			expected: "en",
		},
		{
			name:     "english with region",
			langCode: "en-US",
			expected: "en",
		},
		{
			name:     "other language defaults to english",
			langCode: "fr",
			expected: "en",
		},
		{
			name:     "empty defaults to english",
			langCode: "",
			expected: "en",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectLanguage(tt.langCode)
			if result != tt.expected {
				t.Errorf("detectLanguage(%q) = %v, want %v", tt.langCode, result, tt.expected)
			}
		})
	}
}

func TestManager_GetOrCreateUser(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()
	manager := NewManager(repo)

	// Test creating new user
	user, err := manager.GetOrCreateUser(ctx, 12345, "johndoe", "John", "Doe", "en-US")
	if err != nil {
		t.Fatalf("GetOrCreateUser() error = %v", err)
	}

	if user == nil {
		t.Fatal("GetOrCreateUser() returned nil user")
	}

	if user.TelegramID != 12345 {
		t.Errorf("TelegramID = %v, want %v", user.TelegramID, 12345)
	}

	if user.Username != "johndoe" {
		t.Errorf("Username = %v, want %v", user.Username, "johndoe")
	}

	if user.Language != "en" {
		t.Errorf("Language = %v, want %v", user.Language, "en")
	}

	// Test getting existing user
	user2, err := manager.GetOrCreateUser(ctx, 12345, "johndoe", "John", "Doe", "en-US")
	if err != nil {
		t.Fatalf("GetOrCreateUser() error = %v", err)
	}

	if user2.TelegramID != user.TelegramID {
		t.Error("GetOrCreateUser() should return existing user")
	}
}

func TestManager_SetKindleEmail(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()
	manager := NewManager(repo)

	// Create a user first
	user, _ := manager.GetOrCreateUser(ctx, 12345, "johndoe", "John", "Doe", "en")

	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "valid email",
			email:   "johndoe@kindle.com",
			wantErr: false,
		},
		{
			name:    "invalid email",
			email:   "johndoe@gmail.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.SetKindleEmail(ctx, user.TelegramID, tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetKindleEmail() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				// Verify email was set
				updatedUser, _ := repo.GetUser(ctx, user.TelegramID)
				if updatedUser.KindleEmail != tt.email {
					t.Errorf("KindleEmail = %v, want %v", updatedUser.KindleEmail, tt.email)
				}
			}
		})
	}
}

func TestManager_SetLanguage(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()
	manager := NewManager(repo)

	// Create a user first
	user, _ := manager.GetOrCreateUser(ctx, 12345, "johndoe", "John", "Doe", "en")

	// Set language to Russian
	err := manager.SetLanguage(ctx, user.TelegramID, "ru")
	if err != nil {
		t.Fatalf("SetLanguage() error = %v", err)
	}

	// Verify language was set
	updatedUser, _ := repo.GetUser(ctx, user.TelegramID)
	if updatedUser.Language != "ru" {
		t.Errorf("Language = %v, want %v", updatedUser.Language, "ru")
	}
}

func TestManager_RecordBookSent(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()
	manager := NewManager(repo)

	// Create a user first
	user, _ := manager.GetOrCreateUser(ctx, 12345, "johndoe", "John", "Doe", "en")

	// Record book sent
	err := manager.RecordBookSent(ctx, user.TelegramID)
	if err != nil {
		t.Fatalf("RecordBookSent() error = %v", err)
	}

	// Verify counter was incremented
	updatedUser, _ := repo.GetUser(ctx, user.TelegramID)
	if updatedUser.BooksSent != 1 {
		t.Errorf("BooksSent = %v, want %v", updatedUser.BooksSent, 1)
	}

	// Record another book
	manager.RecordBookSent(ctx, user.TelegramID)
	updatedUser, _ = repo.GetUser(ctx, user.TelegramID)
	if updatedUser.BooksSent != 2 {
		t.Errorf("BooksSent = %v, want %v", updatedUser.BooksSent, 2)
	}
}

func TestMemoryRepository_GetUser_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()

	_, err := repo.GetUser(ctx, 99999)
	if err != ErrUserNotFound {
		t.Errorf("GetUser() error = %v, want %v", err, ErrUserNotFound)
	}
}

func TestMemoryRepository_SaveUser(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()

	user := &models.User{
		TelegramID: 12345,
		Username:   "testuser",
		FirstName:  "Test",
		LastName:   "User",
		Language:   "en",
	}

	err := repo.SaveUser(ctx, user)
	if err != nil {
		t.Fatalf("SaveUser() error = %v", err)
	}

	// Verify user was saved
	savedUser, err := repo.GetUser(ctx, 12345)
	if err != nil {
		t.Fatalf("GetUser() error = %v", err)
	}

	if savedUser.TelegramID != user.TelegramID {
		t.Errorf("TelegramID = %v, want %v", savedUser.TelegramID, user.TelegramID)
	}
}

func TestMemoryRepository_UpdatePreferences(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()

	// Create user
	user := &models.User{
		TelegramID: 12345,
		Language:   "en",
	}
	repo.SaveUser(ctx, user)

	// Update preferences
	prefs := &models.Preferences{
		KindleEmail: "test@kindle.com",
		Language:    "ru",
	}

	err := repo.UpdatePreferences(ctx, 12345, prefs)
	if err != nil {
		t.Fatalf("UpdatePreferences() error = %v", err)
	}

	// Verify preferences were updated
	updatedUser, _ := repo.GetUser(ctx, 12345)
	if updatedUser.KindleEmail != "test@kindle.com" {
		t.Errorf("KindleEmail = %v, want %v", updatedUser.KindleEmail, "test@kindle.com")
	}
	if updatedUser.Language != "ru" {
		t.Errorf("Language = %v, want %v", updatedUser.Language, "ru")
	}
}

func TestMemoryRepository_IncrementBooksSent(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()

	// Create user
	user := &models.User{
		TelegramID: 12345,
		BooksSent:  0,
	}
	repo.SaveUser(ctx, user)

	// Increment counter
	err := repo.IncrementBooksSent(ctx, 12345)
	if err != nil {
		t.Fatalf("IncrementBooksSent() error = %v", err)
	}

	// Verify counter was incremented
	updatedUser, _ := repo.GetUser(ctx, 12345)
	if updatedUser.BooksSent != 1 {
		t.Errorf("BooksSent = %v, want %v", updatedUser.BooksSent, 1)
	}
}

func TestMemoryRepository_ExportData(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()

	// Add some users
	repo.SaveUser(ctx, &models.User{TelegramID: 1, Username: "user1"})
	repo.SaveUser(ctx, &models.User{TelegramID: 2, Username: "user2"})

	data, err := repo.ExportData()
	if err != nil {
		t.Fatalf("ExportData() error = %v", err)
	}

	if data == "" {
		t.Error("ExportData() returned empty string")
	}

	// Verify JSON contains user data
	if !contains(data, "user1") || !contains(data, "user2") {
		t.Error("ExportData() does not contain expected user data")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
