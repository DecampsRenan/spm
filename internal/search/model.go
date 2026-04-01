package search

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/decampsrenan/spm/internal/ui"
)

const (
	maxResults      = 10
	debounceDelay   = 300 * time.Millisecond
	maxVersionRows  = 10
	recentThreshold = 7 * 24 * time.Hour
)

type viewState int

const (
	viewSearch viewState = iota
	viewDetails
)

// model is the bubbletea model for interactive package search.
type model struct {
	// Search view
	textInput  textinput.Model
	initCmd    tea.Cmd
	results    []Result
	cursor     int
	loading    bool
	saveDev    bool
	selected   bool
	err        error
	lastQuery  string
	debounceID int
	width      int

	// Details view
	view             viewState
	details          *PackageDetails
	detailsLoading   bool
	detailsErr       error
	versionInput     textinput.Model
	versionCursor    int
	filteredVersions []VersionInfo
	selectedVersion  string
}

// Selection holds the user's final choice from the search TUI.
type Selection struct {
	Package string
	SaveDev bool
}

// Messages

type searchResultMsg struct {
	query   string
	results []Result
}

type searchErrMsg struct {
	query string
	err   error
}

type debounceMsg struct {
	id    int
	query string
}

type detailsResultMsg struct {
	details *PackageDetails
	err     error
}

func newModel() model {
	ti := textinput.New()
	ti.Placeholder = "Type to search npm packages..."
	ti.CharLimit = 100
	ti.SetWidth(50)
	blinkCmd := ti.Focus()

	vi := textinput.New()
	vi.Placeholder = "Filter versions..."
	vi.CharLimit = 50
	vi.SetWidth(50)

	return model{
		textInput:    ti,
		versionInput: vi,
		initCmd:      blinkCmd,
		width:        80,
	}
}

func (m model) Init() tea.Cmd {
	return m.initCmd
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		// "  Search packages: " = 19 visible chars; "  Filter: " = 10
		m.textInput.SetWidth(max(20, msg.Width-20))
		m.versionInput.SetWidth(max(20, msg.Width-12))
		return m, nil

	case detailsResultMsg:
		if m.view == viewDetails && m.detailsLoading {
			m.detailsLoading = false
			if msg.err != nil {
				m.detailsErr = msg.err
			} else {
				m.details = msg.details
				m.filteredVersions = msg.details.Versions
				m.versionCursor = 0
				for i, v := range m.filteredVersions {
					if v.Version == msg.details.Latest {
						m.versionCursor = i
						break
					}
				}
			}
		}
		return m, nil

	case tea.KeyPressMsg:
		// Global keys
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.view == viewDetails {
				m.view = viewSearch
				blinkCmd := m.textInput.Focus()
				m.versionInput.Blur()
				return m, blinkCmd
			}
			return m, tea.Quit
		}

		if m.view == viewDetails {
			switch msg.String() {
			case "up":
				if m.versionCursor > 0 {
					m.versionCursor--
				}
				return m, nil
			case "down":
				if m.versionCursor < len(m.filteredVersions)-1 {
					m.versionCursor++
				}
				return m, nil
			case "tab":
				m.saveDev = !m.saveDev
				return m, nil
			case "enter":
				if !m.detailsLoading && m.detailsErr == nil && len(m.filteredVersions) > 0 {
					m.selectedVersion = m.filteredVersions[m.versionCursor].Version
					m.selected = true
					return m, tea.Quit
				}
				return m, nil
			}
			// Non-navigation keys fall through to versionInput update below
		} else {
			switch msg.String() {
			case "up":
				if m.cursor > 0 {
					m.cursor--
				}
				return m, nil
			case "down":
				if m.cursor < len(m.results)-1 {
					m.cursor++
				}
				return m, nil
			case "tab":
				m.saveDev = !m.saveDev
				return m, nil
			case "enter":
				if len(m.results) > 0 && m.cursor < len(m.results) {
					m.view = viewDetails
					m.detailsLoading = true
					m.details = nil
					m.detailsErr = nil
					m.filteredVersions = nil
					m.versionCursor = 0
					m.versionInput.SetValue("")
					blinkCmd := m.versionInput.Focus()
					m.textInput.Blur()
					return m, tea.Batch(blinkCmd, doFetchDetails(m.results[m.cursor].Name))
				}
				return m, nil
			}
			// Non-navigation keys fall through to textInput update below
		}

	case searchResultMsg:
		if msg.query == m.textInput.Value() {
			m.results = msg.results
			m.loading = false
			m.cursor = 0
		}
		return m, nil

	case searchErrMsg:
		if msg.query == m.textInput.Value() {
			m.err = msg.err
			m.loading = false
		}
		return m, nil

	case debounceMsg:
		if msg.id == m.debounceID && msg.query != "" && msg.query == m.textInput.Value() {
			m.loading = true
			m.err = nil
			return m, doSearch(msg.query)
		}
		return m, nil
	}

	// Route remaining messages to the active text input
	if m.view == viewDetails && m.details != nil {
		prevFilter := m.versionInput.Value()
		var cmd tea.Cmd
		m.versionInput, cmd = m.versionInput.Update(msg)
		if m.versionInput.Value() != prevFilter {
			m.filteredVersions = filterVersions(m.details.Versions, m.versionInput.Value())
			m.versionCursor = 0
		}
		return m, cmd
	}

	if m.view == viewSearch {
		prevValue := m.textInput.Value()
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)

		if m.textInput.Value() != prevValue {
			m.debounceID++
			id := m.debounceID
			query := m.textInput.Value()

			if query == "" {
				m.results = nil
				m.loading = false
				m.err = nil
				return m, cmd
			}

			debounceCmd := tea.Tick(debounceDelay, func(t time.Time) tea.Msg {
				return debounceMsg{id: id, query: query}
			})
			return m, tea.Batch(cmd, debounceCmd)
		}

		return m, cmd
	}

	return m, nil
}

