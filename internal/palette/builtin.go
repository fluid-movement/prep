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
// muted, subtle, accent, border, selection, success, warning, error; the
// states open, defined, ready, in_progress, done, dropped; the kinds code,
// manual, research, decision. Only default carries hand-picked fallbacks.
var builtins = []Palette{
	{
		Name: "default",
		Dark: picked(
			c("#E4E4EA", "254", "15"), c("#A0A0AE", "247", "7"), c("#5E5E6C", "240", "8"), c("#FF9E5E", "215", "11"),
			c("#3A3A48", "237", "8"), c("#2A2F3A", "236", "0"), c("#5BD68A", "78", "10"), c("#F2C14E", "221", "11"), c("#FF6B81", "204", "9"),
			c("#A0A0AE", "247", "7"), c("#6CB6FF", "75", "12"), c("#3DD6C6", "43", "14"), c("#F2C14E", "221", "11"), c("#5BD68A", "78", "10"), c("#6E6E7C", "242", "8"),
			c("#8AB4F8", "111", "12"), c("#F58FC6", "211", "13"), c("#62C7F5", "81", "14"), c("#E3C58E", "180", "11"),
		),
		Light: picked(
			c("#1E1E26", "235", "0"), c("#5C5C6A", "241", "8"), c("#A4A4B2", "248", "7"), c("#C2410C", "166", "3"),
			c("#D2D2DC", "252", "7"), c("#E8ECF2", "255", "15"), c("#167A3E", "28", "2"), c("#9A5B00", "130", "3"), c("#BE123C", "161", "1"),
			c("#5C5C6A", "241", "8"), c("#1F5FAD", "25", "4"), c("#0B7A70", "30", "6"), c("#9A5B00", "130", "3"), c("#167A3E", "28", "2"), c("#9A9AA8", "247", "7"),
			c("#1D5FBF", "26", "4"), c("#A3246C", "125", "5"), c("#0F6A99", "24", "6"), c("#7A5A12", "94", "3"),
		),
	},
	{
		Name: "high-contrast",
		Dark: hex(
			"#FFFFFF", "#D0D0D0", "#A0A0A0", "#FFB000", "#C0C0C0", "#003D99", "#00FF66", "#FFE000", "#FF4D4D",
			"#D0D0D0", "#66B3FF", "#00FFFF", "#FFE000", "#00FF66", "#A0A0A0",
			"#66B3FF", "#FF66CC", "#00FFFF", "#FFB000",
		),
		Light: hex(
			"#000000", "#303030", "#5A5A5A", "#A33F00", "#303030", "#C7DCFF", "#00662B", "#704600", "#B00020",
			"#303030", "#003D99", "#005F5F", "#704600", "#00662B", "#5A5A5A",
			"#003D99", "#8A005C", "#005F5F", "#A33F00",
		),
	},
	{
		Name: "monochrome",
		Dark: hex(
			"#E6E6E6", "#A8A8A8", "#6A6A6A", "#7AA2F7", "#3C3C3C", "#2C2C2C", "#C8C8C8", "#E6E6E6", "#7AA2F7",
			"#9E9E9E", "#BDBDBD", "#E0E0E0", "#7AA2F7", "#D6D6D6", "#5C5C5C",
			"#BDBDBD", "#9E9E9E", "#D6D6D6", "#E6E6E6",
		),
		Light: hex(
			"#1A1A1A", "#555555", "#9A9A9A", "#2F5FD0", "#D0D0D0", "#EBEBEB", "#444444", "#222222", "#2F5FD0",
			"#666666", "#444444", "#222222", "#2F5FD0", "#333333", "#A0A0A0",
			"#444444", "#666666", "#333333", "#222222",
		),
	},
	{
		Name: "pastel",
		Dark: hex(
			"#ECE7F2", "#B3ACC2", "#6E6880", "#F7A8C4", "#45405A", "#332F45", "#A8E6C1", "#F9D9A0", "#F4A6A6",
			"#B3ACC2", "#A8C8F7", "#A0E3DC", "#F9D9A0", "#A8E6C1", "#7D778F",
			"#B5C8F7", "#F7B5DA", "#A8DDF0", "#EAD7A8",
		),
		Light: hex(
			"#3B3548", "#6E6680", "#A8A0B8", "#C2537E", "#DDD6E8", "#F1ECF7", "#3E8A62", "#A0702A", "#B5525A",
			"#6E6680", "#4F74B8", "#3A8F87", "#A0702A", "#3E8A62", "#A8A0B8",
			"#5574B5", "#B0558A", "#3F86A6", "#8F7340",
		),
	},
	{
		// Catppuccin Mocha (dark) and Latte (light).
		Name: "catppuccin",
		Dark: hex(
			"#CDD6F4", "#A6ADC8", "#6C7086", "#FAB387", "#45475A", "#313244", "#A6E3A1", "#F9E2AF", "#F38BA8",
			"#A6ADC8", "#89B4FA", "#94E2D5", "#F9E2AF", "#A6E3A1", "#7F849C",
			"#B4BEFE", "#F5C2E7", "#89DCEB", "#EBA0AC",
		),
		Light: hex(
			"#4C4F69", "#6C6F85", "#9CA0B0", "#FE640B", "#BCC0CC", "#CCD0DA", "#40A02B", "#DF8E1D", "#D20F39",
			"#6C6F85", "#1E66F5", "#179299", "#DF8E1D", "#40A02B", "#8C8FA1",
			"#7287FD", "#EA76CB", "#04A5E5", "#E64553",
		),
	},
	{
		// Nord: Polar Night (dark) and Snow Storm (light) backgrounds.
		Name: "nord",
		Dark: hex(
			"#ECEFF4", "#A5ADBA", "#616E88", "#88C0D0", "#434C5E", "#3B4252", "#A3BE8C", "#EBCB8B", "#BF616A",
			"#A5ADBA", "#81A1C1", "#8FBCBB", "#EBCB8B", "#A3BE8C", "#616E88",
			"#5E81AC", "#B48EAD", "#88C0D0", "#D08770",
		),
		Light: hex(
			"#2E3440", "#4C566A", "#8A94A6", "#5E81AC", "#D8DEE9", "#E5E9F0", "#5F7F45", "#A07A1F", "#BF616A",
			"#4C566A", "#5E81AC", "#4C8C8A", "#A07A1F", "#5F7F45", "#8A94A6",
			"#5E81AC", "#8F6A88", "#3B7F94", "#B0603A",
		),
	},
	{
		Name: "gruvbox",
		Dark: hex(
			"#EBDBB2", "#A89984", "#7C6F64", "#FE8019", "#504945", "#3C3836", "#B8BB26", "#FABD2F", "#FB4934",
			"#A89984", "#83A598", "#8EC07C", "#FABD2F", "#B8BB26", "#7C6F64",
			"#83A598", "#D3869B", "#8EC07C", "#D79921",
		),
		Light: hex(
			"#3C3836", "#7C6F64", "#A89984", "#AF3A03", "#D5C4A1", "#EBDBB2", "#79740E", "#B57614", "#9D0006",
			"#7C6F64", "#076678", "#427B58", "#B57614", "#79740E", "#A89984",
			"#076678", "#8F3F71", "#427B58", "#AF3A03",
		),
	},
}
