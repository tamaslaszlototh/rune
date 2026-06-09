package tui

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"rune/internal/config"
	"rune/internal/store"

	"github.com/charmbracelet/bubbles/textinput"
)

const (
	autoSaveDelay     = 2 * time.Second
	statusClearDelay  = 1500 * time.Millisecond
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFF")).
			Background(lipgloss.Color("#3B82F6")).
			Padding(0, 1)
	entryTimeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))
	statsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF"))
	sectionHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#D1D5DB")).
				Background(lipgloss.Color("#1F2937")).
				Padding(0, 1)
	emptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF"))
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))
	filterPillStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Padding(0, 1)
	filterPillActiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFF")).
				Background(lipgloss.Color("#3B82F6")).
				Padding(0, 1)
	matchStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#FDE047")).
			Foreground(lipgloss.Color("#000"))
	tagStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B"))
	linkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8B5CF6"))
	branchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#06B6D4"))
)

var (
	tagRenderRe   = regexp.MustCompile(`#[\w][\w-]*`)
	linkRenderRe  = regexp.MustCompile(`@\w+/[\w][\w/-]*`)
)

type model struct {
	entries         []store.Entry
	date            time.Time
	err             error
	store           EntryStore
	git             GitDetector
	loaded          bool
	input           textinput.Model
	project         string
	branch          string
	projects        []string
	filterIndex     int
	searching       bool
	savedDraft      string
	savedFilterIndex int
	statusMsg       string
	cfg             *config.Config
	filterProject   string
}

func Run(s EntryStore, g GitDetector, cfg *config.Config, filterProject string) error {
	p := tea.NewProgram(initialModel(s, g, cfg, filterProject))
	_, err := p.Run()
	return err
}

func initialModel(s EntryStore, g GitDetector, cfg *config.Config, filterProject string) model {
	ti := textinput.New()
	ti.Placeholder = "What did you work on?"
	ti.Focus()

	project, branch, _ := g.Detect()

	ti.Prompt = "> "
	if project != "" {
		ti.Prompt = fmt.Sprintf("> [%s] ", project)
	}

	m := model{
		date:          time.Now(),
		store:         s,
		git:           g,
		input:         ti,
		project:       project,
		branch:        branch,
		cfg:           cfg,
		filterProject: filterProject,
	}

	draft, err := s.ReadDraft(m.date)
	if err == nil && draft != "" {
		m.input.SetValue(draft)
	}

	return m
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

type autoSaveMsg struct{}

type clearStatusMsg struct{}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "/":
			if !m.searching {
				return m.enterSearchMode()
			}
		case "enter":
			if m.searching {
				return m, nil
			}
			return m.handleEnter()
		case "esc":
			if m.searching {
				return m.exitSearchMode()
			}
			m.input.SetValue("")
			m.input.SetCursor(0)
			return m, nil
		case "ctrl+s":
			m.statusMsg = ""
			if m.input.Value() != "" {
				if err := m.store.SaveDraft(m.date, m.input.Value()); err != nil {
					m.err = err
				} else {
					m.statusMsg = "Draft saved"
				}
			}
			return m, tea.Tick(statusClearDelay, func(t time.Time) tea.Msg {
				return clearStatusMsg{}
			})
		case "tab":
			return m.cycleFilter(1), nil
		case "shift+tab":
			return m.cycleFilter(-1), nil
		}
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		if m.searching {
			return m, cmd
		}
		return m, tea.Batch(cmd, m.scheduleAutoSave())
	case entriesLoaded:
		m.entries = msg.entries
		m.loaded = true
		m.projects = extractProjects(m.entries)
		if m.filterProject != "" {
			m.filterIndex = -1
			for i, p := range m.projects {
				if p == m.filterProject {
					m.filterIndex = i
					break
				}
			}
		}
		if m.filterIndex >= len(m.projects) {
			m.filterIndex = -1
		}
	case errMsg:
		m.err = msg.err
		m.loaded = true
	case autoSaveMsg:
		if m.input.Value() != "" {
			if err := m.store.SaveDraft(m.date, m.input.Value()); err != nil {
				m.err = err
			}
		}
	case clearStatusMsg:
		m.statusMsg = ""
	}
	return m, nil
}

