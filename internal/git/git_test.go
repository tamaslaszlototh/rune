package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestDetect_InGitRepo(t *testing.T) {
	resetCache()
	// Set up a temporary git repo
	dir := t.TempDir()

	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}

	run("init", "--initial-branch=main")
	run("remote", "add", "origin", "https://github.com/octocat/hello-world.git")

	// Need at least one commit for rev-parse to work
	echo := exec.Command("sh", "-c", "echo test > dummy")
	echo.Dir = dir
	_ = echo.Run()
	run("add", ".")
	run("commit", "-m", "initial")

	// Change to the temp dir for the test
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	project, branch, err := Detect()
	if err != nil {
		t.Fatal(err)
	}
	if project != "hello-world" {
		t.Errorf("project = %q, want %q", project, "hello-world")
	}
	if branch != "main" {
		t.Errorf("branch = %q, want %q", branch, "main")
	}
}

func TestDetect_OutsideGitRepo(t *testing.T) {
	resetCache()
	dir := t.TempDir()

	// TempDir creates a random path; the basename is unpredictable.
	// We create a subdirectory with a known name to control the project.
	projectDir := filepath.Join(dir, "my-project")
	if err := os.Mkdir(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	os.Chdir(projectDir)
	defer os.Chdir(origDir)

	project, branch, err := Detect()
	if err != nil {
		t.Fatal(err)
	}
	if project != "my-project" {
		t.Errorf("project = %q, want %q", project, "my-project")
	}
	if branch != "" {
		t.Errorf("branch = %q, want empty", branch)
	}
}

func TestDetect_IsCached(t *testing.T) {
	resetCache()
	dir := t.TempDir()
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}
	run("init", "--initial-branch=main")
	run("remote", "add", "origin", "https://github.com/owner/my-repo.git")
	echo := exec.Command("sh", "-c", "echo test > dummy")
	echo.Dir = dir
	_ = echo.Run()
	run("add", ".")
	run("commit", "-m", "initial")

	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	// First call populates cache
	p1, b1, err := Detect()
	if err != nil {
		t.Fatal(err)
	}

	// Move to a directory without git — cached values should survive
	nonGitDir := t.TempDir()
	os.Chdir(nonGitDir)

	p2, b2, err := Detect()
	if err != nil {
		t.Fatal(err)
	}
	if p1 != p2 {
		t.Errorf("project changed after cache: %q → %q", p1, p2)
	}
	if b1 != b2 {
		t.Errorf("branch changed after cache: %q → %q", b1, b2)
	}
}
