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

// The built-in themes, in display order. Values follow Tokens: text,
// muted, subtle, accent, selection, success, warning, error. Only default
// carries hand-picked fallbacks.
var builtins = []Palette{
	{
		Name: "default",
		Dark: picked(
			c("#E4E4EA", "254", "15"), c("#A0A0AE", "247", "7"), c("#5E5E6C", "240", "8"), c("#FF9E5E", "215", "11"),
			c("#2A2F3A", "236", "0"), c("#5BD68A", "78", "10"), c("#F2C14E", "221", "11"), c("#FF6B81", "204", "9"),
		),
		Light: picked(
			c("#1E1E26", "235", "0"), c("#5C5C6A", "241", "8"), c("#A4A4B2", "248", "7"), c("#C2410C", "166", "3"),
			c("#E8ECF2", "255", "15"), c("#167A3E", "28", "2"), c("#9A5B00", "130", "3"), c("#BE123C", "161", "1"),
		),
	},
	{
		Name: "high-contrast",
		Dark: hex(
			"#FFFFFF", "#D0D0D0", "#A0A0A0", "#FFB000",
			"#003D99", "#00FF66", "#FFE000", "#FF4D4D",
		),
		Light: hex(
			"#000000", "#303030", "#5A5A5A", "#A33F00",
			"#C7DCFF", "#00662B", "#704600", "#B00020",
		),
	},
	{
		Name: "monochrome",
		Dark: hex(
			"#E6E6E6", "#A8A8A8", "#6A6A6A", "#7AA2F7",
			"#2C2C2C", "#C8C8C8", "#E6E6E6", "#7AA2F7",
		),
		Light: hex(
			"#1A1A1A", "#555555", "#9A9A9A", "#2F5FD0",
			"#EBEBEB", "#444444", "#222222", "#2F5FD0",
		),
	},
	{
		Name: "pastel",
		Dark: hex(
			"#ECE7F2", "#B3ACC2", "#6E6880", "#F7A8C4",
			"#332F45", "#A8E6C1", "#F9D9A0", "#F4A6A6",
		),
		Light: hex(
			"#3B3548", "#6E6680", "#A8A0B8", "#C2537E",
			"#F1ECF7", "#3E8A62", "#A0702A", "#B5525A",
		),
	},
	{
		// Catppuccin Mocha (dark) and Latte (light).
		Name: "catppuccin",
		Dark: hex(
			"#CDD6F4", "#A6ADC8", "#6C7086", "#FAB387",
			"#313244", "#A6E3A1", "#F9E2AF", "#F38BA8",
		),
		Light: hex(
			"#4C4F69", "#6C6F85", "#9CA0B0", "#FE640B",
			"#CCD0DA", "#40A02B", "#DF8E1D", "#D20F39",
		),
	},
	{
		// Nord: Polar Night (dark) and Snow Storm (light) backgrounds.
		Name: "nord",
		Dark: hex(
			"#ECEFF4", "#A5ADBA", "#616E88", "#88C0D0",
			"#3B4252", "#A3BE8C", "#EBCB8B", "#BF616A",
		),
		Light: hex(
			"#2E3440", "#4C566A", "#8A94A6", "#5E81AC",
			"#E5E9F0", "#5F7F45", "#A07A1F", "#BF616A",
		),
	},
	{
		Name: "gruvbox",
		Dark: hex(
			"#EBDBB2", "#A89984", "#7C6F64", "#FE8019",
			"#3C3836", "#B8BB26", "#FABD2F", "#FB4934",
		),
		Light: hex(
			"#3C3836", "#7C6F64", "#A89984", "#AF3A03",
			"#EBDBB2", "#79740E", "#B57614", "#9D0006",
		),
	},
}
