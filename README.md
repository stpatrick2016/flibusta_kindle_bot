# Flibusta Kindle Bot

A Telegram bot written in Go that searches for books on flibusta.is and sends them directly to your Kindle device.

## Features

- 🌍 **Multi-language Support**: Greetings and interface in English and Russian
- 🔍 **Smart Search**: Search books by title or author
- 📚 **Multiple Results**: Interactive selection when multiple books are found
- 📧 **Kindle Delivery**: Direct book delivery to your Kindle email address
- 🤖 **User-Friendly**: Conversational interface with inline keyboards

## Architecture

### High-Level Design

```
┌─────────────┐         ┌──────────────┐         ┌─────────────┐
│   Telegram  │────────▶│  Bot Server  │────────▶│ Flibusta.is │
│    User     │◀────────│   (Go App)   │◀────────│   Scraper   │
└─────────────┘         └──────────────┘         └─────────────┘
                              │
                              │
                              ▼
                        ┌──────────────┐
                        │    SMTP      │
                        │   Server     │
                        │  (to Kindle) │
                        └──────────────┘
```

### Components

#### 1. **Telegram Bot Handler** (`internal/bot`)
- Receives and processes user messages
- Manages conversation state
- Handles inline keyboard callbacks
- Multi-language message formatting

#### 2. **Search Engine** (`internal/search`)
- Web scraping interface for flibusta.is
- Search by title and author
- Parse search results and book metadata
- Handle pagination for multiple results

#### 3. **Book Downloader** (`internal/downloader`)
- Download book files from flibusta.is
- Support multiple formats (MOBI, EPUB, etc.)
- Format conversion if needed (EPUB to MOBI)
- Temporary file management

#### 4. **Kindle Sender** (`internal/kindle`)
- SMTP client for sending emails
- Format books as email attachments
- Handle Kindle email address validation
- Retry logic for failed deliveries

#### 5. **User Manager** (`internal/user`)
- Store user preferences (language, Kindle email)
- Session management
- User state tracking (search context)

#### 6. **Localization** (`internal/i18n`)
- Multi-language message templates
- Language detection and switching
- Currently supports: English, Russian

### Tech Stack

- **Language**: Go 1.21+
- **Telegram API**: [telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api)
- **Web Scraping**: [colly](https://github.com/gocolly/colly) or [goquery](https://github.com/PuerkitoBio/goquery)
- **Email**: Go's built-in `net/smtp`
- **Storage**: SQLite for user data (or Redis for session management)
- **Configuration**: Environment variables / `.env` file

### Data Flow

#### User Journey

1. **Start**: User sends `/start` command
   - Bot responds with greeting in user's language
   - Requests Kindle email if not set

2. **Search**: User sends book title or author name
   - Bot queries flibusta.is
   - If single result: shows book details with "Send to Kindle" button
   - If multiple results: displays list with inline keyboard

3. **Selection**: User selects a book from results
   - Bot downloads the book
   - Sends book to user's Kindle email
   - Confirms delivery

4. **Settings**: User can update preferences
   - Change language
   - Update Kindle email address

### Project Structure

```
flibusta_kindle_bot/
├── cmd/
│   └── bot/
│       └── main.go              # Application entry point
├── internal/
│   ├── bot/
│   │   ├── handler.go           # Message handlers
│   │   ├── callbacks.go         # Inline keyboard callbacks
│   │   └── middleware.go        # Middleware (logging, auth)
│   ├── search/
│   │   ├── flibusta.go          # Flibusta scraper
│   │   └── parser.go            # HTML parsing logic
│   ├── downloader/
│   │   ├── downloader.go        # Book download logic
│   │   └── converter.go         # Format conversion
│   ├── kindle/
│   │   └── sender.go            # SMTP email sender
│   ├── user/
│   │   ├── repository.go        # User data storage
│   │   └── session.go           # Session management
│   └── i18n/
│       ├── locales/             # Translation files
│       │   ├── en.json
│       │   └── ru.json
│       └── i18n.go              # Localization logic
├── pkg/
│   └── models/
│       ├── book.go              # Book data structure
│       └── user.go              # User data structure
├── configs/
│   └── config.yaml              # Configuration template
├── migrations/
│   └── 001_init.sql             # Database schema
├── .env.example                 # Environment variables template
├── .gitignore
├── go.mod
├── go.sum
├── Dockerfile                   # Docker container setup
├── docker-compose.yml           # Docker compose for local dev
└── README.md
```

### Configuration

The bot requires the following environment variables:

```env
# Telegram
TELEGRAM_BOT_TOKEN=your_bot_token_here

# SMTP (for Kindle delivery)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your_email@gmail.com
SMTP_PASSWORD=your_app_password
SMTP_FROM=your_email@gmail.com

# Application
DATABASE_PATH=./data/bot.db
TEMP_DIR=./tmp
LOG_LEVEL=info
```

### Bot Commands

- `/start` - Initialize bot and set preferences
- `/search <query>` - Search for books by title or author
- `/settings` - Update user preferences
- `/language` - Change interface language
- `/kindle` - Set or update Kindle email address
- `/help` - Show help information

### Development Phases

#### Phase 1: Basic Bot Setup ✅ (Planning)
- [x] Initialize Git repository
- [x] Create project structure
- [x] Document architecture
- [ ] Setup Go modules
- [ ] Basic Telegram bot connection

#### Phase 2: Search Functionality
- [ ] Implement Flibusta scraper
- [ ] Parse search results
- [ ] Handle single/multiple results
- [ ] Display book information

#### Phase 3: Download & Send
- [ ] Download book from Flibusta
- [ ] Format conversion (if needed)
- [ ] SMTP integration for Kindle delivery
- [ ] Delivery confirmation

#### Phase 4: User Management
- [ ] Database setup (SQLite)
- [ ] User preferences storage
- [ ] Session management
- [ ] Settings commands

#### Phase 5: Localization
- [ ] i18n framework setup
- [ ] English translations
- [ ] Russian translations
- [ ] Language switching

#### Phase 6: Production Ready
- [ ] Error handling and logging
- [ ] Rate limiting
- [ ] Docker containerization
- [ ] Deployment documentation

## Security Considerations

- **API Keys**: Never commit API keys or passwords to git
- **Email**: Use app-specific passwords for SMTP
- **User Data**: Encrypt sensitive user information (Kindle emails)
- **Rate Limiting**: Implement rate limiting to prevent abuse
- **Input Validation**: Sanitize all user inputs

## Legal Notice

⚠️ **Important**: This bot is for educational purposes. Make sure you comply with:
- Flibusta.is terms of service
- Copyright laws in your jurisdiction
- Telegram Bot API terms of service
- Amazon Kindle terms of service

Users are responsible for ensuring they have the right to download and distribute the books they search for.

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Git
- A Telegram Bot Token (from [@BotFather](https://t.me/botfather))
- SMTP credentials (Gmail, etc.)
- A Kindle device with an email address

### Installation

```bash
# Clone the repository
git clone <repository-url>
cd flibusta_kindle_bot

# Install dependencies
go mod download

# Copy environment template
cp .env.example .env

# Edit .env with your credentials
nano .env

# Run the bot
go run cmd/bot/main.go
```

### Docker

```bash
# Build and run with docker-compose
docker-compose up -d

# View logs
docker-compose logs -f
```

## Contributing

Contributions are welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## License

MIT License - see LICENSE file for details

## Support

For issues and questions, please open an issue on GitHub.

---

**Happy Reading! 📚**
