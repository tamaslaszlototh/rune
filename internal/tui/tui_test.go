package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"rune/internal/config"
	"rune/internal/store"
)

func defaultConfig() *config.Config {
	return &config.Config{Projects: map[string]string{}}
}

func TestModel_EnterSavesEntry(t *testing.T) {
	s := newMemStore()
	m := 	initialModel(s, &fakeGit{}, defaultConfig(), "")
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
	s := newMemStore()
	m := 	initialModel(s, &fakeGit{}, defaultConfig(), "")
	m.loaded = true

	m.input.SetValue("hello world")

	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlW})
	updated := newM.(model)
	if updated.input.Value() != "hello " {
		t.Errorf("after Ctrl+W got %q, want %q", updated.input.Value(), "hello ")
	}
}

func TestModel_CtrlUClearsLine(t *testing.T) {
	s := newMemStore()
	m := 	initialModel(s, &fakeGit{}, defaultConfig(), "")
	m.loaded = true

	m.input.SetValue("hello world")

	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
	updated := newM.(model)
	if updated.input.Value() != "" {
		t.Errorf("after Ctrl+U got %q, want %q", updated.input.Value(), "")
	}
}

func TestModel_EscClearsInput(t *testing.T) {
	s := newMemStore()
	m := 	initialModel(s, &fakeGit{}, defaultConfig(), "")
	m.loaded = true

	m.input.SetValue("hello world")

	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	updated := newM.(model)
	if updated.input.Value() != "" {
		t.Errorf("after Esc got %q, want %q", updated.input.Value(), "")
	}
}

func TestModel_TypeTriggersDraftSave(t *testing.T) {
	s := newMemStore()
	m := 	initialModel(s, &fakeGit{}, defaultConfig(), "")
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
	s := newMemStore()
	m := 	initialModel(s, &fakeGit{}, defaultConfig(), "")
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
	s := newMemStore()
	// Create once to get the date, save a draft
	m1 := 	initialModel(s, &fakeGit{}, defaultConfig(), "")
	s.SaveDraft(m1.date, "Draft from previous session")

	// Re-create model — should load draft
	m2 := 	initialModel(s, &fakeGit{}, defaultConfig(), "")
	if m2.input.Value() != "Draft from previous session" {
		t.Errorf("input = %q, want %q", m2.input.Value(), "Draft from previous session")
	}
}

func TestModel_InputAreaRendered(t *testing.T) {
	s := newMemStore()
	m := 	initialModel(s, &fakeGit{}, defaultConfig(), "")

	m.loaded = true

	v := m.View()
	if !strings.Contains(v, "> ") {
		t.Errorf("view should contain input prompt '> ', got:\n%s", v)
	}
}

func TestModel_FilterNarrowsEntries(t *testing.T) {
	s := newMemStore()
	m := 	initialModel(s, &fakeGit{}, defaultConfig(), "")
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
	s := newMemStore()
	m := 	initialModel(s, &fakeGit{}, defaultConfig(), "")
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
	s := newMemStore()
	m := 	initialModel(s, &fakeGit{}, defaultConfig(), "")
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
	s := newMemStore()
	m := 	initialModel(s, &fakeGit{}, defaultConfig(), "")

	m.loaded = true
	m.entries = nil

	v := m.View()
	if !strings.Contains(v, "Type below") {
		t.Errorf("empty state should contain 'Type below', got:\n%s", v)
	}
}

func TestModel_InitialFilterProjectView(t *testing.T) {
	s := newMemStore()
	m := initialModel(s, &fakeGit{}, defaultConfig(), "proj-b")
	m.loaded = true
	m.entries = []store.Entry{
		{Project: "proj-a", Body: "entry from a", Timestamp: time.Now()},
		{Project: "proj-b", Body: "entry from b", Timestamp: time.Now()},
	}

	newM, _ := m.Update(entriesLoaded{entries: m.entries})
	updated := newM.(model)

	v := updated.View()
	if !strings.Contains(v, "entry from b") {
		t.Errorf("view should show entry from proj-b, got:\n%s", v)
	}
	if strings.Contains(v, "entry from a") {
		t.Errorf("view should NOT show entry from proj-a when filtered to proj-b, got:\n%s", v)
	}
}

