package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"rune/internal/standup"
	"rune/internal/store"
	"rune/internal/tui"
)

func main() {
	if len(os.Args) < 2 {
		if err := tui.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	switch os.Args[1] {
	case "standup":
		runStandup(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Usage: rune [standup]\n")
		os.Exit(1)
	}
}

func runStandup(args []string) {
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
