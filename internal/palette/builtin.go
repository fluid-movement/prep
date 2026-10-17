package palette

// c is a color with hand-picked 256- and 16-color values.
func c(hex, ansi256, ansi string) Color { return Color{Hex: hex, ANSI256: ansi256, ANSI: ansi} }

// picked builds a variant from colors in Tokens order.
func picked(colors ...Color) Variant {
	toks := Tokens()
	v := make(Variant, len(toks))
	for k, t := range toks {
		v[t] = colors[k]
	}
	return v
}

// The built-in themes, in display order: the dark theme, then the light
// theme. Values follow Tokens: text, muted, subtle, accent, selection,
// success, warning, error. Selections are a step lighter than common dark
// terminal backgrounds (around #1E–#2A), so the selected row stands out.
// Both themes carry hand-picked fallbacks.
var builtins = []Palette{
	{Name: "default", Colors: picked(
		c("#E4E4EA", "254", "15"), c("#A0A0AE", "247", "7"), c("#5E5E6C", "240", "8"), c("#FF9E5E", "215", "11"),
		c("#3A4254", "238", "8"), c("#5BD68A", "78", "10"), c("#F2C14E", "221", "11"), c("#FF6B81", "204", "9"),
	)},
	{Name: Light, Light: true, Colors: picked(
		c("#1E1E26", "235", "0"), c("#5C5C6A", "241", "8"), c("#A4A4B2", "248", "7"), c("#C2410C", "166", "3"),
		c("#E8ECF2", "255", "15"), c("#167A3E", "28", "2"), c("#9A5B00", "130", "3"), c("#BE123C", "161", "1"),
	)},
}
