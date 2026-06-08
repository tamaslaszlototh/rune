package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"rune/internal/config"
	"rune/internal/search"
	"rune/internal/standup"
	"rune/internal/store"
	"rune/internal/tui"
)

func main() {
	cfg := loadConfig()

	if len(os.Args) < 2 {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		s := store.NewStore(filepath.Join(home, ".rune"))
		if err := tui.Run(tui.NewStoreAdapter(s), tui.NewGitAdapter(), cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	switch os.Args[1] {
	case "config":
		runConfig(cfg)
	case "standup":
		runStandup(os.Args[2:], cfg)
	case "search":
		runSearch(os.Args[2:], cfg)
	default:
		fmt.Fprintf(os.Stderr, "Usage: rune [config|standup|search]\n")
		os.Exit(1)
	}
}

func configPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".rune", "config.yml")
}

func runConfig(cfg *config.Config) {
	path := configPath()
	if path == "" {
		fmt.Fprintln(os.Stderr, "Error: cannot determine home directory")
		os.Exit(1)
	}
	if err := config.Edit(path, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func loadConfig() *config.Config {
	path := configPath()
	if path == "" {
		return &config.Config{Projects: map[string]string{}}
	}
	cfg, err := config.Load(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: invalid config file: %v\n", err)
		return &config.Config{Projects: map[string]string{}}
	}
	return cfg
}

func runStandup(args []string, cfg *config.Config) {
	since := time.Now().Add(-24 * time.Hour)

	// Parse --since flag
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--since":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "Error: --since requires a value")
				os.Exit(1)
			}
			i++
			parsed, err := parseSince(args[i])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			since = parsed
		default:
			fmt.Fprintf(os.Stderr, "Error: unknown flag %q\n", args[i])
			os.Exit(1)
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	s := store.NewStore(filepath.Join(home, ".rune"))
	entries, err := s.ReadRange(since, time.Now())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	out := standup.FormatStandup(entries, since)
	fmt.Println(out)
}

func runSearch(args []string, cfg *config.Config) {
	project := ""

	// Parse -p flag
	remaining := args[:0]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-p":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "Error: -p requires a project name")
				os.Exit(1)
			}
			i++
			project = args[i]
		default:
			remaining = append(remaining, args[i])
		}
	}
	args = remaining

	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: search query is required")
		os.Exit(1)
	}
	query := strings.Join(args, " ")

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	s := store.NewStore(filepath.Join(home, ".rune"))
	entries, err := s.ReadRange(time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local), time.Now())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	entries = search.FilterByProject(entries, project)
	matches := search.Search(entries, query)

	for _, m := range matches {
		t := m.Entry.Timestamp
		projectStr := m.Entry.Project
		if projectStr == "" {
			projectStr = "general"
		}
		fmt.Printf("%s %s [%s] %s\n",
			t.Format("2006-01-02"),
			t.Format("15:04"),
			projectStr,
			m.Entry.Body,
		)
	}
}

func parseSince(raw string) (time.Time, error) {
	now := time.Now()

	// Relative format: -3d, -1d, etc.
	if strings.HasPrefix(raw, "-") && strings.HasSuffix(raw, "d") {
		nStr := raw[1 : len(raw)-1]
		n, err := strconv.Atoi(nStr)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid relative format %q (expected e.g. -3d)", raw)
		}
		return now.AddDate(0, 0, -n), nil
	}

	// ISO date: 2025-06-08
	if t, err := time.Parse("2006-01-02", raw); err == nil {
		return t, nil
	}

	// Day name: monday, tuesday, wednesday, thursday, friday, saturday, sunday
	dayNames := map[string]time.Weekday{
		"sunday":    time.Sunday,
		"monday":    time.Monday,
		"tuesday":   time.Tuesday,
		"wednesday": time.Wednesday,
		"thursday":  time.Thursday,
		"friday":    time.Friday,
		"saturday":  time.Saturday,
	}
	if targetDay, ok := dayNames[strings.ToLower(raw)]; ok {
		daysBack := (int(now.Weekday()) - int(targetDay) + 7) % 7
		if daysBack == 0 {
			daysBack = 7 // most recent previous occurrence
		}
		return now.AddDate(0, 0, -daysBack), nil
	}

	return time.Time{}, fmt.Errorf("unrecognized --since value %q (try ISO date, day name, or -Nd)", raw)
}
