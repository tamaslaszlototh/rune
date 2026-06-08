package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Store struct {
	Dir string
}

func NewStore(dir string) *Store {
	return &Store{Dir: dir}
}

func (s *Store) filePath(date time.Time) string {
	name := fmt.Sprintf("%s.md", date.Format("2006-01-02"))
	return filepath.Join(s.Dir, "entries", name)
}

func (s *Store) ReadDay(date time.Time) ([]Entry, error) {
	data, err := os.ReadFile(s.filePath(date))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return parseEntries(date, string(data))
}

func (s *Store) ReadRange(from, to time.Time) ([]Entry, error) {
	var entries []Entry
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		dayEntries, err := s.ReadDay(d)
		if err != nil {
			return nil, err
		}
		entries = append(entries, dayEntries...)
	}
	if entries == nil {
		return []Entry{}, nil
	}
	return entries, nil
}

func (s *Store) entriesDir() string {
	return filepath.Join(s.Dir, "entries")
}

func (s *Store) AppendEntry(date time.Time, entry Entry) error {
	dir := s.entriesDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(s.filePath(date), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintln(f, formatEntry(entry))
	return err
}

func parseEntries(date time.Time, content string) ([]Entry, error) {
	lines := strings.Split(strings.TrimSpace(content), "\n")
	var entries []Entry
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		entry, err := ParseEntryLine(date, line)
		if err != nil {
			return nil, err
		}
		if entry.Timestamp.IsZero() {
			continue
		}
		entries = append(entries, entry)
	}
	if entries == nil {
		return []Entry{}, nil
	}
	return entries, nil
}
