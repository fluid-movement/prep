package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/projects"
	"github.com/fluid-movement/prep/internal/tui/ui"
)

// Switch is where the person asked to go: another project, optionally on
// one of its issues. Run returns it, and prep tui starts over on that
// project.
type Switch struct {
	Root  string
	Issue string
	Entry string // a knowledge entry to open
}

// projectName is the header's name for the project: its own name, else
// the app's.
func (m *Model) projectName() string {
	if m.opts.Project != "" {
		return m.opts.Project
	}
	return "prep"
}

// openProjects opens the project switcher: the registry's projects, newest
// first, filtered as the person types.
func (m *Model) openProjects() tea.Cmd {
	if m.opts.Projects == nil {
		return m.flash("no other projects to switch to")
	}
	ps, err := m.opts.Projects()
	if err != nil {
		return m.flash("projects: " + err.Error())
	}
	m.modal = &modal{kind: modalProjects, inputs: []textinput.Model{m.newInput("type to filter", "")}, projects: ps}
	m.modal.picks = m.projectCandidates("")
	for k, root := range m.modal.picks {
		if root == m.opts.Root {
			m.modal.cursor = k
		}
	}
	return m.modal.inputs[0].Focus()
}

// projectCandidates lists the roots of the projects whose name, remote or
// root contains the filter.
func (m *Model) projectCandidates(filter string) []string {
	filter = strings.ToLower(strings.TrimSpace(filter))
	var out []string
	for _, p := range m.modal.projects {
		if filter != "" && !strings.Contains(strings.ToLower(p.Name+" "+p.Remote+" "+p.Root), filter) {
			continue
		}
		out = append(out, p.Root)
	}
	return out
}

// projectKey handles keys in the switcher.
func (m *Model) projectKey(k tea.KeyPressMsg) tea.Cmd {
	d := m.modal
	switch k.String() {
	case "up", "ctrl+p":
		d.cursor, d.scrolled = clamp(d.cursor-1, 0, len(d.picks)-1), false
		return nil
	case "down", "ctrl+n":
		d.cursor, d.scrolled = clamp(d.cursor+1, 0, len(d.picks)-1), false
		return nil
	case "enter":
		if len(d.picks) == 0 {
			return nil
		}
		return m.switchProject(Switch{Root: d.picks[d.cursor]})
	}
	var cmd tea.Cmd
	d.inputs[0], cmd = d.inputs[0].Update(k)
	d.picks = m.projectCandidates(d.inputs[0].Value())
	d.cursor = clamp(d.cursor, 0, len(d.picks)-1)
	return cmd
}

// switchProject ends this program with the place to go; choosing the
// current project just closes the switcher.
func (m *Model) switchProject(s Switch) tea.Cmd {
	m.modal = nil
	if s.Root == m.opts.Root && s.Issue == "" {
		return nil
	}
	m.switchTo = &s
	return tea.Quit
}

// pickRow is one rendered row of a dialog list and its click target.
type pickRow struct{ id, line string }

// projectRows renders the switcher's visible rows.
func (m *Model) projectRows(inner, rows int) []pickRow {
	d := m.modal
	byRoot := map[string]projects.Project{}
	for _, p := range d.projects {
		byRoot[p.Root] = p
	}
	if len(d.picks) == 0 {
		return []pickRow{{line: m.th.S.Subtle.Render("no project matches")}}
	}
	start := clamp(d.cursor-rows/2, 0, max(0, len(d.picks)-rows))
	if d.scrolled {
		start = clamp(d.scroll, 0, max(0, len(d.picks)-rows))
	}
	d.scroll = start
	var out []pickRow
	for k := start; k < len(d.picks) && k < start+rows; k++ {
		p := byRoot[d.picks[k]]
		label := p.Name
		if p.Collision && p.Remote != "" {
			label += " (" + p.Remote + ")"
		}
		note := p.Root
		if p.Root == m.opts.Root {
			note = "current · " + note
		}
		out = append(out, pickRow{fmt.Sprintf("pick:%d", k), ui.MenuRow(m.th, "", label, note, true, k == d.cursor, inner)})
	}
	return out
}

// xDepLine renders a dependency on another project's issue: the link's
// name before the ID, the state from that project, and for one not
// checked out here the clone hint in place of a title.
func (m *Model) xDepLine(ref string, k int, selected bool, key string, inner int) string {
	x := m.tree.XIssueOf(ref)
	sub := m.th.S.Subtle
	lead := sub.Render("← needs ")
	if x.Status == domain.XFound && !x.State.Terminal() {
		lead = lipgloss.NewStyle().Foreground(m.th.C.Warning).Render("← needs ")
	}
	l := ui.Link{Lead: lead + m.th.S.Muted.Render(x.Link+" › "), State: x.State, ID: shortID(x.ID), Title: x.Title, Key: key}
	switch x.Status {
	case domain.XMissing:
		l.Title = sub.Render("no such issue in " + x.Project)
	case domain.XUnknown:
		l.Title = sub.Render("not checked out here")
		if x.Remote != "" {
			l.Title = sub.Render("not checked out here · clone " + x.Remote)
		}
	}
	return ui.LinkLine(m.th, l, selected, inner)
}

// followElsewhere opens an issue or a knowledge entry of a linked project:
// the TUI starts over there, and backspace comes back.
func (m *Model) followElsewhere(label, link, rest string) tea.Cmd {
	if label != "knowledge" {
		x := m.tree.XIssueOf(link + ":" + rest)
		if x.Status != domain.XFound {
			return m.flash(link + " is not checked out here, or has no such issue")
		}
		return m.switchProject(Switch{Root: x.Root, Issue: x.ID})
	}
	root := m.linkRoot(link)
	if root == "" {
		return m.flash(link + " is not checked out here")
	}
	return m.switchProject(Switch{Root: root, Entry: rest})
}

// linkRoot is the newest checkout of the linked project, or "".
func (m *Model) linkRoot(name string) string {
	l, ok := m.tree.Project.Link(name)
	if !ok || m.opts.Projects == nil {
		return ""
	}
	ps, err := m.opts.Projects()
	if err != nil {
		return ""
	}
	for _, p := range ps {
		if p.ID == l.ID {
			return p.Root
		}
	}
	return ""
}
