package bot

import (
	"context"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/stpatrick2016/flibusta_kindle_bot/internal/i18n"
	usermanager "github.com/stpatrick2016/flibusta_kindle_bot/internal/user"
	"github.com/stpatrick2016/flibusta_kindle_bot/pkg/models"
)

// Handler handles Telegram bot updates.
type Handler struct {
	bot         *tgbotapi.BotAPI
	i18n        *i18n.I18n
	userManager *usermanager.Manager
}

// NewHandler creates a new bot handler.
func NewHandler(bot *tgbotapi.BotAPI, i18n *i18n.I18n, userManager *usermanager.Manager) *Handler {
	return &Handler{
		bot:         bot,
		i18n:        i18n,
		userManager: userManager,
	}
}

// HandleUpdate processes incoming Telegram updates.
func (h *Handler) HandleUpdate(ctx context.Context, update *tgbotapi.Update) error {
	// Handle callback queries (inline keyboard button clicks)
	if update.CallbackQuery != nil {
		return h.handleCallbackQuery(ctx, update.CallbackQuery)
	}

	// Handle messages
	if update.Message != nil {
		return h.handleMessage(ctx, update.Message)
	}

	return nil
}

// handleMessage processes incoming messages.
func (h *Handler) handleMessage(ctx context.Context, message *tgbotapi.Message) error {
	// Get or create user
	user, err := h.userManager.GetOrCreateUser(ctx, message.From.ID, message.From.FirstName, message.From.LastName, message.From.UserName, message.From.LanguageCode)
	if err != nil {
		log.Printf("Failed to get/create user: %v", err)

		return err
	}

	// Check if it's a command
	if message.IsCommand() {
		return h.handleCommand(ctx, message, user)
	}

	// Otherwise, treat as search query
	return h.handleSearchQuery(ctx, message, user)
}

// handleCommand processes bot commands.
func (h *Handler) handleCommand(ctx context.Context, message *tgbotapi.Message, user *models.User) error {
	command := message.Command()

	switch command {
	case "start":
		return h.handleStart(message, user)
	case "help":
		return h.handleHelp(message, user)
	case "kindle":
		return h.handleKindle(ctx, message, user)
	case "language":
		return h.handleLanguage(ctx, message, user)
	case "whitelist":
		return h.handleWhitelist(message, user)
	case "settings":
		return h.handleSettings(message, user)
	case "cancel":
		return h.handleCancel(message, user)
	default:
		return h.sendMessage(message.Chat.ID, user.Language, "unknown_command", nil)
	}
}

// handleStart handles /start command.
func (h *Handler) handleStart(message *tgbotapi.Message, user *models.User) error {
	// Send welcome message
	welcomeMsg := h.i18n.T(user.Language, "welcome", user.FirstName)
	msg := tgbotapi.NewMessage(message.Chat.ID, welcomeMsg)
	if _, err := h.bot.Send(msg); err != nil {
		return err
	}

	// Send whitelist instructions if Kindle email not set
	if !user.HasKindleEmail() {
		whitelistMsg := h.i18n.T(user.Language, "whitelist_instructions")
		msg := tgbotapi.NewMessage(message.Chat.ID, whitelistMsg)
		msg.ParseMode = "Markdown"
		if _, err := h.bot.Send(msg); err != nil {
			return err
		}

		// Prompt for Kindle email
		promptMsg := h.i18n.T(user.Language, "kindle_email_prompt")
		msg = tgbotapi.NewMessage(message.Chat.ID, promptMsg)
		if _, err := h.bot.Send(msg); err != nil {
			return err
		}
	} else {
		// User already has Kindle email set
		searchMsg := h.i18n.T(user.Language, "search_prompt")
		msg := tgbotapi.NewMessage(message.Chat.ID, searchMsg)
		if _, err := h.bot.Send(msg); err != nil {
			return err
		}
	}

	return nil
}

// handleHelp handles /help command.
func (h *Handler) handleHelp(message *tgbotapi.Message, user *models.User) error {
	return h.sendMessage(message.Chat.ID, user.Language, "help_message", nil)
}

// handleKindle handles /kindle command (set Kindle email).
func (h *Handler) handleKindle(ctx context.Context, message *tgbotapi.Message, user *models.User) error {
	args := message.CommandArguments()

	// If no arguments, show current Kindle email or prompt to set one
	if args == "" {
		if user.HasKindleEmail() {
			return h.sendMessage(message.Chat.ID, user.Language, "kindle_email_current", user.KindleEmail)
		}
		return h.sendMessage(message.Chat.ID, user.Language, "kindle_email_prompt", nil)
	}

	// Set Kindle email
	email := strings.TrimSpace(args)
	if err := h.userManager.SetKindleEmail(ctx, user.TelegramID, email); err != nil {
		if err == usermanager.ErrInvalidEmail {
			return h.sendMessage(message.Chat.ID, user.Language, "kindle_email_invalid", nil)
		}
		return err
	}

	// Send confirmation
	confirmMsg := h.i18n.T(user.Language, "kindle_email_set", email)
	msg := tgbotapi.NewMessage(message.Chat.ID, confirmMsg)
	if _, err := h.bot.Send(msg); err != nil {
		return err
	}

	// Send whitelist reminder
	return h.sendMessage(message.Chat.ID, user.Language, "whitelist_reminder", nil)
}

