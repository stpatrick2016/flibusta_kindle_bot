package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stpatrick2016/flibusta_kindle_bot/internal/bot"
	"github.com/stpatrick2016/flibusta_kindle_bot/internal/config"
	"github.com/stpatrick2016/flibusta_kindle_bot/internal/i18n"
	"github.com/stpatrick2016/flibusta_kindle_bot/internal/user"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set up logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("Starting Flibusta Kindle Bot...")
	log.Printf("Bot mode: %s", cfg.BotMode)
	log.Printf("Database type: %s", cfg.DBType)

	// Initialize i18n
	i18nInstance, err := i18n.NewI18n("internal/i18n/locales", "en")
	if err != nil {
		log.Fatalf("Failed to initialize i18n: %v", err)
	}
	log.Printf("Loaded translations for languages: %v", i18nInstance.GetSupportedLanguages())

	// Initialize user repository
	var userRepo user.Repository
	switch cfg.DBType {
	case "memory":
		userRepo = user.NewMemoryRepository()
		log.Printf("Using in-memory user repository")
	case "postgres":
		// TODO: Implement PostgreSQL repository
		log.Fatalf("PostgreSQL repository not implemented yet")
	case "cosmos":
		// TODO: Implement Cosmos DB repository
		log.Fatalf("Cosmos DB repository not implemented yet")
	default:
		log.Fatalf("Unknown database type: %s", cfg.DBType)
	}

	// Initialize user manager
	userManager := user.NewManager(userRepo)

	// Initialize Telegram bot
	botAPI, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		log.Fatalf("Failed to create Telegram bot: %v", err)
	}

	botAPI.Debug = cfg.LogLevel == "debug"
	log.Printf("Authorized on account @%s", botAPI.Self.UserName)

	// Initialize bot handler
	handler := bot.NewHandler(botAPI, i18nInstance, userManager)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start bot based on mode
	switch cfg.BotMode {
	case "polling":
		go runPollingMode(ctx, botAPI, handler)
	case "webhook":
		go runWebhookMode(ctx, cfg, botAPI, handler)
	default:
		cancel() // Cancel context before fatal
		log.Fatalf("Unknown bot mode: %s", cfg.BotMode)
	}

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutting down gracefully...")
	cancel()

	// Give the bot some time to finish processing
	time.Sleep(2 * time.Second)
	log.Println("Bot stopped")
}

// runPollingMode runs the bot in polling mode (long polling).
func runPollingMode(ctx context.Context, botAPI *tgbotapi.BotAPI, handler *bot.Handler) {
	log.Println("Starting bot in polling mode...")

	// Create update config
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	// Get update channel
	updates := botAPI.GetUpdatesChan(updateConfig)

	// Process updates
	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping polling...")
			botAPI.StopReceivingUpdates()

			return
		case update := <-updates:
			// Process update in background to avoid blocking
			go func(update tgbotapi.Update) {
				if err := handler.HandleUpdate(ctx, update); err != nil {
					log.Printf("Error handling update: %v", err)
				}
			}(update)
		}
	}
}

// runWebhookMode runs the bot in webhook mode.
func runWebhookMode(ctx context.Context, cfg *config.Config, botAPI *tgbotapi.BotAPI, handler *bot.Handler) {
	log.Println("Starting bot in webhook mode...")

	// Set webhook
	webhookConfig, wErr := tgbotapi.NewWebhook(cfg.WebhookURL)
	if wErr != nil {
		log.Fatalf("Failed to create webhook config: %v", wErr)
	}

	// Note: SecretToken is available in newer versions of telegram-bot-api.
	// If needed, upgrade to v6+ for webhook secret token support.

	if _, err := botAPI.Request(webhookConfig); err != nil {
		log.Fatalf("Failed to set webhook: %v", err)
	}

	info, err := botAPI.GetWebhookInfo()
	if err != nil {
		log.Fatalf("Failed to get webhook info: %v", err)
	}

	if info.LastErrorDate != 0 {
		log.Printf("Webhook last error: %s", info.LastErrorMessage)
	}

	log.Printf("Webhook set to: %s", cfg.WebhookURL)

	// Create HTTP server
	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		// Verify secret token if configured
		if cfg.WebhookSecret != "" {
			token := r.Header.Get("X-Telegram-Bot-Api-Secret-Token")
			if token != cfg.WebhookSecret {
				log.Printf("Invalid webhook secret token")
				w.WriteHeader(http.StatusUnauthorized)

				return
			}
		}

		// Parse update
		update, err := botAPI.HandleUpdate(r)
		if err != nil {
			log.Printf("Error parsing webhook update: %v", err)
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		// Handle update
		if err := handler.HandleUpdate(ctx, *update); err != nil {
			log.Printf("Error handling webhook update: %v", err)
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		w.WriteHeader(http.StatusOK)
	})

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	// Start HTTP server
	addr := fmt.Sprintf(":%s", cfg.Port)
	server := &http.Server{
		Addr:         addr,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Starting HTTP server on %s", addr)

	// Run server in goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Shutdown server
	log.Println("Shutting down HTTP server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	// Remove webhook
	if _, err := botAPI.Request(tgbotapi.DeleteWebhookConfig{}); err != nil {
		log.Printf("Failed to delete webhook: %v", err)
	}

	log.Println("Webhook mode stopped")
}
