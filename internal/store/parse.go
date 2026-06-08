package store

import (
	"fmt"
	"regexp"
	"time"
)

type Entry struct {
	Timestamp time.Time
	Project   string
	Body      string
	Tags      []string
	Links     []string
	Branch    string
}

var entryRe = regexp.MustCompile(`^\- \[@(\d{2}:\d{2})\] \[([^\]]+)\] (.+) \(branch: ([^)]+)\)$`)
var tagRe = regexp.MustCompile(`#([\w][\w-]*)`)
var linkRe = regexp.MustCompile(`@(\w+/([\w][\w/-]*))`)

func ParseEntryLine(date time.Time, line string) (Entry, error) {
	m := entryRe.FindStringSubmatch(line)
	if m == nil {
		return Entry{}, nil
	}

	timeStr := m[1]
	project := m[2]
	body := m[3]
	branch := m[4]

	t, _ := time.Parse("15:04", timeStr)
	timestamp := time.Date(date.Year(), date.Month(), date.Day(), t.Hour(), t.Minute(), 0, 0, date.Location())

	tags := extractTags(body)
	links := extractLinks(body)

	return Entry{
		Timestamp: timestamp,
		Project:   project,
		Body:      body,
		Tags:      tags,
		Links:     links,
		Branch:    branch,
	}, nil
}

func extractTags(body string) []string {
	matches := tagRe.FindAllStringSubmatch(body, -1)
	tags := make([]string, 0, len(matches))
	for _, m := range matches {
		tags = append(tags, m[1])
	}
	return tags
}

func extractLinks(body string) []string {
	matches := linkRe.FindAllStringSubmatch(body, -1)
	links := make([]string, 0, len(matches))
	for _, m := range matches {
		links = append(links, m[1])
	}
	return links
}

func formatEntry(e Entry) string {
	return fmt.Sprintf("- [@%s] [%s] %s (branch: %s)",
		e.Timestamp.Format("15:04"),
		e.Project,
		e.Body,
		e.Branch,
	)
}
