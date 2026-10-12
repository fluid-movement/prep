package transcript

import (
	"path"
	"regexp"
	"strings"
)

// Command is one invocation of prep found in a shell command.
type Command struct {
	// Name is the prep command ("list", "knowledge").
	Name string `json:"name"`
	// Args are the words after the name, flags included, in order.
	Args []string `json:"args,omitempty"`
	// Project is the --project the call named.
	Project string `json:"project,omitempty"`
	// By is the --by the call named.
	By string `json:"by,omitempty"`
	// Root is the --root the call named.
	Root string `json:"root,omitempty"`
	// Raw is the text of the invocation.
	Raw string `json:"raw"`
}

// Is reports whether the command is the named one, optionally followed by
// leading words: Is("knowledge find") is true for `prep knowledge find x`.
func (c Command) Is(spec string) bool {
	words := strings.Fields(spec)
	if len(words) == 0 || words[0] != c.Name {
		return false
	}
	pos := c.positional()
	for i, w := range words[1:] {
		if i >= len(pos) || pos[i] != w {
			return false
		}
	}
	return true
}

// positional returns the args that are not flags or flag values.
func (c Command) positional() []string {
	var out []string
	for i := 0; i < len(c.Args); i++ {
		a := c.Args[i]
		if strings.HasPrefix(a, "-") && a != "-" {
			if !strings.Contains(a, "=") && i+1 < len(c.Args) && !strings.HasPrefix(c.Args[i+1], "-") && flagTakesValue[strings.TrimLeft(a, "-")] {
				i++
			}
			continue
		}
		out = append(out, a)
	}
	return out
}

// flagTakesValue lists prep flags whose next word is their value, so a
// value is not mistaken for a positional word.
var flagTakesValue = map[string]bool{
	"title": true, "kind": true, "parent": true, "depends-on": true, "tag": true, "priority": true,
	"body": true, "body-file": true, "state": true, "under": true, "text": true, "max": true, "actor": true,
	"add": true, "check": true, "uncheck": true, "remove": true, "reason": true, "note": true, "commit": true,
	"docs": true, "no-impact": true, "type": true, "description": true, "supersedes": true, "as": true,
	"opt-out": true, "view": true, "scope": true, "status": true, "harness": true, "by": true,
	"project": true, "root": true,
}

// Flag returns the value of --name (or --name=value) in the command's
// arguments and whether it was given.
func (c Command) Flag(name string) (string, bool) {
	for i, a := range c.Args {
		if a == "--"+name {
			if i+1 < len(c.Args) {
				return c.Args[i+1], true
			}
			return "", true
		}
		if strings.HasPrefix(a, "--"+name+"=") {
			return a[len(name)+3:], true
		}
	}
	return "", false
}

// ParseShell finds the prep invocations in a shell command line: it handles
// quoting, && || ; | and newlines, env prefixes, `cd x &&`, paths to the
// binary and heredoc bodies. Text it cannot make sense of yields no command.
func ParseShell(line string) []Command {
	var out []Command
	for _, seg := range splitSegments(stripHeredocs(line)) {
		words := tokenize(seg)
		for len(words) > 0 && isAssign(words[0]) {
			words = words[1:]
		}
		// `command prep`, `time prep`, `env X=1 prep`, `sudo`-free forms
		for len(words) > 0 && (words[0] == "command" || words[0] == "time" || words[0] == "exec" || words[0] == "env" || words[0] == "nohup") {
			words = words[1:]
			for len(words) > 0 && isAssign(words[0]) {
				words = words[1:]
			}
		}
		if len(words) == 0 || path.Base(words[0]) != "prep" {
			continue
		}
		if c, ok := parseCommand(words[1:], strings.TrimSpace(seg)); ok {
			out = append(out, c)
		}
	}
	return out
}

func isAssign(w string) bool {
	i := strings.Index(w, "=")
	if i <= 0 {
		return false
	}
	for _, r := range w[:i] {
		if !(r == '_' || r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9') {
			return false
		}
	}
	return true
}

func parseCommand(words []string, raw string) (Command, bool) {
	c := Command{Raw: raw}
	i := 0
	for ; i < len(words); i++ {
		w := words[i]
		if !strings.HasPrefix(w, "-") {
			break
		}
		switch {
		case w == "--project" || w == "--by" || w == "--root":
			if i+1 < len(words) {
				c.setGlobal(w[2:], words[i+1])
				i++
			}
		case strings.HasPrefix(w, "--project=") || strings.HasPrefix(w, "--by=") || strings.HasPrefix(w, "--root="):
			k := strings.Index(w, "=")
			c.setGlobal(w[2:k], w[k+1:])
		default: // --json and other value-less globals
			c.Args = append(c.Args, w)
		}
	}
	if i >= len(words) {
		return c, false
	}
	c.Name = words[i]
	rest := words[i+1:]
	// global flags may follow the command too
	var args []string
	for j := 0; j < len(rest); j++ {
		w := rest[j]
		switch {
		case (w == "--project" || w == "--by" || w == "--root") && j+1 < len(rest):
			c.setGlobal(w[2:], rest[j+1])
			args = append(args, w, rest[j+1])
			j++
		case strings.HasPrefix(w, "--project="):
			c.setGlobal("project", w[len("--project="):])
			args = append(args, w)
		case strings.HasPrefix(w, "--by="):
			c.setGlobal("by", w[len("--by="):])
			args = append(args, w)
		default:
			args = append(args, w)
		}
	}
	c.Args = append(c.Args, args...)
	return c, true
}

