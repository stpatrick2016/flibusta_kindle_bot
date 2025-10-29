package user

import (
	"context"
	"errors"
	"time"

	"github.com/stpatrick2016/flibusta_kindle_bot/pkg/models"
)

var (
	// ErrUserNotFound is returned when user is not found
	ErrUserNotFound = errors.New("user not found")
	// ErrInvalidEmail is returned when Kindle email format is invalid
	ErrInvalidEmail = errors.New("invalid Kindle email format")
)

// Repository defines the interface for user storage
type Repository interface {
	// GetUser retrieves a user by Telegram ID
	GetUser(ctx context.Context, telegramID int64) (*models.User, error)
	
	// SaveUser creates or updates a user
	SaveUser(ctx context.Context, user *models.User) error
	
	// UpdatePreferences updates user preferences
	UpdatePreferences(ctx context.Context, telegramID int64, prefs *models.Preferences) error
	
	// IncrementBooksSent increments the books sent counter
	IncrementBooksSent(ctx context.Context, telegramID int64) error
	
	// UpdateLastActive updates the last active timestamp
	UpdateLastActive(ctx context.Context, telegramID int64) error
}

// Manager handles user operations
type Manager struct {
	repo Repository
}

// NewManager creates a new user manager
func NewManager(repo Repository) *Manager {
	return &Manager{
		repo: repo,
	}
}

// GetOrCreateUser gets an existing user or creates a new one
func (m *Manager) GetOrCreateUser(ctx context.Context, telegramID int64, username, firstName, lastName, langCode string) (*models.User, error) {
	user, err := m.repo.GetUser(ctx, telegramID)
	if err == nil {
		// User exists, update last active
		user.LastActive = time.Now()
		if err := m.repo.UpdateLastActive(ctx, telegramID); err != nil {
			// Log error but don't fail
			return user, nil
		}
		return user, nil
	}

	if !errors.Is(err, ErrUserNotFound) {
		return nil, err
	}

	// Create new user
	user = &models.User{
		ID:         telegramID,
		TelegramID: telegramID,
		Username:   username,
		FirstName:  firstName,
		LastName:   lastName,
		Language:   detectLanguage(langCode),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		LastActive: time.Now(),
		IsActive:   true,
		BooksSent:  0,
	}

	if err := m.repo.SaveUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// SetKindleEmail sets the user's Kindle email
func (m *Manager) SetKindleEmail(ctx context.Context, telegramID int64, email string) error {
	if err := ValidateKindleEmail(email); err != nil {
		return err
	}

	prefs := &models.Preferences{
		KindleEmail: email,
	}

	return m.repo.UpdatePreferences(ctx, telegramID, prefs)
}

// SetLanguage sets the user's preferred language
func (m *Manager) SetLanguage(ctx context.Context, telegramID int64, language string) error {
	prefs := &models.Preferences{
		Language: language,
	}

	return m.repo.UpdatePreferences(ctx, telegramID, prefs)
}

// RecordBookSent increments the books sent counter
func (m *Manager) RecordBookSent(ctx context.Context, telegramID int64) error {
	return m.repo.IncrementBooksSent(ctx, telegramID)
}

// ValidateKindleEmail validates the format of a Kindle email
func ValidateKindleEmail(email string) error {
	if email == "" {
		return ErrInvalidEmail
	}

	// Must end with @kindle.com
	if len(email) < 11 || email[len(email)-11:] != "@kindle.com" {
		return ErrInvalidEmail
	}

	// Must have at least one character before @
	if len(email) == 11 {
		return ErrInvalidEmail
	}

	return nil
}

// detectLanguage detects user's language from Telegram language code
func detectLanguage(langCode string) string {
	// Simple detection based on language code
	if len(langCode) >= 2 {
		lang := langCode[:2]
		if lang == "ru" {
			return "ru"
		}
	}
	return "en" // Default to English
}