func TestModel_InitialFilterProject_NonExistent(t *testing.T) {
	s := newMemStore()
	m := initialModel(s, &fakeGit{}, defaultConfig(), "nonexistent")
	m.loaded = true
	m.entries = []store.Entry{
		{Project: "proj-a", Body: "entry from a", Timestamp: time.Now()},
		{Project: "proj-b", Body: "entry from b", Timestamp: time.Now()},
	}

	newM, _ := m.Update(entriesLoaded{entries: m.entries})
	updated := newM.(model)

	if updated.filterIndex != -1 {
		t.Errorf("filterIndex = %d, want -1 (All) for non-existent project", updated.filterIndex)
	}
}

func TestModel_InitialFilterProject(t *testing.T) {
	s := newMemStore()
	m := initialModel(s, &fakeGit{}, defaultConfig(), "proj-b")
	m.loaded = true
	m.entries = []store.Entry{
		{Project: "proj-a", Body: "entry from a", Timestamp: time.Now()},
		{Project: "proj-b", Body: "entry from b", Timestamp: time.Now()},
	}

	newM, _ := m.Update(entriesLoaded{entries: m.entries})
	updated := newM.(model)

	if updated.filterIndex != 1 {
		t.Errorf("filterIndex = %d, want 1 (proj-b)", updated.filterIndex)
	}
}

func TestModel_ShiftTabCyclesBackward(t *testing.T) {
	s := newMemStore()
	m := 	initialModel(s, &fakeGit{}, defaultConfig(), "")
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
	s := newMemStore()
	m := 	initialModel(s, &fakeGit{}, defaultConfig(), "")
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

func TestModel_SlashKeyEntersSearchMode(t *testing.T) {
	s := newMemStore()
	m := 	initialModel(s, &fakeGit{}, defaultConfig(), "")
	m.loaded = true
	m.input.SetValue("some draft text")

	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	updated := newM.(model)

	if !updated.searching {
		t.Fatal("expected searching to be true after pressing /")
	}
	if updated.input.Value() != "" {
		t.Errorf("search input should be empty, got %q", updated.input.Value())
	}
	if updated.savedDraft != "some draft text" {
		t.Errorf("savedDraft = %q, want %q", updated.savedDraft, "some draft text")
	}
	if updated.input.Prompt != "[/] " {
		t.Errorf("prompt = %q, want %q", updated.input.Prompt, "[/] ")
	}
}

func TestModel_SearchFiltersEntries(t *testing.T) {
	s := newMemStore()
	m := 	initialModel(s, &fakeGit{}, defaultConfig(), "")
	m.loaded = true
	m.entries = []store.Entry{
		{Project: "proj-a", Body: "Fixed the login bug", Timestamp: time.Now()},
		{Project: "proj-b", Body: "Added logout feature", Timestamp: time.Now()},
		{Project: "proj-a", Body: "Updated login page styles", Timestamp: time.Now()},
	}
	m.projects = []string{"proj-a", "proj-b"}
	m.filterIndex = -1

	// Enter search mode
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	updated := newM.(model)

	// Type search query
	for _, ch := range "login" {
		var cmd tea.Cmd
		next, cmd := updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		if cmd != nil {
			cmd()
		}
		updated = next.(model)
	}

	v := updated.View()
	if !strings.Contains(v, "Fixed the login bug") {
		t.Errorf("view should show entry matching 'login', got:\n%s", v)
	}
	if !strings.Contains(v, "Updated login page styles") {
		t.Errorf("view should show entry matching 'login', got:\n%s", v)
	}
	if strings.Contains(v, "Added logout feature") {
		t.Errorf("view should NOT show non-matching entry, got:\n%s", v)
	}
}

func TestModel_SearchEscExitsSearchMode(t *testing.T) {
	s := newMemStore()
	m := 	initialModel(s, &fakeGit{}, defaultConfig(), "")
	m.loaded = true
	m.input.SetValue("my draft")
	m.projects = []string{"proj-a", "proj-b"}
	m.filterIndex = 0

	// Enter search mode
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	updated := newM.(model)

	if !updated.searching {
		t.Fatal("should be in search mode after /")
	}

	// Exit search mode with Esc
	next, _ := updated.Update(tea.KeyMsg{Type: tea.KeyEscape})
	exited := next.(model)

	if exited.searching {
		t.Fatal("should not be searching after Esc")
	}
	if exited.input.Value() != "my draft" {
		t.Errorf("draft should be restored, got %q", exited.input.Value())
	}
	if exited.filterIndex != 0 {
		t.Errorf("filterIndex should be restored to 0, got %d", exited.filterIndex)
	}
	if !strings.HasPrefix(exited.input.Prompt, "> ") {
		t.Errorf("prompt should be restored to '> ', got %q", exited.input.Prompt)
	}
}

func TestModel_SearchHighlightMatches(t *testing.T) {
	s := newMemStore()
	m := 	initialModel(s, &fakeGit{}, defaultConfig(), "")
	m.loaded = true
	m.entries = []store.Entry{
		{Project: "proj-a", Body: "Fixed the login bug", Timestamp: time.Now()},
	}
	m.projects = []string{"proj-a"}
	m.filterIndex = -1

	// Enter search mode and type query
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	updated := newM.(model)
	for _, ch := range "login" {
		next, cmd := updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		if cmd != nil {
			cmd()
		}
		updated = next.(model)
	}

	v := updated.View()
	// The entry body should be present
	if !strings.Contains(v, "Fixed the login bug") {
		t.Errorf("view should contain the entry body, got:\n%s", v)
	}
}

func TestRenderBody(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "tag only", body: "Fixed bug #api-gateway"},
		{name: "link only", body: "Fixed bug @pr/142"},
		{name: "both", body: "Fixed #bug @pr/142"},
		{name: "no tags or links", body: "plain text"},
		{name: "multiple tags", body: "Fixed #bug #api-gateway"},
		{name: "multiple links", body: "Fixed @pr/142 @issue/42"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderBody(tt.body)
			if !strings.Contains(got, tt.body) {
				t.Errorf("renderBody(%q) should contain body text %q, got %q", tt.body, tt.body, got)
			}
		})
	}
}

