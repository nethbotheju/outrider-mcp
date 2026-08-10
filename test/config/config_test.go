package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nethbotheju/outrider-mcp/config"
)

func TestDefaults(t *testing.T) {
	cfg := config.Defaults()

	if cfg.Provider != "duckduckgo" {
		t.Errorf("expected default provider duckduckgo, got %q", cfg.Provider)
	}
	if !cfg.Fetch.Jina.Enabled {
		t.Error("expected Jina enabled by default")
	}
	if !cfg.Fetch.Browser.Enabled {
		t.Error("expected browser enabled by default")
	}
	if cfg.Fetch.MaxLength != 10000 {
		t.Errorf("expected default maxLength 10000, got %d", cfg.Fetch.MaxLength)
	}
	if cfg.Answer.Enabled {
		t.Error("expected answer disabled by default")
	}
	if cfg.Tools.WebSearch.Enabled != true || cfg.Tools.WebFetch.Enabled != true || cfg.Tools.WebAnswer.Enabled != false {
		t.Errorf("unexpected default tool toggles: %+v", cfg.Tools)
	}
}

func TestEnvOverrides(t *testing.T) {
	// Point to a non-existent config file so only env vars and defaults are used.
	tmpDir := t.TempDir()
	missingConfig := filepath.Join(tmpDir, "missing.json")

	setenv(t, "OUTRIDER_CONFIG", missingConfig)
	setenv(t, "SEARCH_PROVIDER", "searxng")
	setenv(t, "SEARXNG_URL", "http://localhost:8080")
	setenv(t, "SEARXNG_API_KEY", "secret")
	setenv(t, "JINA_ENABLED", "false")
	setenv(t, "FETCH_BROWSER_ENABLED", "false")
	setenv(t, "ANSWER_LLM_ENABLED", "true")
	setenv(t, "ANSWER_LLM_MODEL", "qwen2.5:7b")
	setenv(t, "TOOLS_WEB_ANSWER_ENABLED", "true")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load() error: %v", err)
	}

	if cfg.Provider != "searxng" {
		t.Errorf("expected provider searxng, got %q", cfg.Provider)
	}
	if cfg.SearXNG == nil || cfg.SearXNG.BaseURL != "http://localhost:8080" || cfg.SearXNG.APIKey != "secret" {
		t.Errorf("unexpected searxng config: %+v", cfg.SearXNG)
	}
	if cfg.Fetch.Jina.Enabled {
		t.Error("expected Jina disabled from env")
	}
	if cfg.Fetch.Browser.Enabled {
		t.Error("expected browser disabled from env")
	}
	if !cfg.Answer.Enabled {
		t.Error("expected answer enabled from env")
	}
	if cfg.Answer.Model != "qwen2.5:7b" {
		t.Errorf("expected answer model qwen2.5:7b, got %q", cfg.Answer.Model)
	}
	if !cfg.Tools.WebAnswer.Enabled {
		t.Error("expected web_answer tool enabled from env")
	}
}

func TestConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "outrider.json")
	data := []byte(`{
  "provider": "searxng",
  "searxng": {
    "baseUrl": "http://searxng.local",
    "apiKey": "key-from-file"
  },
  "fetch": {
    "maxLength": 5000,
    "browser": { "enabled": false }
  },
  "answer": {
    "enabled": true,
    "baseUrl": "http://localhost:11434/v1",
    "model": "llama3.1:8b"
  },
  "tools": {
    "web_answer": { "enabled": true }
  }
}`)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("writing test config: %v", err)
	}

	setenv(t, "OUTRIDER_CONFIG", path)
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load() error: %v", err)
	}

	if cfg.Provider != "searxng" {
		t.Errorf("expected provider from file, got %q", cfg.Provider)
	}
	if cfg.Fetch.MaxLength != 5000 {
		t.Errorf("expected maxLength 5000 from file, got %d", cfg.Fetch.MaxLength)
	}
	if cfg.Fetch.Browser.Enabled {
		t.Error("expected browser disabled from file")
	}
	if !cfg.Answer.Enabled || cfg.Answer.Model != "llama3.1:8b" {
		t.Errorf("unexpected answer config: %+v", cfg.Answer)
	}
}

func TestEnvOverridesFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "outrider.json")
	data := []byte(`{"provider": "duckduckgo", "fetch": {"maxLength": 5000}}`)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("writing test config: %v", err)
	}

	setenv(t, "OUTRIDER_CONFIG", path)
	setenv(t, "FETCH_MAX_LENGTH", "15000")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load() error: %v", err)
	}
	if cfg.Fetch.MaxLength != 15000 {
		t.Errorf("expected env to override file maxLength, got %d", cfg.Fetch.MaxLength)
	}
}

func setenv(t *testing.T, key, value string) {
	t.Helper()
	prev, existed := os.LookupEnv(key)
	os.Setenv(key, value)
	t.Cleanup(func() {
		if existed {
			os.Setenv(key, prev)
		} else {
			os.Unsetenv(key)
		}
	})
}
