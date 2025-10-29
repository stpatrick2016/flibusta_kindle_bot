# Azure Deployment Guide

This guide walks you through deploying the Flibusta Kindle Bot to Azure using Terraform.

## Prerequisites

1. **Azure Account**: Active Azure subscription
2. **Azure CLI**: Install from https://docs.microsoft.com/en-us/cli/azure/install-azure-cli
3. **Terraform**: Install from https://www.terraform.io/downloads (v1.5+)
4. **Docker**: For building container images
5. **Telegram Bot Token**: Get from [@BotFather](https://t.me/botfather)

## Step-by-Step Deployment

### 1. Login to Azure

```powershell
# Login to Azure
az login

# Set your subscription (if you have multiple)
az account set --subscription "YOUR_SUBSCRIPTION_ID"

# Verify
az account show
```

### 2. Configure Terraform Variables

```powershell
cd terraform

# Copy the example file
cp terraform.tfvars.example terraform.tfvars

# Edit with your values
notepad terraform.tfvars
```

Update `terraform.tfvars`:
```hcl
telegram_bot_token = "1234567890:ABCdefGHIjklMNOpqrsTUVwxyz"
project_name       = "flibusta-bot"
location           = "eastus"
```

### 3. Initialize Terraform

```powershell
terraform init
```

This downloads the Azure provider and sets up the backend.

### 4. Preview Infrastructure

```powershell
# See what will be created
terraform plan
```

Review the plan to ensure everything looks correct.

### 5. Deploy Infrastructure

```powershell
# Apply the configuration
terraform apply

# Type 'yes' when prompted
```

This will create:
- Resource Group
- Container Registry
- Container Apps Environment
- Azure Communication Services
- Cosmos DB
- Storage Account
- Key Vault
- Application Insights

**Deployment time**: ~5-10 minutes

### 6. Get Deployment Outputs

```powershell
# View important values
terraform output

# Get specific values
terraform output container_registry_login_server
terraform output acs_sender_address
```

**Important**: Save the `acs_sender_address` - users will need to whitelist this email!

### 7. Build and Push Docker Image

```powershell
# Return to project root
cd ..

# Get ACR login credentials
$ACR_NAME = terraform -chdir=terraform output -raw container_registry_login_server | ForEach-Object { $_.Split('.')[0] }
az acr login --name $ACR_NAME

# Build the Docker image
docker build -t flibusta-bot:latest .

# Tag for Azure Container Registry
$ACR_LOGIN_SERVER = terraform -chdir=terraform output -raw container_registry_login_server
docker tag flibusta-bot:latest "$ACR_LOGIN_SERVER/flibusta-bot:latest"

# Push to ACR
docker push "$ACR_LOGIN_SERVER/flibusta-bot:latest"
```

### 8. Update Container App

The container app will automatically pull the new image, or you can force an update:

```powershell
$RESOURCE_GROUP = terraform -chdir=terraform output -raw resource_group_name
$CONTAINER_APP = terraform -chdir=terraform output -raw container_app_name

az containerapp update `
  --name $CONTAINER_APP `
  --resource-group $RESOURCE_GROUP `
  --image "$ACR_LOGIN_SERVER/flibusta-bot:latest"
```

### 9. Verify Deployment

```powershell
# Check container app status
az containerapp show `
  --name $CONTAINER_APP `
  --resource-group $RESOURCE_GROUP `
  --query "properties.runningStatus"

# View logs
az containerapp logs show `
  --name $CONTAINER_APP `
  --resource-group $RESOURCE_GROUP `
  --follow
```

### 10. Configure Kindle Email Whitelist

**Critical Step**: Inform your users to whitelist the sender email!

1. Get sender email:
   ```powershell
   terraform -chdir=terraform output acs_sender_address
   ```
   Output example: `DoNotReply@12345678-1234-1234-1234-123456789abc.azurecomm.net`

2. Users must add this email to their Amazon Kindle approved senders:
   - Go to: https://www.amazon.com/hz/mycd/myx#/home/settings/payment
   - Navigate to: Preferences → Personal Document Settings
   - Add email to "Approved Personal Document E-mail List"

## Testing the Bot

1. **Start a conversation**: Search for your bot on Telegram
2. **Send `/start`**: Initialize the bot
3. **Set Kindle email**: Use `/kindle your_kindle@kindle.com`
4. **Search for a book**: Type the book title or author name
5. **Send to Kindle**: Select a book and click "Send to Kindle"

## Monitoring

### View Logs in Azure Portal

1. Go to: Azure Portal → Resource Groups → flibusta-bot-rg
2. Click: Container App → Logs
3. Run queries:
   ```kusto
   ContainerAppConsoleLogs_CL
   | where TimeGenerated > ago(1h)
   | order by TimeGenerated desc
   ```

### Application Insights

1. Go to: Application Insights → flibusta-bot-insights
2. View: Live Metrics, Failures, Performance
3. Set up alerts for errors or high response times

## Updating the Bot

### Code Changes

```powershell
# Make your changes
git add .
git commit -m "Description of changes"

# Rebuild and redeploy
docker build -t flibusta-bot:latest .
$ACR_LOGIN_SERVER = terraform -chdir=terraform output -raw container_registry_login_server
docker tag flibusta-bot:latest "$ACR_LOGIN_SERVER/flibusta-bot:latest"
docker push "$ACR_LOGIN_SERVER/flibusta-bot:latest"

# Container App will auto-update within a few minutes
```

### Infrastructure Changes

```powershell
cd terraform

# Edit .tf files as needed
# Then apply changes
terraform plan
terraform apply
```

## Cost Management

### Monitor Costs

```powershell
# View current month costs
az consumption usage list `
  --start-date "2025-10-01" `
  --end-date "2025-10-31" `
  --query "[?contains(instanceName, 'flibusta')]"
```

### Set Budget Alert

1. Azure Portal → Cost Management + Billing
2. Budgets → Add Budget
3. Set threshold: $100/month
4. Add email alerts

### Cost Optimization Tips

- **Container Apps**: Use minimum replicas = 0 for dev/test (scale to zero)
- **Cosmos DB**: Use serverless mode (pay per request)
- **Storage**: Enable lifecycle management to delete old files
- **Communication Services**: Free tier covers 500 emails/month

## Troubleshooting

### Container Won't Start

```powershell
# Check container app logs
az containerapp logs show `
  --name $CONTAINER_APP `
  --resource-group $RESOURCE_GROUP `
  --tail 100

# Check secrets are set
az containerapp secret list `
  --name $CONTAINER_APP `
  --resource-group $RESOURCE_GROUP
```

### Books Not Sending

1. **Check ACS connection**:
   ```powershell
   # Verify Communication Service
   az communication list `
     --resource-group $RESOURCE_GROUP
   ```

2. **Verify email domain**:
   - Azure Portal → Communication Services → Email → Domains
   - Ensure status is "Verified"

3. **Check user whitelisted sender**:
   - User must add sender email to Amazon Kindle approved list

### Database Connection Issues

```powershell
# Test Cosmos DB connection
az cosmosdb show `
  --name flibusta-bot-cosmos `
  --resource-group $RESOURCE_GROUP

# Check firewall rules
az cosmosdb show `
  --name flibusta-bot-cosmos `
  --resource-group $RESOURCE_GROUP `
  --query "ipRules"
```

## Cleanup (Delete Everything)

**Warning**: This will delete all resources and data!

```powershell
cd terraform

# Destroy all infrastructure
terraform destroy

# Type 'yes' when prompted
```

Alternatively, delete the entire resource group:

```powershell
az group delete --name flibusta-bot-rg --yes --no-wait
```

## Security Best Practices

1. **Never commit secrets**: Keep `terraform.tfvars` out of git
2. **Use Key Vault**: Store all secrets in Azure Key Vault
3. **Enable RBAC**: Use managed identities instead of passwords
4. **Rotate secrets**: Periodically rotate Telegram bot token
5. **Monitor access**: Enable Azure AD authentication logs
6. **Network security**: Use VNet integration for production
7. **Data encryption**: Enable encryption at rest (default in Azure)

## Next Steps

1. **Set up CI/CD**: GitHub Actions or Azure DevOps pipelines
2. **Add custom domain**: For email (requires domain verification)
3. **Enable backup**: Configure Cosmos DB backups
4. **Implement caching**: Use Azure Cache for Redis for sessions
5. **Add rate limiting**: Prevent abuse with rate limits
6. **Multi-region**: Deploy to multiple regions for HA

## Support

- **Azure Issues**: https://portal.azure.com → Support
- **Terraform Issues**: https://github.com/hashicorp/terraform/issues
- **Project Issues**: [Your GitHub repo issues]

---

## Quick Reference

### Useful Commands

```powershell
# Check deployment status
az deployment group list --resource-group flibusta-bot-rg

# View container logs (live)
az containerapp logs show --name flibusta-bot --resource-group flibusta-bot-rg --follow

# Scale container app
az containerapp update --name flibusta-bot --resource-group flibusta-bot-rg --min-replicas 2

# Get connection strings
terraform output -raw communication_service_connection_string

# List all resources
az resource list --resource-group flibusta-bot-rg --output table

# Check costs
az consumption usage list --start-date 2025-10-01 --query "[?contains(instanceName, 'flibusta')].[instanceName, usageStart, usageEnd, pretaxCost]" --output table
```

### Environment Variables Reference

| Variable | Source | Description |
|----------|--------|-------------|
| `TELEGRAM_BOT_TOKEN` | BotFather | Telegram bot API token |
| `ACS_CONNECTION_STRING` | Azure Communication Services | Email service connection |
| `ACS_SENDER_ADDRESS` | Azure Communication Services Domain | Sender email address |
| `DATABASE_CONNECTION_STRING` | Cosmos DB | Database connection |
| `AZURE_STORAGE_CONNECTION_STRING` | Storage Account | File storage |
| `APPLICATIONINSIGHTS_CONNECTION_STRING` | Application Insights | Monitoring |

### Estimated Monthly Costs (USD)

| Service | Dev/Test | Production |
|---------|----------|------------|
| Container Apps | $10-20 | $30-50 |
| Cosmos DB (Serverless) | $5-10 | $15-30 |
| Storage Account | $1-2 | $2-5 |
| Communication Services | Free (500/mo) | $2.50/10k emails |
| Application Insights | $5 | $10-20 |
| Container Registry | $5 | $5 |
| **Total** | **$26-42** | **$65-135** |

*Note: Actual costs depend on usage. Free tier covers light usage.*
