package tui

import (
	"sync"
	"time"

	"rune/internal/store"
)

type memStore struct {
	mu      sync.Mutex
	entries map[time.Time][]store.Entry
	drafts  map[time.Time]string
}

func newMemStore() *memStore {
	return &memStore{
		entries: make(map[time.Time][]store.Entry),
		drafts:  make(map[time.Time]string),
	}
}

func (m *memStore) ReadDay(date time.Time) ([]store.Entry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	date = date.Truncate(24 * time.Hour)
	entries := m.entries[date]
	if entries == nil {
		return nil, nil
	}
	return entries, nil
}

func (m *memStore) ReadRange(from, to time.Time) ([]store.Entry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var all []store.Entry
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		d = d.Truncate(24 * time.Hour)
		all = append(all, m.entries[d]...)
	}
	if all == nil {
		return []store.Entry{}, nil
	}
	return all, nil
}

func (m *memStore) AppendEntry(date time.Time, entry store.Entry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	date = date.Truncate(24 * time.Hour)
	m.entries[date] = append(m.entries[date], entry)
	return nil
}

func (m *memStore) SaveDraft(date time.Time, text string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.drafts[date.Truncate(24*time.Hour)] = text
	return nil
}

func (m *memStore) ReadDraft(date time.Time) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.drafts[date.Truncate(24*time.Hour)], nil
}

func (m *memStore) ClearDraft(date time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.drafts, date.Truncate(24*time.Hour))
	return nil
}

type fakeGit struct {
	project string
	branch  string
}

func (f *fakeGit) Detect() (string, string, error) {
	return f.project, f.branch, nil
}
