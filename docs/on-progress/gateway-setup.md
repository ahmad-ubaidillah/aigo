# Gateway Setup Guide

This guide explains how to configure each messaging platform (Telegram, Discord, Slack, WhatsApp) to work with Aigo.

## Telegram Setup

### 1. Get Bot Token
1. Open Telegram and search for "BotFather"
2. Create a new bot with [@BotFather](https://t.me/botfather)
3. Copy the bot token provided

 BotFather

### 2. Configure Aigo
Add the bot token to your configuration:

**Using YAML config:**
```yaml
gateway:
  telegram:
    enabled: true
    token: "YOUR_BOT_TOKEN_HERE"
```

**or using environment variable:**
```bash
export TELEGRAM_BOT_TOKEN="your_bot_token_here"
```

### 3. Test Connection
Run the connection test:
```bash
aigo gateway test telegram
```

## Discord Setup
### 1. Get Bot Token
1. Go to [Discord Developer Portal](https://discord.com/developers/applications)
2. Create a new application
3. Copy the bot token

### 2. Configure Aigo
Add the bot token to your configuration
```yaml
gateway:
  discord:
    enabled: true
    token: "YOUR_BOT_TOKEN_HERE"
```

**or using environment variable:**
```bash
export DISCORD_BOT_TOKEN="your_bot_token_here"
```

### 3. Test connection
```bash
aigo gateway test discord
```

## Slack Setup
### 1. Get Bot Token
1. Create a [Slack App](https://api.slack.com/apps)
2. Install to app in your workspace
2. Go to "OAuth & Permissions"
3. Add the following bot scopes:
   - `chat:write`
   - `files:read`
   - `incoming-webhook`
4. Copy the Bot User OAuth Token and the App settings

### 2. Configure Aigo
Add the bot token to your configuration
```yaml
gateway:
  slack:
    enabled: true
    token: "xoxb-YOUR-TOKEN-HERE"  # Bot User OAuth Token
    signing_secret: "YOUR_SIGNING_SECRET_HERE"  # Optional, for local development
```

**or using environment variable:**
```bash
export SLACK_BOT_TOKEN="xoxb-your-token-here"
export SLACK_SIGNING_SECRET="your-signing-secret-here"
```

### 3. Test connection
```bash
aigo gateway test slack
```

## WhatsApp Setup
### 1. Get Credentials
1. Install [whatsapp-web.js](https://github.com/pedroslope/whatsapp-web.js) (for Node.js)
   ```bash
   npm install whatsapp-web.js
   ```

2. Generate QR code
```bash
aigo gateway setup whatsapp
```
Follow the prompts to scan QR code

3. Save session credentials

### 2. Configure Aigo
Add credentials to your configuration
```yaml
gateway:
  whatsapp:
    enabled: true
    session_path: "/path/to/session.json"
```

### 3. Test connection
```bash
aigo gateway test whatsapp
```

## Common Configuration
All gateways support these common settings:
```yaml
gateway:
  enabled: true  # Master switch for all gateways
  telegram:
    enabled: true
    token: "..."
  discord:
    enabled: true
    token: "..."
  slack:
    enabled: true
    token: "..."
  whatsapp:
    enabled: true
    session_path: "..."
```

## Environment Variables
| Variable | Description |
|----------|-------------|
| `TELEGRAM_BOT_TOKEN` | Telegram bot token |
| `DISCORD_BOT_TOKEN` | Discord bot token |
| `SLACK_BOT_TOKEN` | Slack bot token |
| `SLACK_SIGNING_SECRET` | Slack signing secret |

| `OPENAI_API_KEY` | OpenAI API key (optional, for LLM fallback) |

## Troubleshooting
### Connection Failed
- Verify token is correct
- Check network connectivity
- Review gateway logs: `aigo gateway logs`

### Messages Not Received
- Verify bot is running
- Check webhook URLs
- Test with `aigo gateway test telegram` directly

### Rate Limits
- Check API rate limits
- Review logs for rate limit warnings

### Bot Not Responding
- Verify bot has proper permissions
- Check if bot is in correct channels
- Test mentioning the bot directly

- Test with a small message first

- Invite bot to a test channel

- Check if bot was kicked

- Review audit logs

