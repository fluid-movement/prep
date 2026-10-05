package domain

import "strings"

// Stable error codes returned by commands.
const (
	ErrUsage     = "E_USAGE"
	ErrNotFound  = "E_NOT_FOUND"
	ErrAmbiguous = "E_AMBIGUOUS"
	ErrGate      = "E_GATE"       // a transition's gate conditions are unmet
	ErrInvalid   = "E_INVALID"    // the write would produce invalid state
	ErrConflict  = "E_CONFLICT"   // a file changed since it was read
	ErrNoProject = "E_NO_PROJECT" // no .prep directory found
	ErrSchema    = "E_SCHEMA"
	ErrIO        = "E_IO"
	ErrCheck     = "E_CHECK" // prep check found errors
)

// Error is a domain error with a stable code.
type Error struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Unmet   []Unmet      `json:"unmet,omitempty"`
	Diags   []Diagnostic `json:"diagnostics,omitempty"`
}

func (e *Error) Error() string {
	if len(e.Unmet) == 0 {
		return e.Message
	}
	var b strings.Builder
	b.WriteString(e.Message)
	for _, u := range e.Unmet {
		b.WriteString("\n  - ")
		b.WriteString(u.Message)
	}
	return b.String()
}
