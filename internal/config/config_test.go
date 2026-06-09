package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEdit_CreatesMissingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	cfg := &Config{}

	os.Setenv("EDITOR", "true")
	defer os.Unsetenv("EDITOR")

	if err := Edit(path, cfg); err != nil {
		t.Fatalf("Edit() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Error("Edit() should create file with template content")
	}
}

func TestEdit_ExistingFileNotOverwritten(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	if err := os.WriteFile(path, []byte("editor: code\n"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &Config{}
	os.Setenv("EDITOR", "true")
	defer os.Unsetenv("EDITOR")

	if err := Edit(path, cfg); err != nil {
		t.Fatalf("Edit() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "editor: code\n" {
		t.Errorf("existing file was modified, got %q", string(data))
	}
}

func TestEditor(t *testing.T) {
	t.Run("from config", func(t *testing.T) {
		cfg := &Config{Editor: "nvim"}
		if got := Editor(cfg); got != "nvim" {
			t.Errorf("Editor() = %q, want %q", got, "nvim")
		}
	})

	t.Run("from env", func(t *testing.T) {
		os.Setenv("EDITOR", "code")
		defer os.Unsetenv("EDITOR")
		cfg := &Config{}
		if got := Editor(cfg); got != "code" {
			t.Errorf("Editor() = %q, want %q", got, "code")
		}
	})

	t.Run("default", func(t *testing.T) {
		os.Unsetenv("EDITOR")
		cfg := &Config{}
		if got := Editor(cfg); got != "vim" {
			t.Errorf("Editor() = %q, want %q", got, "vim")
		}
	})

	t.Run("config overrides env", func(t *testing.T) {
		os.Setenv("EDITOR", "code")
		defer os.Unsetenv("EDITOR")
		cfg := &Config{Editor: "nvim"}
		if got := Editor(cfg); got != "nvim" {
			t.Errorf("Editor() = %q, want %q", got, "nvim")
		}
	})
}

func TestLoad_MissingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.yml")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error for missing file = %v", err)
	}
	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}
	if cfg.Projects == nil {
		t.Error("Projects should be non-nil for default config")
	}
	if len(cfg.Projects) != 0 {
		t.Errorf("Projects = %v, want empty map", cfg.Projects)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	content := []byte("projects: [invalid: yaml: broken")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err == nil {
		t.Fatal("Load() expected error for invalid YAML")
	}
	if cfg != nil {
		t.Errorf("Load() returned %+v, want nil on error", cfg)
	}
}

func TestLoad_Success(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	content := []byte("projects:\n  idea001: Idea Project One\n  myapp: My Application\neditor: vim\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}
	if cfg.Projects["idea001"] != "Idea Project One" {
		t.Errorf("Projects[idea001] = %q, want %q", cfg.Projects["idea001"], "Idea Project One")
	}
	if cfg.Projects["myapp"] != "My Application" {
		t.Errorf("Projects[myapp] = %q, want %q", cfg.Projects["myapp"], "My Application")
	}
	if cfg.Editor != "vim" {
		t.Errorf("Editor = %q, want %q", cfg.Editor, "vim")
	}
}
