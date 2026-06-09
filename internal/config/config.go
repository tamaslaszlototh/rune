package config

import (
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const configTemplate = `# rune configuration
# See https://github.com/rune/usage for documentation

# Editor to open for 'rune config' (default: $EDITOR or vim)
# editor: code

# Friendly names for projects (optional)
# projects:
#   idea001: Idea Project One
`

func Edit(path string, cfg *Config) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(configTemplate), 0644); err != nil {
			return err
		}
	}

	editor := Editor(cfg)
	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Editor(cfg *Config) string {
	if cfg.Editor != "" {
		return cfg.Editor
	}
	if e := os.Getenv("EDITOR"); e != "" {
		return e
	}
	return "vim"
}

type Config struct {
	Projects map[string]string `yaml:"projects"`
	Editor   string            `yaml:"editor"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{
				Projects: map[string]string{},
			}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Projects == nil {
		cfg.Projects = map[string]string{}
	}
	return &cfg, nil
}
