package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseEntryLine(t *testing.T) {
	date := time.Date(2025, 6, 8, 0, 0, 0, 0, time.Local)
	line := "- [@14:30] [idea001] Fixed rate limiting bug #api-gateway @pr/142 (branch: main)"

	entry, err := ParseEntryLine(date, line)
	if err != nil {
		t.Fatal(err)
	}

	expected := time.Date(2025, 6, 8, 14, 30, 0, 0, time.Local)
	if !entry.Timestamp.Equal(expected) {
		t.Errorf("Timestamp = %v, want %v", entry.Timestamp, expected)
	}

	if entry.Project != "idea001" {
		t.Errorf("Project = %q, want %q", entry.Project, "idea001")
	}

	if entry.Body != "Fixed rate limiting bug #api-gateway @pr/142" {
		t.Errorf("Body = %q, want %q", entry.Body, "Fixed rate limiting bug #api-gateway @pr/142")
	}

	if entry.Branch != "main" {
		t.Errorf("Branch = %q, want %q", entry.Branch, "main")
	}

	if len(entry.Tags) != 1 || entry.Tags[0] != "api-gateway" {
		t.Errorf("Tags = %v, want [api-gateway]", entry.Tags)
	}

	if len(entry.Links) != 1 || entry.Links[0] != "pr/142" {
		t.Errorf("Links = %v, want [pr/142]", entry.Links)
	}
}

func TestReadDay(t *testing.T) {
	dir := t.TempDir()
	date := time.Date(2025, 6, 8, 0, 0, 0, 0, time.Local)
	content := `- [@09:15] [project-a] Morning standup prep (branch: main)
- [@14:30] [idea001] Fixed rate limiting bug #api-gateway @pr/142 (branch: main)
`
	entriesDir := filepath.Join(dir, "entries")
	os.MkdirAll(entriesDir, 0755)
	filePath := filepath.Join(entriesDir, "2025-06-08.md")
	os.WriteFile(filePath, []byte(content), 0644)

	s := NewStore(dir)
	entries, err := s.ReadDay(date)
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(entries))
	}

	e1 := entries[0]
	expected1 := time.Date(2025, 6, 8, 9, 15, 0, 0, time.Local)
	if !e1.Timestamp.Equal(expected1) {
		t.Errorf("entry[0] Timestamp = %v, want %v", e1.Timestamp, expected1)
	}
	if e1.Project != "project-a" {
		t.Errorf("entry[0] Project = %q, want %q", e1.Project, "project-a")
	}
	if e1.Body != "Morning standup prep" {
		t.Errorf("entry[0] Body = %q, want %q", e1.Body, "Morning standup prep")
	}
	if e1.Branch != "main" {
		t.Errorf("entry[0] Branch = %q, want %q", e1.Branch, "main")
	}

	e2 := entries[1]
	expected2 := time.Date(2025, 6, 8, 14, 30, 0, 0, time.Local)
	if !e2.Timestamp.Equal(expected2) {
		t.Errorf("entry[1] Timestamp = %v, want %v", e2.Timestamp, expected2)
	}
	if e2.Project != "idea001" {
		t.Errorf("entry[1] Project = %q, want %q", e2.Project, "idea001")
	}
	if e2.Body != "Fixed rate limiting bug #api-gateway @pr/142" {
		t.Errorf("entry[1] Body = %q, want %q", e2.Body, "Fixed rate limiting bug #api-gateway @pr/142")
	}
	if e2.Branch != "main" {
		t.Errorf("entry[1] Branch = %q, want %q", e2.Branch, "main")
	}
	if len(e2.Tags) != 1 || e2.Tags[0] != "api-gateway" {
		t.Errorf("entry[1] Tags = %v, want [api-gateway]", e2.Tags)
	}
	if len(e2.Links) != 1 || e2.Links[0] != "pr/142" {
		t.Errorf("entry[1] Links = %v, want [pr/142]", e2.Links)
	}
}

func TestReadDay_MissingDate(t *testing.T) {
	dir := t.TempDir()
	date := time.Date(2025, 6, 9, 0, 0, 0, 0, time.Local)

	s := NewStore(dir)
	entries, err := s.ReadDay(date)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("got %d entries, want 0", len(entries))
	}
}

