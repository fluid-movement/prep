package ui

import (
	"strings"

	"charm.land/glamour/v2"
	"charm.land/glamour/v2/ansi"
	"charm.land/glamour/v2/styles"

	"github.com/fluid-movement/prep/internal/tui/theme"
)

// Markdown renders markdown to width with a Glamour style derived from the
// theme's tokens.
func Markdown(t *theme.Theme, md string, width int) (string, error) {
	r, err := glamour.NewTermRenderer(
		glamour.WithStyles(markdownStyle(t)),
		glamour.WithWordWrap(max(10, width)),
	)
	if err != nil {
		return "", err
	}
	out, err := r.Render(md)
	if err != nil {
		return "", err
	}
	lines := strings.Split(strings.Trim(out, "\n"), "\n")
	for k, l := range lines {
		lines[k] = strings.TrimRight(l, " ")
	}
	return strings.Join(lines, "\n"), nil
}

func markdownStyle(t *theme.Theme) ansi.StyleConfig {
	s := styles.DarkStyleConfig
	if !t.Dark {
		s = styles.LightStyleConfig
	}
	str := func(v string) *string { return &v }
	yes := func() *bool { b := true; return &b }
	zero := uint(0)
	text, muted, subtle, accent := t.Hex("text"), t.Hex("muted"), t.Hex("subtle"), t.Hex("accent")

	s.Document = ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{Color: str(text)}, Margin: &zero}
	s.Paragraph = ansi.StyleBlock{}
	s.Heading = ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{BlockSuffix: "\n", Color: str(text), Bold: yes()}}
	s.H1 = ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{Color: str(accent), Bold: yes()}}
	s.H2 = ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{Color: str(accent), Bold: yes()}}
	s.H3 = ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{Color: str(text), Bold: yes()}}
	s.H4, s.H5, s.H6 = s.H3, s.H3, s.H3
	s.BlockQuote = ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{Color: str(muted)}, Indent: uintp(1), IndentToken: str("│ ")}
	s.HorizontalRule = ansi.StylePrimitive{Color: str(subtle), Format: "\n────────\n"}
	s.Item = ansi.StylePrimitive{BlockPrefix: "• "}
	s.Task = ansi.StyleTask{Ticked: "[x] ", Unticked: "[ ] "}
	s.Link = ansi.StylePrimitive{Color: str(muted), Underline: yes()}
	s.LinkText = ansi.StylePrimitive{Color: str(accent)}
	s.Code = ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{Color: str(accent)}}
	s.CodeBlock.Margin = &zero
	return s
}

func uintp(v uint) *uint { return &v }
