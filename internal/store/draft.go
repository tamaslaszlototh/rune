package store

import (
	"os"
	"path/filepath"
	"time"
)

func (s *Store) draftsDir() string {
	return filepath.Join(s.Dir, "drafts")
}

func (s *Store) draftPath(date time.Time) string {
	name := date.Format("2006-01-02") + ".md"
	return filepath.Join(s.draftsDir(), name)
}

func (s *Store) SaveDraft(date time.Time, text string) error {
	dir := s.draftsDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(s.draftPath(date), []byte(text), 0644)
}

func (s *Store) ReadDraft(date time.Time) (string, error) {
	data, err := os.ReadFile(s.draftPath(date))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

func (s *Store) ClearDraft(date time.Time) error {
	err := os.Remove(s.draftPath(date))
	if err != nil && os.IsNotExist(err) {
		return nil
	}
	return err
}
