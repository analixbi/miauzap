# Chatwoot Integration Guide

## Overview

Miauzap integrates seamlessly with Chatwoot, allowing you to manage WhatsApp conversations through Chatwoot's powerful CRM interface.

## Features

- ✅ **Automatic Contact Sync** - WhatsApp contacts are automatically created in Chatwoot
- ✅ **Bidirectional Messaging** - Send and receive messages between WhatsApp and Chatwoot
- ✅ **Conversation Management** - Conversations are automatically created and managed
- ✅ **Media Support** - Images, videos, audio, and documents are forwarded
- ✅ **Group Support** - Handle group conversations with participant tracking
- ✅ **Profile Picture Sync** - Contact avatars are synchronized
- ✅ **Message Signing** - Optional signature for outgoing messages
- ✅ **Brazilian Phone Numbers** - Smart handling of Brazilian number variations

## Setup

### Prerequisites

1. A running Chatwoot instance
2. Chatwoot API access token
3. Chatwoot account ID
4. Miauzap instance connected to WhatsApp

### Configuration

#### 1. Get Chatwoot Credentials

From your Chatwoot dashboard:
- **Account ID**: Found in Settings → Account Settings
- **API Token**: Profile Settings → Access Token

#### 2. Configure Miauzap

```bash
POST /chatwoot/setup
Headers:
  token: YOUR_MIAUZAP_TOKEN
Body:
{
  "enabled": true,
  "url": "https://your-chatwoot.com",
  "accountId": "1",
  "token": "YOUR_CHATWOOT_API_TOKEN",
  "nameInbox": "WhatsApp",
  "signMsg": true,
  "signDelimiter": "\n\n---\nSent via WhatsApp",
  "reopenConversation": true,
  "conversationPending": false,
  "mergeBrazilContacts": true,
  "autoCreate": true
}
```

**Response:**
```json
{
  "enabled": true,
  "accountId": "1",
  "url": "https://your-chatwoot.com",
  "inboxId": 123,
  "inboxName": "WhatsApp",
  "signMsg": true,
  "signDelimiter": "\n\n---\nSent via WhatsApp",
  "reopenConversation": true,
  "conversationPending": false,
  "mergeBrazilContacts": true,
  "webhookUrl": "http://your-miauzap.com/chatwoot/webhook/USER_ID"
}
```

#### 3. Configure Chatwoot Webhook

The webhook URL is automatically configured when using `autoCreate: true`. If you need to manually configure it:

1. Go to Chatwoot → Settings → Inboxes → Your Inbox
2. Set webhook URL to: `http://your-miauzap.com/chatwoot/webhook/USER_ID`

## Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | boolean | false | Enable/disable Chatwoot integration |
| `url` | string | - | Chatwoot instance URL |
| `accountId` | string | - | Chatwoot account ID |
| `token` | string | - | Chatwoot API access token |
| `nameInbox` | string | instance name | Name for the Chatwoot inbox |
| `signMsg` | boolean | false | Add signature to outgoing messages |
| `signDelimiter` | string | "\n\n---\n" | Signature delimiter text |
| `reopenConversation` | boolean | false | Automatically reopen resolved conversations on new messages |
| `conversationPending` | boolean | false | Create conversations in pending status |
| `mergeBrazilContacts` | boolean | false | Merge Brazilian phone number variations (+5511 vs +55119) |
| `autoCreate` | boolean | false | Automatically create inbox in Chatwoot |

## API Endpoints

### Setup Chatwoot Integration

```bash
POST /chatwoot/setup
```

Configure or update Chatwoot integration for the authenticated user.

### Get Chatwoot Configuration

```bash
GET /chatwoot/config
```

Retrieve current Chatwoot configuration.

### Webhook Endpoint

```bash
POST /chatwoot/webhook/{instance}
```

Receives webhooks from Chatwoot (configured automatically).

## Message Flow

### WhatsApp → Chatwoot

1. User receives WhatsApp message
2. Miauzap creates/finds contact in Chatwoot
3. Miauzap creates/finds conversation in Chatwoot
4. Message is forwarded to Chatwoot
5. Agent sees message in Chatwoot dashboard

### Chatwoot → WhatsApp