func (c *Command) setGlobal(name, v string) {
	switch name {
	case "project":
		c.Project = v
	case "by":
		c.By = v
	case "root":
		c.Root = v
	}
}

// heredocStart matches the start of a heredoc: <<EOF, <<-EOF, <<'EOF', <<"EOF".
var heredocStart = regexp.MustCompile(`<<-?\s*(?:'([^']*)'|"([^"]*)"|\\?([A-Za-z_][A-Za-z0-9_]*))`)

// stripHeredocs drops heredoc bodies and the redirection that starts them,
// keeping the rest of the line: `cat <<'EOF' | prep edit 01A --body-file -`
// is still a prep command.
func stripHeredocs(s string) string {
	lines := strings.Split(s, "\n")
	var out []string
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.Contains(line, "<<<") {
			out = append(out, line)
			continue
		}
		var delims []string
		line = heredocStart.ReplaceAllStringFunc(line, func(m string) string {
			sub := heredocStart.FindStringSubmatch(m)
			delims = append(delims, sub[1]+sub[2]+sub[3])
			return ""
		})
		out = append(out, line)
		for _, d := range delims {
			for i+1 < len(lines) {
				i++
				if strings.TrimSpace(lines[i]) == d {
					break
				}
			}
		}
	}
	return strings.Join(out, "\n")
}

// splitSegments splits on unquoted && || ; | and newlines.
func splitSegments(s string) []string {
	var segs []string
	var cur strings.Builder
	var quote rune
	esc := false
	rs := []rune(s)
	flush := func() {
		if t := strings.TrimSpace(cur.String()); t != "" {
			segs = append(segs, t)
		}
		cur.Reset()
	}
	for i := 0; i < len(rs); i++ {
		r := rs[i]
		switch {
		case esc:
			esc = false
			cur.WriteRune(r)
		case r == '\\' && quote != '\'':
			esc = true
			cur.WriteRune(r)
		case quote != 0:
			if r == quote {
				quote = 0
			}
			cur.WriteRune(r)
		case r == '\'' || r == '"':
			quote = r
			cur.WriteRune(r)
		case r == '&' && i+1 < len(rs) && rs[i+1] == '&', r == '|' && i+1 < len(rs) && rs[i+1] == '|':
			flush()
			i++
		case r == ';' || r == '|' || r == '\n' || r == '(' || r == ')' || r == '{' || r == '}':
			flush()
		case r == '&' && (i > 0 && rs[i-1] == '>' || i+1 < len(rs) && rs[i+1] == '>'):
			cur.WriteRune(r) // 2>&1, &>
		case r == '&':
			flush()
		default:
			cur.WriteRune(r)
		}
	}
	flush()
	return segs
}

// tokenize splits a segment into words, honoring quotes and backslashes,
// and stops at a redirection or comment.
func tokenize(s string) []string {
	var words []string
	var cur strings.Builder
	has := false
	var quote rune
	esc := false
	for _, r := range s {
		switch {
		case esc:
			esc = false
			cur.WriteRune(r)
		case r == '\\' && quote != '\'':
			esc = true
			has = true
		case quote != 0:
			if r == quote {
				quote = 0
			} else {
				cur.WriteRune(r)
			}
		case r == '\'' || r == '"':
			quote = r
			has = true
		case r == ' ' || r == '\t':
			if has {
				words = append(words, cur.String())
				cur.Reset()
				has = false
			}
		default:
			cur.WriteRune(r)
			has = true
		}
	}
	if has {
		words = append(words, cur.String())
	}
	// drop redirections (>, >>, 2>&1, <) and everything after a comment
	var out []string
	for i := 0; i < len(words); i++ {
		w := words[i]
		if strings.HasPrefix(w, "#") {
			break
		}
		if w == ">" || w == ">>" || w == "<" || w == "2>" || w == "&>" {
			i++
			continue
		}
		if strings.HasPrefix(w, ">") || strings.HasPrefix(w, "2>") || strings.HasPrefix(w, "<") || strings.HasPrefix(w, "&>") {
			continue
		}
		out = append(out, w)
	}
	return out
}
