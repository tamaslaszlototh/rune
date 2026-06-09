package tui

import (
	"time"

	"rune/internal/git"
	"rune/internal/store"
)

type storeAdapter struct {
	inner *store.Store
}

func NewStoreAdapter(s *store.Store) EntryStore {
	return &storeAdapter{inner: s}
}

func (a *storeAdapter) ReadDay(date time.Time) ([]store.Entry, error) {
	return a.inner.ReadDay(date)
}

func (a *storeAdapter) ReadRange(from, to time.Time) ([]store.Entry, error) {
	return a.inner.ReadRange(from, to)
}

func (a *storeAdapter) AppendEntry(date time.Time, entry store.Entry) error {
	return a.inner.AppendEntry(date, entry)
}

func (a *storeAdapter) SaveDraft(date time.Time, text string) error {
	return a.inner.SaveDraft(date, text)
}

func (a *storeAdapter) ReadDraft(date time.Time) (string, error) {
	return a.inner.ReadDraft(date)
}

func (a *storeAdapter) ClearDraft(date time.Time) error {
	return a.inner.ClearDraft(date)
}

type gitAdapter struct{}

func NewGitAdapter() GitDetector {
	return &gitAdapter{}
}

func (gitAdapter) Detect() (string, string, error) {
	return git.Detect()
}
