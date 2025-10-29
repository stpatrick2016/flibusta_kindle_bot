package models

import (
	"fmt"
	"strings"
	"time"
)

// Book represents a book from Flibusta
type Book struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Author      string    `json:"author"`
	Format      string    `json:"format"`
	Size        int64     `json:"size"`
	URL         string    `json:"url"`
	Description string    `json:"description,omitempty"`
	Year        int       `json:"year,omitempty"`
	Language    string    `json:"language,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// FormatSize returns a human-readable file size
func (b *Book) FormatSize() string {
	const unit = 1024
	if b.Size < unit {
		return fmt.Sprintf("%d B", b.Size)
	}
	div, exp := int64(unit), 0
	for n := b.Size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(b.Size)/float64(div), "KMGTPE"[exp])
}

// IsValidFormat checks if the book format is supported by Kindle
func (b *Book) IsValidFormat() bool {
	validFormats := map[string]bool{
		"mobi": true,
		"epub": true,
		"pdf":  true,
		"azw3": true,
		"txt":  true,
		"doc":  true,
		"docx": true,
	}
	return validFormats[strings.ToLower(b.Format)]
}

// GetDownloadURL returns the download URL for the book
func (b *Book) GetDownloadURL() string {
	return b.URL
}

// String returns a string representation of the book
func (b *Book) String() string {
	return fmt.Sprintf("%s by %s (%s, %s)", b.Title, b.Author, b.Format, b.FormatSize())
}
