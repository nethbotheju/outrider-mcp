<div align="center">

<img src="https://cdn.simpleicons.org/modelcontextprotocol/6f42c1" width="110" alt="MCP" />

# Outrider

**The web-search MCP server that scouts ahead of your AI agent.**

[![release](https://img.shields.io/github/v/release/nethbotheju/web-search-mcp?style=flat-square&label=latest%20release)](https://github.com/nethbotheju/web-search-mcp/releases)
[![license](https://img.shields.io/badge/license-MIT-blue?style=flat-square)](./LICENSE)
[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white&style=flat-square)](https://go.dev)
[![MCP](https://img.shields.io/badge/MCP-server-6f42c1?style=flat-square)](https://modelcontextprotocol.io)
[![platforms](https://img.shields.io/badge/platforms-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey?style=flat-square)](https://github.com/nethbotheju/web-search-mcp/releases)

</div>

Outrider gives any [Model Context Protocol](https://modelcontextprotocol.io) client — **Claude Desktop, Cursor, OpenCode**, and the rest — three fast, dependable tools for working with the live web: **search**, **fetch**, and a bounded **question-answering** agent.

It ships as a **single static binary**. No Go toolchain at runtime, no `node_modules`, no Python, and **no API keys to get started**.

> Works the moment you download it. DuckDuckGo search is built in by default — just run the binary.

## Why Outrider

- **Zero-config by default** — search works out of the box. No key, no signup, no billing.
- **One static binary** — drop it anywhere and run it. Cross-compiled for macOS, Linux, Windows, and ARM.
- **Pages actually load** — a three-tier fetcher tries the Jina Reader API, then a headless browser, then local extraction, so you get clean Markdown from real-world pages.
- **Research without context bloat** — `web_answer` runs a small side-agent that searches and reads only what's needed, then returns a concise answer with sources.
- **Private & self-hostable search** — swap in your own [SearXNG](https://searxng.org) or [Degoog](https://degoog-org.github.io/docs/) instance for richer results, metadata, and no rate limits.
- **Tunable tool surface** — enable or disable each tool, and point the answer agent at any OpenAI-compatible model.

## How it compares

| | **Outrider** | Most web-search MCPs |
|---|---|---|
| Runtime | One static binary | Node / Python + deps |
| Default search | DuckDuckGo — no API key | Often requires Brave / Tavily / Serper key |
| Page fetching | 3-tier (Jina → browser → local) | Usually a single method |
| Research agent | Bounded side-agent (`web_answer`) | Uncommon |
| Private search | SearXNG / Degoog | Uncommon |
| Distribution | Prebuilt binaries, 6 targets | `npx` / `pip` |

## Quick start

**1. Download** the binary for your platform from the [Releases](https://github.com/nethbotheju/web-search-mcp/releases) page and make it executable.

**2. Add it to your client.** For **Claude Desktop**, open `claude_desktop_config.json` (*Settings → Developer → Edit Config*):

```json
{
  "mcpServers": {
    "outrider": {
      "command": "/absolute/path/to/outrider"
    }
  }
}
```

**3. Restart your client.** You now have `web_search` and `web_fetch`. No further configuration required.

> Want `web_answer`, a self-hosted search engine, or fine-grained tuning? See **[Configuration →](./docs/configuration.md)**.

If something doesn't work, [open an issue](https://github.com/nethbotheju/web-search-mcp/issues).

<details>
<summary><b>Using Cursor, Claude Code, Codex, or OpenCode?</b></summary>

See the [Client integration](./docs/configuration.md#client-integration) guide for setup snippets.

</details>

## Tools

| Tool | What it does |
|---|---|
| **`web_search`** | Search the web. Returns numbered sources with titles, URLs, and descriptions. With SearXNG, also returns published date, source, journal, DOI, PDF link, and authors. |
| **`web_fetch`** | Fetch a URL and return clean Markdown. Three-tier: Jina Reader → headless browser → local extraction. `lean` mode strips links and images to save tokens. |
| **`web_answer`** | Answer a question from the web using a bounded side-agent. Pass a `url` to read a specific page, or omit it to let the agent search and fetch. Returns a concise answer with sources — and keeps your main context window clean. |

<details>
<summary><b>Parameters</b></summary>

**`web_search`**
- `query` *(string, required)* — the search query
- `count` *(int)* — number of results (DuckDuckGo/Degoog: default 10, max 20 · SearXNG: default 5, max 10)
- `category` *(string)* — **SearXNG only:** `general` (default), `science`, `news`, `it`

**`web_fetch`**
- `url` *(string, required)* — the URL to fetch
- `format` *(string)* — `lean` (default, strips links & images) or `markdown` (full)
- `maxLength` *(int)* — max characters, default `10000`, max `50000`

**`web_answer`**
- `question` *(string, required)* — the question to answer
- `url` *(string)* — a specific page to read and answer from. Omit to search the web.

</details>

## Search providers

| Provider | Setup | Notes |
|---|---|---|
| **DuckDuckGo** | None — default | Works out of the box. |
| **SearXNG** | Self-host | Category filtering + rich metadata. Falls back to DuckDuckGo if unreachable. |
| **Degoog** | Self-host | Aggregated results from multiple engines, no API key. Falls back to DuckDuckGo if unreachable. |

Setup steps for SearXNG and Degoog are in [Configuration →](./docs/configuration.md#search-providers).

## Development

Requires **Go 1.26+**.

```bash
go build -o outrider .              # build
go test ./search/...                # unit tests (offline)
go test ./test/... -timeout 120s    # integration tests (live network)
```

Point any MCP client at the built binary, or inspect it with the [MCP Inspector](https://github.com/modelcontextprotocol/inspector):

```bash
npx @modelcontextprotocol/inspector ./outrider
```

For architecture, conventions, and the constraints agents must follow, see **[AGENTS.md](./AGENTS.md)**.

## Contributing

Pull requests are welcome — from humans and coding agents alike.

- **Using an AI coding agent?** Read [AGENTS.md](./AGENTS.md) first. It documents the project structure, conventions, and the hard constraints every change must respect (for example, stdout is reserved for the MCP JSON-RPC stream, so all logging must go to stderr).
- **Spotted a bug or have an idea?** [Open an issue](https://github.com/nethbotheju/web-search-mcp/issues).

## License

This project is licensed under the [MIT License](./LICENSE).

Copyright © 2026 nethbotheju.