func (m model) View() tea.View {
	var b strings.Builder

	if m.view == viewDetails {
		b.WriteString(m.viewDetailsPane())
	} else {
		b.WriteString(m.viewSearchPane())
	}

	// Footer
	b.WriteString("\n")
	devLabel := "no"
	if m.saveDev {
		devLabel = ui.StyleSuccess.Render("yes")
	}

	if m.view == viewDetails {
		if !m.detailsLoading && m.detailsErr == nil {
			footer := ui.Dim("  enter") + " select" +
				"  " + ui.Dim("•  tab") + " devDependency: " + devLabel +
				"  " + ui.Dim("•  esc") + " back"
			b.WriteString(footer + "\n")
		} else {
			b.WriteString(ui.Dim("  esc") + " back\n")
		}
	} else {
		footer := ui.Dim("  tab") + " devDependency: " + devLabel +
			ui.Dim("  •  enter") + " details" +
			ui.Dim("  •  esc") + " cancel"
		b.WriteString(footer + "\n")
	}

	return tea.NewView(b.String())
}

func (m model) viewSearchPane() string {
	var b strings.Builder

	prompt := ui.StyleBold.Render("Search packages: ")
	b.WriteString("  " + prompt + m.textInput.View() + "\n\n")

	if m.textInput.Value() == "" {
		// Placeholder in the input already says "Type to search npm packages..."
	} else if m.loading {
		b.WriteString(ui.Dim("  Searching...") + "\n")
	} else if m.err != nil {
		b.WriteString("  " + ui.Warning(fmt.Sprintf("Search failed: %v", m.err)) + "\n")
	} else if len(m.results) == 0 && m.lastQuery != "" {
		b.WriteString(ui.Dim("  Aucun résultat.") + "\n")
	} else if len(m.results) > 0 {
		// Compute column widths from actual data
		pkgColWidth := len("Package")
		for _, r := range m.results {
			if len(r.Name) > pkgColWidth {
				pkgColWidth = len(r.Name)
			}
		}
		pkgColWidth += 2 // right padding

		const verColWidth = 10 // enough for "v19.2.4  "

		// indent(2) + cursor(2) + pkg + gap(2) + ver + gap(2) + desc
		descColWidth := m.width - 2 - 2 - pkgColWidth - 2 - verColWidth - 2
		if descColWidth < 15 {
			descColWidth = 15
		}

		// Header row (aligned with data rows: 4-char prefix for indent+cursor)
		b.WriteString("    " +
			ui.Dim(padRight("Package", pkgColWidth)) + "  " +
			ui.Dim(padRight("Latest", verColWidth)) + "  " +
			ui.Dim("Description") + "\n")

		// Separator
		sepWidth := 2 + pkgColWidth + 2 + verColWidth + 2 + descColWidth
		if sepWidth > m.width-2 {
			sepWidth = m.width - 2
		}
		b.WriteString("  " + ui.Dim(strings.Repeat("─", sepWidth)) + "\n")

		// Data rows
		for i, r := range m.results {
			cursor := "  "
			nameStyle := lipgloss.NewStyle().Foreground(ui.ColorPrimary)
			if i == m.cursor {
				cursor = ui.StyleSuccess.Render("❯ ")
				nameStyle = nameStyle.Bold(true)
			}

			name := nameStyle.Render(padRight(r.Name, pkgColWidth))
			version := ui.Dim(padRight("v"+r.Version, verColWidth))

			desc := r.Description
			if len(desc) > descColWidth {
				desc = desc[:descColWidth-1] + "…"
			}

			b.WriteString(fmt.Sprintf("  %s%s  %s  %s\n", cursor, name, version, ui.Dim(desc)))
		}
	}

	return b.String()
}

