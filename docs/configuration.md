# Configuration

Outrider can be configured with a **JSON config file**, **environment variables**, or **both**. Environment variables always take precedence over the config file.

If you just want to search and fetch with DuckDuckGo, **you can skip this entirely** — Outrider works with zero configuration. This document covers everything beyond the defaults.

## Table of contents

- [Config file](#config-file)
- [Defaults](#defaults)
- [Environment variables](#environment-variables)
- [The two `answer` toggles](#the-two-answer-toggles)
- [Search providers](#search-providers)
- [Client integration](#client-integration)

---

## Config file

Outrider looks for a config file in this order:

1. Path from the `OUTRIDER_CONFIG` environment variable
2. `~/.config/outrider/config.json`
3. `./outrider.json` (current directory)

If no file is found, the defaults below are used. Missing fields in a file fall back to defaults too.

A complete config file (most fields are optional):

```json
{
  "provider": "duckduckgo",
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

> The `answer` block accepts any **OpenAI-compatible** endpoint — a local Ollama instance, OpenAI, Google Gemini, LM Studio, etc.

---

## Defaults

| Field | Default | Notes |
|-------|---------|-------|
| `provider` | `duckduckgo` | Auto-switches to `searxng` if `searxng.baseUrl` is set, or `degoog` if `degoog.baseUrl` is set. |
| `searxng` | unset | Only used when `provider` is `searxng`. |
| `degoog` | unset | Only used when `provider` is `degoog`. |
| `fetch.jina.enabled` | `true` | Jina Reader is tried first. |
| `fetch.maxLength` | `10000` | Clamped to a maximum of `50000`. |
| `fetch.timeoutMs` | `30000` | HTTP timeout for Jina and local fetch. |
| `fetch.browser.enabled` | `true` | Headless Chrome fallback tier. |
| `answer.enabled` | `false` | Must be `true` for `web_answer` to work. |
| `answer.maxTokens` | `800` | Applied to final-answer LLM calls. |
| `answer.temperature` | `0` | Applied to all answer LLM calls. |
| `answer.maxTurns` | `4` | Max agent-loop turns when no URL is given. |
| `answer.maxSearches` | `2` | Max `web_search` calls per answer. |
| `answer.maxFetches` | `2` | Max `web_fetch` calls per answer. |
| `tools.web_search.enabled` | `true` | |
| `tools.web_fetch.enabled` | `true` | |
| `tools.web_answer.enabled` | `false` | |

---

## Environment variables

Every config field has an equivalent environment variable. Env vars override the config file.

| Variable | Maps to | Description |
|----------|---------|-------------|
| `OUTRIDER_CONFIG` | — | Path to the config file |
| `SEARCH_PROVIDER` | `provider` | `duckduckgo`, `searxng`, or `degoog` |
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
| `ANSWER_LLM_MAX_TURNS` | `answer.maxTurns` | Max agent-loop turns |
| `ANSWER_LLM_MAX_SEARCHES` | `answer.maxSearches` | Max searches per answer |
| `ANSWER_LLM_MAX_FETCHES` | `answer.maxFetches` | Max fetches per answer |
| `TOOLS_WEB_SEARCH_ENABLED` | `tools.web_search.enabled` | `true` / `false` |
| `TOOLS_WEB_FETCH_ENABLED` | `tools.web_fetch.enabled` | `true` / `false` |
| `TOOLS_WEB_ANSWER_ENABLED` | `tools.web_answer.enabled` | `true` / `false` |

**Backward compatibility:** if `ANSWER_LLM_API_KEY` is set but no base URL is configured, Outrider defaults to the Google Gemini OpenAI-compatible endpoint.

---

## The two `answer` toggles

`web_answer` is controlled by **two separate switches**, and both must be on for the tool to appear:

- `answer.enabled` — turns the **side LLM** on or off. When `true`, the model settings (`baseUrl`, `apiKey`, `model`, …) must also be valid.
- `tools.web_answer.enabled` — controls whether the **`web_answer` tool** is advertised to the agent.

This lets you configure the side LLM once and decide per-project whether the agent may use it.

---

## Search providers

### DuckDuckGo (default)

No setup. This is the default provider and works with no configuration.

### SearXNG

Point Outrider at a self-hosted [SearXNG](https://searxng.org) instance:

```json
{
  "provider": "searxng",
  "searxng": {
    "baseUrl": "http://localhost:8080"
  }
}
```

…or with environment variables:

```bash
export SEARCH_PROVIDER=searxng
export SEARXNG_URL=http://localhost:8080
```

**SearXNG requirement:** add `json` to `search.formats` in your SearXNG `settings.yml`, otherwise every request returns HTTP 403:

```yaml
search:
  formats:
    - html
    - json
```

SearXNG unlocks the `category` parameter (`general` / `science` / `news` / `it`) and richer metadata. If SearXNG is selected but unreachable, Outrider falls back to DuckDuckGo automatically.

### Degoog

Point Outrider at a self-hosted [Degoog](https://degoog-org.github.io/docs/) instance:

```json
{
  "provider": "degoog",
  "degoog": {
    "baseUrl": "http://localhost:14444"
  }
}
```

…or with environment variables:

```bash
export SEARCH_PROVIDER=degoog
export DEGOOG_URL=http://localhost:14444
```

Degoog runs on port `4444` by default and needs no authentication on a local instance. If Degoog is selected but unreachable, Outrider falls back to DuckDuckGo automatically.

---

## Client integration

Replace `/path/to/outrider` with the absolute path to the binary in every example.

### Claude Desktop

Edit `claude_desktop_config.json` (*Settings → Developer → Edit Config*):

```json
{
  "mcpServers": {
    "outrider": {
      "command": "/path/to/outrider"
    }
  }
}
```

### Cursor

Create or edit `.cursor/mcp.json` in your project:

```json
{
  "mcpServers": {
    "outrider": {
      "command": "/path/to/outrider"
    }
  }
}
```

### Claude Code

Add it from your terminal with the `claude mcp add` command:

```bash
claude mcp add outrider -- /path/to/outrider
```

By default this registers the server at `local` scope (current project, private to you). Use `--scope user` for every project, or `--scope project` to commit a shareable `.mcp.json` into the repo.

### OpenAI Codex

Add a server block to `~/.codex/config.toml`:

```toml
[mcp_servers.outrider]
command = "/path/to/outrider"
```

### OpenCode

Add an entry under `mcp` in your OpenCode config:

```json
{
  "mcp": {
    "outrider": {
      "type": "local",
      "command": ["/path/to/outrider"],
      "env": {
        "OUTRIDER_CONFIG": "/path/to/config.json"
      }
    }
  }
}
```

If `web_answer` is not configured, the server starts with only `web_search` and `web_fetch`.
