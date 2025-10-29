# Architecture Documentation

This document provides detailed technical architecture and design decisions for the Flibusta Kindle Bot.

## Table of Contents

- [Overview](#overview)
- [System Architecture](#system-architecture)
- [Components](#components)
- [Data Flow](#data-flow)
- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [Design Decisions](#design-decisions)
- [Security](#security)
- [Performance](#performance)
- [Future Enhancements](#future-enhancements)

## Overview

The Flibusta Kindle Bot is a Telegram-based service that enables users to search for books on flibusta.is and have them delivered directly to their Kindle devices via email.

### Key Characteristics

- **Serverless Architecture**: Runs on Azure Container Apps with auto-scaling
- **Event-Driven**: Responds to Telegram updates (webhook or polling)
- **Stateless Application**: Session data stored externally (database)
- **Microservices Pattern**: Internal packages organized by responsibility
- **Multi-language Support**: i18n architecture for extensibility

## System Architecture

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
                        │  Email (ACS) │
                        │  to Kindle   │
                        └──────────────┘
```

### Detailed Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                        Telegram API                           │
└────────────────┬─────────────────────────────────────────────┘
                 │ Webhook / Long Polling
                 ▼
┌──────────────────────────────────────────────────────────────┐
│                    Azure Container Apps                       │
│                                                               │
│  ┌────────────────────────────────────────────────────────┐ │
│  │              Flibusta Kindle Bot (Go)                  │ │
│  │                                                         │ │
│  │  ┌──────────┐  ┌──────────┐  ┌────────────┐          │ │
│  │  │ Bot      │  │ Search   │  │ Downloader │          │ │
│  │  │ Handler  │─▶│ Engine   │─▶│            │          │ │
│  │  └──────────┘  └──────────┘  └─────┬──────┘          │ │
│  │       │                              │                 │ │
│  │       │  ┌──────────┐  ┌─────────────▼──────┐        │ │
│  │       └─▶│ User     │  │ Kindle              │        │ │
│  │          │ Manager  │  │ Sender              │        │ │
│  │          └──────────┘  └────────────┬────────┘        │ │
│  │               │                      │                 │ │
│  │  ┌────────────▼──────────┐          │                 │ │
│  │  │ i18n                  │          │                 │ │
│  │  │ (Localization)        │          │                 │ │
│  │  └───────────────────────┘          │                 │ │
│  └─────────────────────────────────────┼─────────────────┘ │
└────────────────────────────────────────┼───────────────────┘
                 │                        │
     ┌───────────┴──────┐        ┌───────▼─────────┐
     ▼                  ▼        ▼                 ▼
┌──────────┐      ┌──────────┐  ┌──────────┐  ┌──────────┐
│ Cosmos DB│      │ Azure    │  │  Azure   │  │  Kindle  │
│ (Users)  │      │ Storage  │  │   ACS    │  │  Device  │
└──────────┘      │ (Books)  │  │ (Email)  │  └──────────┘
                  └──────────┘  └──────────┘
```

## Components

### 1. Bot Handler (`internal/bot`)

**Responsibility**: Telegram API integration and message routing

**Key Functions**:
- Receive Telegram updates (messages, callbacks)
- Route commands to appropriate handlers
- **Treat non-command text as search queries**
- Manage conversation state
- Generate inline keyboards
- Send responses with formatting

**Design Pattern**: Command Pattern + Router

```go
type Handler struct {
    bot          *tgbotapi.BotAPI
    searchEngine *search.Engine
    userManager  *user.Manager
    kindleSender *kindle.Sender
    i18n         *i18n.I18n
}

func (h *Handler) HandleUpdate(update tgbotapi.Update) error {
    if update.Message != nil {
        return h.handleMessage(update.Message)
    }
    if update.CallbackQuery != nil {
        return h.handleCallback(update.CallbackQuery)
    }
    return nil
}

func (h *Handler) handleMessage(msg *tgbotapi.Message) error {
    // Commands start with /
    if msg.IsCommand() {
        return h.handleCommand(msg)
    }
    // Everything else is a search query
    return h.handleSearchQuery(msg)
}
```

**State Management**:
- User preferences stored in database
- Active searches cached in memory (with TTL)
- Callback data includes context (book ID, action)

### 2. Search Engine (`internal/search`)

**Responsibility**: Web scraping and book discovery

**Key Functions**:
- HTTP requests to flibusta.is
- HTML parsing (using goquery or colly)
- Extract book metadata (title, author, format, size)
- Handle search pagination
- Return structured results

**Design Pattern**: Repository Pattern

```go
type Engine struct {
    client   *http.Client
    parser   *Parser
    cache    *Cache
}

type SearchResult struct {
    Books      []Book
    TotalCount int
    HasMore    bool
}

type Book struct {
    ID       string
    Title    string
    Author   string
    Format   string
    Size     int64
    URL      string
}

func (e *Engine) Search(query string) (*SearchResult, error) {
    // Check cache first
    if cached := e.cache.Get(query); cached != nil {
        return cached, nil
    }
    
    // Fetch from flibusta.is
    html, err := e.fetchSearchResults(query)
    if err != nil {
        return nil, err
    }
    
    // Parse HTML
    results := e.parser.ParseSearchPage(html)
    
    // Cache results
    e.cache.Set(query, results, 15*time.Minute)
    
    return results, nil
}
```

**Challenges**:
- Website structure may change (brittle)
- Rate limiting (add delays between requests)
- Captchas (use headless browser if needed)

### 3. Book Downloader (`internal/downloader`)

**Responsibility**: Download and prepare book files

**Key Functions**:
- Download book from URL
- Verify file integrity
- Convert formats if needed (EPUB → MOBI)
- Handle temporary file storage
- Clean up after sending

**Design Pattern**: Strategy Pattern (for format conversion)

```go
type Downloader struct {
    client     *http.Client
    storage    *storage.Manager
    converters map[string]Converter
}

func (d *Downloader) Download(bookURL string) (*BookFile, error) {
    // Download to temp file
    tempFile, err := d.storage.CreateTemp("book-*")
    if err != nil {
        return nil, err
    }
    
    resp, err := d.client.Get(bookURL)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    // Copy to file
    _, err = io.Copy(tempFile, resp.Body)
    if err != nil {
        return nil, err
    }
    
    return &BookFile{
        Path:   tempFile.Name(),
        Format: detectFormat(tempFile),
        Size:   getFileSize(tempFile),
    }, nil
}

func (d *Downloader) Convert(file *BookFile, toFormat string) error {
    converter, ok := d.converters[toFormat]
    if !ok {
        return fmt.Errorf("no converter for %s", toFormat)
    }
    return converter.Convert(file)
}
```

**File Management**:
- Use `/tmp` for temporary storage
- Auto-cleanup after 1 hour (or immediate after send)
- Monitor disk usage

### 4. Kindle Sender (`internal/kindle`)

**Responsibility**: Email delivery to Kindle devices

**Key Functions**:
- Send email via Azure Communication Services
- Attach book files
- Handle SMTP errors
- Retry logic for transient failures
- Validate Kindle email format

**Design Pattern**: Adapter Pattern (for email service)

```go
type Sender struct {
    emailClient EmailClient
    senderEmail string
}

type EmailClient interface {
    Send(to, subject, body string, attachments []Attachment) error
}

func (s *Sender) SendToKindle(userEmail string, book *BookFile) error {
    // Validate email format
    if err := s.validateKindleEmail(userEmail); err != nil {
        return err
    }
    
    // Check file size
    if book.Size > 50*1024*1024 {
        return errors.New("file too large (max 50MB)")
    }
    
    // Prepare attachment
    attachment := Attachment{
        Name:        fmt.Sprintf("book.%s", book.Format),
        ContentType: getMimeType(book.Format),
        Data:        book.Data,
    }
    
    // Send with retries
    return retry.Do(func() error {
        return s.emailClient.Send(
            userEmail,
            "Your Book",
            "Your requested book is attached.",
            []Attachment{attachment},
        )
    }, retry.Attempts(3), retry.Delay(time.Second))
}

func (s *Sender) validateKindleEmail(email string) error {
    if !strings.HasSuffix(email, "@kindle.com") {
        return errors.New("not a valid Kindle email")
    }
    return nil
}
```

**Important**: Cannot verify if sender is whitelisted!

### 5. User Manager (`internal/user`)

**Responsibility**: User data and preferences

**Key Functions**:
- Store user preferences (language, Kindle email)
- Track user state (active search, selected book)
- Session management
- User analytics (optional)

**Design Pattern**: Repository Pattern

```go
type Manager struct {
    repo Repository
}

type Repository interface {
    GetUser(telegramID int64) (*User, error)
    SaveUser(user *User) error
    UpdatePreferences(telegramID int64, prefs *Preferences) error
}

type User struct {
    TelegramID   int64
    KindleEmail  string
    Language     string
    CreatedAt    time.Time
    UpdatedAt    time.Time
    ActiveSearch *SearchContext
}

type SearchContext struct {
    Query      string
    Results    []Book
    ExpiresAt  time.Time
}

func (m *Manager) GetOrCreateUser(telegramID int64, langCode string) (*User, error) {
    user, err := m.repo.GetUser(telegramID)
    if err == ErrUserNotFound {
        // Create new user
        user = &User{
            TelegramID: telegramID,
            Language:   detectLanguage(langCode),
            CreatedAt:  time.Now(),
        }
        if err := m.repo.SaveUser(user); err != nil {
            return nil, err
        }
    }
    return user, nil
}
```

**Database Schema** (Cosmos DB or PostgreSQL):
```json
{
  "id": "user_123456789",
  "telegram_id": 123456789,
  "kindle_email": "user@kindle.com",
  "language": "en",
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-15T12:30:00Z",
  "statistics": {
    "books_sent": 42,
    "last_search": "2025-01-15T12:30:00Z"
  }
}
```

### 6. Localization (`internal/i18n`)

**Responsibility**: Multi-language support

**Key Functions**:
- Load translation files
- Translate messages by key
- Pluralization
- Template variable substitution

**Design Pattern**: Strategy Pattern

```go
type I18n struct {
    translations map[string]map[string]string // lang -> key -> value
}

func (i *I18n) T(lang, key string, args ...interface{}) string {
    template := i.translations[lang][key]
    if template == "" {
        // Fallback to English
        template = i.translations["en"][key]
    }
    return fmt.Sprintf(template, args...)
}

func (i *I18n) LoadTranslations(lang string, path string) error {
    data, err := os.ReadFile(path)
    if err != nil {
        return err
    }
    
    var translations map[string]string
    if err := json.Unmarshal(data, &translations); err != nil {
        return err
    }
    
    i.translations[lang] = translations
    return nil
}
```

**Translation Files** (`internal/i18n/locales/`):

`en.json`:
```json
{
  "welcome": "Welcome to Flibusta Kindle Bot!",
  "search_prompt": "Send me a book title or author name to search.",
  "book_sent": "✅ Book sent to %s!",
  "whitelist_required": "⚠️ Please whitelist %s in your Amazon account."
}
```

`ru.json`:
```json
{
  "welcome": "Добро пожаловать в Flibusta Kindle Bot!",
  "search_prompt": "Отправьте мне название книги или имя автора.",
  "book_sent": "✅ Книга отправлена на %s!",
  "whitelist_required": "⚠️ Пожалуйста, добавьте %s в белый список Amazon."
}
```

## Data Flow

### Search Flow

```
User types: "1984"
    │
    ▼
Handler detects non-command text
    │
    ▼
SearchEngine.Search("1984")
    │
    ├─▶ Check cache
    │   └─▶ Cache hit? Return results
    │
    ├─▶ HTTP GET flibusta.is/search?q=1984
    │
    ├─▶ Parse HTML
    │
    └─▶ Return []Book
    │
    ▼
If single result:
    Show book with "Send to Kindle" button
Else:
    Show list with inline keyboard
    │
    ▼
User clicks book or "Send" button
    │
    ▼
Callback handler
    │
    ▼
Downloader.Download(book.URL)
    │
    ▼
KindleSender.SendToKindle(user.KindleEmail, bookFile)
    │
    ▼
User receives book on Kindle
```

### User Onboarding Flow

```
User: /start
    │
    ▼
Handler checks if user exists
    │
    ├─▶ New user
    │   ├─▶ Create user record
    │   ├─▶ Detect language (from Telegram)
    │   └─▶ Show welcome + whitelist instructions
    │
    └─▶ Existing user
        └─▶ Show welcome back message
    │
    ▼
Handler asks for Kindle email
    │
    ▼
User: username@kindle.com
    │
    ▼
Handler validates format
    │
    ├─▶ Valid: Save to database
    └─▶ Invalid: Show error + example
    │
    ▼
Ready to search!
```

## Tech Stack

### Core Technologies

| Technology | Purpose | Version |
|-----------|---------|---------|
| **Go** | Application language | 1.21+ |
| **Telegram Bot API** | Bot framework | [go-telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api) |
| **goquery** | HTML parsing | [PuerkitoBio/goquery](https://github.com/PuerkitoBio/goquery) |
| **Azure SDK** | Cloud services | [Azure SDK for Go](https://github.com/Azure/azure-sdk-for-go) |

### Infrastructure

| Service | Purpose |
|---------|---------|
| **Azure Container Apps** | Application hosting |
| **Azure Communication Services** | Email delivery |
| **Azure Cosmos DB** | User data storage |
| **Azure Storage** | Temporary file storage |
| **Azure Key Vault** | Secrets management |
| **Azure App Insights** | Monitoring |

## Project Structure

```
flibusta_kindle_bot/
├── cmd/
│   └── bot/
│       └── main.go              # Entry point
├── internal/                    # Private application code
│   ├── bot/
│   │   ├── handler.go
│   │   ├── callbacks.go
│   │   └── middleware.go
│   ├── search/
│   │   ├── engine.go
│   │   └── parser.go
│   ├── downloader/
│   │   ├── downloader.go
│   │   └── converter.go
│   ├── kindle/
│   │   └── sender.go
│   ├── user/
│   │   ├── manager.go
│   │   └── repository.go
│   └── i18n/
│       ├── i18n.go
│       └── locales/
├── pkg/                         # Public libraries (if needed)
│   └── models/
│       ├── book.go
│       └── user.go
├── docs/                        # Documentation
├── terraform/                   # Infrastructure as Code
└── .github/workflows/           # CI/CD
```

### Package Organization

**`internal/`**: Private packages, cannot be imported by other projects

**`pkg/`**: Public packages, can be imported (minimal for this project)

**`cmd/`**: Application entry points

## Design Decisions

### Why Go?

- ✅ Excellent performance
- ✅ Simple concurrency (goroutines)
- ✅ Strong standard library
- ✅ Fast compilation
- ✅ Easy deployment (single binary)

### Why Azure Container Apps?

- ✅ Serverless (auto-scale to zero)
- ✅ Built-in HTTPS
- ✅ Easy deployment
- ✅ Cost-effective for low traffic
- ✅ Managed infrastructure

### Why Not AWS Lambda?

- ❌ Cold start issues for Go
- ❌ More complex setup
- ✅ Could be alternative deployment target

### Why Cosmos DB over PostgreSQL?

For **small scale**: Either works

For **large scale**:
- ✅ Cosmos DB: Global distribution, auto-scaling
- ❌ PostgreSQL: Better for complex queries, cheaper

**Recommendation**: Start with PostgreSQL, migrate to Cosmos if needed

### Why Azure Communication Services?

- ✅ Native Azure integration
- ✅ No domain required
- ✅ Free tier (500 emails/month)
- ✅ High deliverability

**Alternative**: SendGrid, Mailgun (more mature, higher costs)

## Security

### Secrets Management

- ✅ Azure Key Vault for all secrets
- ✅ No secrets in code or config files
- ✅ Managed identity for Azure services
- ✅ Environment variables for local dev

### Input Validation

- ✅ Sanitize user inputs
- ✅ Validate email formats
- ✅ Limit file sizes
- ✅ Rate limiting per user

### Data Privacy

- ✅ Encrypt Kindle emails at rest
- ✅ No book content stored long-term
- ✅ GDPR compliance (data deletion on request)
- ✅ Minimal logging of user data

## Performance

### Optimization Strategies

1. **Caching**: Search results cached for 15 minutes
2. **Concurrent Downloads**: Use goroutines for parallel processing
3. **Connection Pooling**: HTTP client with keep-alive
4. **Lazy Loading**: Only download book when user confirms
5. **Rate Limiting**: Prevent abuse and resource exhaustion

### Scaling

- **Horizontal**: Container Apps auto-scale based on HTTP requests
- **Database**: Cosmos DB serverless auto-scales
- **Storage**: Azure Storage handles any load

### Monitoring

- Application Insights for:
  - Request rates
  - Error rates
  - Response times
  - Custom metrics (books sent, searches performed)

## Future Enhancements

### Phase 1 (MVP) - Current
- ✅ Basic search and send
- ✅ English and Russian support
- ✅ Azure deployment

### Phase 2
- [ ] More formats support (FB2, DJVU)
- [ ] Format conversion (using Calibre)
- [ ] Book preview before sending
- [ ] Favorites/reading list

### Phase 3
- [ ] More languages (Spanish, French, German)
- [ ] User statistics dashboard
- [ ] Recommendations based on reading history
- [ ] Integration with Goodreads

### Phase 4
- [ ] Mobile app (using Telegram Mini Apps)
- [ ] Premium features (larger file sizes, priority sending)
- [ ] Social features (share recommendations)

## Related Documentation

- [Deployment Guide](DEPLOYMENT.md)
- [CI/CD Pipeline](CI_CD.md)
- [Kindle Setup](KINDLE_SETUP.md)

---

**Questions or suggestions?** Open an issue on GitHub!