func (m model) scheduleAutoSave() tea.Cmd {
	return tea.Tick(autoSaveDelay, func(t time.Time) tea.Msg {
		return autoSaveMsg{}
	})
}

func countPerProject(entries []store.Entry) map[string]int {
	counts := make(map[string]int)
	for _, e := range entries {
		p := e.Project
		if p == "" {
			p = "general"
		}
		counts[p]++
	}
	return counts
}

func extractProjects(entries []store.Entry) []string {
	seen := make(map[string]bool)
	var projects []string
	for _, e := range entries {
		if e.Project != "" && !seen[e.Project] {
			seen[e.Project] = true
			projects = append(projects, e.Project)
		}
	}
	return projects
}

func (m model) cycleFilter(direction int) model {
	n := len(m.projects)
	if n == 0 {
		return m
	}
	m.filterIndex += direction
	// wrap beyond last project → -1 (All)
	if m.filterIndex >= n {
		m.filterIndex = -1
	}
	// wrap before -1 → last project
	if m.filterIndex < -1 {
		m.filterIndex = n - 1
	}
	return m
}

func highlightMatch(body, query string) string {
	if query == "" {
		return body
	}
	lowerBody := strings.ToLower(body)
	lowerQuery := strings.ToLower(query)
	var b strings.Builder
	start := 0
	for {
		idx := strings.Index(lowerBody[start:], lowerQuery)
		if idx == -1 {
			b.WriteString(body[start:])
			break
		}
		absIdx := start + idx
		b.WriteString(body[start:absIdx])
		b.WriteString(matchStyle.Render(body[absIdx : absIdx+len(query)]))
		start = absIdx + len(query)
	}
	return b.String()
}

func renderBody(body string) string {
	type match struct {
		start, end int
		style      lipgloss.Style
	}

	var matches []match
	for _, m := range tagRenderRe.FindAllStringIndex(body, -1) {
		matches = append(matches, match{m[0], m[1], tagStyle})
	}
	for _, m := range linkRenderRe.FindAllStringIndex(body, -1) {
		matches = append(matches, match{m[0], m[1], linkStyle})
	}
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].start < matches[j].start
	})

	var b strings.Builder
	pos := 0
	for _, m := range matches {
		if m.start < pos {
			continue
		}
		b.WriteString(body[pos:m.start])
		b.WriteString(m.style.Render(body[m.start:m.end]))
		pos = m.end
	}
	b.WriteString(body[pos:])
	return b.String()
}

func (m model) filterBarView() string {
	if len(m.projects) == 0 {
		return ""
	}
	var b strings.Builder
	// "All" pill
	allStyle := filterPillStyle
	if m.filterIndex == -1 {
		allStyle = filterPillActiveStyle
	}
	b.WriteString(allStyle.Render("All"))
	for i, name := range m.projects {
		style := filterPillStyle
		if i == m.filterIndex {
			style = filterPillActiveStyle
		}
		b.WriteString(style.Render(name))
	}
	return b.String()
}

func (m model) enterSearchMode() (tea.Model, tea.Cmd) {
	m.savedDraft = m.input.Value()
	m.savedFilterIndex = m.filterIndex
	m.filterIndex = -1
	m.input.SetValue("")
	m.input.SetCursor(0)
	m.input.Placeholder = "Search entries..."
	m.input.Prompt = "[/] "
	m.searching = true
	return m, nil
}

func (m model) exitSearchMode() (tea.Model, tea.Cmd) {
	m.input.SetValue(m.savedDraft)
	m.input.SetCursor(len(m.savedDraft))
	m.savedDraft = ""
	m.filterIndex = m.savedFilterIndex
	m.input.Placeholder = "What did you work on?"
	m.searching = false
	// Restore project prompt if applicable
	m.input.Prompt = "> "
	if m.project != "" {
		m.input.Prompt = fmt.Sprintf("> [%s] ", m.project)
	}
	return m, nil
}

