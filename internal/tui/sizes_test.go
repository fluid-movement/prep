package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/fluid-movement/prep/internal/activity"
)

// TestTinyTerminals renders every screen, in both layouts, at sizes down to
// nothing: no screen may panic however little room it gets.
func TestTinyTerminals(t *testing.T) {
	var events []activity.Event
	m, _, _ := agentModel(t, &events, 110, 28)
	for _, single := range []bool{false, true} {
		m.single = single
		for _, screen := range []string{"", "b", "w", "c", "s"} {
			for _, w := range []int{0, 1, 2, 5, 10, 20, 30, 44, 60, 92, 120} {
				for h := 0; h <= 8; h++ {
					m.screen = screenIssues
					m.Update(tea.WindowSizeMsg{Width: w, Height: h})
					if screen != "" {
						run(m, screen)
					}
					func() {
						defer func() {
							if r := recover(); r != nil {
								t.Fatalf("screen %q at %dx%d (one pane %v) panicked: %v", screen, w, h, single, r)
							}
						}()
						m.View()
					}()
				}
			}
		}
	}
}

// After the wheel scrolled a list, moving the selection by other means
// than a key (a tab click, a jump, a number key) makes the view follow it
// again.
func TestViewFollowsSelectionAfterWheel(t *testing.T) {
	p, ids := sample(t)
	m := openModel(t, p, 110, 28)
	m.wheeled = true
	m.switchTab(1)
	if m.wheeled {
		t.Fatal("switching tabs kept the wheel's offset")
	}
	m.wheeled = true
	m.jump(ids["doc"], false)
	if m.wheeled {
		t.Fatal("a jump kept the wheel's offset")
	}
}

func TestEmptyKnowledgeListShowsNoEntry(t *testing.T) {
	p, _ := knowledgeSample(t)
	m := withCheck(openModel(t, p, 120, 30), p)
	run(m, "b")
	run(m, "f")
	typeIn(m, "--type pitfall")
	run(m, "enter")
	if paths, _ := m.knowledgePaths(); len(paths) != 0 {
		t.Fatalf("filter left %v", paths)
	}
	if m.shownEntry() != "" || m.copyTarget() != "" {
		t.Fatalf("an empty list still shows %q", m.shownEntry())
	}
	if v := m.View().Content; !strings.Contains(ansi.Strip(v), "Nothing selected") {
		t.Fatalf("entry pane is not empty:\n%s", ansi.Strip(v))
	}
}

func TestScreenKeys(t *testing.T) {
	var events []activity.Event
	m, _, _ := agentModel(t, &events, 110, 28)
	run(m, "w")
	if got := m.copyTarget(); got == "" || m.tree.Issues[got] == nil {
		t.Fatalf("y on the Agent screen copies %q", got)
	}
	tree := m.treeMode
	run(m, "t")
	if m.treeMode != tree {
		t.Fatal("t toggled the issue tree from the Agent screen")
	}
	run(m, "2")
	if m.screen != screenIssues || m.active != 1 {
		t.Fatalf("2 left screen %d on tab %d", m.screen, m.active)
	}
}