func TestAppendEntry(t *testing.T) {
	dir := t.TempDir()
	date := time.Date(2025, 6, 8, 0, 0, 0, 0, time.Local)

	s := NewStore(dir)

	e1 := Entry{
		Timestamp: time.Date(2025, 6, 8, 9, 15, 0, 0, time.Local),
		Project:   "project-a",
		Body:      "Morning standup prep",
		Branch:    "main",
	}
	if err := s.AppendEntry(date, e1); err != nil {
		t.Fatal(err)
	}

	e2 := Entry{
		Timestamp: time.Date(2025, 6, 8, 14, 30, 0, 0, time.Local),
		Project:   "idea001",
		Body:      "Fixed rate limiting bug #api-gateway @pr/142",
		Tags:      []string{"api-gateway"},
		Links:     []string{"pr/142"},
		Branch:    "main",
	}
	if err := s.AppendEntry(date, e2); err != nil {
		t.Fatal(err)
	}

	entries, err := s.ReadDay(date)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(entries))
	}

	if entries[0].Project != "project-a" || entries[0].Body != "Morning standup prep" {
		t.Errorf("first entry mismatch: %+v", entries[0])
	}
	if entries[1].Project != "idea001" || entries[1].Body != "Fixed rate limiting bug #api-gateway @pr/142" {
		t.Errorf("second entry mismatch: %+v", entries[1])
	}
}

func TestReadRange(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)

	date1 := time.Date(2025, 6, 8, 0, 0, 0, 0, time.Local)
	date2 := time.Date(2025, 6, 9, 0, 0, 0, 0, time.Local)
	date3 := time.Date(2025, 6, 10, 0, 0, 0, 0, time.Local)

	s.AppendEntry(date1, Entry{
		Timestamp: time.Date(2025, 6, 8, 9, 0, 0, 0, time.Local),
		Project:   "project-a", Body: "Entry day 1", Branch: "main",
	})
	s.AppendEntry(date2, Entry{
		Timestamp: time.Date(2025, 6, 9, 10, 0, 0, 0, time.Local),
		Project:   "project-b", Body: "Entry day 2", Branch: "dev",
	})
	s.AppendEntry(date3, Entry{
		Timestamp: time.Date(2025, 6, 10, 11, 0, 0, 0, time.Local),
		Project:   "project-a", Body: "Entry day 3", Branch: "main",
	})

	entries, err := s.ReadRange(date1, date3)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("got %d entries, want 3", len(entries))
	}

	if entries[0].Body != "Entry day 1" {
		t.Errorf("entry[0].Body = %q", entries[0].Body)
	}
	if entries[1].Body != "Entry day 2" {
		t.Errorf("entry[1].Body = %q", entries[1].Body)
	}
	if entries[2].Body != "Entry day 3" {
		t.Errorf("entry[2].Body = %q", entries[2].Body)
	}
}

func TestSaveAndReadDraft(t *testing.T) {
	dir := t.TempDir()
	date := time.Date(2025, 6, 8, 0, 0, 0, 0, time.Local)
	s := NewStore(dir)

	// Save draft
	err := s.SaveDraft(date, "Working on the login bug")
	if err != nil {
		t.Fatal(err)
	}

	// Read draft back
	draft, err := s.ReadDraft(date)
	if err != nil {
		t.Fatal(err)
	}
	if draft != "Working on the login bug" {
		t.Errorf("got %q, want %q", draft, "Working on the login bug")
	}
}

func TestClearDraft(t *testing.T) {
	dir := t.TempDir()
	date := time.Date(2025, 6, 8, 0, 0, 0, 0, time.Local)
	s := NewStore(dir)

	s.SaveDraft(date, "draft text")
	if err := s.ClearDraft(date); err != nil {
		t.Fatal(err)
	}

	draft, err := s.ReadDraft(date)
	if err != nil {
		t.Fatal(err)
	}
	if draft != "" {
		t.Errorf("draft should be empty after clear, got %q", draft)
	}
}

func TestReadDraft_NoFile(t *testing.T) {
	dir := t.TempDir()
	date := time.Date(2025, 6, 9, 0, 0, 0, 0, time.Local)
	s := NewStore(dir)

	draft, err := s.ReadDraft(date)
	if err != nil {
		t.Fatal(err)
	}
	if draft != "" {
		t.Errorf("expected empty draft for missing file, got %q", draft)
	}
}

func TestReadRange_Partial(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)

	date1 := time.Date(2025, 6, 8, 0, 0, 0, 0, time.Local)
	date2 := time.Date(2025, 6, 9, 0, 0, 0, 0, time.Local)
	date3 := time.Date(2025, 6, 10, 0, 0, 0, 0, time.Local)

	s.AppendEntry(date1, Entry{
		Timestamp: time.Date(2025, 6, 8, 9, 0, 0, 0, time.Local),
		Project:   "project-a", Body: "Entry day 1", Branch: "main",
	})
	s.AppendEntry(date2, Entry{
		Timestamp: time.Date(2025, 6, 9, 10, 0, 0, 0, time.Local),
		Project:   "project-b", Body: "Entry day 2", Branch: "dev",
	})
	s.AppendEntry(date3, Entry{
		Timestamp: time.Date(2025, 6, 10, 11, 0, 0, 0, time.Local),
		Project:   "project-a", Body: "Entry day 3", Branch: "main",
	})

	entries, err := s.ReadRange(date2, date2)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(entries))
	}
	if entries[0].Body != "Entry day 2" {
		t.Errorf("entry[0].Body = %q", entries[0].Body)
	}
}
