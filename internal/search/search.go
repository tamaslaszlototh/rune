package search

import (
	"strings"

	"rune/internal/store"
)

type Match struct {
	Entry store.Entry
}

func Search(entries []store.Entry, query string) []Match {
	queryLower := strings.ToLower(query)
	var matches []Match
	for _, e := range entries {
		if strings.Contains(strings.ToLower(e.Body), queryLower) {
			matches = append(matches, Match{Entry: e})
		}
	}
	return matches
}

func FilterByProject(entries []store.Entry, project string) []store.Entry {
	if project == "" {
		return entries
	}
	var filtered []store.Entry
	for _, e := range entries {
		if e.Project == project {
			filtered = append(filtered, e)
		}
	}
	return filtered
}
