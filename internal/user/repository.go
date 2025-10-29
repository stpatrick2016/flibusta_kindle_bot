package user

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/yourusername/flibusta_kindle_bot/pkg/models"
)

// MemoryRepository is an in-memory implementation of Repository
// This is for development/testing. In production, use database implementation.
type MemoryRepository struct {
	users map[int64]*models.User
	mu    sync.RWMutex
}

// NewMemoryRepository creates a new in-memory repository
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		users: make(map[int64]*models.User),
	}
}

// GetUser retrieves a user by Telegram ID
func (r *MemoryRepository) GetUser(ctx context.Context, telegramID int64) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[telegramID]
	if !exists {
		return nil, ErrUserNotFound
	}

	// Return a copy to prevent external modifications
	userCopy := *user
	return &userCopy, nil
}

// SaveUser creates or updates a user
func (r *MemoryRepository) SaveUser(ctx context.Context, user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user.UpdatedAt = time.Now()
	
	// Store a copy
	userCopy := *user
	r.users[user.TelegramID] = &userCopy

	return nil
}

// UpdatePreferences updates user preferences
func (r *MemoryRepository) UpdatePreferences(ctx context.Context, telegramID int64, prefs *models.Preferences) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[telegramID]
	if !exists {
		return ErrUserNotFound
	}

	if prefs.KindleEmail != "" {
		user.KindleEmail = prefs.KindleEmail
	}
	if prefs.Language != "" {
		user.Language = prefs.Language
	}

	user.UpdatedAt = time.Now()
	return nil
}

// IncrementBooksSent increments the books sent counter
func (r *MemoryRepository) IncrementBooksSent(ctx context.Context, telegramID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[telegramID]
	if !exists {
		return ErrUserNotFound
	}

	user.BooksSent++
	user.UpdatedAt = time.Now()
	return nil
}

// UpdateLastActive updates the last active timestamp
func (r *MemoryRepository) UpdateLastActive(ctx context.Context, telegramID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[telegramID]
	if !exists {
		return ErrUserNotFound
	}

	user.LastActive = time.Now()
	return nil
}

// ExportData exports all users as JSON (for debugging)
func (r *MemoryRepository) ExportData() (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data, err := json.MarshalIndent(r.users, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal users: %w", err)
	}

	return string(data), nil
}
