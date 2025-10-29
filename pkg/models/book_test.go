package models

import (
	"testing"
)

func TestBook_FormatSize(t *testing.T) {
	tests := []struct {
		name     string
		size     int64
		expected string
	}{
		{
			name:     "bytes",
			size:     500,
			expected: "500 B",
		},
		{
			name:     "kilobytes",
			size:     1024,
			expected: "1.00 KB",
		},
		{
			name:     "kilobytes with decimals",
			size:     1536,
			expected: "1.50 KB",
		},
		{
			name:     "megabytes",
			size:     1048576,
			expected: "1.00 MB",
		},
		{
			name:     "megabytes with decimals",
			size:     5242880,
			expected: "5.00 MB",
		},
		{
			name:     "gigabytes",
			size:     1073741824,
			expected: "1.00 GB",
		},
		{
			name:     "zero",
			size:     0,
			expected: "0 B",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			book := &Book{Size: tt.size}
			result := book.FormatSize()
			if result != tt.expected {
				t.Errorf("FormatSize() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestBook_IsValidFormat(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		expected bool
	}{
		{
			name:     "mobi format",
			format:   "mobi",
			expected: true,
		},
		{
			name:     "epub format",
			format:   "epub",
			expected: true,
		},
		{
			name:     "pdf format",
			format:   "pdf",
			expected: true,
		},
		{
			name:     "azw3 format",
			format:   "azw3",
			expected: true,
		},
		{
			name:     "txt format",
			format:   "txt",
			expected: true,
		},
		{
			name:     "doc format",
			format:   "doc",
			expected: true,
		},
		{
			name:     "docx format",
			format:   "docx",
			expected: true,
		},
		{
			name:     "uppercase format",
			format:   "MOBI",
			expected: true,
		},
		{
			name:     "mixed case format",
			format:   "EpUb",
			expected: true,
		},
		{
			name:     "invalid format",
			format:   "exe",
			expected: false,
		},
		{
			name:     "empty format",
			format:   "",
			expected: false,
		},
		{
			name:     "fb2 format (not supported)",
			format:   "fb2",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			book := &Book{Format: tt.format}
			result := book.IsValidFormat()
			if result != tt.expected {
				t.Errorf("IsValidFormat() for format %q = %v, want %v", tt.format, result, tt.expected)
			}
		})
	}
}

func TestBook_GetDownloadURL(t *testing.T) {
	tests := []struct {
		name     string
		book     *Book
		expected string
	}{
		{
			name: "complete URL",
			book: &Book{
				URL: "https://flibusta.is/b/123456",
			},
			expected: "https://flibusta.is/b/123456",
		},
		{
			name: "empty URL",
			book: &Book{
				URL: "",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.book.GetDownloadURL()
			if result != tt.expected {
				t.Errorf("GetDownloadURL() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestBook_String(t *testing.T) {
	book := &Book{
		ID:     "123",
		Title:  "Test Book",
		Author: "John Doe",
		Format: "epub",
		Size:   1048576,
	}

	result := book.String()
	expected := "Test Book by John Doe (epub, 1.00 MB)"

	if result != expected {
		t.Errorf("String() = %v, want %v", result, expected)
	}
}
