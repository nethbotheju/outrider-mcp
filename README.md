# Outrider

MCP server that rides ahead of your AI agent — searching the web, fetching pages, and extracting clean content. Ships as a single binary -- no Go installation, no `node_modules`, no runtime dependencies.

## Tools

### `web_search`

Search the web. Returns a numbered list of sources with titles, URLs, descriptions, and (with SearXNG) optional metadata such as published date, source, journal, DOI, PDF link, and authors. No API key required.

**Providers:**
- **DuckDuckGo** (default) — works out of the box, no configuration needed.
- **SearXNG** (optional) — self-hosted metasearch engine. Supports category filtering and richer metadata.
- **Degoog** (optional) — self-hosted search aggregator. Returns aggregated web results from multiple engines. No API key required.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `query` | string | Yes | The search query |
| `count` | int | No | Number of results. DuckDuckGo: default 10, max 20. SearXNG: default 5, max 10. Degoog: default 10, max 20. |
| `category` | string | No | **SearXNG only.** `general` (default), `science`, `news`, or `it`. |

### `web_fetch`

Fetch a web page and return its content as clean Markdown. Tries the Jina Reader API first, then a headless browser (if enabled), then local readability extraction.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `url` | string | Yes | The URL to fetch |
| `format` | string | No | `lean` (default) strips link URLs and images for minimal tokens. `markdown` preserves full links and images. |
| `maxLength` | int | No | Max content length in characters. Default 10000, max 50000. |

### `web_answer`

Answer a question from the web **without flooding the main context window**. Pass a `question` and optionally a `url`. If a URL is given, the page is read and answered directly. If no URL is given, a bounded side-agent loop searches and fetches only the most relevant pages, then returns a concise answer with sources. Requires a configured side LLM.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `question` | string | Yes | The question to answer |
| `url` | string | No | Optional specific URL to read and answer from. Omit to search the web. |

When `web_answer` is enabled, prefer it over manually calling `web_search` + `web_fetch` for factual questions.

## Setup

### 1. Download the binary

Download the latest release for your platform from the [Releases](../../releases) page.

### 2. Configure Outrider

Outrider can be configured with a JSON config file, environment variables, or both. **Environment variables override config file values.**

#### Config file

Outrider looks for a config file in this order:

1. Path from the `OUTRIDER_CONFIG` environment variable
2. `~/.config/outrider/config.json`
3. `./outrider.json`

If the file is missing, sensible defaults are used.

Example `config.json` (all fields shown; most are optional):

```json
{
  "provider": "searxng",
  "searxng": {
    "baseUrl": "http://localhost:8080",
    "apiKey": ""
  },
  "degoog": {
    "baseUrl": "http://localhost:14444"
  },
  "fetch": {
    "jina": {
      "enabled": true,
      "apiKey": ""
    },
    "maxLength": 10000,
    "timeoutMs": 30000,
    "browser": {
      "enabled": true
    }
  },
  "answer": {
    "enabled": true,
    "baseUrl": "http://localhost:11434/v1",
    "apiKey": "ollama",
    "model": "llama3.1:8b",
    "maxTokens": 800,
    "temperature": 0,
    "maxTurns": 4,
    "maxSearches": 2,
    "maxFetches": 2
  },
  "tools": {
    "web_search": { "enabled": true },
    "web_fetch": { "enabled": true },
    "web_answer": { "enabled": true }
  }
}
```

#### Defaults

If a field is omitted, these defaults are used:

| Field | Default | Notes |
|-------|---------|-------|
| `provider` | `duckduckgo` | Switches to `searxng` automatically if `searxng.baseUrl` is set. |
| `searxng` | unset | Only used when `provider` is `searxng`. |
| `fetch.jina.enabled` | `true` | Jina Reader is tried first. |
| `fetch.maxLength` | `10000` | Default fetch length. Clamped to max `50000`. |
| `fetch.timeoutMs` | `30000` | HTTP timeout for Jina and local fetch. |
| `fetch.browser.enabled` | `true` | Headless Chrome fallback. |
| `answer.enabled` | `false` | Must be `true` for `web_answer` to work. |
| `answer.maxTokens` | `800` | Applied to final-answer LLM calls. |
| `answer.temperature` | `0` | Applied to all answer LLM calls. |
| `answer.maxTurns` | `4` | Max agent-loop turns when no URL is given. |
| `answer.maxSearches` | `2` | Max `web_search` calls per answer. |
| `answer.maxFetches` | `2` | Max `web_fetch` calls per answer. |
| `tools.web_search.enabled` | `true` | |
| `tools.web_fetch.enabled` | `true` | |
| `tools.web_answer.enabled` | `false` | |

#### `answer.enabled` vs `tools.web_answer.enabled`

These two toggles do different things:

- `answer.enabled` — turns the **side LLM** on or off. When `true`, the model settings (`baseUrl`, `apiKey`, `model`, ...) must also be valid.
- `tools.web_answer.enabled` — controls whether the **`web_answer` MCP tool** is advertised to the agent.

`web_answer` is only registered when **both** are `true`. This lets you configure the side LLM once and decide per-project whether the agent may use it.

#### Environment variables

| Variable | Maps to | Description |
|----------|---------|-------------|
| `OUTRIDER_CONFIG` | — | Path to the config file |
| `SEARCH_PROVIDER` | `provider` | `duckduckgo` or `searxng` or `degoog` |
| `SEARXNG_URL` | `searxng.baseUrl` | SearXNG base URL |
| `SEARXNG_API_KEY` | `searxng.apiKey` | SearXNG Bearer token |
| `DEGOOG_URL` | `degoog.baseUrl` | Degoog base URL |
| `JINA_ENABLED` | `fetch.jina.enabled` | `true` / `false` |
| `JINA_API_KEY` | `fetch.jina.apiKey` | Jina Reader API key |
| `FETCH_MAX_LENGTH` | `fetch.maxLength` | Default fetch max length |
| `FETCH_TIMEOUT_MS` | `fetch.timeoutMs` | HTTP timeout in milliseconds |
| `FETCH_BROWSER_ENABLED` | `fetch.browser.enabled` | `true` / `false` |
| `ANSWER_LLM_ENABLED` | `answer.enabled` | `true` / `false` |
| `ANSWER_LLM_BASE_URL` | `answer.baseUrl` | OpenAI-compatible endpoint |
| `ANSWER_LLM_API_KEY` | `answer.apiKey` | API key |
| `ANSWER_LLM_MODEL` | `answer.model` | Model name |
| `ANSWER_LLM_MAX_TOKENS` | `answer.maxTokens` | Max tokens for final answers |
| `ANSWER_LLM_TEMPERATURE` | `answer.temperature` | Sampling temperature |
| `ANSWER_LLM_MAX_TURNS` | `answer.maxTurns` | Max agent loop turns |
| `ANSWER_LLM_MAX_SEARCHES` | `answer.maxSearches` | Max searches per answer |
| `ANSWER_LLM_MAX_FETCHES` | `answer.maxFetches` | Max fetches per answer |
| `TOOLS_WEB_SEARCH_ENABLED` | `tools.web_search.enabled` | `true` / `false` |
| `TOOLS_WEB_FETCH_ENABLED` | `tools.web_fetch.enabled` | `true` / `false` |
| `TOOLS_WEB_ANSWER_ENABLED` | `tools.web_answer.enabled` | `true` / `false` |

Backward compatibility: if `ANSWER_LLM_API_KEY` is set but no base URL is configured, Outrider defaults to the Google Gemini OpenAI-compatible endpoint (as before).

### 3. Configure SearXNG (optional)

To use a self-hosted SearXNG instance instead of DuckDuckGo:

```json
{
  "provider": "searxng",
  "searxng": {
    "baseUrl": "http://localhost:8080"
  }
}
```

Or with environment variables:

```bash
export SEARCH_PROVIDER=searxng
export SEARXNG_URL=http://localhost:8080
```

