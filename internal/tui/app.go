package tui

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/tui/theme"
	"github.com/fluid-movement/prep/internal/tui/ui"
)

// Options connect the TUI to a project without importing storage adapters.
type Options struct {
	// Load reads the project, the same way the CLI does.
	Load func() (*domain.Tree, error)
	// Watch is the directory to watch for changes (.prep); empty disables it.
	Watch string
}

// Run starts the issue views.
func Run(opts Options, in io.Reader, out io.Writer) error {
	m := NewModel(theme.New(lipgloss.NewRenderer(out)), opts)
	if opts.Watch != "" {
		ch, stop, err := watch(opts.Watch)
		if err != nil {
			m.err = fmt.Errorf("live reload is off: %v", err)
		} else {
			defer stop()
			m.changes = ch
		}
	}
	_, err := tea.NewProgram(m, tea.WithInput(in), tea.WithOutput(out), tea.WithAltScreen(), tea.WithMouseCellMotion()).Run()
	return err
}

// Layout constants, in cells.
const (
	listRatio   = 0.55
	minListW    = 52
	minDetailW  = 40
	headerLines = 1
	footerLines = 1
	noticeFor   = 2 * time.Second
)

type focus int

const (
	focusList focus = iota
	focusDetail
)

// tab is one saved view and its current result.
type tab struct {
	name  string
	flags string
	ids   []string
	err   error
}

// Model is the issue views screen. It owns all state; looks come from ui.
type Model struct {
	th      *theme.Theme
	opts    Options
	tree    *domain.Tree
	err     error // last load error; the last good tree stays shown
	tabs    []tab
	active  int
	cursor  map[string]int // per tab name
	offset  map[string]int // first visible row per tab name
	focus   focus
	w, h    int
	vp      viewport.Model
	docKey  string // issue, width and generation of the rendered detail
	gen     int    // increases on every successful load
	notice  string
	changes <-chan struct{}
}

// NewModel loads the project once and builds the screen state.
func NewModel(th *theme.Theme, opts Options) *Model {
	m := &Model{th: th, opts: opts, cursor: map[string]int{}, offset: map[string]int{}, vp: viewport.New(0, 0)}
	m.apply(opts.Load())
	return m
}

type loadedMsg struct {
	tree *domain.Tree
	err  error
}
type changedMsg struct{}
type clearNoticeMsg struct{ notice string }

func (m *Model) Init() tea.Cmd { return m.waitForChange() }

func (m *Model) waitForChange() tea.Cmd {
	if m.changes == nil {
		return nil
	}
	ch := m.changes
	return func() tea.Msg {
		if _, ok := <-ch; !ok {
			return nil
		}
		return changedMsg{}
	}
}

func (m *Model) reload() tea.Cmd {
	return func() tea.Msg {
		t, err := m.opts.Load()
		return loadedMsg{t, err}
	}
}

// apply installs a load result, keeping the selection by issue ID.
func (m *Model) apply(t *domain.Tree, err error) {
	if err != nil {
		m.err = err
		return
	}
	selected := map[string]string{}
	for _, tb := range m.tabs {
		if c := m.cursor[tb.name]; c < len(tb.ids) {
			selected[tb.name] = tb.ids[c]
		}
	}
	m.tree, m.err = t, nil
	m.gen++
	m.tabs = buildTabs(t)
	if m.active >= len(m.tabs) {
		m.active = 0
	}
	for _, tb := range m.tabs {
		c := m.cursor[tb.name]
		if id, ok := selected[tb.name]; ok {
			for k, x := range tb.ids {
				if x == id {
					c = k
				}
			}
		}
		m.cursor[tb.name] = clamp(c, 0, len(tb.ids)-1)
	}
}

func buildTabs(t *domain.Tree) []tab {
	names := domain.ViewNames(t.Project.Config)
	if len(names) == 0 {
		names = []string{"All"}
	}
	var tabs []tab
	for _, n := range names {
		flags := t.Project.Config.Views[n]
		tb := tab{name: n, flags: flags}
		f, err := domain.ParseFilter(strings.Fields(flags))
		if err == nil {
			tb.ids, err = t.Query(f)
		}
		tb.err = err
		tabs = append(tabs, tb)
	}
	return tabs
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		return m, nil
	case changedMsg:
		return m, tea.Batch(m.reload(), m.waitForChange())
	case loadedMsg:
		m.apply(msg.tree, msg.err)
		return m, nil
	case clearNoticeMsg:
		if m.notice == msg.notice {
			m.notice = ""
		}
		return m, nil
	case tea.KeyMsg:
		return m, m.key(msg)
	case tea.MouseMsg:
		if m.focus == focusDetail {
			var cmd tea.Cmd
			m.vp, cmd = m.vp.Update(msg)
			return m, cmd
		}
	}
	return m, nil
}

