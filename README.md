# Web Search MCP

An MCP server that gives your coding agent the ability to search the web and fetch page content. Ships as a single binary -- no Go installation, no `node_modules`, no runtime dependencies.

## Tools

### `web_search`

Search the web via Brave Search. Returns a numbered list of sources with titles, URLs, and descriptions.

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

### 1. Get a Brave Search API key

Sign up at [brave.com/search/api](https://brave.com/search/api/) -- the free tier includes 1,000 queries/month.

### 2. Download the binary

Download the latest release for your platform from the [Releases](../../releases) page.

### 3. Make it executable (macOS/Linux)

```bash
chmod +x web-search-mcp
```

### 4. Add to your coding agent

Pick your agent below and add the config. Replace `/path/to/web-search-mcp` with the actual path to the binary and `your_api_key_here` with your Brave API key.

---

### Claude Code (CLI)

Add to your project's `.mcp.json` or your global `~/.claude/settings.json`:

```json
{
  "mcpServers": {
    "web-search": {
      "command": "/path/to/web-search-mcp",
      "env": {
        "BRAVE_SEARCH_API_KEY": "your_api_key_here"
      }
    }
  }
}
```

Or use the CLI:

```bash
claude mcp add web-search /path/to/web-search-mcp -e BRAVE_SEARCH_API_KEY=your_api_key_here
```

### Claude Desktop

Edit `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or `%APPDATA%\Claude\claude_desktop_config.json` (Windows):

```json
{
  "mcpServers": {
    "web-search": {
      "command": "/path/to/web-search-mcp",
      "env": {
        "BRAVE_SEARCH_API_KEY": "your_api_key_here"
      }
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
      "command": "/path/to/web-search-mcp",
      "env": {
        "BRAVE_SEARCH_API_KEY": "your_api_key_here"
      }
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
        "command": "/path/to/web-search-mcp",
        "env": {
          "BRAVE_SEARCH_API_KEY": "your_api_key_here"
        }
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
      "command": "/path/to/web-search-mcp",
      "env": {
        "BRAVE_SEARCH_API_KEY": "your_api_key_here"
      }
    }
  }
}
```

## Verify It Works

Use the [MCP Inspector](https://github.com/modelcontextprotocol/inspector) to test the server:

```bash
npx @modelcontextprotocol/inspector /path/to/web-search-mcp
```

This opens a web UI where you can call both tools and inspect the responses.

## License

MIT
