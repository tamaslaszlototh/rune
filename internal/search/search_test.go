package search

import (
	"testing"
	"time"

	"rune/internal/store"
)

func TestSearch(t *testing.T) {
	entries := []store.Entry{
		{
			Timestamp: time.Date(2025, 6, 8, 9, 15, 0, 0, time.Local),
			Project:   "project-a",
			Body:      "Morning standup prep",
			Branch:    "main",
		},
		{
			Timestamp: time.Date(2025, 6, 8, 14, 30, 0, 0, time.Local),
			Project:   "idea001",
			Body:      "Fixed rate limiting bug #api-gateway @pr/142",
			Branch:    "main",
		},
		{
			Timestamp: time.Date(2025, 6, 9, 10, 0, 0, 0, time.Local),
			Project:   "project-b",
			Body:      "Reviewed PR #42",
			Branch:    "dev",
		},
	}

	tests := []struct {
		name       string
		query      string
		wantBodies []string
	}{
		{
			name:       "match body substring",
			query:      "standup",
			wantBodies: []string{"Morning standup prep"},
		},
		{
			name:       "case insensitive",
			query:      "STANDUP",
			wantBodies: []string{"Morning standup prep"},
		},
		{
			name:       "match multiple entries",
			query:      "bug",
			wantBodies: []string{"Fixed rate limiting bug #api-gateway @pr/142"},
		},
		{
			name:       "match tag syntax",
			query:      "#api-gateway",
			wantBodies: []string{"Fixed rate limiting bug #api-gateway @pr/142"},
		},
		{
			name:       "match link syntax",
			query:      "@pr/142",
			wantBodies: []string{"Fixed rate limiting bug #api-gateway @pr/142"},
		},
		{
			name:       "no match returns empty",
			query:      "nonexistent",
			wantBodies: nil,
		},
		{
			name:       "empty query matches all",
			query:      "",
			wantBodies: []string{"Morning standup prep", "Fixed rate limiting bug #api-gateway @pr/142", "Reviewed PR #42"},
		},
		{
			name:       "partial word match",
			query:      "review",
			wantBodies: []string{"Reviewed PR #42"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := Search(entries, tt.query)
			if len(matches) != len(tt.wantBodies) {
				t.Fatalf("got %d matches, want %d", len(matches), len(tt.wantBodies))
			}
			for i, want := range tt.wantBodies {
				if matches[i].Entry.Body != want {
					t.Errorf("match[%d].Body = %q, want %q", i, matches[i].Entry.Body, want)
				}
			}
		})
	}
}

func TestSearch_PreservesEntry(t *testing.T) {
	entry := store.Entry{
		Timestamp: time.Date(2025, 6, 8, 14, 30, 0, 0, time.Local),
		Project:   "idea001",
		Body:      "Fixed rate limiting bug #api-gateway @pr/142",
		Branch:    "main",
		Tags:      []string{"api-gateway"},
		Links:     []string{"pr/142"},
	}
	matches := Search([]store.Entry{entry}, "bug")
	if len(matches) != 1 {
		t.Fatalf("got %d matches, want 1", len(matches))
	}
	got := matches[0].Entry
	if !got.Timestamp.Equal(entry.Timestamp) {
		t.Errorf("Timestamp = %v, want %v", got.Timestamp, entry.Timestamp)
	}
	if got.Project != entry.Project {
		t.Errorf("Project = %q, want %q", got.Project, entry.Project)
	}
	if got.Body != entry.Body {
		t.Errorf("Body = %q, want %q", got.Body, entry.Body)
	}
	if got.Branch != entry.Branch {
		t.Errorf("Branch = %q, want %q", got.Branch, entry.Branch)
	}
	if len(got.Tags) != 1 || got.Tags[0] != "api-gateway" {
		t.Errorf("Tags = %v, want [api-gateway]", got.Tags)
	}
	if len(got.Links) != 1 || got.Links[0] != "pr/142" {
		t.Errorf("Links = %v, want [pr/142]", got.Links)
	}
}

func TestSearch_EmptyEntries(t *testing.T) {
	matches := Search(nil, "anything")
	if len(matches) != 0 {
		t.Errorf("got %d matches, want 0", len(matches))
	}
}

func TestFilterByProject(t *testing.T) {
	entries := []store.Entry{
		{Project: "project-a", Body: "First"},
		{Project: "project-b", Body: "Second"},
		{Project: "project-a", Body: "Third"},
	}

	tests := []struct {
		name      string
		project   string
		wantCount int
	}{
		{"filter by project-a", "project-a", 2},
		{"filter by project-b", "project-b", 1},
		{"filter by nonexistent", "other", 0},
		{"empty filter returns all", "", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := FilterByProject(entries, tt.project)
			if len(filtered) != tt.wantCount {
				t.Errorf("got %d entries, want %d", len(filtered), tt.wantCount)
			}
		})
	}
}
