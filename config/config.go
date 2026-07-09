package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	defaultGoogleBaseURL = "https://generativelanguage.googleapis.com/v1beta/openai/"
	defaultGoogleModel   = "gemma-4-31b-it"
)

// SearXNGConfig holds provider-specific settings for a SearXNG instance.
type SearXNGConfig struct {
	BaseURL string `json:"baseUrl"`
	APIKey  string `json:"apiKey,omitempty"`
}

// DegoogConfig holds provider-specific settings for a Degoog instance.
type DegoogConfig struct {
	BaseURL string `json:"baseUrl"`
}

// JinaConfig controls the Jina Reader fetch tier.
type JinaConfig struct {
	Enabled bool   `json:"enabled"`
	APIKey  string `json:"apiKey,omitempty"`
}

// BrowserConfig controls the optional chromedp headless-browser tier.
type BrowserConfig struct {
	Enabled bool `json:"enabled"`
}

// FetchConfig controls how web pages are fetched and converted.
type FetchConfig struct {
	Jina      JinaConfig    `json:"jina"`
	MaxLength int           `json:"maxLength"`
	TimeoutMs int           `json:"timeoutMs"`
	Browser   BrowserConfig `json:"browser"`
}

// AnswerConfig controls the optional side LLM used by web_answer.
type AnswerConfig struct {
	Enabled     bool    `json:"enabled"`
	BaseURL     string  `json:"baseUrl"`
	APIKey      string  `json:"apiKey"`
	Model       string  `json:"model"`
	MaxTokens   int     `json:"maxTokens"`
	Temperature float64 `json:"temperature"`
	MaxTurns    int     `json:"maxTurns"`
	MaxSearches int     `json:"maxSearches"`
	MaxFetches  int     `json:"maxFetches"`
}

// ToolToggle is a simple on/off switch for a tool.
type ToolToggle struct {
	Enabled bool `json:"enabled"`
}

// ToolsConfig lets users enable or expose individual tools.
type ToolsConfig struct {
	WebSearch ToolToggle `json:"web_search"`
	WebFetch  ToolToggle `json:"web_fetch"`
	WebAnswer ToolToggle `json:"web_answer"`
}

// Config is the top-level global configuration for Outrider.
type Config struct {
	Provider string         `json:"provider"`
	SearXNG  *SearXNGConfig `json:"searxng,omitempty"`
	Degoog   *DegoogConfig  `json:"degoog,omitempty"`
	Fetch    FetchConfig    `json:"fetch"`
	Answer   AnswerConfig   `json:"answer"`
	Tools    ToolsConfig    `json:"tools"`
}

// Defaults returns a fully populated default configuration.
func Defaults() *Config {
	return &Config{
		Provider: "duckduckgo",
		Fetch: FetchConfig{
			Jina:      JinaConfig{Enabled: true},
			MaxLength: 10000,
			TimeoutMs: 30000,
			Browser:   BrowserConfig{Enabled: true},
		},
		Answer: AnswerConfig{
			Enabled:     false,
			BaseURL:     "",
			APIKey:      "",
			Model:       "",
			MaxTokens:   800,
			Temperature: 0,
			MaxTurns:    4,
			MaxSearches: 2,
			MaxFetches:  2,
		},
		Tools: ToolsConfig{
			WebSearch: ToolToggle{Enabled: true},
			WebFetch:  ToolToggle{Enabled: true},
			WebAnswer: ToolToggle{Enabled: false},
		},
	}
}

// Load reads the configuration file and applies environment overrides.
func Load() (*Config, error) {
	cfg := Defaults()

	path, err := resolveConfigPath()
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("resolving config path: %w", err)
	}

	if path != "" {
		data, err := os.ReadFile(path)
		if err == nil {
			if err := json.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("parsing config file %q: %w", path, err)
			}
			if cfg.Provider == "" && cfg.SearXNG != nil && cfg.SearXNG.BaseURL != "" {
				cfg.Provider = "searxng"
			}
			if cfg.Provider == "" && cfg.Degoog != nil && cfg.Degoog.BaseURL != "" {
				cfg.Provider = "degoog"
			}
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("reading config file %q: %w", path, err)
		}
	}

	applyEnvOverrides(cfg)
	return cfg, nil
}

