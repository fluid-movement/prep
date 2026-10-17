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

// listOf makes n rows of the same shape and the width at which each is cut
// by extra characters while every column is shown.
func listOf(th *theme.Theme, n, extra int) ([]Row, int) {
	var rows []Row
	for k := 0; k < n; k++ {
		rows = append(rows, Row{ID: "PWJH2F", State: domain.StateInProgress, Kind: domain.KindCode, Title: "Findings from the first pass of the agent behavior test suite"})
	}
	return rows, rowWidth(th, rows[0], AllColumns) - extra
}

// A list whose titles fit keeps every column, even in a narrow pane.
func TestFitColumnsKeepsAllWhenTitlesFit(t *testing.T) {
	th := testTheme(true)
	rows, _ := listOf(th, 4, 0)
	for _, w := range []int{120, rowWidth(th, rows[0], AllColumns)} {
		if c := FitColumns(th, rows, w); c != AllColumns {
			t.Errorf("width %d: columns %+v, want all", w, c)
		}
	}
	// Short titles in a narrow pane fit too.
	short := []Row{{ID: "PWJH2F", State: domain.StateDone, Kind: domain.KindCode, Title: "Short"}}
	if c := FitColumns(th, short, 35); c != AllColumns {
		t.Errorf("short title: columns %+v, want all", c)
	}
}

// When the cut-off title characters (each row up to 10) exceed 30, the list
// drops the ID, then the state label, then the kind.
func TestFitColumnsDropsInOrder(t *testing.T) {
	th := testTheme(true)
	cases := []struct {
		extra int // characters each of 4 rows is cut by with every column
		want  Columns
	}{
		{7, AllColumns},                    // 4*7 = 28, within the budget
		{8, Columns{false, true, true}},    // 32: the ID goes, 4 cut remains
		{15, Columns{false, false, true}},  // the ID alone leaves 4*8 = 32
		{26, Columns{false, false, true}},  // 4*7 = 28 after ID and label
		{27, Columns{false, false, false}}, // 4*8 = 32 after both
	}
	for _, c := range cases {
		rows, w := listOf(th, 4, c.extra)
		if got := FitColumns(th, rows, w); got != c.want {
			t.Errorf("cut %d: columns %+v, want %+v", c.extra, got, c.want)
		}
	}
}

// One very long title alone adds at most 10 to the total and drops nothing.
func TestFitColumnsCapsOneLongTitle(t *testing.T) {
	th := testTheme(true)
	rows := []Row{{ID: "PWJH2F", State: domain.StateOpen, Kind: domain.KindCode, Title: strings.Repeat("long ", 60)}}
	if c := FitColumns(th, rows, 60); c != AllColumns {
		t.Errorf("columns %+v, want all", c)
	}
}

// Every row is drawn with the columns the list chose, whatever its tree
// lines or note, so the glyphs stay aligned; the state glyph always stays.
func TestListRowUsesGivenColumns(t *testing.T) {
	th := testTheme(true)
	plain := Row{ID: "PWJH2F", State: domain.StateDone, Kind: domain.KindDecision, Title: "Choose the default format"}
	busy := Row{ID: "EPMJ6R", State: domain.StateDefined, Kind: domain.KindCode, Title: "JSON writer", Tree: "└─ ", Note: "■■■■■ 0/1"}
	rows := []Row{plain, busy}
	for _, width := range []int{100, 61, 50, 40} {
		c := FitColumns(th, rows, width)
		for _, dimmed := range []bool{false, true} {
			for _, r := range rows {
				r.Dimmed = dimmed
				line := ansi.Strip(ListRow(th, r, c, false, width))
				if got := strings.Contains(line, r.ID); got != c.ID {
					t.Errorf("width %d: ID shown %v, want %v: %q", width, got, c.ID, line)
				}
				if got := strings.Contains(line, StateLabel(r.State)); got != c.Label {
					t.Errorf("width %d: label shown %v, want %v: %q", width, got, c.Label, line)
				}
				if got := strings.Contains(line, string(r.Kind)); got != c.Kind {
					t.Errorf("width %d: kind shown %v, want %v: %q", width, got, c.Kind, line)
				}
				if !strings.Contains(line, stateGlyphs[r.State]) {
					t.Errorf("width %d: state glyph dropped: %q", width, line)
				}
				if w := ansi.StringWidth(line); w > width {
					t.Errorf("width %d: row is %d cells wide: %q", width, w, line)
				}
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