func TestModel_TagsAndLinksRendered(t *testing.T) {
	s := newMemStore()
	m := 	initialModel(s, &fakeGit{}, defaultConfig(), "")
	m.loaded = true
	m.entries = []store.Entry{
		{
			Timestamp: time.Date(2025, 6, 8, 14, 30, 0, 0, time.Local),
			Project:   "idea001",
			Body:      "Fixed bug #api-gateway @pr/142",
			Branch:    "main",
		},
	}

	v := m.View()
	if !strings.Contains(v, "#api-gateway") {
		t.Errorf("view should contain #api-gateway, got:\n%s", v)
	}
	if !strings.Contains(v, "@pr/142") {
		t.Errorf("view should contain @pr/142, got:\n%s", v)
	}
}

func TestHighlightMatch(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		query    string
		wantCont string
		noMatch  string
	}{
		{name: "simple match", body: "Fixed the login bug", query: "login", wantCont: "login", noMatch: ""},
		{name: "no match", body: "Fixed the logout bug", query: "login", wantCont: "Fixed the logout bug", noMatch: ""},
		{name: "empty query", body: "some text", query: "", wantCont: "some text", noMatch: ""},
		{name: "case insensitive", body: "Login button", query: "login", wantCont: "Login", noMatch: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := highlightMatch(tt.body, tt.query)
			if !strings.Contains(got, tt.wantCont) {
				t.Errorf("highlightMatch(%q, %q) should contain %q, got %q", tt.body, tt.query, tt.wantCont, got)
			}
			if tt.noMatch != "" && strings.Contains(got, tt.noMatch) {
				t.Errorf("highlightMatch(%q, %q) should not contain %q, got %q", tt.body, tt.query, tt.noMatch, got)
			}
		})
	}
}

func TestModel_SearchEmptyResults(t *testing.T) {
	s := newMemStore()
	m := 	initialModel(s, &fakeGit{}, defaultConfig(), "")
	m.loaded = true
	m.entries = []store.Entry{
		{Project: "proj-a", Body: "Fixed the login bug", Timestamp: time.Now()},
		{Project: "proj-b", Body: "Added logout feature", Timestamp: time.Now()},
	}
	m.projects = []string{"proj-a", "proj-b"}
	m.filterIndex = -1

	// Enter search mode
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	updated := newM.(model)

	// Type a query that won't match
	for _, ch := range "zzzzz" {
		var cmd tea.Cmd
		next, cmd := updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		if cmd != nil {
			cmd()
		}
		updated = next.(model)
	}

	v := updated.View()
	if !strings.Contains(v, "No entries matching") {
		t.Errorf("should show 'No entries matching' for empty search results, got:\n%s", v)
	}
}

