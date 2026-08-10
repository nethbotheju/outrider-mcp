package search

import (
	"testing"
)

func TestParseDegoogResults(t *testing.T) {
	data := []byte(`{
		"results": [
			{
				"title": "The Go Programming Language",
				"url": "https://go.dev",
				"snippet": "Go is an open source programming language.",
				"source": "Bing",
				"sources": ["Bing", "Brave Search"],
				"content": "ignored when snippet is present"
			},
			{
				"title": "Go on Wikipedia",
				"url": "https://en.wikipedia.org/wiki/Go_(programming_language)",
				"snippet": "",
				"content": "Go is a statically typed, compiled language."
			},
			{
				"title": "",
				"url": "https://example.com/empty-title"
			},
			{
				"title": "No URL",
				"url": ""
			}
		]
	}`)

	results, err := parseDegoogResults(data, 10)
	if err != nil {
		t.Fatalf("parseDegoogResults error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results (empty title/url skipped), got %d", len(results))
	}

	r0 := results[0]
	if r0.Title != "The Go Programming Language" || r0.URL != "https://go.dev" {
		t.Errorf("unexpected first result: %+v", r0)
	}
	if r0.Description != "Go is an open source programming language." {
		t.Errorf("expected snippet as description, got %q", r0.Description)
	}
	if r0.Source != "" {
		t.Errorf("Source must be left empty, got %q", r0.Source)
	}

	r1 := results[1]
	if r1.Description != "Go is a statically typed, compiled language." {
		t.Errorf("expected content fallback as description, got %q", r1.Description)
	}
}

func TestParseDegoogResultsCount(t *testing.T) {
	data := []byte(`{"results":[
		{"title":"a","url":"https://a.example","snippet":"s"},
		{"title":"b","url":"https://b.example","snippet":"s"},
		{"title":"c","url":"https://c.example","snippet":"s"}
	]}`)

	results, err := parseDegoogResults(data, 2)
	if err != nil {
		t.Fatalf("parseDegoogResults error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected count to cap results at 2, got %d", len(results))
	}
}

func TestParseDegoogResultsEmpty(t *testing.T) {
	results, err := parseDegoogResults([]byte(`{"results":[]}`), 10)
	if err != nil {
		t.Fatalf("parseDegoogResults error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}