// handleLanguage handles /language command.
func (h *Handler) handleLanguage(ctx context.Context, message *tgbotapi.Message, user *models.User) error {
	// Create inline keyboard with language options
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🇬🇧 English", "lang_en"),
			tgbotapi.NewInlineKeyboardButtonData("🇷🇺 Русский", "lang_ru"),
		),
	)

	msg := tgbotapi.NewMessage(message.Chat.ID, h.i18n.T(user.Language, "language_prompt"))
	msg.ReplyMarkup = keyboard

	_, err := h.bot.Send(msg)
	return err
}

// handleWhitelist handles /whitelist command.
func (h *Handler) handleWhitelist(message *tgbotapi.Message, user *models.User) error {
	msg := tgbotapi.NewMessage(message.Chat.ID, h.i18n.T(user.Language, "whitelist_instructions"))
	msg.ParseMode = "Markdown"
	_, err := h.bot.Send(msg)
	return err
}

// handleSettings handles /settings command.
func (h *Handler) handleSettings(message *tgbotapi.Message, user *models.User) error {
	// Display current settings
	kindleEmail := user.KindleEmail
	if kindleEmail == "" {
		kindleEmail = h.i18n.T(user.Language, "not_set")
	}

	language := user.Language
	if language == "en" {
		language = "English"
	} else if language == "ru" {
		language = "Русский"
	}

	return h.sendMessage(message.Chat.ID, user.Language, "settings_display", kindleEmail, language, user.BooksSent)
}

// handleCancel handles /cancel command.
func (h *Handler) handleCancel(message *tgbotapi.Message, user *models.User) error {
	// Cancel any active search context
	// This will be implemented when we add search functionality
	return h.sendMessage(message.Chat.ID, user.Language, "operation_cancelled", nil)
}

// handleSearchQuery handles text messages as book search queries.
func (h *Handler) handleSearchQuery(ctx context.Context, message *tgbotapi.Message, user *models.User) error {
	query := strings.TrimSpace(message.Text)

	// Check if user has Kindle email set
	if !user.HasKindleEmail() {
		// Try to parse the message as a Kindle email
		if strings.Contains(query, "@kindle.com") {
			if err := h.userManager.SetKindleEmail(ctx, user.TelegramID, query); err != nil {
				return h.sendMessage(message.Chat.ID, user.Language, "kindle_email_invalid", nil)
			}

			// Send confirmation
			confirmMsg := h.i18n.T(user.Language, "kindle_email_set", query)
			msg := tgbotapi.NewMessage(message.Chat.ID, confirmMsg)
			if _, err := h.bot.Send(msg); err != nil {
				return err
			}

			// Send whitelist reminder
			return h.sendMessage(message.Chat.ID, user.Language, "whitelist_reminder", nil)
		}

		// User needs to set Kindle email first
		return h.sendMessage(message.Chat.ID, user.Language, "kindle_email_required", nil)
	}

	// Send "searching..." message
	searchingMsg := h.i18n.T(user.Language, "searching", query)
	statusMsg := tgbotapi.NewMessage(message.Chat.ID, searchingMsg)
	sentMsg, err := h.bot.Send(statusMsg)
	if err != nil {
		return err
	}

	// TODO: Implement actual search functionality
	// For now, just send a placeholder message
	_ = sentMsg // Will use this to update the message later

	return h.sendMessage(message.Chat.ID, user.Language, "search_not_implemented", nil)
}

// handleCallbackQuery handles inline keyboard button clicks.
func (h *Handler) handleCallbackQuery(ctx context.Context, query *tgbotapi.CallbackQuery) error {
	// Get user
	user, err := h.userManager.GetOrCreateUser(ctx, query.From.ID, query.From.FirstName, query.From.LastName, query.From.UserName, query.From.LanguageCode)
	if err != nil {
		log.Printf("Failed to get/create user: %v", err)
		return err
	}

	// Parse callback data
	data := query.Data

	// Handle language selection
	if strings.HasPrefix(data, "lang_") {
		lang := strings.TrimPrefix(data, "lang_")

		// Update user language
		if err := h.userManager.SetLanguage(ctx, user.TelegramID, lang); err != nil {
			return err
		}

		// Send confirmation
		callback := tgbotapi.NewCallback(query.ID, h.i18n.T(lang, "language_changed"))
		if _, err := h.bot.Request(callback); err != nil {
			return err
		}

		// Edit the original message
		editMsg := tgbotapi.NewEditMessageText(
			query.Message.Chat.ID,
			query.Message.MessageID,
			h.i18n.T(lang, "language_changed"),
		)
		_, err := h.bot.Send(editMsg)
		return err
	}

	// Handle book selection (will implement with search)
	if strings.HasPrefix(data, "book_") {
		// TODO: Implement book selection and download
		callback := tgbotapi.NewCallback(query.ID, h.i18n.T(user.Language, "feature_coming_soon"))
		_, err := h.bot.Request(callback)
		return err
	}

	// Unknown callback
	callback := tgbotapi.NewCallback(query.ID, "")
	_, err = h.bot.Request(callback)
	return err
}

// sendMessage is a helper to send localized messages.
func (h *Handler) sendMessage(chatID int64, language, key string, args ...interface{}) error {
	text := h.i18n.T(language, key, args...)
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := h.bot.Send(msg)
	return err
}

// sendError sends an error message to the user.
func (h *Handler) sendError(chatID int64, language string, err error) error {
	log.Printf("Error: %v", err)
	return h.sendMessage(chatID, language, "error_occurred", nil)
}
