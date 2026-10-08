package palette

import "github.com/fluid-movement/prep/internal/domain"

// Borders and kinds are drawn from the gray ramp: kinds are written out,
// so they stay muted.
const (
	BorderToken = "subtle"
	KindToken   = "muted"
)

// stateTokens draws each state from the palette: the glyph tells states
// apart, the color only marks what needs attention (ready, in progress)
// and lets finished work fade.
var stateTokens = map[domain.State]string{
	domain.StateOpen:       "muted",
	domain.StateDefined:    "text",
	domain.StateReady:      "success",
	domain.StateInProgress: "accent",
	domain.StateDone:       "subtle",
	domain.StateDropped:    "subtle",
}

// StateToken returns the token an issue state is drawn with; unknown
// states are muted. The TUI and prep theme colors both use it, so a
// harness never repeats the mapping.
func StateToken(s domain.State) string {
	if tok, ok := stateTokens[s]; ok {
		return tok
	}
	return "muted"
}
