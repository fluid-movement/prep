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
	listRatio    = 0.55
	minListW     = 52
	minDetailW   = 40
	headerLines  = 1
	footerLines  = 1
	maxRelations = 6
	noticeFor    = 2 * time.Second
)

type focus int

const (
	focusList focus = iota
	focusDetail
)

// row is one displayed list row.
type row struct {
	id      string
	context bool // shown for structure, not part of the view's result
	prefix  string
}

// tab is one saved view and its current rows.
type tab struct {
	name  string
	flags string
	count int // matching issues in the whole view, ignoring parent focus
	rows  []row
	err   error
}

// relation is one navigable link in the detail pane.
type relation struct {
	label string
	id    string
}

// Model is the issue views screen. It owns all state; looks come from ui.
type Model struct {
	th       *theme.Theme
	opts     Options
	tree     *domain.Tree
	err      error // last load error; the last good tree stays shown
	tabs     []tab
	active   int
	treeMode bool
	scope    map[string][]string // per tab name: stack of focused parents
	cursor   map[string]int      // per tab name
	offset   map[string]int      // first visible row per tab name
	focus    focus
	w, h     int
	vp       viewport.Model
	docKey   string // issue, width and generation of the rendered detail
	gen      int    // increases on every successful load
	relFor   string // issue whose relations relCursor refers to
	relIdx   int
	back     []string // previously shown issues, for backspace
	notice   string
	changes  <-chan struct{}
}

