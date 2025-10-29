# Flibusta Kindle Bot

A Telegram bot written in Go that searches for books on flibusta.is and sends them directly to your Kindle device.

## Features

- 🌍 **Multi-language Support**: Greetings and interface in English and Russian
- 🔍 **Natural Search**: Just type book title or author - no commands needed!
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
- **Treats non-command text as search queries automatically**
- Manages conversation state
- Handles inline keyboard callbacks
- Multi-language message formatting
- **Displays whitelist instructions during onboarding**
- **Provides `/whitelist` command for repeat instructions**

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
- Handle Kindle email address validation (format only)
- Retry logic for failed deliveries
- **Error handling for bounce-backs (when available)**
- **Cannot verify if sender is whitelisted by user**

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
- **Email**: Azure Communication Services Email API or `net/smtp`
- **Storage**: Azure Cosmos DB (SQLite for local dev) or Azure Cache for Redis
- **Configuration**: Environment variables / Azure Key Vault
- **Deployment**: Azure Container Apps
- **IaC**: Terraform or Bicep

### Data Flow

#### User Journey

1. **Start**: User sends `/start` command
   - Bot responds with greeting in user's language
   - **CRITICAL**: Displays instructions to whitelist sender email on Amazon
   - Shows step-by-step guide with link to Amazon settings
   - Requests Kindle email if not set
   - ⚠️ **Note**: Bot cannot verify if user whitelisted the sender - user must do this manually

2. **Search**: User types **any text message** (not a command)
   - Bot automatically treats it as a search query
   - Queries flibusta.is for the book by title or author
   - If single result: shows book details with "Send to Kindle" button
   - If multiple results: displays list with inline keyboard for selection
   - **No need for `/search` command** - just type naturally!

3. **Selection**: User selects a book from results (via inline keyboard)
   - Bot downloads the book
   - Sends book to user's Kindle email
   - Confirms email was sent (but cannot guarantee delivery)
   - ⚠️ If book doesn't arrive, reminds user to check whitelist settings

4. **Settings**: User can update preferences via commands
   - `/language` - Change language
   - `/kindle` - Update Kindle email address
   - `/whitelist` - View whitelist instructions again
   - `/help` - Show all available commands

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
│   │   └── sender.go            # Email sender (Azure Communication Services)
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
├── terraform/                   # Infrastructure as Code
│   ├── main.tf                  # Main Terraform configuration
│   ├── variables.tf             # Variable definitions
│   ├── outputs.tf               # Output values
│   ├── container-app.tf         # Azure Container Apps
│   ├── communication-services.tf # Azure Communication Services
│   ├── storage.tf               # Azure Storage & Database
│   └── monitoring.tf            # Application Insights
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

# Azure Communication Services (for Kindle delivery)
ACS_CONNECTION_STRING=endpoint=https://...;accesskey=...
ACS_SENDER_ADDRESS=DoNotReply@your-domain.azurecomm.net

# Alternative: SMTP (if not using ACS)
# SMTP_HOST=smtp.gmail.com
# SMTP_PORT=587
# SMTP_USERNAME=your_email@gmail.com
# SMTP_PASSWORD=your_app_password

# Azure Storage (for user data and book cache)
AZURE_STORAGE_CONNECTION_STRING=DefaultEndpointsProtocol=https;...
DATABASE_CONNECTION_STRING=your_cosmos_db_or_postgres_connection

