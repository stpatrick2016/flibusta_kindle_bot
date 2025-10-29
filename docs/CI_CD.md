# CI/CD Pipeline Documentation

This document describes the continuous integration and deployment pipeline for the Flibusta Kindle Bot using GitHub Actions.

## Table of Contents

- [Overview](#overview)
- [GitHub Actions Workflows](#github-actions-workflows)
- [Required Secrets](#required-secrets)
- [Branch Protection](#branch-protection)
- [Development Workflow](#development-workflow)
- [Monitoring](#monitoring)
- [Troubleshooting](#troubleshooting)

## Overview

The CI/CD pipeline automates:
- ✅ Code quality checks (linting)
- ✅ Unit testing with coverage reporting
- ✅ Security scanning
- ✅ Docker image building
- ✅ Deployment to Azure Container Apps
- ✅ Automatic rollback on failure

### Pipeline Flow

```
┌─────────────┐
│ Git Push/PR │
└──────┬──────┘
       │
       ├──────────────────┐
       ▼                  ▼
┌─────────────┐    ┌──────────────┐
│  PR Checks  │    │ Push to main │
│             │    │              │
│ • Lint      │    │ Triggers:    │
│ • Test      │    │ • PR Checks  │
│ • Security  │    │ • Build      │
│ • Build     │    │ • Deploy     │
└─────────────┘    └──────┬───────┘
                          │
                          ▼
                   ┌──────────────┐
                   │ Deploy to    │
                   │ Azure        │
                   │              │
                   │ If fails →   │
                   │ Rollback     │
                   └──────────────┘
```

## GitHub Actions Workflows

### 1. PR Checks Workflow

**File**: `.github/workflows/pr-checks.yml`

**Triggers**:
- Pull requests to `main` or `develop`
- Push to `main` or `develop`

**Jobs**:

#### Lint Job
- Runs `golangci-lint` with comprehensive rules
- Checks code quality, style, and best practices
- Configuration in `.golangci.yml`

```yaml
- golangci/golangci-lint-action@v4
  with:
    version: latest
    args: --timeout=5m
```

#### Test Job
- Runs unit tests with race detector
- Generates coverage report
- Uploads to Codecov
- **Fails if coverage < 60%**

```yaml
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
```

Coverage threshold check:
```bash
coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
if (( $(echo "$coverage < 60" | bc -l) )); then
  exit 1
fi
```

#### Build Job
- Compiles Go application
- Verifies no build errors
- Checks binary creation

```yaml
go build -v -o bin/bot ./cmd/bot
```

#### Security Job
- Runs Gosec security scanner
- Scans for common security issues
- Uploads results as SARIF file
- Integrates with GitHub Security tab

```yaml
securego/gosec@master
  with:
    args: '-no-fail -fmt sarif -out gosec-results.sarif ./...'
```

#### Docker Build Job
- Tests Docker image build
- Verifies Dockerfile syntax
- Uses build cache for speed
- Does NOT push image (only on deploy)

```yaml
docker/build-push-action@v5
  with:
    push: false
    cache-from: type=gha
    cache-to: type=gha,mode=max
```

#### Summary Job
- Aggregates all check results
- Posts summary to PR
- Shows ✅ or ❌ for each check

### 2. Deployment Workflow

**File**: `.github/workflows/deploy.yml`

**Triggers**:
- Push to `main` branch (automatic)
- Manual workflow dispatch (for staging/production)

**Jobs**:

#### Build and Push Job
- Builds Docker image
- Tags with multiple formats:
  - `main-<sha>` (branch + commit SHA)
  - `main` (latest on branch)
  - `v1.2.3` (if semantic version tag)
- Pushes to Azure Container Registry
- Uses layer caching for faster builds

```yaml
docker/metadata-action@v5
  tags: |
    type=ref,event=branch
    type=sha,prefix={{branch}}-
    type=semver,pattern={{version}}
```

#### Deploy to Azure Job
- Logs into Azure with service principal
- Deploys new container image
- Configures environment variables from secrets
- References Key Vault secrets
- Performs health check
- Generates deployment summary

```yaml
azure/container-apps-deploy-action@v1
  with:
    containerAppName: flibusta-bot
    imageToDeploy: ${{ needs.build-and-push.outputs.image-tag }}
    environmentVariables: |
      TELEGRAM_BOT_TOKEN=secretref:telegram-bot-token
      ACS_CONNECTION_STRING=secretref:acs-connection-string
```

Health check:
```bash
curl -f -s "$APP_URL/health" | grep "200"
```

#### Rollback Job
- **Only runs if deployment fails**
- Gets previous revision
- Activates last known good version
- Minimizes downtime

```yaml
az containerapp revision activate \
  --revision ${{ steps.get-revision.outputs.previous-revision }}
```

#### Notify Job
- Reports deployment status
- Can integrate with Slack/Teams/Discord
- Currently logs to GitHub summary

### Workflow Configuration

#### PR Checks Permissions
```yaml
permissions:
  contents: read
  pull-requests: write
  checks: write
```

#### Deployment Permissions
```yaml
permissions:
  contents: read
  id-token: write  # For Azure OIDC
```

## Required Secrets

Configure these in GitHub: **Settings** → **Secrets and variables** → **Actions**

### Azure Secrets

| Secret | Description | How to Get |
|--------|-------------|------------|
| `AZURE_CREDENTIALS` | Service Principal JSON | `az ad sp create-for-rbac --sdk-auth` |
| `AZURE_CONTAINER_REGISTRY` | Registry URL | `yourregistry.azurecr.io` |
| `AZURE_REGISTRY_USERNAME` | ACR username | From Terraform output or Azure Portal |
| `AZURE_REGISTRY_PASSWORD` | ACR password | From Terraform output or Azure Portal |

### Application Secrets

| Secret | Description | How to Get |
|--------|-------------|------------|
| `TELEGRAM_BOT_TOKEN` | Bot API token | [@BotFather](https://t.me/botfather) |
| `ACS_CONNECTION_STRING` | Azure Communication Services | Azure Portal → ACS → Keys |
| `ACS_SENDER_ADDRESS` | Email sender address | `DoNotReply@yourbot.azurecomm.net` |
| `DATABASE_CONNECTION_STRING` | Database connection | From Terraform output |
| `AZURE_APPINSIGHTS_KEY` | Application Insights | From Terraform output |

### Optional Secrets

| Secret | Description | Required? |
|--------|-------------|-----------|
| `CODECOV_TOKEN` | Codecov integration | No (public repos work without) |
| `LOG_LEVEL` | Logging level | No (defaults to 'info') |
| `SLACK_WEBHOOK_URL` | Slack notifications | No |

### Setting Secrets

```bash
# Via GitHub CLI
gh secret set AZURE_CREDENTIALS < azure-credentials.json
gh secret set TELEGRAM_BOT_TOKEN --body "your_token_here"

# Via GitHub Web UI
# Repository → Settings → Secrets → New repository secret
```

## Branch Protection

Recommended settings for `main` branch:

### Required Settings

Go to: **Settings** → **Branches** → **Add rule**

```yaml
Branch name pattern: main

☑ Require a pull request before merging
  ☑ Require approvals: 1
  ☑ Dismiss stale pull request approvals when new commits are pushed
  ☑ Require review from Code Owners

☑ Require status checks to pass before merging
  ☑ Require branches to be up to date before merging
  Required checks:
    • Lint Code
    • Run Tests
    • Build Application
    • Security Scan
    • Docker Build Test

☑ Require conversation resolution before merging
☑ Require linear history
☑ Do not allow bypassing the above settings
☐ Allow force pushes (keep unchecked)
☐ Allow deletions (keep unchecked)
```

### Optional Settings

```yaml
☑ Require deployments to succeed before merging
  Environments: staging

☑ Lock branch (for production-only branches)

☑ Require signed commits
```

## Development Workflow

### 1. Start New Feature

```bash
# Update main branch
git checkout main
git pull origin main

# Create feature branch
git checkout -b feature/your-feature-name
```

### 2. Development

```bash
# Make changes
# Write tests
# Run locally

# Check code locally (optional but recommended)
golangci-lint run
go test ./... -race -coverprofile=coverage.out
go build -o bin/bot ./cmd/bot
```

### 3. Commit Changes

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```bash
git add .

# Format: <type>: <description>
# Types: feat, fix, docs, style, refactor, test, chore

git commit -m "feat: add book format selection"
git commit -m "fix: handle connection timeout"
git commit -m "docs: update API documentation"
```

### 4. Push and Create PR

```bash
# Push to remote
git push origin feature/your-feature-name

# Create PR via CLI
gh pr create --title "Add book format selection" --body "Implements #123"

# Or via GitHub web interface
```

### 5. PR Checks

Automatically runs:
1. ✅ Lint check
2. ✅ Unit tests (must have 60%+ coverage)
3. ✅ Build verification
4. ✅ Security scan
5. ✅ Docker build test

Monitor progress: GitHub → Pull Requests → Your PR → Checks tab

### 6. Review Process

- Wait for checks to pass (all must be ✅)
- Request review from team members
- Address review comments
- Update PR with changes

```bash
# Make changes
git add .
git commit -m "fix: address review comments"
git push origin feature/your-feature-name
# Checks run again automatically
```

### 7. Merge to Main

After approval and passing checks:

```bash
# Merge via GitHub UI (Squash and merge recommended)
# Or via CLI
gh pr merge --squash --delete-branch
```

### 8. Automatic Deployment

Once merged to `main`:
1. 🚀 Deployment workflow triggers automatically
2. 🐳 Docker image is built and pushed to ACR
3. ☁️ Deploys to Azure Container Apps
4. ✅ Health check verifies deployment
5. 📊 Deployment summary posted to PR

## Monitoring

### GitHub Actions Dashboard

**View workflow runs**:
```
Repository → Actions tab
```

Features:
- See all workflow runs
- Filter by event type, branch, status
- View logs for each job
- Download artifacts (coverage reports)
- Re-run failed workflows

### Deployment Status

**Check deployment in Azure**:
```bash
# Container App status
az containerapp show \
  --name flibusta-bot \
  --resource-group flibusta-bot-rg \
  --query "properties.provisioningState"

# View revisions
az containerapp revision list \
  --name flibusta-bot \
  --resource-group flibusta-bot-rg
```

### Logs and Metrics

**Application logs**:
```bash
# Stream logs
az containerapp logs show \
  --name flibusta-bot \
  --resource-group flibusta-bot-rg \
  --follow

# Recent logs
az containerapp logs show \
  --name flibusta-bot \
  --resource-group flibusta-bot-rg \
  --tail 100
```

**GitHub Actions logs**:
- Available in Actions tab
- Logs retained for 90 days
- Download as ZIP for archival

## Troubleshooting

### PR Checks Failing

#### Lint Errors
```bash
# Run locally to see issues
golangci-lint run

# Auto-fix some issues
golangci-lint run --fix
```

#### Test Failures
```bash
# Run tests locally
go test ./... -v

# Run specific test
go test ./internal/bot -v -run TestHandlerFunction

# Check coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

#### Coverage Below Threshold
- Write more tests
- Test edge cases
- Don't test external dependencies (use mocks)

#### Security Issues
```bash
# Run Gosec locally
gosec ./...

# Check specific issue
gosec -include=G401 ./...
```

### Deployment Failures

#### Image Build Fails
- Check Dockerfile syntax
- Verify all dependencies are available
- Check build logs in GitHub Actions

#### Deployment Fails
- Verify Azure secrets are correct
- Check Container App logs
- Ensure resource group exists
- Verify service principal has permissions

#### Rollback Not Working
- Check if previous revision exists
- Manually activate revision:
```bash
az containerapp revision activate \
  --name flibusta-bot \
  --resource-group flibusta-bot-rg \
  --revision <revision-name>
```

### GitHub Actions Issues

#### Workflow Not Triggering
- Check branch protection rules
- Verify workflow file syntax (YAML)
- Check workflow permissions

#### Secrets Not Available
- Verify secrets are set in repository settings
- Check secret names match workflow file
- Ensure you're not in a forked repo (secrets don't transfer)

#### Rate Limiting
- GitHub Actions has usage limits
- Check: Settings → Billing → GitHub Actions
- Public repos: unlimited for public repos

## Best Practices

### Code Quality

1. ✅ Run linter before committing
2. ✅ Write tests for new features
3. ✅ Keep functions small and focused
4. ✅ Use meaningful variable names
5. ✅ Document complex logic

### Git Workflow

1. ✅ Keep commits atomic and focused
2. ✅ Write descriptive commit messages
3. ✅ Rebase feature branches before merging
4. ✅ Delete merged branches
5. ✅ Use conventional commits format

### Testing

1. ✅ Test happy path and edge cases
2. ✅ Use table-driven tests
3. ✅ Mock external dependencies
4. ✅ Test error handling
5. ✅ Aim for >60% coverage (80%+ for critical code)

### Deployment

1. ✅ Test in staging before production
2. ✅ Monitor logs after deployment
3. ✅ Keep rollback plan ready
4. ✅ Document deployment issues
5. ✅ Use feature flags for risky changes

## Related Documentation

- [Deployment Guide](DEPLOYMENT.md) - Azure infrastructure setup
- [Architecture](ARCHITECTURE.md) - System design and components
- [Kindle Setup](KINDLE_SETUP.md) - Email configuration

---

**Need help?** Open an issue on GitHub or check the troubleshooting section above.