// padRight pads s with spaces on the right to reach width.
func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

func (m model) viewDetailsPane() string {
	var b strings.Builder

	pkgName := ""
	if m.cursor < len(m.results) {
		pkgName = m.results[m.cursor].Name
	}

	if m.detailsLoading {
		b.WriteString("  " + lipgloss.NewStyle().Foreground(ui.ColorPrimary).Bold(true).Render(pkgName) + "\n\n")
		b.WriteString(ui.Dim("  Loading package details...") + "\n")
		return b.String()
	}

	if m.detailsErr != nil {
		b.WriteString("  " + lipgloss.NewStyle().Foreground(ui.ColorPrimary).Bold(true).Render(pkgName) + "\n\n")
		b.WriteString("  " + ui.Warning(fmt.Sprintf("Failed to load: %v", m.detailsErr)) + "\n")
		return b.String()
	}

	d := m.details
	nameStyle := lipgloss.NewStyle().Foreground(ui.ColorPrimary).Bold(true)

	// ── Line 1: name  vLatest  License ──────────────────────────────────
	line1 := nameStyle.Render(d.Name) + "  " + ui.Dim("v"+d.Latest)
	if d.License != "" {
		line1 += "  " + ui.StyleInfo.Render(d.License)
	}
	b.WriteString("  " + line1 + "\n")

	// ── Line 2: description ──────────────────────────────────────────────
	if d.Description != "" {
		desc := d.Description
		maxLen := m.width - 4
		if maxLen < 20 {
			maxLen = 20
		}
		if len(desc) > maxLen {
			desc = desc[:maxLen-1] + "…"
		}
		b.WriteString("  " + ui.Dim(desc) + "\n")
	}

	// ── Line 3: author ───────────────────────────────────────────────────
	if d.Author != "" {
		b.WriteString("  " + ui.Dim("by "+d.Author) + "\n")
	}

	b.WriteString("\n")

	// ── Stats line: downloads · stars · last published · repo ────────────
	var stats []string
	if d.WeeklyDownloads > 0 {
		stats = append(stats, ui.Dim("downloads ")+FormatCount(d.WeeklyDownloads)+ui.Dim("/wk"))
	}
	if d.Stars > 0 {
		stats = append(stats, ui.Dim("stars ")+FormatCount(int64(d.Stars)))
	}
	if len(d.Versions) > 0 {
		stats = append(stats, ui.Dim("published ")+d.Versions[0].PublishedAt.Format("Jan 2, 2006"))
	}
	if len(stats) > 0 {
		b.WriteString("  " + strings.Join(stats, ui.Dim("  ·  ")) + "\n")
	}

	// ── Repo link ────────────────────────────────────────────────────────
	repoDisplay := d.Repository
	if repoDisplay == "" {
		repoDisplay = d.Homepage
	}
	if repoDisplay != "" {
		maxR := m.width - 4
		if maxR < 20 {
			maxR = 20
		}
		if len(repoDisplay) > maxR {
			repoDisplay = repoDisplay[:maxR-1] + "…"
		}
		b.WriteString("  " + ui.StyleInfo.Render(repoDisplay) + "\n")
	}

	b.WriteString("\n")

	// ── Separator ────────────────────────────────────────────────────────
	sepLen := m.width - 4
	if sepLen < 10 {
		sepLen = 10
	}
	b.WriteString("  " + ui.Dim(strings.Repeat("─", sepLen)) + "\n\n")

	// ── Version filter ───────────────────────────────────────────────────
	b.WriteString("  " + ui.StyleBold.Render("Filter: ") + m.versionInput.View() + "\n\n")

	// ── Version table ────────────────────────────────────────────────────
	if len(m.filteredVersions) == 0 {
		b.WriteString(ui.Dim("  Aucun résultat.\n"))
	} else {
		// Compute version column width from visible data (cap at 30)
		verColWidth := len("Version")
		start, end := scrollWindow(m.versionCursor, len(m.filteredVersions), maxVersionRows)
		for i := start; i < end; i++ {
			if l := len(m.filteredVersions[i].Version); l > verColWidth {
				verColWidth = l
			}
		}
		const maxVerColWidth = 30
		if verColWidth > maxVerColWidth {
			verColWidth = maxVerColWidth
		}
		verColWidth += 2

		// Header row (4-char prefix = indent + cursor placeholder)
		b.WriteString("    " +
			ui.Dim(padRight("Version", verColWidth)) + "  " +
			ui.Dim("Published") + "\n")
		sepWidth := 2 + verColWidth + 2 + 28 // date(12) + tags(~16)
		if sepWidth > m.width-2 {
			sepWidth = m.width - 2
		}
		b.WriteString("  " + ui.Dim(strings.Repeat("─", sepWidth)) + "\n")

		// Build inverse dist-tag map: version → []tag, with "latest" always first.
		versionTags := make(map[string][]string)
		for tag, ver := range d.DistTags {
			versionTags[ver] = append(versionTags[ver], tag)
		}
		for ver, tags := range versionTags {
			sort.Slice(tags, func(i, j int) bool {
				if tags[i] == "latest" {
					return true
				}
				if tags[j] == "latest" {
					return false
				}
				return tags[i] < tags[j]
			})
			versionTags[ver] = tags
		}

		// Data rows
		for i := start; i < end; i++ {
			v := m.filteredVersions[i]

			cursor := "  "
			vStyle := lipgloss.NewStyle()
			if i == m.versionCursor {
				cursor = ui.StyleSuccess.Render("❯ ")
				vStyle = vStyle.Bold(true)
			}

			// Truncate version if needed
			vStr := v.Version
			if len(vStr) > maxVerColWidth {
				vStr = vStr[:maxVerColWidth-1] + "…"
			}
			vStr = vStyle.Render(padRight(vStr, verColWidth))

			// Published date
			published := v.PublishedAt.Format("Jan 2, 2006")

			// Tags: dist-tags first (latest always first), then recent warning
			var tags []string
			for _, tag := range versionTags[v.Version] {
				tags = append(tags, ui.StyleSuccess.Render(tag))
			}
			age := time.Since(v.PublishedAt)
			if age < recentThreshold && age >= 0 {
				days := int(age.Hours() / 24)
				var ageStr string
				switch {
				case days == 0:
					ageStr = "today"
				case days == 1:
					ageStr = "1 day ago"
				default:
					ageStr = fmt.Sprintf("%d days ago", days)
				}
				tags = append(tags, ui.StyleWarning.Render("⚠  "+ageStr))
			}

			publishedCell := ui.Dim(published)
			if len(tags) > 0 {
				publishedCell += "  " + strings.Join(tags, "  ")
			}

			b.WriteString(fmt.Sprintf("  %s%s  %s\n", cursor, vStr, publishedCell))
		}

		if len(m.filteredVersions) > maxVersionRows {
			b.WriteString(ui.Dim(fmt.Sprintf("\n  %d versions total", len(m.filteredVersions))) + "\n")
		}
	}

	return b.String()
}

