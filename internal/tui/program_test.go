package tui

import (
	"fmt"
	"io"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

// TestProgramReadsMouseInput runs the real program on the escape sequences
// a terminal sends (SGR mouse reports, 1-based), so parsing, hit-testing
// and the handlers are tested together.
func TestProgramReadsMouseInput(t *testing.T) {
	p, ids := sample(t)
	m := NewModel(testTheme(), Options{Load: p.load})
	m.Update(tea.WindowSizeMsg{Width: 110, Height: 28})
	x, y := find(t, m, "All", 0, 110) // before the program owns the model
	keys(m, "6")
	rx, ry := find(t, m, "CSV writer", 0, 60)
	keys(m, "1")
	in, w := io.Pipe()
	prog := tea.NewProgram(m, tea.WithInput(in), tea.WithOutput(io.Discard), tea.WithWindowSize(110, 28))
	done := make(chan error, 1)
	go func() { _, err := prog.Run(); done <- err }()
	send := func(s string) {
		w.Write([]byte(s))
		time.Sleep(150 * time.Millisecond) // let the program render between inputs
	}
	time.Sleep(300 * time.Millisecond)
	send(fmt.Sprintf("\x1b[<0;%d;%dM\x1b[<0;%d;%dm", x+1, y+1, x+1, y+1))     // click the All tab
	send(fmt.Sprintf("\x1b[<0;%d;%dM\x1b[<0;%d;%dm", rx+1, ry+1, rx+1, ry+1)) // click the CSV writer row
	send("\x1b[<65;6;6M")                                                     // a wheel notch keeps the selection
	send("q")
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the program did not quit")
	}
	if m.current().name != "All" || m.selected() != ids["csv"] {
		t.Fatalf("after clicks on All and CSV writer and a wheel notch: tab %q, selected %s", m.current().name, m.selected())
	}
}
