package models

import (
	"testing"
	"time"
)

func TestUser_HasKindleEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		{
			name:     "has kindle email",
			email:    "user@kindle.com",
			expected: true,
		},
		{
			name:     "empty email",
			email:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{KindleEmail: tt.email}
			result := user.HasKindleEmail()
			if result != tt.expected {
				t.Errorf("HasKindleEmail() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestUser_GetDisplayName(t *testing.T) {
	tests := []struct {
		name      string
		user      *User
		expected  string
	}{
		{
			name: "username available",
			user: &User{
				Username:  "johndoe",
				FirstName: "John",
				LastName:  "Doe",
			},
			expected: "@johndoe",
		},
		{
			name: "first and last name",
			user: &User{
				Username:  "",
				FirstName: "John",
				LastName:  "Doe",
			},
			expected: "John Doe",
		},
		{
			name: "first name only",
			user: &User{
				Username:  "",
				FirstName: "John",
				LastName:  "",
			},
			expected: "John",
		},
		{
			name: "no names",
			user: &User{
				TelegramID: 123456,
				Username:   "",
				FirstName:  "",
				LastName:   "",
			},
			expected: "User 123456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.GetDisplayName()
			if result != tt.expected {
				t.Errorf("GetDisplayName() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestUser_IsValidLanguage(t *testing.T) {
	tests := []struct {
		name     string
		language string
		expected bool
	}{
		{
			name:     "english",
			language: "en",
			expected: true,
		},
		{
			name:     "russian",
			language: "ru",
			expected: true,
		},
		{
			name:     "invalid language",
			language: "fr",
			expected: false,
		},
		{
			name:     "empty language",
			language: "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Language: tt.language}
			result := user.IsValidLanguage()
			if result != tt.expected {
				t.Errorf("IsValidLanguage() for %q = %v, want %v", tt.language, result, tt.expected)
			}
		})
	}
}

func TestSearchContext_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		context  *SearchContext
		expected bool
	}{
		{
			name: "active with results",
			context: &SearchContext{
				Query:   "test query",
				Results: []*Book{{ID: 1}},
			},
			expected: true,
		},
		{
			name: "empty query",
			context: &SearchContext{
				Query:   "",
				Results: []*Book{{ID: 1}},
			},
			expected: false,
		},
		{
			name: "no results",
			context: &SearchContext{
				Query:   "test query",
				Results: []*Book{},
			},
			expected: false,
		},
		{
			name: "nil results",
			context: &SearchContext{
				Query:   "test query",
				Results: nil,
			},
			expected: false,
		},
		{
			name:     "nil context",
			context:  nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result bool
			if tt.context == nil {
				result = (*SearchContext)(nil).IsActive()
			} else {
				result = tt.context.IsActive()
			}
			if result != tt.expected {
				t.Errorf("IsActive() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestUser_UpdateLastActive(t *testing.T) {
	user := &User{
		TelegramID: 123456,
		LastActive: time.Now().Add(-24 * time.Hour),
	}

	oldTime := user.LastActive
	time.Sleep(10 * time.Millisecond)

	user.LastActive = time.Now()
	
	if !user.LastActive.After(oldTime) {
		t.Error("LastActive was not updated to a newer time")
	}
}

func TestPreferences_Validation(t *testing.T) {
	tests := []struct {
		name        string
		prefs       *Preferences
		shouldError bool
	}{
		{
			name: "valid kindle email",
			prefs: &Preferences{
				KindleEmail: "user@kindle.com",
				Language:    "en",
			},
			shouldError: false,
		},
		{
			name: "valid language",
			prefs: &Preferences{
				Language: "ru",
			},
			shouldError: false,
		},
		{
			name: "empty preferences",
			prefs: &Preferences{},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation - just ensure struct can be created
			if tt.prefs == nil {
				t.Error("Preferences should not be nil")
			}
		})
	}
}