# Application
TEMP_DIR=./tmp
LOG_LEVEL=info
AZURE_APPINSIGHTS_KEY=your_application_insights_key
```

### Bot Commands

- `/start` - Initialize bot, show whitelist instructions, and set preferences
- `/language` - Change interface language
- `/kindle` - Set or update Kindle email address
- `/whitelist` - Show Amazon whitelist instructions again
- `/settings` - View and update all preferences
- `/help` - Show help information and setup guide

**Note**: You don't need a `/search` command - just type the book title or author name directly!

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

## Azure Deployment

### Infrastructure Overview

The bot is deployed on Azure using the following services:

- **Azure Container Apps**: Hosts the Go application (serverless, auto-scaling)
- **Azure Communication Services**: Email delivery to Kindle devices
- **Azure Cosmos DB** or **Azure Database for PostgreSQL**: User preferences and session data
- **Azure Storage Account**: Temporary book file storage
- **Azure Key Vault**: Secure storage for secrets and API keys
- **Azure Application Insights**: Monitoring, logging, and diagnostics
- **Azure Container Registry**: Docker image storage

### Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                         Azure Cloud                          │
│                                                               │
│  ┌────────────────────┐          ┌──────────────────────┐  │
│  │  Container Apps    │          │   Key Vault          │  │
│  │  (Go Bot)          │◀────────▶│  (Secrets)           │  │
│  └────────────────────┘          └──────────────────────┘  │
│           │                                                  │
│           ├──────────────────┬──────────────────┬─────────┐│
│           ▼                  ▼                  ▼          ││
│  ┌─────────────────┐ ┌─────────────────┐ ┌──────────────┐││
│  │ Communication   │ │  Cosmos DB /    │ │   Storage    │││
│  │ Services (Email)│ │  PostgreSQL     │ │   Account    │││
│  └─────────────────┘ └─────────────────┘ └──────────────┘││
│           │                                                  │
│           ▼                                                  │
│  ┌─────────────────┐          ┌──────────────────────┐    │
│  │ App Insights    │          │  Container Registry  │    │
│  │ (Monitoring)    │          │  (Docker Images)     │    │
│  └─────────────────┘          └──────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
         │
         ▼
   ┌──────────────┐
   │   Kindle     │
   │   Devices    │
   └──────────────┘
```

### Deployment Options

#### Option 1: Terraform (Recommended)

Terraform provides infrastructure as code with excellent Azure support.

**Pros:**
- Multi-cloud support (if you need it later)
- Large community and module ecosystem
- State management
- Plan/preview before apply

**Setup:**
```bash
cd terraform
terraform init
terraform plan
terraform apply
```

#### Option 2: Bicep (Azure-Native)

Bicep is Microsoft's domain-specific language for Azure.

**Pros:**
- Native Azure support
- Simpler syntax than ARM templates
- Better IntelliSense in VS Code
- Directly integrated with Azure CLI

**Setup:**
```bash
az deployment group create \
  --resource-group flibusta-bot-rg \
  --template-file main.bicep
```

### Kindle Email Configuration

⚠️ **CRITICAL**: Users must whitelist your sender email before they can receive books!

**Important**: There is **no programmatic way** to verify if a user has whitelisted your sender email. The bot must:
1. Display clear instructions during `/start`
2. Provide a `/whitelist` command to show instructions again
3. Remind users if they report books not arriving
4. Handle email bounce-backs gracefully (though Amazon may not send them)

#### What Users Must Do:

The bot should display these instructions clearly during onboarding:

```
📧 IMPORTANT: Before you can receive books, you MUST whitelist our sender email!

Follow these steps:
1️⃣ Go to: https://www.amazon.com/hz/mycd/myx#/home/settings/payment
2️⃣ Click "Preferences" → "Personal Document Settings"
3️⃣ Under "Approved Personal Document E-mail List", click "Add a new approved e-mail address"
4️⃣ Enter: DoNotReply@yourbot.azurecomm.net
5️⃣ Click "Add Address"

✅ Done! Now you can receive books on your Kindle.

Your Kindle email address (found in the same page) looks like: username@kindle.com
```

#### Steps for Users (Detailed):

1. **Go to Amazon Account Settings**
   - Visit: https://www.amazon.com/hz/mycd/myx#/home/settings/payment
   - Or: Amazon → Account & Lists → Manage Your Content and Devices

2. **Navigate to Personal Document Settings**
   - Click "Preferences" tab
   - Find "Personal Document Settings"

3. **Add Approved Email**
   - Under "Approved Personal Document E-mail List"
   - Click "Add a new approved e-mail address"
   - Enter your bot's sender email (e.g., `DoNotReply@yourbot.azurecomm.net`)
   - Click "Add Address"

4. **Find Kindle Email**
   - Under "Send-to-Kindle E-Mail Settings"
   - Each device has a unique email like `username@kindle.com`
   - Users provide this email to the bot via `/kindle` command

