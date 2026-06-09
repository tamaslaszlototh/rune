package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

var (
	cacheOnce     sync.Once
	cachedProject string
	cachedBranch  string
	cachedErr     error
)

func resetCache() {
	cacheOnce = sync.Once{}
	cachedProject = ""
	cachedBranch = ""
	cachedErr = nil
}

func Detect() (project string, branch string, err error) {
	cacheOnce.Do(func() {
		cachedProject, cachedBranch, cachedErr = detect()
	})
	return cachedProject, cachedBranch, cachedErr
}

func detect() (project string, branch string, err error) {
	// Check if we're in a git repo first
	if _, err := exec.Command("git", "rev-parse", "--show-toplevel").Output(); err != nil {
		// Outside a git repo: fall back to directory basename
		cwd, err := os.Getwd()
		if err != nil {
			return "", "", fmt.Errorf("getwd: %w", err)
		}
		return filepath.Base(cwd), "", nil
	}

	remoteOut, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err != nil {
		// Has a git repo but no remote — use directory basename with branch
		cwd, err := os.Getwd()
		if err != nil {
			return "", "", fmt.Errorf("getwd: %w", err)
		}
		project = filepath.Base(cwd)
	} else {
		remote := strings.TrimSpace(string(remoteOut))
		project = extractRepoName(remote)
	}

	branchOut, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err == nil {
		branch = strings.TrimSpace(string(branchOut))
	}

	return project, branch, nil
}

func extractRepoName(remote string) string {
	// Handle both HTTPS and SSH URL formats:
	// https://github.com/owner/repo.git
	// git@github.com:owner/repo.git
	remote = strings.TrimSuffix(remote, ".git")

	if idx := strings.LastIndex(remote, "/"); idx != -1 {
		return remote[idx+1:]
	}
	if idx := strings.LastIndex(remote, ":"); idx != -1 {
		return remote[idx+1:]
	}
	return remote
}
