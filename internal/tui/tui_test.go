package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"rune/internal/store"
)

func TestModel_EnterSavesEntry(t *testing.T) {
	s := store.NewStore(t.TempDir())
	m := initialModel(s)
	m.loaded = true

	m.input.SetValue("Fixed the login bug")

	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		// If a command is returned, run it to process side effects
		msg := cmd()
		newM, _ = newM.Update(msg)
	}

	// Check entry was saved to store
	entries, err := s.ReadDay(m.date)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(entries))
	}
	if entries[0].Body != "Fixed the login bug" {
		t.Errorf("body = %q, want %q", entries[0].Body, "Fixed the login bug")
	}

	// Check input was cleared
	updated := newM.(model)
	if updated.input.Value() != "" {
		t.Errorf("input not cleared, got %q", updated.input.Value())
	}

	// Check entries list in model is updated
	if len(updated.entries) != 1 {
		t.Errorf("model has %d entries, want 1", len(updated.entries))
	}
}

func TestModel_CtrlWDeletesLastWord(t *testing.T) {
	s := store.NewStore(t.TempDir())
	m := initialModel(s)
	m.loaded = true

	m.input.SetValue("hello world")

	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlW})
	updated := newM.(model)
	if updated.input.Value() != "hello " {
		t.Errorf("after Ctrl+W got %q, want %q", updated.input.Value(), "hello ")
	}
}

func TestModel_CtrlUClearsLine(t *testing.T) {
	s := store.NewStore(t.TempDir())
	m := initialModel(s)
	m.loaded = true

	m.input.SetValue("hello world")

	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
	updated := newM.(model)
	if updated.input.Value() != "" {
		t.Errorf("after Ctrl+U got %q, want %q", updated.input.Value(), "")
	}
}

func TestModel_EscClearsInput(t *testing.T) {
	s := store.NewStore(t.TempDir())
	m := initialModel(s)
	m.loaded = true

	m.input.SetValue("hello world")

	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	updated := newM.(model)
	if updated.input.Value() != "" {
		t.Errorf("after Esc got %q, want %q", updated.input.Value(), "")
	}
}

func TestModel_TypeTriggersDraftSave(t *testing.T) {
	s := store.NewStore(t.TempDir())
	m := initialModel(s)
	m.loaded = true

	m.input.SetValue("Working on login bug")

	// Simulate the debounce timer firing
	nextM, cmd := m.Update(autoSaveMsg{})
	if cmd != nil {
		msg := cmd()
		nextM, _ = nextM.Update(msg)
	}
	m = nextM.(model)

	draft, err := s.ReadDraft(m.date)
	if err != nil {
		t.Fatal(err)
	}
	if draft != "Working on login bug" {
		t.Errorf("draft = %q, want %q", draft, "Working on login bug")
	}

	// Typing more should update the draft
	m.input.SetValue("Working on login bug and tests")
	nextM, cmd = m.Update(autoSaveMsg{})
	if cmd != nil {
		msg := cmd()
		nextM, _ = nextM.Update(msg)
	}
	m = nextM.(model)

	draft, err = s.ReadDraft(m.date)
	if err != nil {
		t.Fatal(err)
	}
	if draft != "Working on login bug and tests" {
		t.Errorf("draft = %q, want %q", draft, "Working on login bug and tests")
	}
}

func TestModel_EnterClearsDraft(t *testing.T) {
	s := store.NewStore(t.TempDir())
	m := initialModel(s)
	m.loaded = true

	// Save a draft via auto-save
	m.input.SetValue("Draft text")
	nextM, cmd := m.Update(autoSaveMsg{})
	if cmd != nil {
		msg := cmd()
		nextM, _ = nextM.Update(msg)
	}
	m = nextM.(model)

	// Press Enter to save the entry
	m.input.SetValue("Draft text")
	nextM, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		msg := cmd()
		nextM, _ = nextM.Update(msg)
	}
	updated := nextM.(model)

	// Draft should be cleared
	draft, err := s.ReadDraft(m.date)
	if err != nil {
		t.Fatal(err)
	}
	if draft != "" {
		t.Errorf("draft should be empty after Enter, got %q", draft)
	}

	if len(updated.entries) != 1 {
		t.Errorf("model has %d entries, want 1", len(updated.entries))
	}
}

func TestModel_DraftLoadedOnInit(t *testing.T) {
	s := store.NewStore(t.TempDir())
	// Create once to get the date, save a draft
	m1 := initialModel(s)
	s.SaveDraft(m1.date, "Draft from previous session")

	// Re-create model — should load draft
	m2 := initialModel(s)
	if m2.input.Value() != "Draft from previous session" {
		t.Errorf("input = %q, want %q", m2.input.Value(), "Draft from previous session")
	}
}

func TestModel_InputAreaRendered(t *testing.T) {
	s := store.NewStore(t.TempDir())
	m := initialModel(s)

	m.loaded = true

	v := m.View()
	if !strings.Contains(v, "> ") {
		t.Errorf("view should contain input prompt '> ', got:\n%s", v)
	}
}

func TestModel_EmptyState(t *testing.T) {
	s := store.NewStore(t.TempDir())
	m := initialModel(s)

	m.loaded = true
	m.entries = nil

	v := m.View()
	if !strings.Contains(v, "Type below") {
		t.Errorf("empty state should contain 'Type below', got:\n%s", v)
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