func (m *Model) key(k tea.KeyMsg) tea.Cmd {
	tb := m.current()
	switch s := k.String(); s {
	case "q", "ctrl+c":
		return tea.Quit
	case "tab":
		m.switchTab(m.active + 1)
	case "shift+tab":
		m.switchTab(m.active - 1)
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		if n := int(s[0] - '1'); n < len(m.tabs) {
			m.switchTab(n)
		}
	case "enter", "right", "l":
		if m.selected() != "" {
			m.focus = focusDetail
		}
	case "esc", "left", "h":
		m.focus = focusList
	case "r":
		return m.reload()
	case "y":
		if id := m.selected(); id != "" {
			m.th.R.Output().Copy(id)
			return m.flash("copied " + id)
		}
	default:
		if m.focus == focusDetail {
			var cmd tea.Cmd
			m.vp, cmd = m.vp.Update(k)
			return cmd
		}
		if tb == nil {
			return nil
		}
		c := m.cursor[tb.name]
		switch s {
		case "down", "j":
			c++
		case "up", "k":
			c--
		case "pgdown", "ctrl+d":
			c += m.listHeight()
		case "pgup", "ctrl+u":
			c -= m.listHeight()
		case "home", "g":
			c = 0
		case "end", "G":
			c = len(tb.ids) - 1
		}
		m.cursor[tb.name] = clamp(c, 0, len(tb.ids)-1)
	}
	return nil
}

func (m *Model) flash(n string) tea.Cmd {
	m.notice = n
	return tea.Tick(noticeFor, func(time.Time) tea.Msg { return clearNoticeMsg{n} })
}

func (m *Model) switchTab(n int) {
	if len(m.tabs) == 0 {
		return
	}
	m.active = (n + len(m.tabs)) % len(m.tabs)
	m.focus = focusList
}

func (m *Model) current() *tab {
	if m.active < len(m.tabs) {
		return &m.tabs[m.active]
	}
	return nil
}

func (m *Model) selected() string {
	tb := m.current()
	if tb == nil || len(tb.ids) == 0 {
		return ""
	}
	return tb.ids[clamp(m.cursor[tb.name], 0, len(tb.ids)-1)]
}

// --- view ---

func (m *Model) bodyHeight() int { return ui.Stack(m.h, headerLines, footerLines) }

func (m *Model) listHeight() int { return max(1, m.bodyHeight()-2) }

func (m *Model) View() string {
	if m.w == 0 || m.h == 0 {
		return ui.Loading(m.th, "Starting …")
	}
	header := m.header()
	footer := m.footer()
	bodyH := m.bodyHeight()
	var body string
	switch {
	case m.tree == nil:
		body = ui.Empty(m.th, "Could not load .prep", fmt.Sprint(m.err), m.w, bodyH)
	default:
		listW, detailW := ui.Split(m.w, listRatio, minListW, minDetailW)
		switch {
		case detailW == 0 && m.focus == focusDetail:
			body = m.detailPane(m.w, bodyH)
		case detailW == 0:
			body = m.listPane(m.w, bodyH)
		default:
			body = lipgloss.JoinHorizontal(lipgloss.Top, m.listPane(listW, bodyH), m.detailPane(detailW, bodyH))
		}
	}
	return header + "\n" + body + "\n" + footer
}

func (m *Model) header() string {
	name := m.th.S.Title.Render("prep") + " "
	var tabs []ui.Tab
	for _, tb := range m.tabs {
		tabs = append(tabs, ui.Tab{Label: tb.name, Count: len(tb.ids)})
	}
	return ui.Fit(name+ui.Tabs(m.th, tabs, m.active, m.w-lipgloss.Width(name)), m.w)
}

