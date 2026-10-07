package tui

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"slices"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/palette"
	"github.com/fluid-movement/prep/internal/tui/theme"
	"github.com/fluid-movement/prep/internal/tui/ui"
	"github.com/fluid-movement/prep/internal/watch"
)

// Options connect the TUI to a project without importing storage adapters.
type Options struct {
	// Load reads the project, the same way the CLI does.
	Load func() (*domain.Tree, error)
	// Watch is the directory to watch for changes (.prep); empty disables it.
	Watch string
	// Check returns every diagnostic, including knowledge drift. It is slower
	// than Load (git), so it runs only for the check screen.
	Check func() ([]domain.Diagnostic, error)
	// Write runs a planned change through the CLI's write pipeline (plan,
	// CheckWrite, store, stage) and returns the changed issue's ID. Nil
	// makes the TUI read-only.
	Write func(plan func(*domain.Tree) (*domain.Change, error)) (string, error)
	// Actor records who writes, human:<name> for the TUI.
	Actor string
	// NoMouse starts with mouse capture off (the user's tui.mouse choice).
	NoMouse bool
	// SaveMouse records the user's mouse capture choice; nil keeps a
	// toggle for this session only.
	SaveMouse func(on bool) error
}

// Run starts the issue views.
func Run(opts Options, in io.Reader, out io.Writer) error {
	m := NewModel(theme.New(true, colorprofile.Detect(out, os.Environ())), opts)
	if opts.Watch != "" {
		ch, stop, err := watch.Dir(opts.Watch)
		if err != nil {
			m.err = fmt.Errorf("live reload is off: %v", err)
		} else {
			defer stop()
			m.changes = ch
		}
	}
	_, err := tea.NewProgram(m, tea.WithInput(in), tea.WithOutput(out)).Run()
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

type screen int

const (
	screenIssues screen = iota
	screenCheck
	screenSettings
	screenKnowledge
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
	docKey   string   // issue, width and generation of the rendered detail
	gen      int      // increases on every successful load
	back     []string // previously shown issues, for backspace
	notice   string
	changes  <-chan struct{}

	filters   map[string]string // per tab name: the applied filter text
	filtering bool              // the filter bar has the keyboard
	input     textinput.Model
	filterErr string
	flagInfo  []domain.FlagInfo // query flags and their values, read when the filter bar opens
	comp      completion        // candidates for the word being typed
	pick      int               // highlighted candidate

	screen   screen
	page     viewport.Model // check and settings screens
	diags    []domain.Diagnostic
	diagErr  error
	diagDone bool

	modal         *modal
	know          knowState
	confirm       string            // a pending second key press: delete|view in settings
	moving        bool              // settings: the selected view moves with j/k
	setIdx        int               // selected settings row: the theme, the views, then the mouse
	termLight     bool              // the terminal reported a light background
	pending       map[string]string // issue|field: edited text a rejected write left
	pendingSelect string            // issue to select after the next load
	noticeTone    ui.Tone
	mouse         bool              // capture the mouse: clicks and the wheel act
	wheeled       bool              // the wheel scrolled a list: its view stops following the selection until a key
	hits          []*lipgloss.Layer // clickable regions of the last render, by ID
	panes         []*lipgloss.Layer // the panes of the last render, for the wheel
	at            point             // where the pane being rendered starts
	runEditor     func(path string, done func(error) tea.Msg) tea.Cmd
	after         func(time.Duration, func(time.Time) tea.Msg) tea.Cmd // tea.Tick; tests drop timers
}

// NewModel loads the project once and builds the screen state.
func NewModel(th *theme.Theme, opts Options) *Model {
	in := textinput.New()
	in.Prompt = "/ "
	in.Placeholder = "--kind code --state open or words in titles"
	in.SetStyles(inputStyles(th))
	m := &Model{th: th, opts: opts, treeMode: true, scope: map[string][]string{}, cursor: map[string]int{}, offset: map[string]int{},
		vp: newViewport(), page: newViewport(), know: knowState{vp: newViewport()}, filters: map[string]string{}, input: in,
		pending: map[string]string{}, mouse: !opts.NoMouse, runEditor: execEditor, after: tea.Tick}
	m.apply(opts.Load())
	return m
}

// newViewport scrolls vertically only: everything shown in a viewport is
// wrapped to its width, so left and right keep their navigation meaning.
func newViewport() viewport.Model {
	vp := viewport.New()
	vp.SetHorizontalStep(0)
	return vp
}

type loadedMsg struct {
	tree *domain.Tree
	err  error
}
type changedMsg struct{}
type clearNoticeMsg struct{ notice string }
type checkedMsg struct {
	diags []domain.Diagnostic
	err   error
}

func (m *Model) Init() tea.Cmd { return tea.Batch(tea.RequestBackgroundColor, m.waitForChange()) }

// retheme follows the terminal's reported background (it picks the theme
// only when none is configured) and color profile.
func (m *Model) retheme(dark bool, profile colorprofile.Profile) {
	m.termLight = !dark
	if profile != m.th.Profile {
		m.setTheme(m.th.Rebuild(profile))
	}
	m.syncTheme()
}

// syncTheme shows the configured theme, or the default for the terminal's
// background: a new choice or an edited custom theme re-renders
// everything. An invalid theme keeps the current one; the check screen
// reports it.
func (m *Model) syncTheme() {
	var cfg domain.Config
	if m.tree != nil {
		cfg = m.tree.Project.Config
	}
	p, err := palette.Resolve(palette.Pick(cfg.Theme, !m.termLight), cfg.Themes)
	if err != nil || reflect.DeepEqual(p, m.th.Palette) {
		return
	}
	m.setTheme(theme.From(p, m.th.Profile))
}

// setTheme replaces the theme in place; components hold the same pointer,
// inputs copy their styles, and cached renders start over.
func (m *Model) setTheme(th *theme.Theme) {
	*m.th = *th
	m.input.SetStyles(inputStyles(m.th))
	m.docKey, m.know.docKey = "", ""
}

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

func (m *Model) runCheck() tea.Cmd {
	if m.opts.Check == nil {
		return nil
	}
	return func() tea.Msg {
		ds, err := m.opts.Check()
		return checkedMsg{ds, err}
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
	m.syncTheme()
	m.rebuild()
	m.syncKnowledge()
	if id := m.pendingSelect; id != "" && t.Issues[id] != nil {
		m.pendingSelect = ""
		m.jump(id, false)
	}
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
		tb := buildTab(m.tree, n, m.tree.Project.Config.Views[n], m.filters[n], sc, m.treeMode)
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
func buildTab(t *domain.Tree, name, flags, filter string, scope []string, treeMode bool) tab {
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
	if filter != "" {
		ff, err := parseFilterText(filter)
		var keep []string
		if err == nil {
			keep, err = t.Query(ff)
		}
		if err != nil {
			tb.err = err
			return tb
		}
		in := map[string]bool{}
		for _, id := range keep {
			in[id] = true
		}
		var kept []string
		for _, id := range ids {
			if in[id] {
				kept = append(kept, id)
			}
		}
		ids = kept
	}

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

// parseFilterText reads filter bar input: prep list flags, with bare words
// joined into one --text phrase (repeated --text flags would OR).
func parseFilterText(s string) (domain.Filter, error) {
	var args, words []string
	fields := strings.Fields(s)
	for k := 0; k < len(fields); k++ {
		f := fields[k]
		if !strings.HasPrefix(f, "--") {
			words = append(words, f)
			continue
		}
		args = append(args, f)
		name := strings.TrimPrefix(f, "--")
		kind := domain.FlagKind(name)
		takesValue := !strings.Contains(name, "=") && (kind == domain.FlagList || kind == domain.FlagText)
		if takesValue && k+1 < len(fields) {
			k++
			args = append(args, fields[k])
		}
	}
	if len(words) > 0 {
		args = append(args, "--text", strings.Join(words, " "))
	}
	return domain.ParseFilter(args)
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
		if (m.screen == screenCheck || m.screen == screenKnowledge) && msg.err == nil {
			return m, m.runCheck()
		}
		return m, nil
	case checkedMsg:
		m.diags, m.diagErr, m.diagDone = msg.diags, msg.err, true
		return m, nil
	case wroteMsg:
		return m, m.handleWrote(msg)
	case editedMsg:
		return m, m.handleEdited(msg)
	case areaEditedMsg:
		return m, m.handleAreaEdited(msg)
	case clearNoticeMsg:
		if m.notice == msg.notice {
			m.notice = ""
		}
		return m, nil
	case tea.BackgroundColorMsg:
		m.retheme(msg.IsDark(), m.th.Profile)
		return m, nil
	case tea.ColorProfileMsg:
		m.retheme(m.th.Dark, msg.Profile)
		return m, nil
	case tea.PasteMsg:
		return m, m.paste(msg)
	case tea.KeyPressMsg:
		m.wheeled = false // keys move the selection, so the view follows it again
		if m.modal != nil {
			m.modal.scrolled = false
			return m, m.modalKey(msg)
		}
		if m.filtering {
			return m, m.filterKey(msg)
		}
		return m, m.key(msg)
	case tea.MouseMsg:
		return m, m.mouseMsg(msg)
	}
	return m, nil
}

func (m *Model) key(k tea.KeyPressMsg) tea.Cmd {
	s := k.String()
	if s == "?" {
		return m.openHelp()
	}
	// A second press confirms; any other key cancels the question.
	if m.confirm != "" && !(m.screen == screenSettings && s == "d") {
		m.confirm, m.notice = "", ""
	}
	// The settings screen takes digits for its rows (view n is tab n) and
	// t/T for the theme.
	if m.screen == screenSettings && m.tree != nil && (len(s) == 1 && s >= "1" && s <= "9" || s == "t" || s == "T") {
		return m.settingsKey(s)
	}
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
			return tea.Batch(tea.SetClipboard(id), m.flash("copied "+id))
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
	case "b":
		if m.tree == nil {
			return nil
		}
		if m.screen == screenKnowledge {
			m.screen = screenIssues
			return nil
		}
		return m.openKnowledge("")
	case "c", "s":
		target := screenCheck
		if s == "s" {
			target = screenSettings
		}
		if m.screen == target {
			m.screen = screenIssues
			return nil
		}
		m.screen = target
		m.page.GotoTop()
		if target == screenCheck {
			return m.runCheck()
		}
		return nil
	}
	if m.screen != screenIssues {
		if s == "esc" && !m.moving && m.screen != screenKnowledge {
			m.screen = screenIssues
			return nil
		}
		if m.screen == screenSettings && m.tree != nil {
			return m.settingsKey(s)
		}
		if m.screen == screenKnowledge && m.tree != nil {
			return m.knowledgeKey(s, k)
		}
		var cmd tea.Cmd
		m.page, cmd = m.page.Update(k)
		return cmd
	}
	switch s {
	case "p":
		return m.jumpToParent()
	case "a":
		if m.tree != nil {
			return m.openMenu()
		}
		return nil
	case "n":
		if m.tree != nil {
			return m.openCreate()
		}
		return nil
	case "e":
		return m.openEditMenu()
	case "i":
		if m.tree != nil && m.selected() != "" {
			return m.openPriority()
		}
		return nil
	case "o":
		return m.openLinks()
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
		} else if s == "esc" && m.filters[tb.name] != "" {
			m.setFilter(tb.name, "")
		}
		return nil
	case "f", "/": // f is the melody; / stays for US-keyboard habits
		if tb != nil {
			m.filtering, m.filterErr = true, ""
			m.input.SetValue(m.filters[tb.name])
			m.input.CursorEnd()
			m.flagInfo = m.tree.QueryFlags()
			m.refreshCompletion()
			return m.input.Focus()
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

// filterKey handles keys while the filter bar is open: up and down move
// through the candidates, tab inserts one, enter applies, esc clears.
func (m *Model) filterKey(k tea.KeyPressMsg) tea.Cmd {
	tb := m.current()
	switch k.String() {
	case "ctrl+c":
		return tea.Quit
	case "up":
		m.pick = clamp(m.pick-1, 0, max(0, len(m.comp.items)-1))
		return nil
	case "down":
		m.pick = clamp(m.pick+1, 0, max(0, len(m.comp.items)-1))
		return nil
	case "tab":
		m.insertSuggestion(m.pick)
		return nil
	case "esc":
		m.filtering, m.filterErr = false, ""
		m.input.Blur()
		if m.screen == screenKnowledge {
			m.know.filter = ""
			m.syncKnowledge()
			return nil
		}
		if tb != nil {
			m.setFilter(tb.name, "")
		}
		return nil
	case "enter":
		text := strings.TrimSpace(m.input.Value())
		if m.screen == screenKnowledge {
			m.knowledgeFilterKey(text)
			return nil
		}
		if text != "" {
			if _, err := parseFilterText(text); err != nil {
				m.filterErr = err.Error()
				return nil
			}
		}
		m.filtering, m.filterErr = false, ""
		m.input.Blur()
		if tb != nil {
			m.setFilter(tb.name, text)
		}
		return nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(k)
	m.refreshCompletion()
	return cmd
}

// refreshCompletion recomputes the candidates for the issue filter bar
// after the input changed; the knowledge filter has none.
func (m *Model) refreshCompletion() {
	m.comp, m.pick = completion{}, 0
	if m.screen != screenIssues {
		return
	}
	m.comp = complete(m.flagInfo, m.input.Value(), m.input.Position())
	if len(m.comp.items) > maxSuggestions {
		m.comp.items = m.comp.items[:maxSuggestions]
	}
}

// insertSuggestion replaces the typed part with candidate n.
func (m *Model) insertSuggestion(n int) {
	if n < 0 || n >= len(m.comp.items) {
		return
	}
	runes := []rune(m.input.Value())
	ins := []rune(m.comp.items[n].insert)
	out := append(append(append([]rune{}, runes[:m.comp.start]...), ins...), runes[m.comp.end:]...)
	m.input.SetValue(string(out))
	m.input.SetCursor(m.comp.start + len(ins))
	m.refreshCompletion()
}

func (m *Model) setFilter(tab, text string) {
	if text == "" {
		delete(m.filters, tab)
	} else {
		m.filters[tab] = text
	}
	m.rebuild()
}

func (m *Model) detailKey(s string, k tea.KeyPressMsg) tea.Cmd {
	switch s {
	case "esc", "left", "h":
		m.focus = focusList
		return nil
	case "tab":
		m.switchTab(m.active + 1)
		return nil
	case "shift+tab":
		m.switchTab(m.active - 1)
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
	m.notice, m.noticeTone = n, ui.ToneAccent
	return m.after(noticeFor, func(time.Time) tea.Msg { return clearNoticeMsg{n} })
}

// flashErr shows an error in the footer for longer than a notice.
func (m *Model) flashErr(err error) tea.Cmd {
	n := strings.ReplaceAll(err.Error(), "\n", " ")
	m.notice, m.noticeTone = n, ui.ToneError
	return m.after(errorFor, func(time.Time) tea.Msg { return clearNoticeMsg{n} })
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
	// Ancestors, root first: the breadcrumb.
	anc := m.tree.Ancestors(id)
	for k := len(anc) - 1; k >= 0; k-- {
		if m.tree.Issues[anc[k]] != nil {
			out = append(out, relation{"path", anc[k]})
		}
	}
	_ = i.Parent
	for _, c := range m.tree.ByPriority(m.tree.Children(id)) {
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
	for _, p := range m.tree.IssueEntries(id) {
		out = append(out, relation{"knowledge", p})
	}
	return out
}

// --- view ---

func (m *Model) bodyHeight() int { return ui.Stack(m.h, headerLines, footerLines) }

func (m *Model) listHeight() int { return max(1, m.bodyHeight()-2) }

func (m *Model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	if m.mouse {
		v.MouseMode = tea.MouseModeCellMotion
	}
	return v
}

func (m *Model) render() string {
	m.hits, m.panes = nil, nil
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
	case m.screen == screenCheck:
		m.at = point{0, headerLines}
		body = m.checkPane(m.w, bodyH)
	case m.screen == screenSettings:
		m.at = point{0, headerLines}
		body = m.settingsPane(m.w, bodyH)
	case m.screen == screenKnowledge:
		m.at = point{0, headerLines}
		body = m.knowledgePane(m.w, bodyH)
	default:
		m.at = point{0, headerLines}
		listW, detailW := ui.Split(m.w, listRatio, minListW, minDetailW)
		switch {
		case detailW == 0 && m.focus == focusDetail:
			body = m.detailPane(m.w, bodyH)
		case detailW == 0:
			body = m.listPane(m.w, bodyH)
		default:
			list := m.listPane(listW, bodyH)
			m.at.x = listW
			body = lipgloss.JoinHorizontal(lipgloss.Top, list, m.detailPane(detailW, bodyH))
		}
	}
	if m.modal != nil && m.tree != nil {
		// Dialogs float over the screen they act on.
		block := m.modalView(m.w, bodyH)
		bw, bh := lipgloss.Width(block), lipgloss.Height(block)
		x, y := ui.OverlayAt(bw, bh, m.w, bodyH)
		y += headerLines
		m.markAt("dialog", x, y, bw, bh, zDialog)
		for _, e := range m.modal.marks {
			m.markAt(e.id, x+paneInner.x+e.x, y+paneInner.y+e.line, e.w, e.h, zEntry)
		}
		body = ui.Overlay(m.th, body, block, m.w, bodyH)
	}
	return header + "\n" + body + "\n" + footer
}

func (m *Model) header() string {
	name := m.th.S.Title.Render("prep") + " "
	var tabs []ui.Tab
	for _, tb := range m.tabs {
		tabs = append(tabs, ui.Tab{Label: tb.name, Count: tb.count})
	}
	status := ""
	if m.diagDone {
		errs, warns := 0, 0
		for _, d := range m.diags {
			if d.Severity == domain.SevError {
				errs++
			} else {
				warns++
			}
		}
		status = " " + ui.Note(m.th, fmt.Sprintf("✕ %d", errs), toneIf(errs > 0, ui.ToneError)) + " " + ui.Note(m.th, fmt.Sprintf("▲ %d", warns), toneIf(warns > 0, ui.ToneWarning))
	}
	tabsW := m.w - lipgloss.Width(name) - lipgloss.Width(status)
	line := name + ui.Tabs(m.th, tabs, m.active, tabsW)
	xs, ws := ui.TabSpans(tabs, tabsW)
	for k := range xs {
		m.markAt(fmt.Sprintf("tab:%d", k), lipgloss.Width(name)+xs[k], 0, ws[k], 1, zRow)
	}
	if status != "" {
		line += strings.Repeat(" ", max(0, m.w-lipgloss.Width(line)-lipgloss.Width(status))) + status
	}
	return ui.Fit(line, m.w)
}

func toneIf(cond bool, t ui.Tone) ui.Tone {
	if cond {
		return t
	}
	return ui.ToneMuted
}

// binding is one entry of a screen's keymap. Essential ones show in the
// footer; ? shows them all, grouped.
type binding struct {
	group     string
	key       ui.Key
	essential bool
}

func bind(group, keys, desc string, essential bool) binding {
	return binding{group, ui.Key{Keys: keys, Desc: desc}, essential}
}

var (
	issueBindings = []binding{
		bind("Issue", "a", "actions (letters run them)", true),
		bind("Issue", "e", "edit: r requirement · c context · t title", true),
		bind("Issue", "i", "priority: c critical · h high · m medium · l low", true),
		bind("Issue", "n", "new issue (under the focused parent)", true),
		bind("Issue", "y", "copy the ID", false),
		bind("Issue", "p", "go to the parent", false),
		bind("Issue", "o", "go to a linked issue (numbered menu)", false),
	}
	viewBindings = []binding{
		bind("Views", "f", "filter (prep list flags; words match titles)", true),
		bind("Views", "tab 1-9", "switch view", false),
		bind("Views", "t", "tree or flat", false),
	}
	screenBindings = []binding{
		bind("Screens", "c", "check", false),
		bind("Screens", "s", "settings", false),
		bind("Screens", "b", "knowledge", true),
		bind("Screens", "r", "reload", false),
		bind("Screens", "?", "all keys", true),
		bind("Screens", "q", "quit", true),
	}
	listBindings = concat([]binding{
		bind("Move", "↑↓ j k", "move", false),
		bind("Move", "g G pgup pgdn", "top, bottom, page", false),
		bind("Move", "enter → l", "details, or into a parent", false),
		bind("Move", "← h esc", "out of a parent; esc clears a filter", false),
	}, issueBindings, viewBindings, screenBindings)
	detailBindings = concat([]binding{
		bind("Move", "↑↓ pgup pgdn", "scroll", false),
		bind("Move", "o", "links: go to one by its number", true),
		bind("Move", "⌫", "back to the previous issue", false),
		bind("Move", "esc", "list (← h too)", true),
	}, nonEssential(issueBindings, "n"), []binding{
		bind("Views", "tab 1-9", "switch view", false),
	}, screenBindings)
	settingsBindings = []binding{
		bind("Settings", "↑↓ j k 1-9", "select (n is view n)", false),
		bind("Settings", "space", "switch the theme or toggle the mouse", true), bind("Settings", "t T", "next or previous theme", true),
		bind("Settings", "enter", "edit the view", true),
		bind("Settings", "n", "add a view", true),
		bind("Settings", "d d", "delete the view", true),
		bind("Settings", "m", "move the view: j k or arrows, enter when done", true),
		bind("Settings", "esc", "back", true),
		bind("Screens", "c", "check", false),
		bind("Screens", "?", "all keys", true),
	}
	pageBindings = []binding{
		bind("Page", "↑↓ pgup pgdn", "scroll", false),
		bind("Page", "esc", "back", true),
		bind("Screens", "c", "check", true),
		bind("Screens", "s", "settings", true),
		bind("Screens", "?", "all keys", true),
		bind("Screens", "q", "quit", true),
	}
	filterKeys = []ui.Key{{Keys: "enter", Desc: "apply"}, {Keys: "esc", Desc: "clear"}, {Keys: "↑/↓ tab", Desc: "pick a value"}, {Keys: "--state --kind --tag --priority --text --stale --blocked --actionable", Desc: "flags; words match titles"}}
)

// nonEssential copies bindings with the given keys left out of the footer.
func nonEssential(bs []binding, keys ...string) []binding {
	out := append([]binding(nil), bs...)
	for k := range out {
		if slices.Contains(keys, out[k].key.Keys) {
			out[k].essential = false
		}
	}
	return out
}

func concat(lists ...[]binding) []binding {
	var out []binding
	for _, l := range lists {
		out = append(out, l...)
	}
	return out
}

var knowledgeBindings = []binding{
	bind("Knowledge", "↑↓ j k", "select an entry", false),
	bind("Knowledge", "enter → l", "read the entry (arrows scroll, esc back)", false),
	bind("Knowledge", "f", "filter: --type --status --scope, words match titles", true),
	bind("Knowledge", "a", "only entries that need an agent's attention", true),
	bind("Knowledge", "o", "go to an issue that changed the entry", true),
	bind("Knowledge", "⌫", "back to the issue you came from", false),
	bind("Knowledge", "esc b", "back to the issues", true),
	bind("Screens", "c", "check", false),
	bind("Screens", "?", "all keys", true),
	bind("Screens", "q", "quit", true),
}

// bindings is the keymap of what the screen shows now.
func (m *Model) bindings() []binding {
	switch {
	case m.screen == screenKnowledge && m.tree != nil:
		return knowledgeBindings
	case m.screen == screenSettings && m.tree != nil:
		return settingsBindings
	case m.screen != screenIssues:
		return pageBindings
	case m.focus == focusDetail:
		return detailBindings
	}
	return listBindings
}

// essentials are the footer's keys, their descriptions cut to the first words.
func essentials(bs []binding) []ui.Key {
	var out []ui.Key
	for _, x := range bs {
		if x.essential {
			k := x.key
			k.Desc = strings.SplitN(k.Desc, ":", 2)[0]
			k.Desc = strings.SplitN(k.Desc, " (", 2)[0]
			out = append(out, k)
		}
	}
	return out
}

func (m *Model) footer() string {
	switch {
	case m.notice != "":
		return ui.Fit(ui.Note(m.th, m.notice, m.noticeTone), m.w)
	case m.modal != nil:
		return ui.KeyHelp(m.th, m.modalKeys(), m.w)
	case m.err != nil:
		return ui.Fit(ui.Error(m.th, m.err.Error()), m.w)
	case m.filtering:
		return ui.KeyHelp(m.th, filterKeys, m.w)
	}
	return ui.KeyHelp(m.th, essentials(m.bindings()), m.w)
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
	if f := m.filters[tb.name]; f != "" {
		title += "  / " + f
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
	var bar []string
	if m.filtering {
		m.input.SetWidth(max(1, inner-lipgloss.Width(m.input.Prompt)-1))
		bar = append(bar, m.input.View())
		if m.filterErr != "" {
			bar = append(bar, ui.Error(m.th, m.filterErr))
		}
		for k, s := range m.comp.items {
			m.mark(fmt.Sprintf("sugg:%d", k), paneInner.x, paneInner.y+len(bar), inner, 1)
			bar = append(bar, ui.Suggestion(m.th, s.value, s.count, s.label, k == m.pick, inner))
		}
		bar = append(bar, "")
		rows = max(1, rows-len(bar))
	}
	var body string
	switch {
	case tb == nil:
	case tb.err != nil:
		body = ui.Error(m.th, tb.err.Error())
	case len(tb.rows) == 0:
		body = ui.Empty(m.th, "No issues in this view", "Issues appear when they match "+orAll(tb.flags), inner, rows)
	default:
		c := m.cursor[tb.name]
		off := listOffset(m.offset[tb.name], c, len(tb.rows), rows, !m.wheeled)
		m.offset[tb.name] = off
		var lines []string
		for k := off; k < len(tb.rows) && k < off+rows; k++ {
			m.mark(fmt.Sprintf("row:%d", k), paneInner.x, paneInner.y+len(bar)+k-off, inner, 1)
			lines = append(lines, ui.ListRow(m.th, m.row(tb.rows[k]), k == c, inner))
		}
		body = strings.Join(lines, "\n")
	}
	if len(bar) > 0 {
		body = strings.Join(bar, "\n") + "\n" + body
	}
	m.pane("list", w, h)
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
	r := ui.Row{ID: shortID(id), State: t.State(id), Kind: i.Kind, Title: i.Title, Tags: i.Tags, Priority: i.Priority}
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

// shortID shows the last 6 characters of an ID, which are random in a
// ULID; any unique suffix resolves in the CLI.
func shortID(id string) string {
	if len(id) > 6 {
		return id[len(id)-6:]
	}
	return id
}

// relationBlock draws the detail's links by their shape: ancestors as a
// breadcrumb, children as a tree under their progress, dependencies as
// arrows ("← needs" what this waits on, "→ unblocks" what waits on it).
// The block is display only; o opens a numbered menu of the same links.
//
// links holds, per returned line, the index into rels it shows, or -1.
func (m *Model) relationBlock(id string, rels []relation, inner int) (out []string, links []int) {
	if len(rels) == 0 {
		return nil, nil
	}
	sub := m.th.S.Subtle
	var path []string
	type line struct {
		text string
		rel  int // index into rels, -1 for a heading
	}
	var lines []line
	link := func(k int, lead string) line {
		rid := rels[k].id
		r := m.issueRow(rid)
		return line{ui.LinkLine(m.th, ui.Link{Lead: lead, State: m.tree.State(rid), ID: shortID(rid), Priority: m.tree.Issues[rid].Priority, Note: r.Note, Title: m.tree.Issues[rid].Title}, false, inner), k}
	}
	var children []int
	for k, r := range rels {
		switch r.label {
		case "path":
			title := ui.Fit(m.tree.Issues[r.id].Title, 32)
			title = m.th.S.Muted.Render(title)
			path = append(path, title)
		case "child":
			children = append(children, k)
		}
	}
	if len(path) > 0 {
		out = append(out, ui.Fit(sub.Render("↑ ")+strings.Join(path, sub.Render(" › ")), inner))
		links = append(links, -1)
	}
	if len(children) > 0 {
		p := m.tree.ChildProgress(id)
		lines = append(lines, line{m.th.S.Muted.Render("Children ") + ui.Progress(m.th, p.Done+p.Dropped, p.Total), -1})
		for n, k := range children {
			lead := "├─ "
			if n == len(children)-1 {
				lead = "└─ "
			}
			lines = append(lines, link(k, sub.Render(lead)))
		}
	}
	for k, r := range rels {
		switch r.label {
		case "depends on":
			lead := sub.Render("← needs ")
			if !m.tree.State(r.id).Terminal() {
				lead = lipgloss.NewStyle().Foreground(m.th.C.Warning).Render("← needs ")
			}
			lines = append(lines, link(k, lead))
		case "blocks":
			lines = append(lines, link(k, sub.Render("→ unblocks ")))
		case "knowledge":
			e := m.tree.Knowledge[r.id]
			lines = append(lines, line{ui.Fit("  "+sub.Render("≡ knows ")+m.th.S.Muted.Render(fmt.Sprintf("%-10s", e.Type))+m.th.S.Body.Render(e.Title), inner), k})
		}
	}
	for n := 0; n < len(lines) && n < maxRelations; n++ {
		out = append(out, lines[n].text)
		links = append(links, lines[n].rel)
	}
	if hidden := len(lines) - maxRelations; hidden > 0 {
		out = append(out, sub.Render(fmt.Sprintf("  … %d more · o lists all", hidden)))
		links = append(links, -1)
	}
	return append(out, ""), append(links, -1)
}

func (m *Model) detailPane(w, h int) string {
	id := m.selected()
	inner := w - 2 - 2*theme.Pad
	focused := m.focus == focusDetail
	m.pane("detail", w, h)
	if id == "" {
		return ui.Pane{Title: "Details", Body: ui.Empty(m.th, "Nothing selected", "", inner, h-2), Focused: focused, Width: w, Height: h}.View(m.th)
	}

	// Relations: a fixed block of rows above the scrolling document.
	rels := m.relations(id)
	block, links := m.relationBlock(id, rels, inner)
	for n, rel := range links {
		if rel >= 0 && paneInner.y+n < h-1 {
			m.mark(fmt.Sprintf("rel:%d", rel), paneInner.x, paneInner.y+n, inner, 1)
		}
	}

	key := fmt.Sprintf("%s|%d|%d", id, inner, m.gen)
	if key != m.docKey {
		doc, err := ui.Markdown(m.th, detailMarkdown(m.tree, id), inner)
		if err != nil {
			doc = ui.Error(m.th, err.Error())
		}
		if d := m.staleDiff(id, inner); d != "" {
			doc = d + "\n\n" + doc
		}
		sameIssue := strings.HasPrefix(m.docKey, id+"|")
		m.vp.SetContent(doc)
		if !sameIssue {
			m.vp.GotoTop()
		}
		m.docKey = key
	}
	m.vp.SetWidth(inner)
	m.vp.SetHeight(max(1, h-2-len(block)))
	title := shortID(id) + " " + m.tree.Issues[id].Title
	if m.vp.TotalLineCount() > m.vp.Height() {
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

// staleDiff renders what changed in a stale issue's requirement and kind
// since its newest baseline.
func (m *Model) staleDiff(id string, width int) string {
	i := m.tree.Issues[id]
	b := i.LatestBaseline()
	if b == nil || !m.tree.Stale(id) {
		return ""
	}
	lines := []string{ui.Note(m.th, "Changed since baseline "+b.Name, ui.ToneWarning)}
	if b.Kind != i.Kind {
		lines = append(lines, ui.Diff(m.th, []ui.DiffLine{{Op: '-', Text: "kind: " + string(b.Kind)}, {Op: '+', Text: "kind: " + string(i.Kind)}}, width))
	}
	if b.Requirement != i.Body {
		lines = append(lines, ui.Diff(m.th, ui.DiffLines(strings.Split(b.Requirement, "\n"), strings.Split(i.Body, "\n")), width))
	}
	lines = append(lines, m.th.S.Subtle.Render("prep ack for a trivial change, prep define to re-enrich"))
	return strings.Join(lines, "\n")
}

// checkPane shows every diagnostic of the last check run.
func (m *Model) checkPane(w, h int) string {
	inner := w - 2 - 2*theme.Pad
	var body string
	switch {
	case m.opts.Check == nil:
		body = ui.Empty(m.th, "Check is not available", "", inner, h-2)
	case !m.diagDone:
		body = ui.Loading(m.th, "Running prep check …")
	case m.diagErr != nil:
		body = ui.Error(m.th, m.diagErr.Error())
	case len(m.diags) == 0:
		body = ui.Empty(m.th, "No problems", "prep check finds no errors or warnings", inner, h-2)
	default:
		var errs, warns []string
		for _, d := range m.diags {
			where := d.Issue
			if d.File != "" {
				if where != "" {
					where += " · "
				}
				where += d.File
			}
			row := ui.DiagnosticRow(m.th, ui.Diagnostic{Error: d.Severity == domain.SevError, Code: d.Code, Where: where, Message: d.Message, Fix: d.Fix, Class: string(d.Class)}, inner)
			if d.Severity == domain.SevError {
				errs = append(errs, row)
			} else {
				warns = append(warns, row)
			}
		}
		var parts []string
		if len(errs) > 0 {
			parts = append(parts, m.th.S.Heading.Render(fmt.Sprintf("Errors (%d)", len(errs))), strings.Join(errs, "\n\n"))
		}
		if len(warns) > 0 {
			if len(parts) > 0 {
				parts = append(parts, "")
			}
			parts = append(parts, m.th.S.Heading.Render(fmt.Sprintf("Warnings (%d)", len(warns))), strings.Join(warns, "\n\n"))
		}
		body = strings.Join(parts, "\n")
	}
	return m.pagePane("Check", body, w, h)
}

// settingsPane lists the editable settings as rows: the theme, the saved
// views numbered like the tabs they show as, then the user's mouse choice.
// The schema and the project's Definition of Done follow, read-only.
func (m *Model) settingsPane(w, h int) string {
	inner := w - 2 - 2*theme.Pad
	p := m.tree.Project
	m.setIdx = clamp(m.setIdx, 0, m.settingsRows()-1)
	row := func(k int, key, label, value string) string {
		sel := k == m.setIdx
		marker := "  "
		if sel {
			marker = lipgloss.NewStyle().Foreground(m.th.C.Accent).Render("▌ ")
		}
		line := marker + lipgloss.NewStyle().Foreground(m.th.C.Accent).Bold(true).Render(fmt.Sprintf("%-3s", key)) +
			m.th.S.Body.Render(fmt.Sprintf("%-18s", ui.Fit(label, 17))) + m.th.S.Muted.Render(value)
		line = ui.Fit(line, inner)
		if sel {
			return lipgloss.NewStyle().Background(m.th.C.Selection).Width(inner).Render(line)
		}
		return line
	}
	var b []string
	rowLines := map[int]int{} // settings row → its line in b, for clicks
	addRow := func(k int, key, label, value string) {
		rowLines[k] = len(b)
		b = append(b, row(k, key, label, value))
	}
	b = append(b, m.th.S.Heading.Render("Look"), "")
	addRow(0, "t", "Theme", m.th.Palette.Name+"  (t next, T previous; prep tui --gallery previews them)")
	b = append(b, "", m.th.S.Heading.Render("Saved views"), "")
	for k, n := range domain.ViewNames(p.Config) {
		flags := p.Config.Views[n]
		if flags == "" {
			flags = "(all issues)"
		}
		addRow(k+1, fmt.Sprint(k+1), n, flags)
	}
	b = append(b, "", m.th.S.Heading.Render("You")+"  "+m.th.S.Subtle.Render("your user configuration, not the project's"), "")
	addRow(m.settingsRows()-1, "␣", "Mouse", map[bool]string{true: "on", false: "off"}[m.mouse]+"  (clicks and the wheel; off lets the terminal select text)")
	b = append(b, "", m.th.S.Subtle.Render(fmt.Sprintf("Schema version %d.", p.Schema)))
	if len(p.DoD) > 0 {
		b = append(b, "", m.th.S.Heading.Render("Definition of Done"), m.th.S.Subtle.Render("edit in .prep/project.md"), "")
		for _, d := range p.DoD {
			b = append(b, ui.Fit("- "+d, inner))
		}
	}
	out := m.pagePane("Settings", strings.Join(b, "\n"), w, h)
	for k, line := range rowLines {
		if y := line - m.page.YOffset(); y >= 0 && y < h-2 {
			m.mark(fmt.Sprintf("set:%d", k), paneInner.x, paneInner.y+y, inner, 1)
		}
	}
	return out
}

func (m *Model) pagePane(title, body string, w, h int) string {
	m.pane("page", w, h)
	m.page.SetWidth(w - 2 - 2*theme.Pad)
	m.page.SetHeight(h - 2)
	m.page.SetContent(body)
	if m.page.TotalLineCount() > m.page.Height() {
		title += fmt.Sprintf("  %d%%", int(m.page.ScrollPercent()*100))
	}
	return ui.Pane{Title: title, Body: m.page.View(), Focused: true, Width: w, Height: h}.View(m.th)
}
