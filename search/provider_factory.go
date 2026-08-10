package search

import (
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/nethbotheju/outrider-mcp/config"
)

// NewProviderFromEnv builds a provider using only environment variables.
func NewProviderFromEnv() Provider {
	// The config package's env override logic is unexported, so replicate the
	// minimal provider-relevant parts here for backward compatibility.
	providerName := strings.ToLower(strings.TrimSpace(os.Getenv("SEARCH_PROVIDER")))
	searxngURL := strings.TrimRight(strings.TrimSpace(os.Getenv("SEARXNG_URL")), "/")
	searxngKey := os.Getenv("SEARXNG_API_KEY")
	degoogURL := strings.TrimRight(strings.TrimSpace(os.Getenv("DEGOOG_URL")), "/")

	if providerName == "" && searxngURL != "" {
		providerName = "searxng"
	} else if providerName == "" && degoogURL != "" {
		providerName = "degoog"
	}

	return newProvider(providerName, searxngURL, searxngKey, degoogURL)
}

// NewProviderFromConfig builds a provider from the loaded configuration.
func NewProviderFromConfig(cfg *config.Config) Provider {
	providerName := strings.ToLower(strings.TrimSpace(cfg.Provider))
	searxngURL := ""
	searxngKey := ""
	if cfg.SearXNG != nil {
		searxngURL = cfg.SearXNG.BaseURL
		searxngKey = cfg.SearXNG.APIKey
	}
	degoogURL := ""
	if cfg.Degoog != nil {
		degoogURL = cfg.Degoog.BaseURL
	}
	return newProvider(providerName, searxngURL, searxngKey, degoogURL)
}

func newProvider(providerName, searxngURL, searxngKey, degoogURL string) Provider {
	if providerName == "" {
		providerName = "duckduckgo"
	}

	switch providerName {
	case "searxng":
		baseURL := strings.TrimRight(strings.TrimSpace(searxngURL), "/")
		if baseURL == "" {
			log.Printf("[search] Warning: SEARCH_PROVIDER=searxng but SEARXNG_URL is not set; falling back to DuckDuckGo")
			return NewDuckDuckGoProvider()
		}
		if _, err := url.ParseRequestURI(baseURL); err != nil {
			log.Printf("[search] Warning: SEARXNG_URL=%q is not a valid URL (%v); falling back to DuckDuckGo", baseURL, err)
			return NewDuckDuckGoProvider()
		}
		log.Printf("[search] Using SearXNG at %s", baseURL)
		return NewSearXNGProvider(baseURL, searxngKey)

	case "degoog":
		baseURL := strings.TrimRight(strings.TrimSpace(degoogURL), "/")
		if baseURL == "" {
			log.Printf("[search] Warning: SEARCH_PROVIDER=degoog but DEGOOG_URL is not set; falling back to DuckDuckGo")
			return NewDuckDuckGoProvider()
		}
		if _, err := url.ParseRequestURI(baseURL); err != nil {
			log.Printf("[search] Warning: DEGOOG_URL=%q is not a valid URL (%v); falling back to DuckDuckGo", baseURL, err)
			return NewDuckDuckGoProvider()
		}
		log.Printf("[search] Using Degoog at %s", baseURL)
		return NewDegoogProvider(baseURL)

	default:
		if providerName != "duckduckgo" {
			log.Printf("[search] Unknown SEARCH_PROVIDER=%q; using DuckDuckGo", providerName)
		}
		return NewDuckDuckGoProvider()
	}
}
