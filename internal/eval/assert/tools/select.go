// Package tools holds the assertions that read the transcript: which prep
// commands and tools the agent used, in what order and how often. Importing
// it registers its kinds.
package tools

import (
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/fluid-movement/prep/internal/eval/assert"
	"github.com/fluid-movement/prep/internal/eval/transcript"
)

// Selector picks steps (or commands within steps). All fields given must
// match.
type Selector struct {
	// Command is a prep command with optional leading words: "guide",
	// "knowledge find".
	Command string `yaml:"command"`
	// Project is the --project the command named.
	Project string `yaml:"project"`
	// Args is a regular expression matched against the command's arguments
	// joined by spaces.
	Args string `yaml:"args"`
	// Tool is the harness tool name (Read, Edit, Bash...).
	Tool string `yaml:"tool"`
	// Path is a glob (with ** across directories) matched against the file
	// the tool targets; for Bash it is matched against each word of the
	// shell command.
	Path string `yaml:"path"`
	// Shell is a regular expression matched against a Bash command line.
	Shell string `yaml:"shell"`
	// Subagent restricts the match to steps run by a subagent (true) or by
	// the main session (false).
	Subagent *bool `yaml:"subagent"`

	args  *regexp.Regexp
	path  *regexp.Regexp
	shell *regexp.Regexp
}

// parseSelector reads a selector from a scalar (a command) or a mapping.
// extra are keys the caller's struct also allows and strips first.
func parseSelector(n *yaml.Node) (Selector, error) {
	var s Selector
	switch n.Kind {
	case yaml.ScalarNode:
		s.Command = n.Value
	case yaml.MappingNode:
		if err := assert.DecodeStrict(n, &s); err != nil {
			return s, err
		}
	default:
		return s, fmt.Errorf("line %d: expected a command or a mapping", n.Line)
	}
	return s, s.compile()
}

func (s *Selector) compile() error {
	var err error
	if s.Command == "" && s.Tool == "" && s.Path == "" && s.Shell == "" && s.Subagent == nil {
		return fmt.Errorf("selector names nothing: give command, tool, path, shell or subagent")
	}
	if s.Args != "" {
		if s.args, err = regexp.Compile(s.Args); err != nil {
			return fmt.Errorf("args: %w", err)
		}
	}
	if s.Shell != "" {
		if s.shell, err = regexp.Compile(s.Shell); err != nil {
			return fmt.Errorf("shell: %w", err)
		}
	}
	if s.Path != "" {
		if s.path, err = globRegexp(s.Path); err != nil {
			return fmt.Errorf("path: %w", err)
		}
	}
	return nil
}

// globRegexp turns a glob into an expression that matches it at the end of
// a path, on a directory boundary: ".prep/**" matches "/w/.prep/a/b.md".
func globRegexp(g string) (*regexp.Regexp, error) {
	var b strings.Builder
	if !strings.HasPrefix(g, "/") {
		b.WriteString(`(?:^|/)`)
	} else {
		b.WriteString(`^`)
	}
	for i := 0; i < len(g); i++ {
		switch c := g[i]; c {
		case '*':
			if i+1 < len(g) && g[i+1] == '*' {
				b.WriteString(`.*`)
				i++
			} else {
				b.WriteString(`[^/]*`)
			}
		case '?':
			b.WriteString(`[^/]`)
		default:
			b.WriteString(regexp.QuoteMeta(string(c)))
		}
	}
	b.WriteString(`$`)
	return regexp.Compile(b.String())
}

// Describe puts the selector in words.
func (s Selector) Describe() string {
	var parts []string
	if s.Command != "" {
		parts = append(parts, "prep "+s.Command)
	}
	if s.Project != "" {
		parts = append(parts, "--project "+s.Project)
	}
	if s.Args != "" {
		parts = append(parts, "args /"+s.Args+"/")
	}
	if s.Tool != "" {
		parts = append(parts, s.Tool)
	}
	if s.Path != "" {
		parts = append(parts, "on "+s.Path)
	}
	if s.Shell != "" {
		parts = append(parts, "shell /"+s.Shell+"/")
	}
	if s.Subagent != nil {
		parts = append(parts, map[bool]string{true: "in a subagent", false: "in the main session"}[*s.Subagent])
	}
	return strings.Join(parts, " ")
}

// Matches returns how many times the step matches: a prep command selector
// counts each matching command in the step; a tool selector counts the step
// once.
func (s Selector) Matches(st transcript.Step) int {
	if s.Command != "" {
		n := 0
		for _, c := range st.Commands {
			if !c.Is(s.Command) {
				continue
			}
			if s.Project != "" && c.Project != s.Project {
				continue
			}
			if s.args != nil && !s.args.MatchString(strings.Join(c.Args, " ")) {
				continue
			}
			if !s.stepMatches(st) {
				continue
			}
			n++
		}
		return n
	}
	if s.Project != "" || s.args != nil {
		return 0 // those qualify prep commands
	}
	if s.stepMatches(st) {
		return 1
	}
	return 0
}

// stepMatches checks the tool, path and shell parts.
func (s Selector) stepMatches(st transcript.Step) bool {
	if s.Tool != "" && st.Tool != s.Tool {
		return false
	}
	if s.Subagent != nil && (st.Parent != "") != *s.Subagent {
		return false
	}
	if s.shell != nil && !s.shell.MatchString(st.Shell) {
		return false
	}
	if s.path != nil {
		if st.Path != "" {
			if !s.path.MatchString(st.Path) {
				return false
			}
		} else if st.Shell != "" {
			hit := false
			for _, w := range strings.Fields(st.Shell) {
				if s.path.MatchString(strings.Trim(w, `"'`)) {
					hit = true
					break
				}
			}
			if !hit {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

// First returns the index of the first matching step and the total number of
// matches.
func (s Selector) First(t *transcript.Transcript) (first, count int) {
	first = -1
	for _, st := range t.Steps {
		if n := s.Matches(st); n > 0 {
			if first < 0 {
				first = st.Index
			}
			count += n
		}
	}
	return first, count
}
