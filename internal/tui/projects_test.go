package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/fluid-movement/prep/internal/projects"
)

func switcherModel(t *testing.T) *Model {
	t.Helper()
	p, _ := sample(t)
	list := []projects.Project{
		{Root: "/code/web", Name: "web"},
		{Root: "/code/api", Name: "api", Remote: "github.com/acme/api"},
		{Root: "/code/tools", Name: "tools"},
	}
	m := NewModel(testTheme(), Options{Load: p.load, Project: "web", Root: "/code/web", Previous: "/code/old",
		Projects: func() ([]projects.Project, error) { return list, nil }})
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	return m
}

func TestHeaderShowsTheProjectAndPOpensTheSwitcher(t *testing.T) {
	m := switcherModel(t)
	if first := strings.Split(ansi.Strip(m.View().Content), "\n")[0]; !strings.HasPrefix(strings.TrimSpace(first), "web ") {
		t.Fatalf("header: %q", first)
	}
	keys(m, "P")
	if m.modal == nil || m.modal.kind != modalProjects || m.modal.picks[m.modal.cursor] != "/code/web" {
		t.Fatalf("switcher: %+v", m.modal)
	}
	if v := ansi.Strip(m.View().Content); !strings.Contains(v, "current · /code/web") || !strings.Contains(v, "api") {
		t.Fatalf("view:\n%s", v)
	}
	// Typing filters; enter switches and ends the program.
	keys(m, "a", "p", "i")
	if len(m.modal.picks) != 1 {
		t.Fatalf("filtered: %v", m.modal.picks)
	}
	_, cmd := m.Update(keyMsg("enter"))
	if m.switchTo == nil || m.switchTo.Root != "/code/api" || cmd == nil {
		t.Fatalf("switch: %+v", m.switchTo)
	}
}

func TestClickTheProjectNameAndBackspaceReturns(t *testing.T) {
	m := switcherModel(t)
	clickText(t, m, "web", 0, 10)
	if m.modal == nil || m.modal.kind != modalProjects {
		t.Fatal("click did not open the switcher")
	}
	// Choosing the current project only closes it.
	keys(m, "enter")
	if m.modal != nil || m.switchTo != nil {
		t.Fatalf("current: %+v %+v", m.modal, m.switchTo)
	}
	// With nothing else to go back to, backspace returns to the project
	// switched from.
	keys(m, "backspace")
	if m.switchTo == nil || m.switchTo.Root != "/code/old" {
		t.Fatalf("back: %+v", m.switchTo)
	}
}
