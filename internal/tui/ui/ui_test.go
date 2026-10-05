package ui

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/fluid-movement/prep/internal/tui/theme"
)

var update = flag.Bool("update", false, "rewrite golden files")

// testTheme renders with a fixed true-color profile and background, so
// snapshots do not depend on the terminal running the tests.
func testTheme(dark bool) *theme.Theme {
	r := lipgloss.NewRenderer(io.Discard)
	r.SetColorProfile(termenv.TrueColor)
	r.SetHasDarkBackground(dark)
	return theme.New(r)
}

func TestGallerySnapshots(t *testing.T) {
	for _, dark := range []bool{true, false} {
		for _, width := range []int{80, 120} {
			bg := map[bool]string{true: "dark", false: "light"}[dark]
			name := fmt.Sprintf("gallery-%s-%d", bg, width)
			t.Run(name, func(t *testing.T) {
				got := Gallery(testTheme(dark), width)
				for k, l := range strings.Split(got, "\n") {
					if w := lipgloss.Width(l); w > width {
						t.Errorf("line %d is %d cells wide, limit %d: %q", k+1, w, width, l)
					}
				}
				golden(t, name, got)
			})
		}
	}
}

func TestSplit(t *testing.T) {
	cases := []struct {
		total, minA, minB int
		ratio             float64
		a, b              int
	}{
		{100, 30, 30, 0.4, 40, 60},
		{100, 50, 30, 0.4, 50, 50},
		{100, 30, 70, 0.5, 30, 70},
		{50, 30, 30, 0.5, 50, 0},
		{0, 10, 10, 0.5, 0, 0},
	}
	for _, c := range cases {
		a, b := Split(c.total, c.ratio, c.minA, c.minB)
		if a != c.a || b != c.b {
			t.Errorf("Split(%d, %.1f, %d, %d) = %d, %d; want %d, %d", c.total, c.ratio, c.minA, c.minB, a, b, c.a, c.b)
		}
	}
}

func TestPaneClipsToSize(t *testing.T) {
	th := testTheme(true)
	out := Pane{Title: "A very long pane title that will not fit", Body: strings.Repeat("word ", 40) + "\n2\n3\n4\n5\n6", Width: 30, Height: 5}.View(th)
	lines := strings.Split(out, "\n")
	if len(lines) != 5 {
		t.Fatalf("pane has %d lines, want 5:\n%s", len(lines), out)
	}
	for _, l := range lines {
		if w := lipgloss.Width(l); w != 30 {
			t.Fatalf("pane line is %d wide, want 30: %q", w, l)
		}
	}
}

func golden(t *testing.T, name, got string) {
	t.Helper()
	p := filepath.Join("testdata", name+".golden")
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("%v (run go test ./internal/tui/... -update to create it)", err)
	}
	if string(want) != got {
		t.Errorf("%s differs from the golden file; check prep tui --gallery and run go test ./internal/tui/... -update if the change is intended", name)
	}
}
