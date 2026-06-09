package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCLI_UnknownFlag(t *testing.T) {
	binary := buildBinary(t)

	cmd := exec.Command(binary, "--help")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected exit error for unknown flag")
	}
	if string(out) != "Error: Usage: rune [-p <project>] [config|standup|search]\n" {
		t.Errorf("got %q, want %q", string(out), "Error: Usage: rune [-p <project>] [config|standup|search]\n")
	}
}

func TestCLI_Standup_Empty(t *testing.T) {
	binary, runeDir := buildBinaryWithDir(t)

	cmd := exec.Command(binary, "standup")
	cmd.Env = append(os.Environ(), "HOME="+runeDir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("standup failed: %v\n%s", err, out)
	}
	if strings.TrimSpace(string(out)) != "" {
		t.Errorf("expected empty output, got %q", string(out))
	}
}

func TestCLI_Standup_WithEntries(t *testing.T) {
	binary, runeDir := buildBinaryWithDir(t)

	today := time.Now().Format("2006-01-02")
	entryContent := "- [@09:15] [project-a] Morning standup prep (branch: main)\n"
	entriesDir := filepath.Join(runeDir, ".rune", "entries")
	os.MkdirAll(entriesDir, 0755)
	os.WriteFile(filepath.Join(entriesDir, today+".md"), []byte(entryContent), 0644)

	cmd := exec.Command(binary, "standup")
	cmd.Env = append(os.Environ(), "HOME="+runeDir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("standup failed: %v\n%s", err, out)
	}

	output := string(out)
	if !strings.Contains(output, "## Standup") {
		t.Errorf("expected standup header, got:\n%s", output)
	}
	if !strings.Contains(output, "### project-a") {
		t.Errorf("expected project header, got:\n%s", output)
	}
	if !strings.Contains(output, "Morning standup prep") {
		t.Errorf("expected entry body, got:\n%s", output)
	}
}

func TestCLI_Standup_SinceFlag_ExcludesOldEntries(t *testing.T) {
	binary, runeDir := buildBinaryWithDir(t)

	entriesDir := filepath.Join(runeDir, ".rune", "entries")
	os.MkdirAll(entriesDir, 0755)

	// Write entry for today
	today := time.Now().Format("2006-01-02")
	entryContent := "- [@10:00] [project-b] Recent entry from today (branch: main)\n"
	os.WriteFile(filepath.Join(entriesDir, today+".md"), []byte(entryContent), 0644)

	// Using --since set to tomorrow should exclude everything
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	cmd := exec.Command(binary, "standup", "--since", tomorrow)
	cmd.Env = append(os.Environ(), "HOME="+runeDir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("standup failed: %v\n%s", err, out)
	}
	if strings.TrimSpace(string(out)) != "" {
		t.Errorf("expected empty output when --since is in the future, got:\n%s", string(out))
	}
}

func TestCLI_Search_NoQuery(t *testing.T) {
	binary := buildBinary(t)

	cmd := exec.Command(binary, "search")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected exit error when no query given")
	}
	if string(out) != "Error: search query is required\n" {
		t.Errorf("got %q, want %q", string(out), "Error: search query is required\n")
	}
}

func TestCLI_Search_Matches(t *testing.T) {
	binary, runeDir := buildBinaryWithDir(t)

	entriesDir := filepath.Join(runeDir, ".rune", "entries")
	os.MkdirAll(entriesDir, 0755)
	os.WriteFile(filepath.Join(entriesDir, "2025-06-08.md"), []byte("- [@09:15] [project-a] Morning standup prep (branch: main)\n- [@14:30] [idea001] Fixed rate limiting bug #api-gateway @pr/142 (branch: main)\n"), 0644)
	os.WriteFile(filepath.Join(entriesDir, "2025-06-09.md"), []byte("- [@10:00] [project-b] Reviewed PR #42 (branch: dev)\n"), 0644)

	cmd := exec.Command(binary, "search", "standup")
	cmd.Env = append(os.Environ(), "HOME="+runeDir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("search failed: %v\n%s", err, out)
	}

	output := string(out)
	if !strings.Contains(output, "2025-06-08 09:15 [project-a] Morning standup prep") {
		t.Errorf("expected match line, got:\n%s", output)
	}
}

func TestCLI_Search_NoMatch(t *testing.T) {
	binary, runeDir := buildBinaryWithDir(t)

	entriesDir := filepath.Join(runeDir, ".rune", "entries")
	os.MkdirAll(entriesDir, 0755)
	os.WriteFile(filepath.Join(entriesDir, "2025-06-08.md"), []byte("- [@09:15] [project-a] Morning standup prep (branch: main)\n"), 0644)

	cmd := exec.Command(binary, "search", "nonexistent")
	cmd.Env = append(os.Environ(), "HOME="+runeDir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("search failed: %v\n%s", err, out)
	}
	if strings.TrimSpace(string(out)) != "" {
		t.Errorf("expected empty output, got %q", string(out))
	}
}

func TestCLI_Search_ProjectFilter(t *testing.T) {
	binary, runeDir := buildBinaryWithDir(t)

	entriesDir := filepath.Join(runeDir, ".rune", "entries")
	os.MkdirAll(entriesDir, 0755)
	os.WriteFile(filepath.Join(entriesDir, "2025-06-08.md"), []byte("- [@09:15] [project-a] Morning standup prep (branch: main)\n- [@14:30] [idea001] Fixed rate limiting bug (branch: main)\n"), 0644)

	cmd := exec.Command(binary, "search", "-p", "project-a", "bug")
	cmd.Env = append(os.Environ(), "HOME="+runeDir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("search failed: %v\n%s", err, out)
	}
	if strings.TrimSpace(string(out)) != "" {
		t.Errorf("expected no match for bug in project-a, got:\n%s", string(out))
	}
}

func TestCLI_ProjectFlag_NoValue(t *testing.T) {
	binary := buildBinary(t)

	cmd := exec.Command(binary, "-p")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected exit error when -p has no value")
	}
	if string(out) != "Error: -p requires a project name\n" {
		t.Errorf("got %q, want %q", string(out), "Error: -p requires a project name\n")
	}
}

func TestCLI_Config_CreatesFile(t *testing.T) {
	binary, runeDir := buildBinaryWithDir(t)

	cmd := exec.Command(binary, "config")
	cmd.Env = append(os.Environ(), "HOME="+runeDir, "EDITOR=true")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("config failed: %v\n%s", err, out)
	}

	configPath := filepath.Join(runeDir, ".rune", "config.yml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("config file was not created at %s", configPath)
	}
}

func TestCLI_Search_CaseInsensitive(t *testing.T) {
	binary, runeDir := buildBinaryWithDir(t)

	entriesDir := filepath.Join(runeDir, ".rune", "entries")
	os.MkdirAll(entriesDir, 0755)
	os.WriteFile(filepath.Join(entriesDir, "2025-06-08.md"), []byte("- [@09:15] [project-a] Morning Standup Prep (branch: main)\n"), 0644)

	cmd := exec.Command(binary, "search", "standup")
	cmd.Env = append(os.Environ(), "HOME="+runeDir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("search failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "Morning Standup Prep") {
		t.Errorf("expected case-insensitive match, got:\n%s", string(out))
	}
}

func buildBinary(t *testing.T) string {
	t.Helper()
	binary := t.TempDir() + "/rune"
	cmd := exec.Command("go", "build", "-o", binary, ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	return binary
}

func buildBinaryWithDir(t *testing.T) (string, string) {
	t.Helper()
	return buildBinary(t), t.TempDir()
}
