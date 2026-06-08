package main

import (
	"os/exec"
	"testing"
)

func TestCLI_UnknownFlag(t *testing.T) {
	// Build the binary
	binary := t.TempDir() + "/rune"
	cmd := exec.Command("go", "build", "-o", binary, ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}

	// Run with unknown flag
	cmd = exec.Command(binary, "--help")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected exit error for unknown flag")
	}
	if string(out) != "Usage: rune\n" {
		t.Errorf("got %q, want %q", string(out), "Usage: rune\n")
	}
}
