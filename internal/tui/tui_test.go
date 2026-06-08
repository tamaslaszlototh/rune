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

func TestModel_FilterNarrowsEntries(t *testing.T) {
	s := store.NewStore(t.TempDir())
	m := initialModel(s)
	m.loaded = true
	m.entries = []store.Entry{
		{Project: "proj-a", Body: "entry from a", Timestamp: time.Now()},
		{Project: "proj-b", Body: "entry from b", Timestamp: time.Now()},
	}
	m.projects = []string{"proj-a", "proj-b"}
	m.filterIndex = -1

	// Show all
	v := m.View()
	if !strings.Contains(v, "entry from a") {
		t.Errorf("All filter should show entry from a")
	}
	if !strings.Contains(v, "entry from b") {
		t.Errorf("All filter should show entry from b")
	}

	// Filter to proj-a
	m.filterIndex = 0
	v = m.View()
	if !strings.Contains(v, "entry from a") {
		t.Errorf("proj-a filter should show entry from a")
	}
	if strings.Contains(v, "entry from b") {
		t.Errorf("proj-a filter should NOT show entry from b")
	}

	// Filter to proj-b
	m.filterIndex = 1
	v = m.View()
	if strings.Contains(v, "entry from a") {
		t.Errorf("proj-b filter should NOT show entry from a")
	}
	if !strings.Contains(v, "entry from b") {
		t.Errorf("proj-b filter should show entry from b")
	}
}

func TestModel_FilteredEmptyState(t *testing.T) {
	s := store.NewStore(t.TempDir())
	m := initialModel(s)
	m.loaded = true
	m.entries = []store.Entry{
		{Project: "proj-a", Body: "entry from a", Timestamp: time.Now()},
	}
	m.projects = []string{"proj-a", "proj-b"}
	m.filterIndex = 1 // proj-b has no entries

	v := m.View()
	if !strings.Contains(v, "No entries for proj-b") {
		t.Errorf("filtered empty state should mention project, got:\n%s", v)
	}
}

func TestModel_FilterBarShowsProjectPills(t *testing.T) {
	s := store.NewStore(t.TempDir())
	m := initialModel(s)
	m.loaded = true
	m.entries = []store.Entry{
		{Project: "proj-a", Body: "a", Timestamp: time.Now()},
		{Project: "proj-b", Body: "b", Timestamp: time.Now()},
	}
	m.projects = []string{"proj-a", "proj-b"}
	m.filterIndex = -1

	v := m.View()
	if !strings.Contains(v, "proj-a") {
		t.Errorf("view should contain project pill 'proj-a', got:\n%s", v)
	}
	if !strings.Contains(v, "proj-b") {
		t.Errorf("view should contain project pill 'proj-b', got:\n%s", v)
	}
	if !strings.Contains(v, "All") {
		t.Errorf("view should contain 'All' pill, got:\n%s", v)
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

func TestModel_ShiftTabCyclesBackward(t *testing.T) {
	s := store.NewStore(t.TempDir())
	m := initialModel(s)
	m.loaded = true
	m.projects = []string{"proj-a", "proj-b"}
	m.filterIndex = -1

	check := func(presses, want int, desc string) {
		t.Helper()
		cur := model(m)
		for i := 0; i < presses; i++ {
			var cmd tea.Cmd
			var next tea.Model
			next, cmd = cur.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
			if cmd != nil {
				cmd()
			}
			cur = next.(model)
		}
		if cur.filterIndex != want {
			t.Errorf("%s: filterIndex = %d, want %d", desc, cur.filterIndex, want)
		}
	}

	check(1, 1, "first Shift+Tab goes to last project (proj-b)")
	check(2, 0, "second Shift+Tab goes to proj-a")
	check(3, -1, "third Shift+Tab wraps to All")
}

func TestModel_TabCyclesFilterForward(t *testing.T) {
	s := store.NewStore(t.TempDir())
	m := initialModel(s)
	m.loaded = true
	m.projects = []string{"proj-a", "proj-b"}
	m.filterIndex = -1

	check := func(presses, want int, desc string) {
		t.Helper()
		cur := model(m)
		for i := 0; i < presses; i++ {
			var cmd tea.Cmd
			var next tea.Model
			next, cmd = cur.Update(tea.KeyMsg{Type: tea.KeyTab})
			if cmd != nil {
				cmd()
			}
			cur = next.(model)
		}
		if cur.filterIndex != want {
			t.Errorf("%s: filterIndex = %d, want %d", desc, cur.filterIndex, want)
		}
	}

	check(1, 0, "first Tab goes to proj-a")
	check(2, 1, "second Tab goes to proj-b")
	check(3, -1, "third Tab wraps to All")
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
