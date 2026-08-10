# AGENTS.md

## Project Overview

**Outrider** is an MCP (Model Context Protocol) server that exposes three tools to AI agents: `web_search`, `web_fetch`, and `web_answer` (an optional side-LLM research agent). It ships as a single static binary — no Go toolchain or runtime deps required at deploy time.

- **Language**: Go (module `github.com/nethbotheju/outrider-mcp`, requires Go 1.26+)
- **MCP SDK**: `github.com/modelcontextprotocol/go-sdk`
- **Transport**: stdio only — JSON-RPC over stdin/stdout.

### Architecture

```
main.go              Entry point — loads config, wires provider + tools, runs server
config/              Config loading: JSON file (~/.config/outrider/config.json or ./outrider.json) + env overrides
search/              Search providers (DuckDuckGo default, SearXNG + Degoog optional) + factory
fetcher/             Tiered fetch: Jina Reader → chromedp headless browser → local readability/markdown
answer/              Bounded LLM agent loop for web_answer (search/fetch sub-calls, turn + tool caps)
tools/               MCP tool definitions and handlers
test/                Integration tests (LIVE network calls) — *_test packages
```

## Setup & Build

```bash
go build -o outrider .            # local build (reports version "dev")
go build -ldflags="-s -w" .       # release-style stripped build
go vet ./...                      # always run alongside a build
```

The server version is **injected at build time** via `-ldflags "-X main.version=<tag>"`. The `version` var in `main.go` defaults to `"dev"`. **Never hardcode a version string** — update the build/release pipeline instead.

## Testing

Two test tiers, kept in different packages on purpose:

```bash
go test ./search/...                  # unit tests — NO network, safe to run always
go test ./test/...                    # integration tests — LIVE network calls
go test ./... -timeout 120s           # everything (only when network is available)
```

- `search/*_test.go` (package `search`): unit tests, no external calls. These must stay green on every change.
- `test/**/*_test.go` (packages `*_test`): integration tests against live sites/services.
- `test/answer/answer_test.go` requires `ANSWER_LLM_API_KEY` (and a reachable OpenAI-compatible endpoint) or it skips.

Run unit tests after any change to a provider or config. Add or update tests for code you change, even if not asked.

## Critical Constraints

**stdout is reserved for JSON-RPC.** The server runs on the stdio transport, so any stray `fmt.Print` / `log` output to stdout corrupts the protocol and breaks every client. **All logging goes to stderr** (`log.*` defaults to stderr — keep it that way). Do not add stdout writes anywhere except the SDK's own transport.

**Env vars override the config file.** When adding a new option, wire it in `config/config.go` through both the struct field and `applyEnvOverrides`, and document it in both the README defaults table and env-var table. Don't read env vars ad hoc elsewhere.

**Graceful fallbacks.** If a selected provider (SearXNG/Degoog) is unreachable, the server falls back to DuckDuckGo rather than failing. Preserve this when adding providers — never hard-fail the whole server on a provider outage.

**Bounded agent loop.** `web_answer` enforces `maxTurns`, `maxSearches`, `maxFetches`. Any new tool the answer agent can call must respect these caps via `execCounters`.

## Code Style

- `gofmt`/`goimports` formatted — run `gofmt -l .` (should output nothing).
- `go vet ./...` clean.
- **Minimal comments.** Keep inline comments to the absolute minimum; the code should be self-explanatory. Reserve comments for genuinely non-obvious behavior (e.g., the stderr/stdout constraint, build-time version injection).
- Package layout follows the tree above. Keep tool handlers in `tools/`, provider impls in `search/`.

## Configuration

Outrider is configured by a JSON file, environment variables, or both (env wins). Defaults live in `config.Defaults()`. When touching config:

1. Add the field to the relevant struct (with a `json` tag).
2. Set its default in `Defaults()`.
3. Add the env override in `applyEnvOverrides` if it should be env-controllable.
4. Update the README defaults + env-var tables.

Local dev config: copy the README's example JSON to `./outrider.json` or set env vars. `.env` is gitignored and is the conventional place for local secrets.

## Releases & Commits

**Releases are tag-driven.** Pushing a `v*` tag triggers `.github/workflows/release.yml`, which cross-compiles (darwin/linux/windows × amd64/arm64, plus android/arm64) and publishes a GitHub Release with auto-generated notes. The workflow injects the tag as `main.version`, so do not bump the version in source.

**Commit messages** follow Conventional Commits:

- `feat:` / `fix:` / `chore:` / `docs:` / `refactor:` — optional scope, e.g. `fix(search): ...`
- Imperative mood, lowercase after the colon.

**Git operations are manual.** Do not run `git add`, `git commit`, `git push`, or create tags autonomously — surface the exact command for a human to review and run. This includes cutting releases.