#### Bot Implementation Notes:

- **Cannot Validate**: No API exists to check if sender is whitelisted
- **Cannot Guarantee Delivery**: Email might be sent but not delivered
- **User Responsibility**: Make this clear in all communications
- **Error Handling**: Catch SMTP errors, but most delivery failures are silent
- **Help Command**: Always make whitelist instructions easily accessible

#### Kindle Email Limitations:

- **File Size**: Max 50 MB per email
- **Supported Formats**: 
  - ✅ MOBI (native, recommended)
  - ✅ EPUB (auto-converted to AZW3)
  - ✅ PDF (preserves formatting)
  - ✅ DOC, DOCX, TXT
- **Delivery Time**: Usually instant, but can take a few minutes
- **Subject Line**: Use "Convert" in subject to convert formats automatically

### Email Service Comparison

#### Azure Communication Services (Recommended)

**Pricing:**
- Free tier: 500 emails/month
- Paid: $0.00025 per email (~$2.50 per 10,000 emails)

**Setup:**
```bash
# Create Communication Service
az communication create \
  --name flibusta-bot-acs \
  --resource-group flibusta-bot-rg

# Get connection string
az communication list-key \
  --name flibusta-bot-acs \
  --resource-group flibusta-bot-rg
```

**Pros:**
- Native Azure integration
- High deliverability
- Built-in monitoring
- No domain required (use @*.azurecomm.net)

**Cons:**
- Relatively new service
- Limited to Azure ecosystem

#### Alternative: Third-Party SMTP

If you prefer traditional SMTP (SendGrid, Mailgun, etc.):

```env
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USERNAME=apikey
SMTP_PASSWORD=your_sendgrid_api_key
```

### Deployment Steps

1. **Prerequisites**
   ```bash
   # Install Azure CLI
   az login
   
   # Install Terraform
   # Download from: https://www.terraform.io/downloads
   ```

2. **Setup GitHub Repository Secrets**
   
   Navigate to your GitHub repository → Settings → Secrets and variables → Actions
   
   Add the following secrets:
   
   ```
   AZURE_CREDENTIALS              # Service Principal JSON
   AZURE_CONTAINER_REGISTRY       # yourregistry.azurecr.io
   AZURE_REGISTRY_USERNAME        # Registry username
   AZURE_REGISTRY_PASSWORD        # Registry password
   TELEGRAM_BOT_TOKEN             # From @BotFather
   ACS_CONNECTION_STRING          # Azure Communication Services
   ACS_SENDER_ADDRESS             # DoNotReply@yourbot.azurecomm.net
   DATABASE_CONNECTION_STRING     # Cosmos DB or PostgreSQL
   AZURE_APPINSIGHTS_KEY          # Application Insights key
   CODECOV_TOKEN                  # (Optional) For code coverage
   ```

3. **Create Azure Service Principal**
   
   ```bash
   # Create service principal for GitHub Actions
   az ad sp create-for-rbac \
     --name "github-actions-flibusta-bot" \
     --role contributor \
     --scopes /subscriptions/{subscription-id}/resourceGroups/flibusta-bot-rg \
     --sdk-auth
   
   # Copy the output JSON to AZURE_CREDENTIALS secret
   ```

4. **Configure Secrets**
   ```bash
   # Create Key Vault
   az keyvault create \
     --name flibusta-bot-kv \
     --resource-group flibusta-bot-rg \
     --location eastus
   
   # Add secrets
   az keyvault secret set --vault-name flibusta-bot-kv \
     --name telegram-bot-token --value "your_token"
   ```

5. **Deploy Infrastructure (One-time)**
   ```bash
   cd terraform
   terraform init
   terraform plan -out=tfplan
   terraform apply tfplan
   ```

6. **Deploy Application via GitHub Actions**
   
   **Option A: Automatic (Push to main)**
   ```bash
   git checkout main
   git push origin main
   # GitHub Actions will automatically deploy
   ```
   
   **Option B: Manual (Workflow Dispatch)**
   - Go to GitHub → Actions → "Deploy to Azure"
   - Click "Run workflow"
   - Select environment (production/staging)
   - Click "Run workflow"

