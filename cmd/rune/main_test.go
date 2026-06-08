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
	if string(out) != "Usage: rune [standup]\n" {
		t.Errorf("got %q, want %q", string(out), "Usage: rune [standup]\n")
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
