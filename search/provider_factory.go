package search

import (
	"log"
	"net/url"
	"os"
	"strings"
)

type ProviderName string

const (
	ProviderDuckDuckGo ProviderName = "duckduckgo"
	ProviderSearXNG    ProviderName = "searxng"
)

// NewProviderFromEnv returns the search provider configured via environment
func NewProviderFromEnv() Provider {
	providerName := strings.ToLower(strings.TrimSpace(os.Getenv("SEARCH_PROVIDER")))
	if providerName == "" {
		providerName = string(ProviderDuckDuckGo)
	}

	switch ProviderName(providerName) {
	case ProviderSearXNG:
		baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("SEARXNG_URL")), "/")

		if baseURL == "" {
			log.Printf("[search] Warning: SEARCH_PROVIDER=searxng but SEARXNG_URL is not set; falling back to DuckDuckGo")
			return NewDuckDuckGoProvider()
		}

		if _, err := url.ParseRequestURI(baseURL); err != nil {
			log.Printf("[search] Warning: SEARXNG_URL=%q is not a valid URL (%v); falling back to DuckDuckGo", baseURL, err)
			return NewDuckDuckGoProvider()
		}

		log.Printf("[search] Using SearXNG at %s", baseURL)
		return NewSearXNGProvider(baseURL)

	default:
		if providerName != "" {
			log.Printf("[search] Unknown SEARCH_PROVIDER=%q; using DuckDuckGo", providerName)
		}
		return NewDuckDuckGoProvider()
	}
}
