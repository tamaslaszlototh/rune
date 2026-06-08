package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"rune/internal/store"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFF")).
			Background(lipgloss.Color("#3B82F6")).
			Padding(0, 1)
	entryTimeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))
	entryProjectStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#10B981"))
	emptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF"))
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))
)

type model struct {
	entries []store.Entry
	date    time.Time
	err     error
	store   *store.Store
	loaded  bool
}

func Run() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	s := store.NewStore(filepath.Join(home, ".rune"))
	p := tea.NewProgram(initialModel(s))
	_, err = p.Run()
	return err
}

func initialModel(s *store.Store) model {
	return model{
		date:  time.Now(),
		store: s,
	}
}

func (m model) Init() tea.Cmd {
	return m.loadEntries
}

func (m model) loadEntries() tea.Msg {
	entries, err := m.store.ReadDay(m.date)
	if err != nil {
		return errMsg{err}
	}
	return entriesLoaded{entries: entries}
}

type entriesLoaded struct {
	entries []store.Entry
}

type errMsg struct {
	err error
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case entriesLoaded:
		m.entries = msg.entries
		m.loaded = true
	case errMsg:
		m.err = msg.err
		m.loaded = true
	}
	return m, nil
}

func (m model) View() string {
	if !m.loaded {
		return "Loading..."
	}
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress q to quit.", m.err)
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render(fmt.Sprintf(" rune — %s ", m.date.Format("Mon Jan 2, 2006"))))
	b.WriteString("\n\n")

	if len(m.entries) == 0 {
		b.WriteString(emptyStyle.Render("No entries yet. Add entries by editing the markdown file directly."))
		b.WriteString("\n")
	} else {
		for _, e := range m.entries {
			b.WriteString(fmt.Sprintf(" %s ", entryTimeStyle.Render(e.Timestamp.Format("15:04"))))
			b.WriteString(entryProjectStyle.Render("[" + e.Project + "]"))
			b.WriteString(" ")
			b.WriteString(e.Body)
			if e.Branch != "" {
				b.WriteString(fmt.Sprintf(" (%s)", helpStyle.Render("branch: "+e.Branch)))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("Press q or Ctrl+C to quit."))
	b.WriteString("\n")

	return b.String()
}
