package tui

import (
	"time"

	"rune/internal/store"
)

type EntryStore interface {
	ReadDay(date time.Time) ([]store.Entry, error)
	ReadRange(from, to time.Time) ([]store.Entry, error)
	AppendEntry(date time.Time, entry store.Entry) error
	SaveDraft(date time.Time, text string) error
	ReadDraft(date time.Time) (string, error)
	ClearDraft(date time.Time) error
}

type GitDetector interface {
	Detect() (project, branch string, err error)
}