func (m model) handleEnter() (tea.Model, tea.Cmd) {
	body := m.input.Value()
	if body == "" {
		return m, nil
	}

	entry := store.Entry{
		Timestamp: time.Now(),
		Body:      body,
		Project:   m.project,
		Branch:    m.branch,
	}

	if err := m.store.AppendEntry(m.date, entry); err != nil {
		m.err = err
		return m, nil
	}

	m.input.SetValue("")
	if err := m.store.ClearDraft(m.date); err != nil {
		m.err = err
		return m, nil
	}
	return m, m.loadEntries
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
	b.WriteString("\n")
	if len(m.entries) > 0 {
		n := len(m.entries)
		np := len(m.projects)
		lastTime := m.entries[len(m.entries)-1].Timestamp.Format("15:04")
		stats := fmt.Sprintf("  %d entries · %d projects · last entry %s ", n, np, lastTime)
		b.WriteString(statsStyle.Render(stats))
	}
	b.WriteString("\n\n")

	displayEntries := m.entries
	if m.filterIndex >= 0 && m.filterIndex < len(m.projects) {
		filterProject := m.projects[m.filterIndex]
		var filtered []store.Entry
		for _, e := range displayEntries {
			if e.Project == filterProject {
				filtered = append(filtered, e)
			}
		}
		displayEntries = filtered
	}
	if m.searching && m.input.Value() != "" {
		query := strings.ToLower(m.input.Value())
		var filtered []store.Entry
		for _, e := range displayEntries {
			if strings.Contains(strings.ToLower(e.Body), query) {
				filtered = append(filtered, e)
			}
		}
		displayEntries = filtered
	}

	if len(displayEntries) == 0 {
		if m.searching {
			msg := fmt.Sprintf("No entries matching %q.", m.input.Value())
			b.WriteString(emptyStyle.Render(msg))
		} else if m.filterIndex >= 0 && len(m.projects) > 0 {
			msg := fmt.Sprintf("No entries for %s.", m.projects[m.filterIndex])
			b.WriteString(emptyStyle.Render(msg))
		} else {
			msg := "No entries yet. Type below and press Enter to add one."
			b.WriteString(emptyStyle.Render(msg))
		}
		b.WriteString("\n")
	} else {
		counts := countPerProject(displayEntries)
		var currentProject string
		for _, e := range displayEntries {
			if e.Project != currentProject {
				currentProject = e.Project
				section := currentProject
				if section == "" {
					section = "general"
				}
				n := counts[currentProject]
				b.WriteString(sectionHeaderStyle.Render(fmt.Sprintf(" %s (%d) ", section, n)))
				b.WriteString("\n")
			}
			b.WriteString(fmt.Sprintf(" %s ", entryTimeStyle.Render(e.Timestamp.Format("15:04"))))
			if m.searching && m.input.Value() != "" {
				b.WriteString(highlightMatch(e.Body, m.input.Value()))
			} else {
				b.WriteString(renderBody(e.Body))
			}
			if e.Branch != "" {
				b.WriteString(fmt.Sprintf(" %s", branchStyle.Render("["+e.Branch+"]")))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(m.input.View())
	b.WriteString("\n\n")
	b.WriteString(m.filterBarView())
	b.WriteString("\n")
	if m.statusMsg != "" {
		b.WriteString(helpStyle.Render(m.statusMsg))
		b.WriteString("\n")
	}
	if m.searching {
		b.WriteString(helpStyle.Render("ctrl+s: save draft | esc: exit search | ctrl+w: delete word | ctrl+u: clear line"))
	} else {
		b.WriteString(helpStyle.Render("ctrl+c: quit | enter: save | ctrl+s: save draft | /: search | esc: clear | ctrl+w: delete word | ctrl+u: clear line"))
	}
	b.WriteString("\n")

	return b.String()
}
