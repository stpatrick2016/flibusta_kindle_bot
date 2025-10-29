# Flibusta Kindle Bot

A Telegram bot written in Go that searches for books on flibusta.is and sends them directly to your Kindle device.

[![Deploy to Azure](https://github.com/yourusername/flibusta_kindle_bot/actions/workflows/deploy.yml/badge.svg)](https://github.com/yourusername/flibusta_kindle_bot/actions/workflows/deploy.yml)
[![PR Checks](https://github.com/yourusername/flibusta_kindle_bot/actions/workflows/pr-checks.yml/badge.svg)](https://github.com/yourusername/flibusta_kindle_bot/actions/workflows/pr-checks.yml)
[![codecov](https://codecov.io/gh/yourusername/flibusta_kindle_bot/branch/main/graph/badge.svg)](https://codecov.io/gh/yourusername/flibusta_kindle_bot)

## ✨ Features

- 🌍 **Multi-language Support** - Interface in English and Russian
- 🔍 **Natural Search** - Just type book title or author, no commands needed
- 📚 **Smart Results** - Interactive selection when multiple books found
- 📧 **Kindle Delivery** - Direct delivery to your Kindle email address
- 🤖 **User-Friendly** - Conversational interface with inline keyboards
- ☁️ **Cloud-Native** - Deployed on Azure with auto-scaling

## 🚀 Quick Start

### For Users

1. **Start the bot**: Open [@FlibustaKindleBot](https://t.me/your_bot) on Telegram
2. **Setup Kindle**: Follow the instructions to whitelist the sender email → [Setup Guide](docs/KINDLE_SETUP.md)
3. **Set your Kindle email**: Use `/kindle your_email@kindle.com`
4. **Search for books**: Just type the book title or author name
5. **Send to Kindle**: Select the book and click "Send to Kindle"

📖 **Full user guide**: [Kindle Setup Documentation](docs/KINDLE_SETUP.md)

### For Developers

```bash
# Clone the repository
git clone https://github.com/yourusername/flibusta_kindle_bot.git
cd flibusta_kindle_bot

# Install dependencies
go mod download

# Copy environment template
cp .env.example .env
# Edit .env with your credentials

# Run locally
go run cmd/bot/main.go
```

## 📚 Documentation

| Document | Description |
|----------|-------------|
| [Architecture](docs/ARCHITECTURE.md) | System design, components, and technical decisions |
| [Deployment Guide](docs/DEPLOYMENT.md) | Azure deployment with Terraform |
| [CI/CD Pipeline](docs/CI_CD.md) | GitHub Actions workflows and development process |
| [Kindle Setup](docs/KINDLE_SETUP.md) | How to configure Kindle email and whitelist sender |

## 🏗️ Architecture

```
┌─────────────┐         ┌──────────────┐         ┌─────────────┐
│   Telegram  │────────▶│  Bot Server  │────────▶│ Flibusta.is │
│    User     │◀────────│   (Go App)   │◀────────│   Scraper   │
└─────────────┘         └──────────────┘         └─────────────┘
                              │
                              ▼
                        ┌──────────────┐
                        │  Azure ACS   │
                        │   (Email)    │
                        │  to Kindle   │
                        └──────────────┘
```

**Deployed on Azure**:
- Container Apps (serverless, auto-scaling)
- Communication Services (email delivery)
- Cosmos DB (user data)
- Storage Account (temporary files)
- Application Insights (monitoring)

📖 **Detailed architecture**: [Architecture Documentation](docs/ARCHITECTURE.md)

## 🛠️ Tech Stack

- **Language**: Go 1.21+
- **Bot Framework**: [telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api)
- **Web Scraping**: [goquery](https://github.com/PuerkitoBio/goquery)
- **Cloud**: Azure (Container Apps, Communication Services, Cosmos DB)
- **IaC**: Terraform
- **CI/CD**: GitHub Actions

## 💻 Development

### Prerequisites

- Go 1.21+
- Docker (for local testing)
- Azure CLI (for deployment)
- Terraform (for infrastructure)

### Local Development

```bash
# Install dependencies
go mod download

# Run tests
go test ./...

# Run with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Lint code
golangci-lint run

# Build
go build -o bin/bot ./cmd/bot
```

### Project Structure

```
flibusta_kindle_bot/
├── cmd/bot/              # Application entry point
├── internal/             # Private application code
│   ├── bot/              # Telegram bot handlers
│   ├── search/           # Flibusta search engine
│   ├── downloader/       # Book downloader
│   ├── kindle/           # Email sender
│   ├── user/             # User management
│   └── i18n/             # Localization
├── docs/                 # Documentation
├── terraform/            # Infrastructure as Code
└── .github/workflows/    # CI/CD pipelines
```

## 🚢 Deployment

### Option 1: Automatic (Recommended)

Push to `main` branch - GitHub Actions automatically deploys to Azure.

```bash
git push origin main
```

### Option 2: Manual

```bash
# Deploy infrastructure
cd terraform
terraform init
terraform apply

# Build and deploy via GitHub Actions
# Go to Actions → Deploy to Azure → Run workflow
```

📖 **Full deployment guide**: [Deployment Documentation](docs/DEPLOYMENT.md)

## 🤝 Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`go test ./...`)
5. Commit your changes (`git commit -m 'feat: add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

**PR Requirements**:
- ✅ All tests pass
- ✅ Code coverage ≥ 60%
- ✅ Linting passes
- ✅ Security scan passes

📖 **Development workflow**: [CI/CD Documentation](docs/CI_CD.md)

## 🔐 Security

- **Secrets**: Stored in Azure Key Vault
- **Encryption**: Kindle emails encrypted at rest
- **Input Validation**: All user inputs sanitized
- **Rate Limiting**: Prevents abuse
- **GDPR Compliant**: Data deletion on request

## 📊 Bot Commands

| Command | Description |
|---------|-------------|
| `/start` | Initialize bot and setup |
| `/kindle` | Set your Kindle email address |
| `/language` | Change interface language |
| `/whitelist` | Show Amazon whitelist instructions |
| `/settings` | View and update preferences |
| `/help` | Show help and commands |

**No `/search` command needed** - just type the book title or author name!

## ⚠️ Legal Notice

This bot is for **educational purposes** only. Users must ensure they have the right to download and distribute the books they search for. Please comply with:
- Flibusta.is terms of service
- Copyright laws in your jurisdiction
- Telegram Bot API terms of service
- Amazon Kindle terms of service

## 📝 License

MIT License - see [LICENSE](LICENSE) file for details.

## 💬 Support

- **Issues**: [GitHub Issues](https://github.com/yourusername/flibusta_kindle_bot/issues)
- **Documentation**: [docs/](docs/)
- **Telegram**: [@your_support_contact](https://t.me/your_support_contact)

## 🌟 Acknowledgments

- [Flibusta](https://flibusta.is) - Book source
- [Telegram Bot API](https://core.telegram.org/bots/api) - Bot platform
- [Azure](https://azure.microsoft.com) - Cloud infrastructure

---

**Made with ❤️ and Go**

**Happy Reading! 📚**
