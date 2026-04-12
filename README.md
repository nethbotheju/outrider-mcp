# Web Search MCP

An MCP server that gives your coding agent the ability to search the web and fetch page content. Ships as a single binary -- no Go installation, no `node_modules`, no runtime dependencies.

## Tools

### `web_search`

Search the web via DuckDuckGo. Returns a numbered list of sources with titles, URLs, and descriptions. No API key required.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `query` | string | Yes | The search query |
| `count` | int | No | Number of results (default 10, max 20) |

### `fetch`

Fetch a web page and return its content as clean text. Strips scripts, styles, navigation, and other noise -- the agent gets readable content only.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `url` | string | Yes | The URL to fetch |
| `maxLength` | int | No | Max content length in characters (default 50000) |

## Setup

### 1. Download the binary

Download the latest release for your platform from the [Releases](../../releases) page.

### 2. Make it executable (macOS/Linux)

```bash
chmod +x web-search-mcp
```

### 3. Add to your coding agent

Pick your agent below and add the config. Replace `/path/to/web-search-mcp` with the actual path to the binary.

---

### Claude Code (CLI)

Add to your project's `.mcp.json` or your global `~/.claude/settings.json`:

```json
{
  "mcpServers": {
    "web-search": {
      "command": "/path/to/web-search-mcp"
    }
  }
}
```

Or use the CLI:

```bash
claude mcp add web-search /path/to/web-search-mcp
```

### Claude Desktop

Edit `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or `%APPDATA%\Claude\claude_desktop_config.json` (Windows):

```json
{
  "mcpServers": {
    "web-search": {
      "command": "/path/to/web-search-mcp"
    }
  }
}
```

### Cursor

Add to your project's `.cursor/mcp.json` or global Cursor MCP settings:

```json
{
  "mcpServers": {
    "web-search": {
      "command": "/path/to/web-search-mcp"
    }
  }
}
```

### VS Code (Copilot / Continue)

Add to your VS Code `settings.json` under `mcp.servers`:

```json
{
  "mcp": {
    "servers": {
      "web-search": {
        "command": "/path/to/web-search-mcp"
      }
    }
  }
}
```

### Windsurf

Edit `~/.codeium/windsurf/mcp_config.json`:

```json
{
  "mcpServers": {
    "web-search": {
      "command": "/path/to/web-search-mcp"
    }
  }
}
```

### OpenCode

Add to your project's `opencode.json`:

```json
{
  "mcp": {
    "web-search": {
      "type": "local",
      "command": [
        "/path/to/web-search-mcp"
      ]
    }
  }
}
```

## Development

### Prerequisites

- Go 1.25+

### Project Structure

```
web-search-mcp/
├── main.go              # Server entry point — wires tools to the MCP server
├── fetcher/             # HTTP fetching + HTML-to-text extraction
│   └── fetcher.go
├── search/              # Search provider interface and implementations
│   ├── provider.go      # Provider interface, shared types, helpers
│   └── duckduckgo.go    # DuckDuckGo HTML search (free, no API key)
├── tools/               # MCP tool definitions and handlers
│   ├── web_search.go    # web_search tool
│   └── fetch.go         # fetch tool
└── test/                # Integration tests (live network calls)
    ├── search/
    │   └── duckduckgo_test.go
    └── fetch/
        └── fetcher_test.go
```

### Build

```bash
go build -o web-search-mcp .
```

### Run Tests

```bash
go test ./test/search/ ./test/fetch/ -v -count=1 -timeout 60s
```

## Verify It Works

Use the [MCP Inspector](https://github.com/modelcontextprotocol/inspector) to test the server:

```bash
npx @modelcontextprotocol/inspector /path/to/web-search-mcp
```

This opens a web UI where you can call both tools and inspect the responses.

## License

MIT
