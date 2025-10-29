package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// I18n provides internationalization support
type I18n struct {
	translations map[string]map[string]string // lang -> key -> value
	defaultLang  string
}

// New creates a new I18n instance
func New(defaultLang string) *I18n {
	return &I18n{
		translations: make(map[string]map[string]string),
		defaultLang:  defaultLang,
	}
}

// NewI18n creates a new I18n instance and loads translations from the specified directory
func NewI18n(dir, defaultLang string) (*I18n, error) {
	i18n := New(defaultLang)
	if err := i18n.LoadTranslations(dir); err != nil {
		return nil, err
	}
	return i18n, nil
}

// LoadTranslations loads translation files from a directory
func (i *I18n) LoadTranslations(dir string) error {
	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", dir)
	}

	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return fmt.Errorf("failed to list translation files: %w", err)
	}

	for _, file := range files {
		// Extract language code from filename (e.g., "en.json" -> "en")
		lang := strings.TrimSuffix(filepath.Base(file), ".json")

		if err := i.LoadLanguage(lang, file); err != nil {
			return fmt.Errorf("failed to load language %s: %w", lang, err)
		}
	}

	return nil
}

// LoadLanguage loads translations for a specific language
func (i *I18n) LoadLanguage(lang, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var translations map[string]string
	if err := json.Unmarshal(data, &translations); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	i.translations[lang] = translations
	return nil
}

// T translates a key to the specified language
func (i *I18n) T(lang, key string, args ...interface{}) string {
	// Try requested language
	if template, ok := i.translations[lang][key]; ok {
		if len(args) > 0 {
			return fmt.Sprintf(template, args...)
		}
		return template
	}

	// Fallback to default language
	if template, ok := i.translations[i.defaultLang][key]; ok {
		if len(args) > 0 {
			return fmt.Sprintf(template, args...)
		}
		return template
	}

	// Return key if translation not found
	return key
}

// GetSupportedLanguages returns list of loaded languages
func (i *I18n) GetSupportedLanguages() []string {
	langs := make([]string, 0, len(i.translations))
	for lang := range i.translations {
		langs = append(langs, lang)
	}
	return langs
}

// DetectLanguage detects user's language from Telegram language code
func DetectLanguage(telegramLangCode string) string {
	// Telegram sends language codes like "en", "ru", "en-US", etc.
	// Extract the base language code
	lang := strings.ToLower(telegramLangCode)
	if idx := strings.Index(lang, "-"); idx != -1 {
		lang = lang[:idx]
	}

	// Check if we support this language
	supportedLanguages := map[string]bool{
		"en": true,
		"ru": true,
	}

	if supportedLanguages[lang] {
		return lang
	}

	// Default to English
	return "en"
}