// resolveConfigPath returns the path to the config file, or "" if none found.
func resolveConfigPath() (string, error) {
	if p := os.Getenv("OUTRIDER_CONFIG"); p != "" {
		return p, nil
	}

	home, err := os.UserHomeDir()
	if err == nil {
		p := filepath.Join(home, ".config", "outrider", "config.json")
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	p := filepath.Join(".", "outrider.json")
	if _, err := os.Stat(p); err == nil {
		return p, nil
	}

	return "", os.ErrNotExist
}

// applyEnvOverrides lets environment variables take precedence over file values.
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("SEARCH_PROVIDER"); v != "" {
		cfg.Provider = strings.ToLower(strings.TrimSpace(v))
	} else if os.Getenv("SEARXNG_URL") != "" {
		cfg.Provider = "searxng"
	} else if os.Getenv("DEGOOG_URL") != "" {
		cfg.Provider = "degoog"
	}

	if cfg.SearXNG == nil {
		cfg.SearXNG = &SearXNGConfig{}
	}
	if v := os.Getenv("SEARXNG_URL"); v != "" {
		cfg.SearXNG.BaseURL = strings.TrimRight(strings.TrimSpace(v), "/")
	}
	if v, ok := os.LookupEnv("SEARXNG_API_KEY"); ok {
		cfg.SearXNG.APIKey = v
	}

	if cfg.Degoog == nil {
		cfg.Degoog = &DegoogConfig{}
	}
	if v := os.Getenv("DEGOOG_URL"); v != "" {
		cfg.Degoog.BaseURL = strings.TrimRight(strings.TrimSpace(v), "/")
	}

	if v, ok := lookupBool("JINA_ENABLED"); ok {
		cfg.Fetch.Jina.Enabled = v
	}
	if v, ok := os.LookupEnv("JINA_API_KEY"); ok {
		cfg.Fetch.Jina.APIKey = v
	}
	if v, ok := lookupInt("FETCH_MAX_LENGTH"); ok {
		cfg.Fetch.MaxLength = v
	}
	if v, ok := lookupInt("FETCH_TIMEOUT_MS"); ok {
		cfg.Fetch.TimeoutMs = v
	}
	if v, ok := lookupBool("FETCH_BROWSER_ENABLED"); ok {
		cfg.Fetch.Browser.Enabled = v
	}

	if v, ok := lookupBool("ANSWER_LLM_ENABLED"); ok {
		cfg.Answer.Enabled = v
	} else if os.Getenv("ANSWER_LLM_API_KEY") != "" {
		cfg.Answer.Enabled = true
	}
	if v := os.Getenv("ANSWER_LLM_BASE_URL"); v != "" {
		cfg.Answer.BaseURL = strings.TrimRight(strings.TrimSpace(v), "/")
	}
	if v, ok := os.LookupEnv("ANSWER_LLM_API_KEY"); ok {
		cfg.Answer.APIKey = v
	}
	if v := os.Getenv("ANSWER_LLM_MODEL"); v != "" {
		cfg.Answer.Model = strings.TrimSpace(v)
	}
	if v, ok := lookupInt("ANSWER_LLM_MAX_TOKENS"); ok {
		cfg.Answer.MaxTokens = v
	}
	if v, ok := lookupFloat("ANSWER_LLM_TEMPERATURE"); ok {
		cfg.Answer.Temperature = v
	}
	if v, ok := lookupInt("ANSWER_LLM_MAX_TURNS"); ok {
		cfg.Answer.MaxTurns = v
	}
	if v, ok := lookupInt("ANSWER_LLM_MAX_SEARCHES"); ok {
		cfg.Answer.MaxSearches = v
	}
	if v, ok := lookupInt("ANSWER_LLM_MAX_FETCHES"); ok {
		cfg.Answer.MaxFetches = v
	}

	if v, ok := lookupBool("TOOLS_WEB_SEARCH_ENABLED"); ok {
		cfg.Tools.WebSearch.Enabled = v
	}
	if v, ok := lookupBool("TOOLS_WEB_FETCH_ENABLED"); ok {
		cfg.Tools.WebFetch.Enabled = v
	}
	if v, ok := lookupBool("TOOLS_WEB_ANSWER_ENABLED"); ok {
		cfg.Tools.WebAnswer.Enabled = v
	}

	// Backward compatibility: if an answer API key is configured but no base URL
	// is set, default to the Google Gemini endpoint (the previous behaviour).
	if cfg.Answer.APIKey != "" && cfg.Answer.BaseURL == "" {
		cfg.Answer.BaseURL = defaultGoogleBaseURL
		cfg.Answer.Model = defaultGoogleModel
	}
}

func lookupBool(key string) (bool, bool) {
	v, ok := os.LookupEnv(key)
	if !ok {
		return false, false
	}
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "true", "1", "yes", "on":
		return true, true
	case "false", "0", "no", "off":
		return false, true
	default:
		return false, false
	}
}

func lookupInt(key string) (int, bool) {
	v, ok := os.LookupEnv(key)
	if !ok {
		return 0, false
	}
	n, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil {
		return 0, false
	}
	return n, true
}

func lookupFloat(key string) (float64, bool) {
	v, ok := os.LookupEnv(key)
	if !ok {
		return 0, false
	}
	n, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
	if err != nil {
		return 0, false
	}
	return n, true
}
