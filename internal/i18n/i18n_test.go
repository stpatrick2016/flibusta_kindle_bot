package i18n

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewI18n(t *testing.T) {
	// Create temporary test directory with locale files
	tmpDir := t.TempDir()

	// Create test locale files
	enContent := `{
		"welcome": "Welcome!",
		"greeting": "Hello, %s!",
		"test_key": "Test value"
	}`

	ruContent := `{
		"welcome": "Добро пожаловать!",
		"greeting": "Привет, %s!",
		"test_key": "Тестовое значение"
	}`

	if err := os.WriteFile(filepath.Join(tmpDir, "en.json"), []byte(enContent), 0644); err != nil {
		t.Fatalf("Failed to create en.json: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "ru.json"), []byte(ruContent), 0644); err != nil {
		t.Fatalf("Failed to create ru.json: %v", err)
	}

	// Test successful initialization
	i18n, err := NewI18n(tmpDir, "en")
	if err != nil {
		t.Fatalf("NewI18n() error = %v", err)
	}

	if i18n == nil {
		t.Fatal("NewI18n() returned nil")
	}

	// Test that translations were loaded
	if len(i18n.translations) != 2 {
		t.Errorf("Expected 2 languages, got %d", len(i18n.translations))
	}
}

func TestNewI18n_InvalidPath(t *testing.T) {
	_, err := NewI18n("/nonexistent/path", "en")
	if err == nil {
		t.Error("Expected error for invalid path, got nil")
	}
}

func TestI18n_T(t *testing.T) {
	// Create temporary test directory
	tmpDir := t.TempDir()

	enContent := `{
		"simple": "Simple text",
		"with_param": "Hello, %s!",
		"with_multiple": "User %s has %d books",
		"missing_in_ru": "Only in English"
	}`

	ruContent := `{
		"simple": "Простой текст",
		"with_param": "Привет, %s!",
		"with_multiple": "У пользователя %s есть %d книг"
	}`

	os.WriteFile(filepath.Join(tmpDir, "en.json"), []byte(enContent), 0644)
	os.WriteFile(filepath.Join(tmpDir, "ru.json"), []byte(ruContent), 0644)

	i18n, err := NewI18n(tmpDir, "en")
	if err != nil {
		t.Fatalf("Failed to create i18n: %v", err)
	}

	tests := []struct {
		name     string
		lang     string
		key      string
		args     []interface{}
		expected string
	}{
		{
			name:     "simple english",
			lang:     "en",
			key:      "simple",
			args:     nil,
			expected: "Simple text",
		},
		{
			name:     "simple russian",
			lang:     "ru",
			key:      "simple",
			args:     nil,
			expected: "Простой текст",
		},
		{
			name:     "with parameter english",
			lang:     "en",
			key:      "with_param",
			args:     []interface{}{"John"},
			expected: "Hello, John!",
		},
		{
			name:     "with parameter russian",
			lang:     "ru",
			key:      "with_param",
			args:     []interface{}{"Иван"},
			expected: "Привет, Иван!",
		},
		{
			name:     "multiple parameters",
			lang:     "en",
			key:      "with_multiple",
			args:     []interface{}{"Alice", 5},
			expected: "User Alice has 5 books",
		},
		{
			name:     "fallback to default language",
			lang:     "ru",
			key:      "missing_in_ru",
			args:     nil,
			expected: "Only in English",
		},
		{
			name:     "missing key returns key",
			lang:     "en",
			key:      "nonexistent",
			args:     nil,
			expected: "nonexistent",
		},
		{
			name:     "invalid language fallback to default",
			lang:     "fr",
			key:      "simple",
			args:     nil,
			expected: "Simple text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := i18n.T(tt.lang, tt.key, tt.args...)
			if result != tt.expected {
				t.Errorf("T() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestI18n_DetectLanguage(t *testing.T) {
	tests := []struct {
		name     string
		langCode string
		expected string
	}{
		{
			name:     "russian code",
			langCode: "ru",
			expected: "ru",
		},
		{
			name:     "russian with region",
			langCode: "ru-RU",
			expected: "ru",
		},
		{
			name:     "english code",
			langCode: "en",
			expected: "en",
		},
		{
			name:     "english with region",
			langCode: "en-US",
			expected: "en",
		},
		{
			name:     "unsupported language",
			langCode: "fr",
			expected: "en",
		},
		{
			name:     "empty code",
			langCode: "",
			expected: "en",
		},
		{
			name:     "single character",
			langCode: "r",
			expected: "en",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectLanguage(tt.langCode)
			if result != tt.expected {
				t.Errorf("DetectLanguage(%q) = %v, want %v", tt.langCode, result, tt.expected)
			}
		})
	}
}

func TestI18n_AvailableLanguages(t *testing.T) {
	tmpDir := t.TempDir()

	os.WriteFile(filepath.Join(tmpDir, "en.json"), []byte(`{"test": "test"}`), 0644)
	os.WriteFile(filepath.Join(tmpDir, "ru.json"), []byte(`{"test": "тест"}`), 0644)
	os.WriteFile(filepath.Join(tmpDir, "fr.json"), []byte(`{"test": "test"}`), 0644)

	i18n, err := NewI18n(tmpDir, "en")
	if err != nil {
		t.Fatalf("Failed to create i18n: %v", err)
	}

	langs := i18n.GetSupportedLanguages()

	if len(langs) != 3 {
		t.Errorf("Expected 3 languages, got %d", len(langs))
	}

	// Check that expected languages are present
	langMap := make(map[string]bool)
	for _, lang := range langs {
		langMap[lang] = true
	}

	expectedLangs := []string{"en", "ru", "fr"}
	for _, expected := range expectedLangs {
		if !langMap[expected] {
			t.Errorf("Expected language %q not found in available languages", expected)
		}
	}
}

func TestI18n_LoadTranslations_MalformedJSON(t *testing.T) {
	tmpDir := t.TempDir()

	// Create malformed JSON file
	malformed := `{
		"key": "value"
		missing comma
	}`

	os.WriteFile(filepath.Join(tmpDir, "en.json"), []byte(malformed), 0644)

	_, err := NewI18n(tmpDir, "en")
	if err == nil {
		t.Error("Expected error for malformed JSON, got nil")
	}
}