**SearXNG config requirement:** Make sure `json` is listed in your SearXNG `settings.yml` under `search.formats`:

```yaml
search:
  formats:
    - html
    - json
```

Without this, SearXNG returns HTTP 403 on every `/search?format=json` request. If SearXNG is selected but not reachable, the server falls back to DuckDuckGo automatically.

### 3b. Configure Degoog (optional)

To use a self-hosted [Degoog](https://degoog-org.github.io/docs/) instance instead of DuckDuckGo:

```json
{
  "provider": "degoog",
  "degoog": {
    "baseUrl": "http://localhost:14444"
  }
}
```

Or with environment variables:

```bash
export SEARCH_PROVIDER=degoog
export DEGOOG_URL=http://localhost:14444
```

Degoog runs on port `4444` by default. The provider queries its `/api/search` JSON endpoint, which needs no authentication on a local instance. If Degoog is selected but not reachable, the server falls back to DuckDuckGo automatically.

### 4. Add to your coding agent

Replace `/path/to/outrider` with the actual path to the binary.

#### OpenCode (using a config file)

```json
{
  "mcp": {
    "web-search": {
      "type": "local",
      "command": [
        "/path/to/outrider"
      ],
      "env": {
        "OUTRIDER_CONFIG": "/path/to/config.json"
      }
    }
  }
}
```

#### OpenCode (using env variables only)

```json
{
  "mcp": {
    "web-search": {
      "type": "local",
      "command": [
        "/path/to/outrider"
      ],
      "env": {
        "SEARCH_PROVIDER": "searxng",
        "SEARXNG_URL": "http://localhost:8080",
        "ANSWER_LLM_API_KEY": "your-key",
        "ANSWER_LLM_BASE_URL": "https://generativelanguage.googleapis.com/v1beta/openai/",
        "ANSWER_LLM_MODEL": "gemma-4-31b-it"
      }
    }
  }
}
```

If `web_answer` is not configured, the server starts with only `web_search` and `web_fetch`.

## Development

### Prerequisites

- Go 1.25+

### Project Structure

```
outrider-mcp/
├── main.go              # Server entry point — loads config and wires tools
├── config/              # Global config file + env override loader
│   └── config.go
├── answer/              # Side agent for the web_answer tool
│   ├── agent.go         # LLM client, tool schemas, bounded agent loop
│   └── prompts.go       # System prompts for URL and search modes
├── fetcher/             # HTTP fetching + HTML-to-Markdown extraction
│   ├── fetcher.go       # Tiered fetch orchestration
│   ├── jina.go          # Jina Reader tier
│   ├── browser.go       # chromedp headless-browser tier
│   └── extract.go       # Local readability + markdown extraction
├── search/              # Search provider interface and implementations
│   ├── provider.go      # Provider interface, shared types, helpers
│   ├── duckduckgo.go    # DuckDuckGo HTML search
│   ├── searxng.go       # SearXNG JSON API search
│   └── provider_factory.go  # Provider selection from config/env
├── tools/               # MCP tool definitions and handlers
│   ├── web_search.go    # web_search tool (provider-aware schema)
│   ├── web_fetch.go     # web_fetch tool
│   └── web_answer.go    # web_answer tool
└── test/                # Integration tests (live network calls)
    ├── search/
    │   └── duckduckgo_test.go
    ├── fetch/
    │   └── fetcher_test.go
    ├── answer/
    │   └── answer_test.go
    └── config/
        └── config_test.go
```

### Build

```bash
go build -o outrider .
```

### Run Tests

```bash
go test ./... -v -count=1 -timeout 120s
```

For `web_answer` tests (requires `ANSWER_LLM_API_KEY`):

```bash
ANSWER_LLM_API_KEY=your-key go test ./test/answer/ -v -count=1 -timeout 120s
```

## Verify It Works

Use the [MCP Inspector](https://github.com/modelcontextprotocol/inspector) to test the server:

```bash
npx @modelcontextprotocol/inspector /path/to/outrider
```

This opens a web UI where you can call all tools and inspect the responses.

## License

MIT