var (
	listKeys   = []ui.Key{{Keys: "tab/1-9", Desc: "views"}, {Keys: "↑↓", Desc: "move"}, {Keys: "enter", Desc: "details"}, {Keys: "y", Desc: "copy id"}, {Keys: "q", Desc: "quit"}}
	detailKeys = []ui.Key{{Keys: "↑↓ pgup/pgdn", Desc: "scroll"}, {Keys: "esc", Desc: "back"}, {Keys: "tab/1-9", Desc: "views"}, {Keys: "y", Desc: "copy id"}, {Keys: "q", Desc: "quit"}}
)

func (m *Model) footer() string {
	switch {
	case m.notice != "":
		return ui.Fit(ui.Note(m.th, m.notice, ui.ToneAccent), m.w)
	case m.err != nil:
		return ui.Fit(ui.Error(m.th, m.err.Error()), m.w)
	case m.focus == focusDetail:
		return ui.KeyHelp(m.th, detailKeys, m.w)
	}
	return ui.KeyHelp(m.th, listKeys, m.w)
}

func (m *Model) listPane(w, h int) string {
	tb := m.current()
	title := "Issues"
	if tb != nil {
		title = tb.name
		if tb.flags != "" {
			title += "  " + tb.flags
		}
	}
	inner := w - 2 - 2*theme.Pad
	rows := h - 2
	var body string
	switch {
	case tb == nil:
	case tb.err != nil:
		body = ui.Error(m.th, tb.err.Error())
	case len(tb.ids) == 0:
		body = ui.Empty(m.th, "No issues in this view", "Issues appear when they match "+orAll(tb.flags), inner, rows)
	default:
		c := m.cursor[tb.name]
		off := clamp(m.offset[tb.name], 0, max(0, len(tb.ids)-rows))
		if c < off {
			off = c
		}
		if c >= off+rows {
			off = c - rows + 1
		}
		m.offset[tb.name] = off
		var lines []string
		for k := off; k < len(tb.ids) && k < off+rows; k++ {
			lines = append(lines, ui.ListRow(m.th, m.row(tb.ids[k]), k == c, inner))
		}
		body = strings.Join(lines, "\n")
	}
	return ui.Pane{Title: title, Body: body, Focused: m.focus == focusList, Width: w, Height: h}.View(m.th)
}

func orAll(flags string) string {
	if flags == "" {
		return "any issue"
	}
	return flags
}

func (m *Model) row(id string) ui.Row {
	t := m.tree
	i := t.Issues[id]
	r := ui.Row{ID: shortID(id), State: t.State(id), Kind: i.Kind, Title: i.Title}
	switch {
	case len(t.Children(id)) > 0:
		p := t.ChildProgress(id)
		r.Note = ui.Progress(m.th, p.Done+p.Dropped, p.Total)
	case t.Stale(id):
		r.Note = ui.Note(m.th, "stale", ui.ToneError)
	case t.Blocked(id):
		r.Note = ui.Note(m.th, "blocked", ui.ToneWarning)
	}
	return r
}

// shortID shows the time part of an ID; any unique suffix resolves in the CLI.
func shortID(id string) string {
	if k := strings.LastIndexByte(id, '-'); k >= 0 {
		return id[k+1:]
	}
	return id
}

func (m *Model) detailPane(w, h int) string {
	id := m.selected()
	inner := w - 2 - 2*theme.Pad
	if id == "" {
		return ui.Pane{Title: "Details", Body: ui.Empty(m.th, "Nothing selected", "", inner, h-2), Focused: m.focus == focusDetail, Width: w, Height: h}.View(m.th)
	}
	key := fmt.Sprintf("%s|%d|%d", id, inner, m.gen)
	if key != m.docKey {
		doc, err := ui.Markdown(m.th, detailMarkdown(m.tree, id), inner)
		if err != nil {
			doc = ui.Error(m.th, err.Error())
		}
		sameIssue := strings.HasPrefix(m.docKey, id+"|")
		m.vp.SetContent(doc)
		if !sameIssue {
			m.vp.GotoTop()
		}
		m.docKey = key
	}
	m.vp.Width, m.vp.Height = inner, h-2
	title := shortID(id) + " " + m.tree.Issues[id].Title
	if pct := m.vp.ScrollPercent(); m.vp.TotalLineCount() > m.vp.Height {
		title += fmt.Sprintf("  %d%%", int(pct*100))
	}
	return ui.Pane{Title: title, Body: m.vp.View(), Focused: m.focus == focusDetail, Width: w, Height: h}.View(m.th)
}

func clamp(v, lo, hi int) int {
	if hi < lo {
		return lo
	}
	return min(max(v, lo), hi)
}
