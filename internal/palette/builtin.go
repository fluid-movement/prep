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

// The built-in themes, in display order: dark themes, then the one light
// theme. Values follow Tokens: text, muted, subtle, accent, selection,
// success, warning, error. Selections are a step lighter than common dark
// terminal backgrounds (around #1E–#2A), so the selected row stands out.
// Only default and light carry hand-picked fallbacks.
var builtins = []Palette{
	{Name: "default", Colors: picked(
		c("#E4E4EA", "254", "15"), c("#A0A0AE", "247", "7"), c("#5E5E6C", "240", "8"), c("#FF9E5E", "215", "11"),
		c("#3A4254", "238", "8"), c("#5BD68A", "78", "10"), c("#F2C14E", "221", "11"), c("#FF6B81", "204", "9"),
	)},
	{Name: "high-contrast", Colors: hex(
		"#FFFFFF", "#D0D0D0", "#A0A0A0", "#FFB000",
		"#1F4FB0", "#00FF66", "#FFE000", "#FF4D4D",
	)},
	{Name: "monochrome", Colors: hex(
		"#E6E6E6", "#A8A8A8", "#6A6A6A", "#7AA2F7",
		"#444444", "#C8C8C8", "#E6E6E6", "#7AA2F7",
	)},
	{Name: "pastel", Colors: hex(
		"#ECE7F2", "#B3ACC2", "#6E6880", "#F7A8C4",
		"#4A4460", "#A8E6C1", "#F9D9A0", "#F4A6A6",
	)},
	// Catppuccin Mocha.
	{Name: "catppuccin", Colors: hex(
		"#CDD6F4", "#A6ADC8", "#6C7086", "#FAB387",
		"#45475A", "#A6E3A1", "#F9E2AF", "#F38BA8",
	)},
	// Nord on Polar Night.
	{Name: "nord", Colors: hex(
		"#ECEFF4", "#A5ADBA", "#616E88", "#88C0D0",
		"#434C5E", "#A3BE8C", "#EBCB8B", "#BF616A",
	)},
	{Name: "gruvbox", Colors: hex(
		"#EBDBB2", "#A89984", "#7C6F64", "#FE8019",
		"#504945", "#B8BB26", "#FABD2F", "#FB4934",
	)},
	{Name: Light, Light: true, Colors: picked(
		c("#1E1E26", "235", "0"), c("#5C5C6A", "241", "8"), c("#A4A4B2", "248", "7"), c("#C2410C", "166", "3"),
		c("#E8ECF2", "255", "15"), c("#167A3E", "28", "2"), c("#9A5B00", "130", "3"), c("#BE123C", "161", "1"),
	)},
}
