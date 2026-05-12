# go-whatsapp-mcp

A Model Context Protocol (MCP) server written in Go that connects AI agents to WhatsApp. Agents can read messages from group chats and send messages back, enabling task coordination via WhatsApp.

## Features

- **4 MCP tools**: `list_chats`, `get_messages`, `send_message`, `get_chat_info`
- **stdio transport**: compatible with Claude Desktop, Claude Agent SDK, and any MCP client
- **Persistent session**: authenticates once via QR code, stores session in SQLite
- **Allowlist**: restrict agent access to specific group JIDs
- **No CGo**: uses `modernc.org/sqlite` — runs in any container without gcc

---

## Quick Start (local binary)

### Prerequisites

- Go 1.25+

### Build and run

```bash
git clone https://github.com/gfournier/go-whatsapp-mcp
cd go-whatsapp-mcp
go build -o go-whatsapp-mcp .

# First run: scan the QR code printed to stderr
./go-whatsapp-mcp
```

On first run, a QR code URL is printed to `stderr`. Open the URL in a browser and scan the QR code with WhatsApp mobile:

1. Open WhatsApp on your phone
2. Tap **Settings → Linked Devices → Link a Device**
3. Scan the QR code

The session is saved to `whatsapp.db`. Subsequent runs connect automatically without needing a QR scan.

---

## Running with Podman (recommended for production)

### Install Podman on macOS

```bash
# Install via Homebrew
brew install podman

# Initialize and start the VM (required on macOS)
podman machine init
podman machine start

# Verify
podman version

# Optional: install podman-compose for docker-compose.yml support
brew install podman-compose
```

### First run — QR code authentication

The QR code URL is printed to `stderr`. Run interactively so you can see it:

```bash
podman build -t go-whatsapp-mcp .

podman run -it --rm \
  -v whatsapp-data:/data \
  -e WHATSAPP_ALLOWED_JIDS="120363XXXXXX@g.us" \
  go-whatsapp-mcp
```

Scan the QR code, then `Ctrl+C`. The session is saved in the `whatsapp-data` volume.

### Subsequent runs

```bash
podman run -i --rm \
  -v whatsapp-data:/data \
  -e WHATSAPP_ALLOWED_JIDS="120363XXXXXX@g.us" \
  go-whatsapp-mcp
```

The `-i` flag keeps stdin open for MCP JSON-RPC communication.

### With podman-compose

```bash
# Edit WHATSAPP_ALLOWED_JIDS in docker-compose.yml first
podman-compose up --build
```

---

## Configuration

| Variable | Default | Description |
|---|---|---|
| `WHATSAPP_DB_PATH` | `whatsapp.db` | Path to SQLite session file |
| `WHATSAPP_ALLOWED_JIDS` | *(empty = all)* | Comma-separated JIDs the agent may access |
| `WHATSAPP_MAX_MESSAGES` | `500` | Ring buffer size per chat |
| `LOG_LEVEL` | `info` | `debug` / `info` / `warn` |

You can also create a `.env` file in the working directory.

**Finding a group JID**: Use the `list_chats` tool after connecting to list all joined groups with their JIDs.

---

## MCP Tools Reference

### `list_chats`

List all accessible WhatsApp chats and groups.

```json
// No input required
// Output:
[
  {"jid": "120363000000000001@g.us", "name": "Engineering Team", "is_group": true},
  {"jid": "15551234567@s.whatsapp.net", "name": "15551234567@s.whatsapp.net", "is_group": false}
]
```

### `get_messages`

Fetch recent messages from a chat (newest first).

```json
// Input:
{"jid": "120363000000000001@g.us", "limit": 20}

// Output:
[
  {
    "id": "ABC123",
    "jid": "120363000000000001@g.us",
    "sender": "15551234567@s.whatsapp.net",
    "sender_name": "Alice",
    "body": "Can you review the PR?",
    "timestamp": "2026-05-08T10:30:00Z",
    "is_from_me": false
  }
]
```

### `send_message`

Send a text message to a chat.

```json
// Input:
{"jid": "120363000000000001@g.us", "text": "On it! Will review shortly."}

// Output:
"message sent at 2026-05-08T10:31:00Z"
```

### `get_chat_info`

Get metadata about a group.

```json
// Input:
{"jid": "120363000000000001@g.us"}

// Output:
{
  "jid": "120363000000000001@g.us",
  "name": "Engineering Team",
  "description": "Daily standups and task tracking",
  "participant_count": 5,
  "participants": [
    {"jid": "15551234567@s.whatsapp.net", "is_admin": true, "is_super_admin": false}
  ]
}
```

---

## Claude Desktop Integration

Add to `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS):

```json
{
  "mcpServers": {
    "whatsapp": {
      "command": "podman",
      "args": [
        "run", "-i", "--rm",
        "-v", "whatsapp-data:/data",
        "-e", "WHATSAPP_ALLOWED_JIDS=120363XXXXXX@g.us",
        "go-whatsapp-mcp"
      ]
    }
  }
}
```

Or if running the binary directly:

```json
{
  "mcpServers": {
    "whatsapp": {
      "command": "/path/to/go-whatsapp-mcp",
      "env": {
        "WHATSAPP_ALLOWED_JIDS": "120363XXXXXX@g.us"
      }
    }
  }
}
```

---

## Security

- The agent can only read/write chats in `WHATSAPP_ALLOWED_JIDS`
- No tools for adding/removing participants, deleting messages, or accessing media
- `whatsapp.db` contains your session credentials — keep it private (mode 600)
- The QR code URL is only printed to `stderr`, never to `stdout` (which MCP clients read)
