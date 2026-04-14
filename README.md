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

### `answer`

Answer a question by searching the web and reading relevant pages. Returns a concise answer with sources. Uses a side agent (LLM + web_search + fetch) internally, so the main agent's context window stays clean. Requires an API key.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `question` | string | Yes | The question to answer |

When the `answer` tool is enabled, prefer it over manually calling `web_search` + `fetch` for factual questions. It produces the same result using a fraction of the context window.

## Setup

### 1. Download the binary

Download the latest release for your platform from the [Releases](../../releases) page.

### 2. Configure the API key (for the `answer` tool)

The `answer` tool requires an API key for a side LLM. [Google AI Studio](https://aistudio.google.com/apikey) offers a decent free quota for models like **Gemma 4 31B** (`gemma-4-31b-it`) and **Gemma 4 26B** (`gemma-4-26b-a4b-it`). You can also use any OpenAI-compatible endpoint (OpenRouter, Ollama, etc.) by changing the base URL.

The default base URL is set to Google's Gemini endpoint and the default model is `gemma-4-31b-it`. If you want to use a different model or endpoint, set these environment variables alongside the API key:

| Variable | Default | Description |
|----------|---------|-------------|
| `ANSWER_LLM_API_KEY` | — | API key for the side LLM. If not set, the `answer` tool is disabled. |
| `ANSWER_LLM_BASE_URL` | `https://generativelanguage.googleapis.com/v1beta/openai/` | Base URL for the LLM API (any OpenAI-compatible endpoint) |
| `ANSWER_LLM_MODEL` | `gemma-4-31b-it` | Model name |

This key is only needed for the `answer` tool. The `web_search` and `fetch` tools work without it.

### 3. Add to your coding agent

Replace `/path/to/web-search-mcp` with the actual path to the binary.

#### OpenCode

Add to your project's `opencode.json`:

```json
{
  "mcp": {
    "web-search": {
      "type": "local",
      "command": [
        "/path/to/web-search-mcp"
      ],
      "env": {
        "ANSWER_LLM_API_KEY": "your-google-ai-studio-api-key"
      }
    }
  }
}
```

If `ANSWER_LLM_API_KEY` is not set, the server starts with only `web_search` and `fetch` -- fully backward compatible.

## Development

### Prerequisites

- Go 1.25+

### Project Structure

```
web-search-mcp/
├── main.go              # Server entry point — wires tools to the MCP server
├── answer/              # Side agent for the answer tool
│   └── agent.go         # LLM client, tool schemas, agent loop
├── fetcher/             # HTTP fetching + HTML-to-text extraction
│   └── fetcher.go
├── search/              # Search provider interface and implementations
│   ├── provider.go      # Provider interface, shared types, helpers
│   └── duckduckgo.go    # DuckDuckGo HTML search (free, no API key)
├── tools/               # MCP tool definitions and handlers
│   ├── web_search.go    # web_search tool
│   ├── fetch.go         # fetch tool
│   └── answer.go        # answer tool
└── test/                # Integration tests (live network calls)
    ├── search/
    │   └── duckduckgo_test.go
    ├── fetch/
    │   └── fetcher_test.go
    └── answer/
        └── answer_test.go
```

### Build

```bash
go build -o web-search-mcp .
```

### Run Tests

```bash
go test ./test/search/ ./test/fetch/ -v -count=1 -timeout 60s
```

For answer tool tests (requires `ANSWER_LLM_API_KEY`):

```bash
ANSWER_LLM_API_KEY=your-key go test ./test/answer/ -v -count=1 -timeout 120s
```

## Verify It Works

Use the [MCP Inspector](https://github.com/modelcontextprotocol/inspector) to test the server:

```bash
npx @modelcontextprotocol/inspector /path/to/web-search-mcp
```

This opens a web UI where you can call all tools and inspect the responses.

## License

MIT