func TestModel_SearchEnterDoesNotSave(t *testing.T) {
	s := newMemStore()
	m := 	initialModel(s, &fakeGit{}, defaultConfig(), "")
	m.date = time.Date(2026, 6, 9, 10, 0, 0, 0, time.Local)
	// Persist an entry to the store first
	if err := s.AppendEntry(m.date, store.Entry{Body: "Existing entry", Project: "proj-a", Timestamp: m.date}); err != nil {
		t.Fatal(err)
	}
	m.loaded = true
	m.entries, _ = s.ReadDay(m.date)
	m.projects = []string{"proj-a"}
	m.filterIndex = -1

	// Enter search mode and type something
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	updated := newM.(model)
	for _, ch := range "search text" {
		next, cmd := updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		if cmd != nil {
			cmd()
		}
		updated = next.(model)
	}

	// Press Enter — should not save an entry
	next, cmd := updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		msg := cmd()
		next, _ = next.Update(msg)
	}
	afterEnter := next.(model)

	// Check no new entry was saved
	entries, err := s.ReadDay(m.date)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("got %d entries, want 1 (Enter should not save during search)", len(entries))
	}

	// Should still be in search mode
	if !afterEnter.searching {
		t.Error("should still be in search mode after Enter")
	}
}

func TestModel_CtrlSSavesDraft(t *testing.T) {
	s := newMemStore()
	m := initialModel(s, &fakeGit{}, defaultConfig(), "")
	m.loaded = true

	m.input.SetValue("Important note")

	// Send Ctrl+S — should save draft immediately, not via auto-save
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})

	draft, err := s.ReadDraft(m.date)
	if err != nil {
		t.Fatal(err)
	}
	if draft != "Important note" {
		t.Errorf("draft = %q, want %q", draft, "Important note")
	}

	updated := newM.(model)
	if updated.input.Value() != "Important note" {
		t.Errorf("input should not be cleared, got %q", updated.input.Value())
	}
	if updated.statusMsg != "Draft saved" {
		t.Errorf("statusMsg = %q, want %q", updated.statusMsg, "Draft saved")
	}
}

func TestModel_CtrlSInSearchModeSavesSearchQuery(t *testing.T) {
	s := newMemStore()
	m := initialModel(s, &fakeGit{}, defaultConfig(), "")
	m.loaded = true

	// Enter search mode with some saved draft text
	m.input.SetValue("original draft")
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	updated := newM.(model)

	// Type a search query
	for _, ch := range "login bug" {
		next, _ := updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		updated = next.(model)
	}

	// Ctrl+S should save the search query as draft
	next, _ := updated.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	afterSave := next.(model)

	draft, err := s.ReadDraft(m.date)
	if err != nil {
		t.Fatal(err)
	}
	if draft != "login bug" {
		t.Errorf("draft = %q, want %q", draft, "login bug")
	}

	// Should still be in search mode
	if !afterSave.searching {
		t.Error("should still be in search mode after Ctrl+S")
	}
	if afterSave.statusMsg != "Draft saved" {
		t.Errorf("statusMsg = %q, want %q", afterSave.statusMsg, "Draft saved")
	}
}

func TestModel_CtrlSStatusMsgClearsAfterTick(t *testing.T) {
	s := newMemStore()
	m := initialModel(s, &fakeGit{}, defaultConfig(), "")
	m.loaded = true
	m.input.SetValue("some text")

	// Ctrl+S returns a Tick cmd; execute it to send clearStatusMsg
	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if cmd == nil {
		t.Fatal("expected non-nil cmd (Tick)")
	}
	updated := newM.(model)
	if updated.statusMsg != "Draft saved" {
		t.Fatalf("statusMsg = %q, want %q before tick", updated.statusMsg, "Draft saved")
	}

	// Execute the Tick — it will block ~1.5s, then return clearStatusMsg
	msg := cmd()
	nextM, _ := updated.Update(msg)
	cleared := nextM.(model)
	if cleared.statusMsg != "" {
		t.Errorf("statusMsg should be cleared after tick, got %q", cleared.statusMsg)
	}
}

func TestModel_ShowsEntries(t *testing.T) {
	s := newMemStore()
	m := 	initialModel(s, &fakeGit{}, defaultConfig(), "")

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
