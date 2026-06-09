package standup

import (
	"sort"
	"strings"
	"time"

	"rune/internal/store"
)

func FormatStandup(entries []store.Entry, since time.Time) string {
	if len(entries) == 0 {
		return ""
	}

	// Filter by since cutoff
	var filtered []store.Entry
	for _, e := range entries {
		if !e.Timestamp.Before(since) {
			filtered = append(filtered, e)
		}
	}
	if len(filtered) == 0 {
		return ""
	}

	// Sort chronologically
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Timestamp.Before(filtered[j].Timestamp)
	})

	// Group by project
	type projectGroup struct {
		display string
		branch  string
		entries []store.Entry
	}

	var groups []*projectGroup
	groupIndex := make(map[string]*projectGroup)
	for _, e := range filtered {
		display := e.Project
		if display == "" {
			display = "general"
		}
		g, ok := groupIndex[display]
		if !ok {
			g = &projectGroup{
				display: display,
				branch:  e.Branch,
			}
			groupIndex[display] = g
			groups = append(groups, g)
		}
		g.entries = append(g.entries, e)
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].display < groups[j].display
	})

	var b strings.Builder
	b.WriteString("## Standup — ")
	b.WriteString(time.Now().Format("2006-01-02"))
	b.WriteString("\n\n")

	for _, g := range groups {
		b.WriteString("### ")
		b.WriteString(g.display)
		if g.branch != "" {
			b.WriteString(" (")
			b.WriteString(g.branch)
			b.WriteString(")")
		}
		b.WriteString("\n")
		for _, e := range g.entries {
			b.WriteString("• ")
			b.WriteString(e.Body)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	return strings.TrimSuffix(b.String(), "\n")
}
