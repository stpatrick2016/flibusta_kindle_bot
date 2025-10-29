# Azure Deployment Guide

This document covers deploying the Flibusta Kindle Bot to Azure using Terraform and GitHub Actions.

## Table of Contents

- [Infrastructure Overview](#infrastructure-overview)
- [Azure Services](#azure-services)
- [Deployment Options](#deployment-options)
- [Prerequisites](#prerequisites)
- [Initial Setup](#initial-setup)
- [Terraform Deployment](#terraform-deployment)
- [Configuration](#configuration)
- [Monitoring](#monitoring)
- [Cost Estimation](#cost-estimation)
- [Troubleshooting](#troubleshooting)

## Infrastructure Overview

The bot is deployed on Azure using a serverless, auto-scaling architecture:

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

## Azure Services

### Core Services

| Service | Purpose | Tier/SKU |
|---------|---------|----------|
| **Azure Container Apps** | Hosts the Go application | Consumption (serverless) |
| **Azure Communication Services** | Email delivery to Kindle | Free tier (500 emails/month) |
| **Azure Cosmos DB** | User data & sessions | Serverless |
| **Azure Storage Account** | Temporary book storage | Standard LRS |
| **Azure Key Vault** | Secrets management | Standard |
| **Azure Application Insights** | Monitoring & diagnostics | Pay-as-you-go |
| **Azure Container Registry** | Docker images | Basic |

### Why These Services?

- **Container Apps**: Serverless, auto-scales to zero, built-in HTTPS, easy deployment
- **Communication Services**: Native Azure, high deliverability, free tier
- **Cosmos DB**: Global distribution (if needed), serverless billing, automatic scaling
- **Key Vault**: Secure secret storage, integrated with all Azure services
- **App Insights**: Comprehensive monitoring with minimal configuration

## Deployment Options

### Option 1: Terraform (Recommended)

Terraform provides infrastructure as code with excellent Azure support.

**Pros:**
- ✅ Multi-cloud support (if you need it later)
- ✅ Large community and module ecosystem
- ✅ State management
- ✅ Plan/preview before apply
- ✅ Version controlled infrastructure

**Cons:**
- ❌ Requires learning HCL syntax
- ❌ State file management needed

### Option 2: Bicep (Azure-Native Alternative)

Bicep is Microsoft's domain-specific language for Azure.

**Pros:**
- ✅ Native Azure support
- ✅ Simpler syntax than ARM templates
- ✅ Better IntelliSense in VS Code
- ✅ Directly integrated with Azure CLI

**Cons:**
- ❌ Azure-only (vendor lock-in)
- ❌ Smaller community than Terraform

**We recommend Terraform** for this project due to better community support and flexibility.

## Prerequisites

### Required Tools

```bash
# 1. Azure CLI
# Windows (PowerShell):
winget install -e --id Microsoft.AzureCLI

# macOS:
brew install azure-cli

# Linux:
curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash

# 2. Terraform
# Windows (PowerShell):
winget install -e --id Hashicorp.Terraform

# macOS:
brew install terraform

# Linux:
wget -O- https://apt.releases.hashicorp.com/gpg | sudo gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg
echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/hashicorp.list
sudo apt update && sudo apt install terraform

# 3. Git (if not already installed)
git --version
```

### Azure Subscription

1. **Login to Azure**:
   ```bash
   az login
   ```

2. **Set your subscription** (if you have multiple):
   ```bash
   az account list --output table
   az account set --subscription "Your Subscription Name"
   ```

3. **Verify access**:
   ```bash
   az account show
   ```

## Initial Setup

### 1. Create Service Principal for GitHub Actions

```bash
# Create resource group first
az group create --name flibusta-bot-rg --location eastus

# Create service principal with contributor role
az ad sp create-for-rbac \
  --name "github-actions-flibusta-bot" \
  --role contributor \
  --scopes /subscriptions/{subscription-id}/resourceGroups/flibusta-bot-rg \
  --sdk-auth

# Save the JSON output - you'll need it for GitHub secrets
```

The output will look like:
```json
{
  "clientId": "xxx",
  "clientSecret": "xxx",
  "subscriptionId": "xxx",
  "tenantId": "xxx",
  ...
}
```

### 2. Configure GitHub Secrets

Go to your GitHub repository → **Settings** → **Secrets and variables** → **Actions**

Add these secrets:

| Secret Name | Value | How to Get |
|-------------|-------|------------|
| `AZURE_CREDENTIALS` | Full JSON from service principal | Step 1 output |
| `AZURE_CONTAINER_REGISTRY` | `yourregistry.azurecr.io` | After Terraform apply |
| `AZURE_REGISTRY_USERNAME` | Registry username | After Terraform apply |
| `AZURE_REGISTRY_PASSWORD` | Registry password | After Terraform apply |
| `TELEGRAM_BOT_TOKEN` | Your bot token | [@BotFather](https://t.me/botfather) |

Additional secrets (added after infrastructure is created):
- `ACS_CONNECTION_STRING`
- `ACS_SENDER_ADDRESS`
- `DATABASE_CONNECTION_STRING`
- `AZURE_APPINSIGHTS_KEY`

## Terraform Deployment

### 1. Configure Variables

```bash
cd terraform

# Copy example file
cp terraform.tfvars.example terraform.tfvars

# Edit with your values
nano terraform.tfvars
```

Example `terraform.tfvars`:
```hcl
project_name     = "flibusta-bot"
environment      = "production"
location         = "eastus"
telegram_bot_token = "your_token_here"  # Or use Key Vault reference
```

### 2. Initialize Terraform

```bash
terraform init
```

This will:
- Download Azure provider
- Initialize backend (if configured)
- Prepare modules

### 3. Plan Deployment

```bash
terraform plan -out=tfplan
```

Review the output carefully. It will show:
- Resources to be created
- Estimated costs
- Any errors

### 4. Apply Configuration

```bash
terraform apply tfplan
```

This will create:
- Resource Group
- Container Registry
- Container App Environment
- Communication Services
- Storage Account
- Cosmos DB (or PostgreSQL)
- Key Vault
- Application Insights

⏱️ **Takes about 5-10 minutes**

### 5. Get Output Values

```bash
terraform output
```

You'll need these values for GitHub secrets:
- Container registry URL
- Container registry credentials
- Database connection string
- ACS connection string
- App Insights key

## Configuration

### Environment Variables

The bot uses these environment variables (configured in Container App):

```env
# Telegram
TELEGRAM_BOT_TOKEN=<from-keyvault>

# Email Service
ACS_CONNECTION_STRING=<from-keyvault>
ACS_SENDER_ADDRESS=DoNotReply@yourbot.azurecomm.net

# Database
DATABASE_CONNECTION_STRING=<from-keyvault>

# Storage
AZURE_STORAGE_CONNECTION_STRING=<from-keyvault>

# Monitoring
AZURE_APPINSIGHTS_KEY=<from-keyvault>

# Application
LOG_LEVEL=info
TEMP_DIR=/tmp
```

### Key Vault Integration

All secrets are stored in Azure Key Vault and referenced in Container Apps:

```bash
# Add secret to Key Vault
az keyvault secret set \
  --vault-name flibusta-bot-kv \
  --name telegram-bot-token \
  --value "your_token_here"

# Container Apps automatically fetch secrets via managed identity
```

## Monitoring

### Application Insights

View logs and metrics:
```bash
# Open in Azure Portal
az portal open --resource-id "/subscriptions/{sub-id}/resourceGroups/flibusta-bot-rg/providers/Microsoft.Insights/components/flibusta-bot-insights"

# Query logs with CLI
az monitor app-insights query \
  --app flibusta-bot-insights \
  --analytics-query "requests | take 10"
```

### Container App Logs

```bash
# Stream logs
az containerapp logs show \
  --name flibusta-bot \
  --resource-group flibusta-bot-rg \
  --follow

# View recent logs
az containerapp logs show \
  --name flibusta-bot \
  --resource-group flibusta-bot-rg \
  --tail 100
```

### Metrics Dashboard

Access in Azure Portal:
1. Navigate to Container App
2. Click "Metrics"
3. Add charts for:
   - CPU usage
   - Memory usage
   - Request count
   - Response time

## Cost Estimation

Monthly costs (estimated, US East region):

| Service | Tier | Estimated Cost |
|---------|------|----------------|
| Container Apps | Consumption (0.5 vCPU, 1GB RAM) | $20-40 |
| Communication Services | 500 emails/month | Free |
| Cosmos DB | Serverless | $5-25 |
| Storage Account | Standard LRS (100GB) | $2-5 |
| Key Vault | Standard | $0.03 |
| Application Insights | Basic (5GB/month) | $5-10 |
| Container Registry | Basic | $5 |
| **Total** | | **~$37-85/month** |

### Cost Optimization Tips

1. **Use free tiers**: Communication Services, Cosmos DB serverless
2. **Scale to zero**: Container Apps auto-scales to zero when idle
3. **Retention policies**: Set log retention to 30 days
4. **Reserved capacity**: If predictable load, use reserved instances
5. **Monitoring**: Set up cost alerts in Azure

## Troubleshooting

### Common Issues

#### 1. Deployment Fails

```bash
# Check Container App status
az containerapp show \
  --name flibusta-bot \
  --resource-group flibusta-bot-rg \
  --query "properties.provisioningState"

# View events
az containerapp revision list \
  --name flibusta-bot \
  --resource-group flibusta-bot-rg
```

#### 2. Bot Not Responding

```bash
# Check logs for errors
az containerapp logs show \
  --name flibusta-bot \
  --resource-group flibusta-bot-rg \
  --tail 50

# Restart container
az containerapp revision restart \
  --name flibusta-bot \
  --resource-group flibusta-bot-rg
```

#### 3. Email Not Sending

- Verify ACS connection string
- Check ACS service is not suspended
- Verify sender domain is configured
- Check Application Insights for errors

#### 4. Terraform Errors

```bash
# Refresh state
terraform refresh

# Destroy and recreate problematic resource
terraform destroy -target=azurerm_container_app.bot
terraform apply
```

### Get Help

- **Azure Support**: [Azure Portal](https://portal.azure.com) → Support
- **Terraform Issues**: Check [Terraform Azure Provider](https://github.com/hashicorp/terraform-provider-azurerm/issues)
- **Project Issues**: Open issue on GitHub

## Next Steps

After successful deployment:

1. ✅ Configure GitHub secrets (see [CI/CD Guide](CI_CD.md))
2. ✅ Setup Kindle sender email (see [Kindle Setup Guide](KINDLE_SETUP.md))
3. ✅ Test the bot
4. ✅ Monitor logs and metrics
5. ✅ Set up alerts for errors

---

**Need help?** Check out the [Architecture Documentation](ARCHITECTURE.md) or open an issue.
