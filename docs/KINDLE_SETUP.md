# Kindle Setup Guide

This document explains how to configure Amazon Kindle to receive books from the bot.

## Table of Contents

- [Important Notice](#important-notice)
- [User Setup Instructions](#user-setup-instructions)
- [Bot Implementation](#bot-implementation)
- [Kindle Limitations](#kindle-limitations)
- [Email Service Configuration](#email-service-configuration)
- [Troubleshooting](#troubleshooting)

## Important Notice

⚠️ **CRITICAL**: Users MUST whitelist your bot's sender email address in their Amazon account before they can receive books!

### Key Points

- **No Programmatic Verification**: There is NO API to check if a user has whitelisted your sender email
- **Cannot Guarantee Delivery**: The bot can only confirm the email was sent, not that it was delivered
- **User Responsibility**: Users must complete the Amazon setup themselves
- **Silent Failures**: Most delivery failures won't generate error messages

## User Setup Instructions

### Step 1: Find Kindle Email Address

Every Kindle device has a unique email address for receiving documents.

**How to find it**:

1. Go to [Amazon Content and Devices](https://www.amazon.com/hz/mycd/myx#/home/settings/payment)
2. Click **"Preferences"** tab
3. Scroll to **"Personal Document Settings"**
4. Find **"Send-to-Kindle E-Mail Settings"**
5. Your Kindle email looks like: `username@kindle.com` or `username_123@kindle.com`

**Alternative methods**:
- On Kindle device: Settings → Your Account → Send-to-Kindle Email
- In Kindle app: Settings → Send-to-Kindle Email

### Step 2: Whitelist Bot's Sender Email

This is the **MOST IMPORTANT** step!

**Instructions for users**:

1. Go to [Amazon Content and Devices](https://www.amazon.com/hz/mycd/myx#/home/settings/payment)
2. Click **"Preferences"** tab
3. Scroll to **"Personal Document Settings"**
4. Find **"Approved Personal Document E-mail List"**
5. Click **"Add a new approved e-mail address"**
6. Enter: `DoNotReply@yourbot.azurecomm.net` (replace with your actual sender email)
7. Click **"Add Address"**
8. ✅ Done!

### Step 3: Configure Bot

In the Telegram bot:

```
/start
# Follow the greeting instructions

/kindle
# Enter your Kindle email (e.g., username@kindle.com)
```

### Visual Guide

```
Amazon Website
├── Account & Lists
│   └── Content and Devices
│       └── Preferences Tab
│           └── Personal Document Settings
│               ├── Send-to-Kindle E-Mail Settings
│               │   └── [Find your Kindle email here]
│               │
│               └── Approved Personal Document E-mail List
│                   └── [Add bot's sender email here]
```

## Bot Implementation

### Onboarding Flow

When user sends `/start`, the bot should display:

```
👋 Welcome to Flibusta Kindle Bot!

📧 IMPORTANT: Setup Required

Before you can receive books, you must whitelist our sender email in your Amazon account:

1️⃣ Go to: https://www.amazon.com/hz/mycd/myx#/home/settings/payment
2️⃣ Click "Preferences" → "Personal Document Settings"
3️⃣ Under "Approved Personal Document E-mail List", add:
   DoNotReply@yourbot.azurecomm.net
4️⃣ Click "Add Address"

✅ Then come back and tell me your Kindle email!

Use /kindle to set your Kindle email address.
Use /whitelist anytime to see these instructions again.
```

### Bot Commands

#### `/start`
- Show greeting in user's language
- Display whitelist instructions
- Ask for Kindle email if not set

#### `/kindle [email]`
- Set or update Kindle email
- Validate email format (`*@kindle.com`)
- Remind about whitelist if first time

#### `/whitelist`
- Show Amazon whitelist instructions
- Provide direct link to Amazon settings
- Display current sender email address

#### `/help`
- Show all commands
- Include setup guide link
- Remind about whitelist requirement

### Validation Logic

```go
// Validate Kindle email format
func ValidateKindleEmail(email string) error {
    // Check basic format
    if !strings.HasSuffix(email, "@kindle.com") {
        return errors.New("Email must end with @kindle.com")
    }
    
    // Check for @
    if !strings.Contains(email, "@") {
        return errors.New("Invalid email format")
    }
    
    // Check username not empty
    parts := strings.Split(email, "@")
    if len(parts[0]) == 0 {
        return errors.New("Email username cannot be empty")
    }
    
    return nil
}
```

### Error Handling

```go
// Handle delivery attempt
func SendToKindle(email, bookFile string) error {
    // Send email
    err := sendEmail(email, bookFile)
    if err != nil {
        // SMTP error - email couldn't be sent
        return fmt.Errorf("failed to send email: %w", err)
    }
    
    // Email sent successfully, but delivery not guaranteed
    // Amazon may silently reject if sender not whitelisted
    return nil
}
```

**Important**: The bot can only detect SMTP errors, not Amazon rejection!

### User Messages

#### Success Message
```
✅ Book sent to your Kindle!

The book has been sent to: username@kindle.com

📱 It should appear on your Kindle in a few minutes.

❓ Book didn't arrive?
• Check your Kindle is connected to Wi-Fi
• Verify you whitelisted our sender email: /whitelist
• Check Amazon's Personal Documents folder
• Wait a few minutes (delivery can take 2-5 min)
```

#### If Book Doesn't Arrive
```
❌ Book hasn't arrived?

Common issues:
1️⃣ Sender email not whitelisted
   → Use /whitelist to see setup instructions

2️⃣ Wrong Kindle email
   → Use /kindle to update your email

3️⃣ File too large (>50 MB)
   → Try a different format

4️⃣ Wi-Fi not connected
   → Connect your Kindle to Wi-Fi

5️⃣ Format not supported
   → We support MOBI, EPUB, PDF

Need help? Use /help
```

## Kindle Limitations

### File Size
- **Maximum**: 50 MB per email
- **Recommendation**: Keep books under 40 MB for reliability
- **Large books**: Split into volumes or compress

### Supported Formats

| Format | Native Support | Notes |
|--------|---------------|-------|
| **MOBI** | ✅ Yes | Best format, no conversion needed |
| **AZW3** | ✅ Yes | Amazon's format |
| **EPUB** | ⚠️ Converted | Converted to AZW3 automatically |
| **PDF** | ✅ Yes | Preserves formatting, not reflowable |
| **TXT** | ✅ Yes | Plain text only |
| **DOC/DOCX** | ⚠️ Converted | Converted to Kindle format |
| **HTML** | ⚠️ Converted | Converted to Kindle format |
| **FB2** | ❌ No | Not supported, convert to EPUB first |

### Auto-Conversion

Add "Convert" in email subject to force conversion:
```
Subject: Convert
```

Amazon will attempt to convert the document to Kindle format.

### Delivery Time

- **Typical**: Instant to 2 minutes
- **Slow**: Up to 5 minutes during peak times
- **Failed**: No delivery after 10 minutes usually means rejection

### Kindle Email Variants

Different Kindle email formats:
- `username@kindle.com` - Default
- `username_123@kindle.com` - If username taken
- `username@free.kindle.com` - Free Kindle Reading Apps (no WhisperSync)

**Recommendation**: Use `@kindle.com` for full features.

## Email Service Configuration

### Azure Communication Services

**Sender Email Format**:
- Provided by Azure: `DoNotReply@<random>.azurecomm.net`
- Or custom domain: `DoNotReply@yourdomain.com`

**Configuration**:
```go
import "github.com/Azure/azure-sdk-for-go/sdk/communication/azemail"

client, err := azemail.NewEmailClientFromConnectionString(connectionString)

message := azemail.EmailMessage{
    SenderAddress: "DoNotReply@yourbot.azurecomm.net",
    Recipients: azemail.EmailRecipients{
        To: []azemail.EmailAddress{
            {Address: userKindleEmail},
        },
    },
    Subject: "Your Book",
    Content: azemail.EmailContent{
        PlainText: "Your requested book is attached.",
    },
    Attachments: []azemail.EmailAttachment{
        {
            Name: "book.mobi",
            ContentType: "application/x-mobipocket-ebook",
            ContentInBase64: base64EncodedBook,
        },
    },
}
```

### SMTP Alternative

If using traditional SMTP:

```go
import "net/smtp"

// Configure SMTP
auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost)

// Build message
msg := []byte("To: " + kindleEmail + "\r\n" +
    "Subject: Your Book\r\n" +
    "MIME-Version: 1.0\r\n" +
    "Content-Type: multipart/mixed; boundary=boundary\r\n" +
    "\r\n" +
    "--boundary\r\n" +
    "Content-Type: text/plain\r\n" +
    "\r\n" +
    "Your requested book is attached.\r\n" +
    "--boundary\r\n" +
    "Content-Type: application/x-mobipocket-ebook\r\n" +
    "Content-Disposition: attachment; filename=\"book.mobi\"\r\n" +
    "Content-Transfer-Encoding: base64\r\n" +
    "\r\n" +
    base64Book + "\r\n" +
    "--boundary--")

// Send
err := smtp.SendMail(smtpHost+":"+smtpPort, auth, senderEmail, []string{kindleEmail}, msg)
```

## Troubleshooting

### User Issues

#### Book Not Arriving

**Checklist for users**:
1. ✅ Whitelisted sender email in Amazon?
2. ✅ Correct Kindle email address?
3. ✅ Kindle connected to Wi-Fi?
4. ✅ File size under 50 MB?
5. ✅ Waited at least 5 minutes?

**Bot response**:
```
Let's troubleshoot! 🔍

1. Have you whitelisted our sender email?
   Use /whitelist to see instructions

2. Is your Kindle email correct?
   Current: username@kindle.com
   Use /kindle to update

3. Is your Kindle online?
   Connect to Wi-Fi and sync

4. Check "Personal Documents" in Amazon
   Go to: www.amazon.com/mycd
```

#### Wrong Email Format

**Common mistakes**:
- `username@gmail.com` (not a Kindle email)
- `username@amazon.com` (Amazon account, not Kindle)
- `username@free-kindle.com` (typo, should be `@free.kindle.com`)

**Validation**:
```
❌ Invalid Kindle email!

Your Kindle email must end with @kindle.com

Example: username@kindle.com or username_123@kindle.com

Find yours at: https://www.amazon.com/mycd
Click "Preferences" → "Send-to-Kindle E-Mail Settings"
```

### Bot Issues

#### Emails Not Sending

**Check**:
1. ACS connection string valid?
2. Sender email configured in ACS?
3. ACS service not suspended?
4. Attachment size under 50 MB?
5. SMTP credentials correct (if using SMTP)?

**Debug**:
```bash
# Check ACS logs
az communication email show \
  --name flibusta-bot-acs \
  --resource-group flibusta-bot-rg

# View Application Insights
az monitor app-insights query \
  --app flibusta-bot-insights \
  --analytics-query "exceptions | where timestamp > ago(1h)"
```

#### High Bounce Rate

**Possible causes**:
- Users not whitelisting sender
- Users providing wrong Kindle emails
- Amazon rate limiting

**Mitigation**:
- Emphasize whitelist instructions
- Validate email format
- Add rate limiting (e.g., 10 books/hour per user)

### Amazon Issues

#### Sender Blocked

**Symptoms**:
- All emails to Kindle silently fail
- Works with other email providers

**Solution**:
- Contact Amazon Kindle support
- Verify sender domain reputation
- Use reputable email service (Azure ACS recommended)

#### File Rejected

**Symptoms**:
- Email delivered but book doesn't appear

**Causes**:
- File format not supported
- File corrupted
- File too large
- Copyright protection detected

**Solution**:
- Convert to MOBI format
- Verify file integrity
- Compress file
- Use different source

## Testing

### Test Checklist

Before deploying:

1. ✅ Send test email to your own Kindle
2. ✅ Test with different file formats (MOBI, EPUB, PDF)
3. ✅ Test with different file sizes (1MB, 10MB, 40MB)
4. ✅ Test without whitelisting (should fail)
5. ✅ Test with wrong Kindle email format
6. ✅ Test whitelist command display
7. ✅ Test help messages
8. ✅ Test error handling

### Manual Testing

```bash
# Send test email via Azure CLI
az communication email send \
  --sender "DoNotReply@yourbot.azurecomm.net" \
  --to "yourkindle@kindle.com" \
  --subject "Test Book" \
  --text "Test message" \
  --attachment-name "test.txt" \
  --attachment-type "text/plain" \
  --attachment-path "./test.txt"
```

## Related Documentation

- [Architecture](ARCHITECTURE.md) - System design
- [Deployment](DEPLOYMENT.md) - Azure setup
- [CI/CD](CI_CD.md) - Development workflow

---

**Questions?** Open an issue on GitHub or check the troubleshooting section.