## CI/CD Pipeline

### GitHub Actions Workflows

#### 1. PR Checks (`pr-checks.yml`)

Runs on every pull request and push to main/develop branches.

**Jobs:**
- **Lint**: Code quality checks with `golangci-lint`
- **Test**: Unit tests with coverage reporting
  - Runs tests with race detector
  - Generates coverage report
  - Uploads to Codecov
  - Enforces 60% coverage threshold
- **Build**: Compiles the Go application
- **Security**: Scans code with Gosec
- **Docker Build**: Tests Docker image build
- **Summary**: Generates PR check summary

**Status Checks:**
All checks must pass before merging to main.

#### 2. Deployment (`deploy.yml`)

Runs on:
- Push to `main` branch (automatic deployment)
- Manual trigger via workflow dispatch

**Jobs:**
- **Build and Push**: 
  - Builds Docker image
  - Pushes to Azure Container Registry
  - Tags with commit SHA and branch name
  - Caches layers for faster builds
  
- **Deploy to Azure**:
  - Logs into Azure
  - Deploys to Azure Container Apps
  - Configures environment variables
  - Verifies deployment health
  - Generates deployment summary
  
- **Rollback** (on failure):
  - Automatically rolls back to previous revision
  - Activates last known good version
  
- **Notify**:
  - Can be configured to send Slack/Teams notifications
  - Reports deployment status

**Environments:**
- `production`: Main branch deployments
- `staging`: Manual workflow dispatch option

### Workflow Triggers

```yaml
# PR Checks - Automatic
on:
  pull_request:
    branches: [main, develop]
  push:
    branches: [main, develop]

# Deployment - Automatic on main, manual option
on:
  push:
    branches: [main]
  workflow_dispatch:
    inputs:
      environment:
        type: choice
        options: [production, staging]
```

### Branch Protection Rules

Recommended GitHub branch protection settings for `main`:

- ✅ Require pull request before merging
- ✅ Require status checks to pass:
  - Lint Code
  - Run Tests
  - Build Application
  - Security Scan
  - Docker Build Test
- ✅ Require branches to be up to date
- ✅ Require linear history
- ✅ Do not allow force pushes
- ✅ Do not allow deletions

### Development Workflow

1. **Create Feature Branch**
   ```bash
   git checkout -b feature/add-new-feature
   ```

2. **Make Changes & Commit**
   ```bash
   git add .
   git commit -m "feat: add new feature"
   ```

3. **Push & Create PR**
   ```bash
   git push origin feature/add-new-feature
   # Create PR on GitHub
   ```

4. **PR Checks Run Automatically**
   - Linting
   - Tests (must pass with 60%+ coverage)
   - Build verification
   - Security scan
   - Docker build test

5. **Code Review & Merge**
   - After approval and passing checks
   - Merge to `main`

6. **Automatic Deployment**
   - GitHub Actions automatically deploys to Azure
   - Can monitor progress in Actions tab
   - Deployment summary posted to PR

### Monitoring Deployments

**GitHub Actions Dashboard:**
- Go to repository → Actions tab
- View workflow runs
- Check logs for each job
- Download artifacts (coverage reports)

**Azure Portal:**
- Container Apps → flibusta-bot
- View logs and metrics
- Monitor resource usage
- Check application health

**Application Insights:**
- Real-time monitoring
- Request traces
- Error tracking
- Performance metrics

### Monitoring

Azure Application Insights provides:
- Real-time metrics
- Request/response logging
- Error tracking
- Performance monitoring
- Custom telemetry

Access via Azure Portal → Application Insights → flibusta-bot-insights

### Cost Estimation (Monthly)

- **Container Apps**: ~$20-40 (depends on usage)
- **Cosmos DB**: ~$25 (serverless) or $5 (shared)
- **Communication Services**: Free tier (500 emails) or ~$2.50/10k emails
- **Storage Account**: ~$1-5
- **Application Insights**: ~$5-10
- **Key Vault**: ~$0.03

**Total: ~$50-80/month** (with free tiers, can be as low as $30)

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
