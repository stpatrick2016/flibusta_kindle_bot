package models

import "time"

// User represents a Telegram user
type User struct {
	ID          int64     `json:"id"`           // Telegram user ID
	TelegramID  int64     `json:"telegram_id"`  // Same as ID, for clarity
	Username    string    `json:"username"`     // Telegram username
	FirstName   string    `json:"first_name"`   // User's first name
	LastName    string    `json:"last_name"`    // User's last name
	KindleEmail string    `json:"kindle_email"` // User's Kindle email
	Language    string    `json:"language"`     // User's preferred language (en, ru)
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	BooksSent   int       `json:"books_sent"`     // Statistics
	LastActive  time.Time `json:"last_active"`    // Last interaction time
	IsActive    bool      `json:"is_active"`      // Is user active
	IsBanned    bool      `json:"is_banned"`      // Is user banned
}

// Preferences represents user preferences
type Preferences struct {
	KindleEmail    string `json:"kindle_email"`
	Language       string `json:"language"`
	PreferredFormat string `json:"preferred_format"` // mobi, epub, pdf
}

// SearchContext represents an active search session
type SearchContext struct {
	Query     string    `json:"query"`
	Results   []Book    `json:"results"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// HasKindleEmail checks if user has configured their Kindle email
func (u *User) HasKindleEmail() bool {
	return u.KindleEmail != ""
}

// IsValidLanguage checks if the language code is supported
func (u *User) IsValidLanguage() bool {
	validLanguages := map[string]bool{
		"en": true,
		"ru": true,
	}
	return validLanguages[u.Language]
}

// GetDisplayName returns the user's display name
func (u *User) GetDisplayName() string {
	if u.FirstName != "" {
		if u.LastName != "" {
			return u.FirstName + " " + u.LastName
		}
		return u.FirstName
	}
	if u.Username != "" {
		return u.Username
	}
	return "User"
}
