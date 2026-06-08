package standup

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"rune/internal/store"
)

func TestFormatStandup_EmptyEntries(t *testing.T) {
	out := FormatStandup(nil, time.Now())
	if out != "" {
		t.Errorf("got %q, want empty string", out)
	}
}

func TestFormatStandup_EntriesOlderThanSince(t *testing.T) {
	since := time.Date(2025, 6, 8, 12, 0, 0, 0, time.Local)
	entries := []store.Entry{
		{
			Timestamp: time.Date(2025, 6, 8, 9, 0, 0, 0, time.Local),
			Project:   "project-a",
			Body:      "Morning entry",
			Branch:    "main",
		},
	}
	out := FormatStandup(entries, since)
	if out != "" {
		t.Errorf("expected empty output for entries before since, got %q", out)
	}
}

func TestFormatStandup_SingleProject(t *testing.T) {
	since := time.Date(2025, 6, 8, 0, 0, 0, 0, time.Local)
	entries := []store.Entry{
		{
			Timestamp: time.Date(2025, 6, 8, 14, 30, 0, 0, time.Local),
			Project:   "idea001",
			Body:      "Fixed rate limiting bug (#api-gateway @pr/142)",
			Branch:    "main",
		},
		{
			Timestamp: time.Date(2025, 6, 8, 9, 15, 0, 0, time.Local),
			Project:   "idea001",
			Body:      "Morning standup prep",
			Branch:    "main",
		},
	}

	out := FormatStandup(entries, since)

	// Exact header
	expectedHeader := fmt.Sprintf("## Standup — %s", time.Now().Format("2006-01-02"))
	if !strings.HasPrefix(out, expectedHeader) {
		t.Errorf("output should start with %q, got:\n%s", expectedHeader, out)
	}

	// Should have project header with branch
	if !strings.Contains(out, "### idea001 (main)") {
		t.Errorf("expected '### idea001 (main)' in output:\n%s", out)
	}

	// Should have both entry bodies as bullets
	if !strings.Contains(out, "• Fixed rate limiting bug (#api-gateway @pr/142)") {
		t.Errorf("expected entry body in output:\n%s", out)
	}
	if !strings.Contains(out, "• Morning standup prep") {
		t.Errorf("expected entry body in output:\n%s", out)
	}

	// Entries should be in chronological order (earliest first)
	body1 := "• Morning standup prep"
	body2 := "• Fixed rate limiting bug (#api-gateway @pr/142)"
	if strings.Index(out, body1) > strings.Index(out, body2) {
		t.Errorf("entries not in chronological order:\n%s", out)
	}
}

func TestFormatStandup_MultiProject(t *testing.T) {
	since := time.Date(2025, 6, 8, 0, 0, 0, 0, time.Local)
	entries := []store.Entry{
		{
			Timestamp: time.Date(2025, 6, 8, 14, 30, 0, 0, time.Local),
			Project:   "idea001",
			Body:      "Fixed rate limiting bug (#api-gateway @pr/142)",
			Branch:    "main",
		},
		{
			Timestamp: time.Date(2025, 6, 8, 9, 15, 0, 0, time.Local),
			Project:   "project-a",
			Body:      "Morning standup prep",
			Branch:    "main",
		},
		{
			Timestamp: time.Date(2025, 6, 8, 11, 0, 0, 0, time.Local),
			Project:   "project-a",
			Body:      "Reviewed PR #42",
			Branch:    "feature/x",
		},
	}

	out := FormatStandup(entries, since)

	if out == "" {
		t.Fatal("expected non-empty output")
	}

	// Both project headers should be present
	if !strings.Contains(out, "### project-a") {
		t.Errorf("expected '### project-a' in output:\n%s", out)
	}
	if !strings.Contains(out, "### idea001") {
		t.Errorf("expected '### idea001' in output:\n%s", out)
	}

	// project-a uses branch from first entry (main)
	if !strings.Contains(out, "### project-a (main)") {
		t.Errorf("expected '### project-a (main)' in output:\n%s", out)
	}

	// All entry bodies present
	for _, e := range entries {
		if !strings.Contains(out, "• "+e.Body) {
			t.Errorf("expected body %q in output:\n%s", "• "+e.Body, out)
		}
	}
}

func TestFormatStandup_NoBranch(t *testing.T) {
	since := time.Date(2025, 6, 8, 0, 0, 0, 0, time.Local)
	entries := []store.Entry{
		{
			Timestamp: time.Date(2025, 6, 8, 10, 0, 0, 0, time.Local),
			Project:   "",
			Body:      "Worked on something",
			Branch:    "",
		},
	}

	out := FormatStandup(entries, since)

	if !strings.Contains(out, "### general") {
		t.Errorf("expected '### general' for empty project, got:\n%s", out)
	}
	if strings.Contains(out, "(") {
		t.Errorf("expected no branch suffix for empty branch, got:\n%s", out)
	}
}