// scrollWindow returns the visible [start, end) window for a list.
func scrollWindow(cursor, total, windowSize int) (start, end int) {
	if total <= windowSize {
		return 0, total
	}
	start = cursor - windowSize/2
	if start < 0 {
		start = 0
	}
	end = start + windowSize
	if end > total {
		end = total
		start = end - windowSize
		if start < 0 {
			start = 0
		}
	}
	return start, end
}

// filterVersions returns versions whose version string contains the filter.
func filterVersions(versions []VersionInfo, filter string) []VersionInfo {
	if filter == "" {
		return versions
	}
	var result []VersionInfo
	for _, v := range versions {
		if strings.Contains(v.Version, filter) {
			result = append(result, v)
		}
	}
	return result
}

// doSearch fires an async npm registry search.
func doSearch(query string) tea.Cmd {
	return func() tea.Msg {
		results, err := Query(context.Background(), query, maxResults)
		if err != nil {
			return searchErrMsg{query: query, err: err}
		}
		return searchResultMsg{query: query, results: results}
	}
}

// doFetchDetails fires an async npm package details fetch.
func doFetchDetails(name string) tea.Cmd {
	return func() tea.Msg {
		details, err := FetchDetails(context.Background(), name)
		return detailsResultMsg{details: details, err: err}
	}
}