1. Agent replies in Chatwoot
2. Chatwoot sends webhook to Miauzap
3. Miauzap sends message via WhatsApp
4. Optional signature is added if enabled

## Media Handling

Supported media types:
- **Images** (JPEG, PNG, GIF)
- **Videos** (MP4, etc.)
- **Audio** (MP3, OGG, etc.)
- **Documents** (PDF, DOCX, etc.)

Media files are:
1. Downloaded from WhatsApp
2. Uploaded to Chatwoot
3. Displayed in conversation

When sending from Chatwoot:
1. Media URL is extracted from message
2. File is downloaded
3. Uploaded to WhatsApp
4. Sent to recipient

## Group Conversations

Groups are handled specially:
- Group name includes "(GROUP)" suffix
- Participant contacts are created separately
- Group metadata (name, picture) is synced
- Messages show participant information

## Brazilian Phone Numbers

When `mergeBrazilContacts` is enabled:
- Handles both formats: +5511999999999 and +551199999999
- Automatically merges duplicate contacts
- Searches both variations when looking up contacts

## Troubleshooting

### Messages not appearing in Chatwoot

1. Check Chatwoot configuration:
   ```bash
   GET /chatwoot/config
   ```

2. Verify webhook URL is correct
3. Check Miauzap logs for errors
4. Ensure Chatwoot API token is valid

### Messages not sending to WhatsApp

1. Verify WhatsApp connection:
   ```bash
   GET /session/status
   ```

2. Check webhook is configured in Chatwoot
3. Verify user is logged in to WhatsApp
4. Check Miauzap logs for webhook errors

### Duplicate contacts

1. Enable `mergeBrazilContacts` for Brazilian numbers
2. Check contact identifier mapping in database
3. Manually merge contacts in Chatwoot if needed

### Media not forwarding

1. Verify media URL is accessible
2. Check file size limits
3. Ensure supported media type
4. Check Miauzap logs for upload errors

## Database Schema

### chatwoot_config
Stores Chatwoot configuration per user.

### chatwoot_contacts
Maps WhatsApp phone numbers to Chatwoot contact IDs.

### chatwoot_conversations
Maps WhatsApp JIDs to Chatwoot conversation IDs.

### chatwoot_messages
Tracks message synchronization status.

## Performance Considerations

- Contacts and conversations are cached for 5 minutes
- Database lookups are indexed for fast retrieval
- Webhook processing is asynchronous
- Media downloads are streamed (not buffered entirely)

## Security

- API tokens are stored encrypted in database
- Webhook endpoint requires valid instance ID
- Private messages in Chatwoot are not forwarded
- Contact messages (not agent) are ignored

## Example Workflows

### Basic Setup

```bash
# 1. Connect WhatsApp
POST /session/connect
{
  "Subscribe": ["message"]
}

# 2. Setup Chatwoot
POST /chatwoot/setup
{
  "enabled": true,
  "url": "https://chatwoot.example.com",
  "accountId": "1",
  "token": "YOUR_TOKEN",
  "autoCreate": true
}

# 3. Start receiving messages in Chatwoot!
```

### Update Configuration

```bash
# Enable message signing
POST /chatwoot/setup
{
  "enabled": true,
  "signMsg": true,
  "signDelimiter": "\n\n---\nPowered by Miauzap"
}
```

### Disable Integration

```bash
POST /chatwoot/setup
{
  "enabled": false
}
```

## Support

For issues or questions:
1. Check Miauzap logs
2. Check Chatwoot logs
3. Verify network connectivity
4. Review this documentation

## Advanced Configuration

### Custom Webhook URL

If you need to use a custom webhook URL (e.g., behind a proxy):

1. Set `autoCreate: false` in setup
2. Manually create inbox in Chatwoot
3. Configure webhook URL to point to your Miauzap instance
4. Update Miauzap configuration with inbox ID

### Multiple Instances

You can run multiple Miauzap instances, each with their own Chatwoot configuration:
- Each instance has separate database tables
- Each instance can connect to different Chatwoot accounts
- Webhook URLs include instance identifier

## Limitations

- Maximum message length: Chatwoot's limit (typically 10,000 characters)
- Media file size: Limited by WhatsApp (16MB for most media)
- Webhook timeout: 30 seconds
- Contact search: Limited to phone number and identifier