// NewModel loads the project once and builds the screen state.
func NewModel(th *theme.Theme, opts Options) *Model {
	m := &Model{th: th, opts: opts, treeMode: true, scope: map[string][]string{}, cursor: map[string]int{}, offset: map[string]int{}, vp: viewport.New(0, 0)}
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

// apply installs a load result; a failed load keeps the last good tree.
func (m *Model) apply(t *domain.Tree, err error) {
	if err != nil {
		m.err = err
		return
	}
	m.tree, m.err = t, nil
	m.gen++
	m.rebuild()
}

// rebuild recomputes every tab from the tree, keeping each tab's selection
// by issue ID and dropping focused parents that no longer exist.
func (m *Model) rebuild() {
	selected := map[string]string{}
	for _, tb := range m.tabs {
		if c := m.cursor[tb.name]; c < len(tb.rows) {
			selected[tb.name] = tb.rows[c].id
		}
	}
	names := domain.ViewNames(m.tree.Project.Config)
	if len(names) == 0 {
		names = []string{"All"}
	}
	m.tabs = m.tabs[:0]
	for _, n := range names {
		var sc []string
		for _, id := range m.scope[n] {
			if m.tree.Issues[id] != nil {
				sc = append(sc, id)
			}
		}
		m.scope[n] = sc
		tb := buildTab(m.tree, n, m.tree.Project.Config.Views[n], sc, m.treeMode)
		c := m.cursor[n]
		if id, ok := selected[n]; ok {
			if k := tb.index(id); k >= 0 {
				c = k
			}
		}
		m.cursor[n] = clamp(c, 0, len(tb.rows)-1)
		m.tabs = append(m.tabs, tb)
	}
	if m.active >= len(m.tabs) {
		m.active = 0
	}
}

// buildTab queries a view, narrows it to the focused parent's subtree, and
// lays it out flat or as a tree with context ancestors.
func buildTab(t *domain.Tree, name, flags string, scope []string, treeMode bool) tab {
	tb := tab{name: name, flags: flags}
	f, err := domain.ParseFilter(strings.Fields(flags))
	var ids []string
	if err == nil {
		ids, err = t.Query(f)
	}
	if err != nil {
		tb.err = err
		return tb
	}
	tb.count = len(ids)

	root := ""
	inScope := map[string]bool{}
	if len(scope) > 0 {
		root = scope[len(scope)-1]
		inScope[root] = true
		for _, d := range t.Descendants(root) {
			inScope[d] = true
		}
		var kept []string
		for _, id := range ids {
			if inScope[id] {
				kept = append(kept, id)
			}
		}
		ids = kept
	}
	matched := map[string]bool{}
	for _, id := range ids {
		matched[id] = true
	}

	var list []string
	if treeMode {
		all := t.WithAncestors(ids)
		if root != "" {
			all = append(all, root)
			var kept []string
			for _, id := range all {
				if inScope[id] {
					kept = append(kept, id)
				}
			}
			all = kept
		}
		list = t.TreeOrder(all)
	} else {
		if root != "" && !matched[root] {
			list = append(list, root)
		}
		list = append(list, ids...)
	}

	inList := map[string]bool{}
	for _, id := range list {
		inList[id] = true
	}
	depths := make([]int, len(list))
	if treeMode {
		for k, id := range list {
			for _, a := range t.Ancestors(id) {
				if inList[a] {
					depths[k]++
				}
			}
		}
	}
	prefixes := ui.TreePrefixes(depths)
	for k, id := range list {
		tb.rows = append(tb.rows, row{id: id, context: !matched[id], prefix: prefixes[k]})
	}
	return tb
}

func (tb *tab) index(id string) int {
	for k, r := range tb.rows {
		if r.id == id {
			return k
		}
	}
	return -1
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
	s := k.String()
	// Keys that work in both panes.
	switch s {
	case "q", "ctrl+c":
		return tea.Quit
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		if n := int(s[0] - '1'); n < len(m.tabs) {
			m.switchTab(n)
		}
		return nil
	case "r":
		return m.reload()
	case "y":
		if id := m.selected(); id != "" {
			m.th.R.Output().Copy(id)
			return m.flash("copied " + id)
		}
		return nil
	case "t":
		if m.tree == nil {
			return nil
		}
		m.treeMode = !m.treeMode
		m.rebuild()
		if m.treeMode {
			return m.flash("tree view")
		}
		return m.flash("flat view")
	case "p":
		return m.jumpToParent()
	}
	if m.focus == focusDetail {
		return m.detailKey(s, k)
	}
	return m.listKey(s)
}

func (m *Model) listKey(s string) tea.Cmd {
	tb := m.current()
	switch s {
	case "tab":
		m.switchTab(m.active + 1)
		return nil
	case "shift+tab":
		m.switchTab(m.active - 1)
		return nil
	case "enter":
		if m.selected() != "" {
			m.focus = focusDetail
		}
		return nil
	case "right", "l":
		id := m.selected()
		if id != "" && len(m.tree.Children(id)) > 0 && m.scopeTop() != id {
			m.scope[tb.name] = append(m.scope[tb.name], id)
			m.rebuild()
			m.cursor[tb.name] = 0
			return nil
		}
		if id != "" {
			m.focus = focusDetail
		}
		return nil
	case "left", "h", "esc":
		if sc := m.scope[tb.name]; len(sc) > 0 {
			up := sc[len(sc)-1]
			m.scope[tb.name] = sc[:len(sc)-1]
			m.rebuild()
			m.selectInCurrent(up)
		}
		return nil
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
		c = len(tb.rows) - 1
	}
	m.cursor[tb.name] = clamp(c, 0, len(tb.rows)-1)
	return nil
}

func (m *Model) detailKey(s string, k tea.KeyMsg) tea.Cmd {
	rels := m.relations(m.selected())
	switch s {
	case "esc", "left", "h":
		m.focus = focusList
		return nil
	case "tab", "shift+tab":
		if len(rels) > 0 {
			step := 1
			if s == "shift+tab" {
				step = -1
			}
			m.relIdx = (m.relIdx + step + len(rels)) % len(rels)
		}
		return nil
	case "enter":
		if len(rels) > 0 {
			return m.jump(rels[clamp(m.relIdx, 0, len(rels)-1)].id, true)
		}
		return nil
	case "backspace":
		if n := len(m.back); n > 0 {
			id := m.back[n-1]
			m.back = m.back[:n-1]
			return m.jump(id, false)
		}
		return m.flash("no earlier issue")
	}
	var cmd tea.Cmd
	m.vp, cmd = m.vp.Update(k)
	return cmd
}

func (m *Model) jumpToParent() tea.Cmd {
	id := m.selected()
	if id == "" {
		return nil
	}
	p := m.tree.Issues[id].Parent
	if p == "" || m.tree.Issues[p] == nil {
		return m.flash("top-level issue: no parent")
	}
	return m.jump(p, true)
}

// jump selects an issue: in the current tab when it is a row there,
// otherwise in the first view that contains it (views without flags first),
// clearing that view's parent focus.
func (m *Model) jump(id string, remember bool) tea.Cmd {
	from := m.selected()
	if !m.selectInCurrent(id) {
		var order []int
		for k, tb := range m.tabs {
			if tb.flags == "" {
				order = append(order, k)
			}
		}
		for k, tb := range m.tabs {
			if tb.flags != "" {
				order = append(order, k)
			}
		}
		found := false
		for _, k := range order {
			name := m.tabs[k].name
			saved := m.scope[name]
			m.scope[name] = nil
			m.rebuild()
			if i := m.tabs[k].index(id); i >= 0 {
				m.active = k
				m.cursor[name] = i
				found = true
				break
			}
			m.scope[name] = saved
			m.rebuild()
		}
		if !found {
			return m.flash("not in any view")
		}
	}
	if remember && from != "" && from != id {
		m.back = append(m.back, from)
	}
	return nil
}

func (m *Model) selectInCurrent(id string) bool {
	tb := m.current()
	if tb == nil {
		return false
	}
	if k := tb.index(id); k >= 0 {
		m.cursor[tb.name] = k
		return true
	}
	return false
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

func (m *Model) scopeTop() string {
	tb := m.current()
	if tb == nil {
		return ""
	}
	if sc := m.scope[tb.name]; len(sc) > 0 {
		return sc[len(sc)-1]
	}
	return ""
}

func (m *Model) selected() string {
	tb := m.current()
	if tb == nil || len(tb.rows) == 0 {
		return ""
	}
	return tb.rows[clamp(m.cursor[tb.name], 0, len(tb.rows)-1)].id
}

// relations lists the issues the detail pane links to, in a fixed order.
func (m *Model) relations(id string) []relation {
	if m.tree == nil || m.tree.Issues[id] == nil {
		return nil
	}
	i := m.tree.Issues[id]
	var out []relation
	if i.Parent != "" && m.tree.Issues[i.Parent] != nil {
		out = append(out, relation{"parent", i.Parent})
	}
	for _, c := range m.tree.Children(id) {
		out = append(out, relation{"child", c})
	}
	for _, d := range i.DependsOn {
		if m.tree.Issues[d] != nil {
			out = append(out, relation{"depends on", d})
		}
	}
	for _, b := range m.tree.Blocks(id) {
		out = append(out, relation{"blocks", b})
	}
	return out
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
		tabs = append(tabs, ui.Tab{Label: tb.name, Count: tb.count})
	}
	return ui.Fit(name+ui.Tabs(m.th, tabs, m.active, m.w-lipgloss.Width(name)), m.w)
}

var (
	listKeys   = []ui.Key{{Keys: "↑↓", Desc: "move"}, {Keys: "enter", Desc: "details"}, {Keys: "→/←", Desc: "into/out of parent"}, {Keys: "p", Desc: "parent"}, {Keys: "t", Desc: "tree/flat"}, {Keys: "tab/1-9", Desc: "views"}, {Keys: "y", Desc: "copy id"}, {Keys: "q", Desc: "quit"}}
	detailKeys = []ui.Key{{Keys: "tab", Desc: "next link"}, {Keys: "enter", Desc: "open"}, {Keys: "⌫", Desc: "back"}, {Keys: "↑↓", Desc: "scroll"}, {Keys: "esc", Desc: "list"}, {Keys: "p", Desc: "parent"}, {Keys: "y", Desc: "copy id"}, {Keys: "q", Desc: "quit"}}
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

func (m *Model) listTitle(tb *tab) string {
	parts := []string{tb.name}
	for _, id := range m.scope[tb.name] {
		parts = append(parts, ui.Fit(m.tree.Issues[id].Title, 24))
	}
	title := strings.Join(parts, " › ")
	if len(parts) == 1 && tb.flags != "" {
		title += "  " + tb.flags
	}
	if !m.treeMode {
		title += "  (flat)"
	}
	return title
}

func (m *Model) listPane(w, h int) string {
	tb := m.current()
	title := "Issues"
	if tb != nil {
		title = m.listTitle(tb)
	}
	inner := w - 2 - 2*theme.Pad
	rows := h - 2
	var body string
	switch {
	case tb == nil:
	case tb.err != nil:
		body = ui.Error(m.th, tb.err.Error())
	case len(tb.rows) == 0:
		body = ui.Empty(m.th, "No issues in this view", "Issues appear when they match "+orAll(tb.flags), inner, rows)
	default:
		c := m.cursor[tb.name]
		off := clamp(m.offset[tb.name], 0, max(0, len(tb.rows)-rows))
		if c < off {
			off = c
		}
		if c >= off+rows {
			off = c - rows + 1
		}
		m.offset[tb.name] = off
		var lines []string
		for k := off; k < len(tb.rows) && k < off+rows; k++ {
			lines = append(lines, ui.ListRow(m.th, m.row(tb.rows[k]), k == c, inner))
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

func (m *Model) row(r row) ui.Row {
	out := m.issueRow(r.id)
	out.Tree, out.Dimmed = r.prefix, r.context
	return out
}

func (m *Model) issueRow(id string) ui.Row {
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
	focused := m.focus == focusDetail
	if id == "" {
		return ui.Pane{Title: "Details", Body: ui.Empty(m.th, "Nothing selected", "", inner, h-2), Focused: focused, Width: w, Height: h}.View(m.th)
	}

	// Relations: a fixed block of rows above the scrolling document.
	rels := m.relations(id)
	if m.relFor != id {
		m.relFor, m.relIdx = id, 0
	}
	m.relIdx = clamp(m.relIdx, 0, len(rels)-1)
	var block []string
	if len(rels) > 0 {
		hint := ""
		if focused {
			hint = " · tab to move, enter to open"
		}
		block = append(block, m.th.S.Muted.Render(fmt.Sprintf("Related %d/%d", m.relIdx+1, len(rels)))+m.th.S.Subtle.Render(hint))
		start := clamp(m.relIdx-maxRelations/2, 0, max(0, len(rels)-maxRelations))
		for k := start; k < len(rels) && k < start+maxRelations; k++ {
			rid := rels[k].id
			block = append(block, ui.LinkRow(m.th, rels[k].label, m.tree.State(rid), shortID(rid), m.tree.Issues[rid].Title, focused && k == m.relIdx, inner))
		}
		block = append(block, "")
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
	m.vp.Width, m.vp.Height = inner, max(1, h-2-len(block))
	title := shortID(id) + " " + m.tree.Issues[id].Title
	if m.vp.TotalLineCount() > m.vp.Height {
		title += fmt.Sprintf("  %d%%", int(m.vp.ScrollPercent()*100))
	}
	body := strings.Join(append(block, m.vp.View()), "\n")
	return ui.Pane{Title: title, Body: body, Focused: focused, Width: w, Height: h}.View(m.th)
}

func clamp(v, lo, hi int) int {
	if hi < lo {
		return lo
	}
	return min(max(v, lo), hi)
}
