package ui

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"

	"github.com/fluid-movement/prep/internal/domain"
	"github.com/fluid-movement/prep/internal/palette"
	"github.com/fluid-movement/prep/internal/tui/theme"
)

var update = flag.Bool("update", false, "rewrite golden files")

// testTheme renders with a fixed true-color profile and background, so
// snapshots do not depend on the terminal running the tests.
func testTheme(dark bool) *theme.Theme { return theme.New(dark, colorprofile.TrueColor) }

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

// TestThemesFollowTheColorProfile checks what a terminal receives with
// every theme: Bubble Tea writes through a colorprofile.Writer, which
// downsamples for 256- and 16-color terminals and drops colors without
// color support (NO_COLOR).
func TestThemesFollowTheColorProfile(t *testing.T) {
	for _, name := range palette.Builtins() {
		p, _ := palette.Resolve(name)
		for _, c := range []struct {
			profile colorprofile.Profile
			banned  []string
		}{
			{colorprofile.ANSI256, []string{"38;2;"}},
			{colorprofile.ANSI, []string{"38;2;", "38;5;"}},
			{colorprofile.ASCII, []string{"38;", "48;"}},
		} {
			var out bytes.Buffer
			w := &colorprofile.Writer{Forward: &out, Profile: c.profile}
			if _, err := w.WriteString(Gallery(theme.From(p, c.profile), 80)); err != nil {
				t.Fatal(err)
			}
			got := out.String()
			if !strings.Contains(got, "State badges") {
				t.Fatalf("%s with profile %v lost its text", name, c.profile)
			}
			for _, b := range c.banned {
				if strings.Contains(got, b) {
					t.Errorf("%s with profile %v renders %q", name, c.profile, b)
				}
			}
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

// A narrow list row drops the ID first, then the state label, then the kind,
// and stops once the title has 30 cells; the state glyph always stays.
func TestListRowDropsColumnsInOrder(t *testing.T) {
	th := testTheme(true)
	base := Row{ID: "PWJH2F", State: domain.StateInProgress, Kind: domain.KindCode, Title: "Findings from the first pass of the agent behavior test suite"}
	cases := []struct {
		width           int
		id, label, kind bool
	}{
		{100, true, true, true},
		{61, false, true, true},  // 29 cells short with the ID: only the ID goes
		{50, false, false, true}, // the label goes too, the kind still fits
		{40, false, false, false},
	}
	for _, c := range cases {
		for _, dimmed := range []bool{false, true} {
			r := base
			r.Dimmed = dimmed
			line := ansi.Strip(ListRow(th, r, false, c.width))
			if got := strings.Contains(line, "PWJH2F"); got != c.id {
				t.Errorf("width %d dimmed %v: ID shown %v, want %v: %q", c.width, dimmed, got, c.id, line)
			}
			if got := strings.Contains(line, "in progress"); got != c.label {
				t.Errorf("width %d dimmed %v: label shown %v, want %v: %q", c.width, dimmed, got, c.label, line)
			}
			if got := strings.Contains(line, "code"); got != c.kind {
				t.Errorf("width %d dimmed %v: kind shown %v, want %v: %q", c.width, dimmed, got, c.kind, line)
			}
			if !strings.Contains(line, "▶") {
				t.Errorf("width %d dimmed %v: state glyph dropped: %q", c.width, dimmed, line)
			}
			if w := ansi.StringWidth(line); w > c.width {
				t.Errorf("width %d dimmed %v: row is %d cells wide: %q", c.width, dimmed, w, line)
			}
		}
	}
}

func TestTreePrefixes(t *testing.T) {
	got := TreePrefixes([]int{0, 1, 2, 2, 1, 2, 0, 1})
	want := []string{"", "├─ ", "│  ├─ ", "│  └─ ", "└─ ", "   └─ ", "", "└─ "}
	for k := range want {
		if got[k] != want[k] {
			t.Errorf("row %d: %q, want %q", k, got[k], want[k])
		}
	}
}

func TestDiffLines(t *testing.T) {
	got := DiffLines([]string{"a", "b", "c"}, []string{"a", "x", "c", "d"})
	var s []string
	for _, l := range got {
		s = append(s, string(l.Op)+l.Text)
	}
	if strings.Join(s, ",") != " a,-b,+x, c,+d" {
		t.Fatalf("diff = %v", s)
	}
}

func TestMenuRowKeepsLongLabelsOnOneLine(t *testing.T) {
	th := theme.New(true, colorprofile.TrueColor)
	row := MenuRow(th, "1", "├ 152616 TUI: human client next to the harness", "", true, true, 40)
	if strings.Contains(row, "\n") {
		t.Fatalf("long label wrapped:\n%s", row)
	}
	if w := ansi.StringWidth(row); w != 40 {
		t.Fatalf("row width %d, want 40", w)
	}
}

// Every row of a list keeps the same columns, whatever its tree lines or
// note, so the glyphs stay aligned.
func TestListRowColumnsDependOnWidthOnly(t *testing.T) {
	th := testTheme(true)
	plain := Row{ID: "PWJH2F", State: domain.StateDone, Kind: domain.KindDecision, Title: "Choose the default format"}
	busy := Row{ID: "EPMJ6R", State: domain.StateDefined, Kind: domain.KindCode, Title: "JSON writer", Tree: "└─ ", Note: "■■■■■ 0/1"}
	for _, width := range []int{100, 61, 55, 50, 40} {
		a, b := ansi.Strip(ListRow(th, plain, false, width)), ansi.Strip(ListRow(th, busy, false, width))
		if fitColumns(plain, width) != fitColumns(busy, width) {
			t.Errorf("width %d: columns differ between rows:\n%q\n%q", width, a, b)
		}
	}
}
