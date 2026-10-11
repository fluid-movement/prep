package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/fluid-movement/prep/internal/domain"
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

// linkedModel shows an issue that depends on a linked project's issue and
// links its knowledge, with the dependency resolved as the CLI would.
func linkedModel(t *testing.T, status domain.XStatus) (*Model, string) {
	t.Helper()
	p, ids := sample(t)
	const apiID, dep = "01M4BBBBBBBBBBBBBBBBBBBBBB", "01M4DDDDDDDDDDDDDDDDDDDDDD"
	load := func() (*domain.Tree, error) {
		tr, err := p.load()
		if err != nil {
			return nil, err
		}
		tr.Project.Links = []domain.ProjectLink{{Name: "api", ID: apiID, Remote: "github.com/acme/api"}}
		i := *tr.Issues[ids["export"]]
		i.DependsOn = []string{"api:" + dep}
		i.Context = "See [auth](api:/components/auth.md)."
		tr.Issues[i.ID] = &i
		x := domain.XIssue{Ref: "api:" + dep, Link: "api", ID: dep, Status: status, Remote: "github.com/acme/api"}
		if status == domain.XFound {
			x.State, x.Title, x.Project, x.Root = domain.StateInProgress, "Rate limits", "api", "/code/api"
		}
		tr.External = map[string]domain.XIssue{x.Ref: x}
		return tr, nil
	}
	m := NewModel(testTheme(), Options{Load: load, Root: "/code/web",
		Projects: func() ([]projects.Project, error) {
			return []projects.Project{{Root: "/code/api", ID: apiID, Name: "api"}}, nil
		}})
	m.Update(tea.WindowSizeMsg{Width: 140, Height: 40})
	m.jump(ids["export"], false)
	return m, ids["export"]
}

func TestCrossProjectLinksInTheDetail(t *testing.T) {
	m, _ := linkedModel(t, domain.XFound)
	v := ansi.Strip(m.View().Content)
	if !strings.Contains(v, "← needs api › ") || !strings.Contains(v, "Rate limits") || !strings.Contains(v, "≡ knows api › /components/auth.md") {
		t.Fatalf("detail:\n%s", v)
	}
	clickText(t, m, "Rate limits", 50, 140)
	if m.switchTo == nil || m.switchTo.Root != "/code/api" || m.switchTo.Issue != "01M4DDDDDDDDDDDDDDDDDDDDDD" {
		t.Fatalf("dependency: %+v", m.switchTo)
	}
	m, _ = linkedModel(t, domain.XFound)
	clickText(t, m, "api › /components/auth.md", 50, 140)
	if m.switchTo == nil || m.switchTo.Root != "/code/api" || m.switchTo.Entry != "/components/auth.md" {
		t.Fatalf("knowledge: %+v", m.switchTo)
	}
}

func TestUnknownCrossProjectDependency(t *testing.T) {
	m, _ := linkedModel(t, domain.XUnknown)
	if v := ansi.Strip(m.View().Content); !strings.Contains(v, "not checked out here · clone") {
		t.Fatalf("detail:\n%s", v)
	}
	clickText(t, m, "not checked out here", 50, 140)
	if m.switchTo != nil {
		t.Fatalf("switched: %+v", m.switchTo)
	}
}
