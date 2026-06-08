package tui

import (
	"strings"
	"testing"
	"time"

	"rune/internal/store"
)

func TestModel_EmptyState(t *testing.T) {
	s := store.NewStore(t.TempDir())
	m := initialModel(s)

	m.loaded = true
	m.entries = nil

	v := m.View()
	if !strings.Contains(v, "No entries yet") {
		t.Errorf("empty state should contain 'No entries yet', got:\n%s", v)
	}
}

func TestModel_ShowsEntries(t *testing.T) {
	s := store.NewStore(t.TempDir())
	m := initialModel(s)

	m.loaded = true
	m.entries = []store.Entry{
		{
			Timestamp: time.Date(2025, 6, 8, 14, 30, 0, 0, time.Local),
			Project:   "idea001",
			Body:      "Fixed rate limiting bug",
			Branch:    "main",
		},
	}

	v := m.View()
	if !strings.Contains(v, "idea001") {
		t.Errorf("view should contain project, got:\n%s", v)
	}
	if !strings.Contains(v, "Fixed rate limiting bug") {
		t.Errorf("view should contain entry body, got:\n%s", v)
	}
	if !strings.Contains(v, "14:30") {
		t.Errorf("view should contain time, got:\n%s", v)
	}
}
