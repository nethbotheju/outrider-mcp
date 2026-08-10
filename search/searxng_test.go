package search

import (
	"testing"
)

func TestParseSearXNGResults(t *testing.T) {
	data := []byte(`{
		"results": [
			{
				"title": "Paper Title",
				"url": "https://example.com/paper",
				"content": "A short abstract.",
				"publishedDate": "2024-03-15T10:00:00+00:00",
				"source": "arXiv",
				"journal": "Nature",
				"doi": "10.1000/abc",
				"pdf_url": "https://example.com/paper.pdf",
				"authors": ["Alice Smith", "Bob Jones"]
			},
			{
				"title": "News Article",
				"url": "https://example.com/news",
				"content": "Breaking news.",
				"publishedDate": "2024-01-20",
				"source": "BBC",
				"author": "Carol Writer"
			}
		]
	}`)

	results, err := parseSearXNGResults(data, 10)
	if err != nil {
		t.Fatalf("parseSearXNGResults error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	r0 := results[0]
	if r0.Title != "Paper Title" || r0.URL != "https://example.com/paper" {
		t.Errorf("unexpected first result: %+v", r0)
	}
	if len(r0.Authors) != 2 || r0.Authors[0] != "Alice Smith" {
		t.Errorf("unexpected authors: %v", r0.Authors)
	}
	if r0.PDFURL != "https://example.com/paper.pdf" {
		t.Errorf("unexpected pdf url: %q", r0.PDFURL)
	}

	r1 := results[1]
	if len(r1.Authors) != 1 || r1.Authors[0] != "Carol Writer" {
		t.Errorf("unexpected author parsing: %v", r1.Authors)
	}
}

func TestFormatResultsRichMetadata(t *testing.T) {
	results := []SearchResult{
		{
			Title:         "Paper",
			URL:           "https://example.com/paper",
			Description:   "Abstract.",
			PublishedDate: "2024-03-15T10:00:00+00:00",
			Source:        "arXiv",
			Journal:       "Nature",
			DOI:           "10.1000/abc",
			PDFURL:        "https://example.com/paper.pdf",
			Authors:       []string{"Alice Smith"},
		},
	}
	formatted := FormatResults(results)
	for _, expected := range []string{"Published: 2024-03-15", "Source: arXiv", "Journal: Nature", "DOI: 10.1000/abc", "PDF:", "Authors: Alice Smith"} {
		if !contains(formatted, expected) {
			t.Errorf("expected formatted output to contain %q, got:\n%s", expected, formatted)
		}
	}
}

func contains(s, substr string) bool {
	return len(substr) <= len(s) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
