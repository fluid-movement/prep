package cli

import (
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

type hookShellWord struct {
	text string
	op   bool
}

var hookShellAssignment = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=`)

// hookShellWrappers run the command that follows them.
var hookShellWrappers = map[string]bool{"sudo": true, "env": true, "command": true, "exec": true, "nice": true, "nohup": true, "time": true, "timeout": true, "xargs": true, "stdbuf": true}

func isNumber(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if (r < '0' || r > '9') && r != '.' && r != 's' && r != 'm' {
			return false
		}
	}
	return true
}

// Parse literal shell words and operators without expanding or executing them.
// Heredoc contents are values, so consume them before looking for commands.
func hookShellSegments(command string) [][]hookShellWord {
	var segments [][]hookShellWord
	var words []hookShellWord
	var word strings.Builder
	var quote byte
	started := false
	var heredocs []string
	flush := func() {
		if started {
			value := word.String()
			if len(words) > 0 && words[len(words)-1].op && words[len(words)-1].text == "<<" {
				heredocs = append(heredocs, value)
			}
			words = append(words, hookShellWord{text: value})
			word.Reset()
			started = false
		}
	}
	for i := 0; i < len(command); i++ {
		c := command[i]
		if c == '\\' && quote != '\'' && i+1 < len(command) {
			next := command[i+1]
			if quote == 0 || strings.ContainsRune("$`\"\\\n", rune(next)) {
				i++
				if next != '\n' {
					word.WriteByte(next)
					started = true
				}
				continue
			}
		}
		if quote != 0 {
			if c == quote {
				quote = 0
			} else {
				word.WriteByte(c)
			}
			continue
		}
		if c == '\'' || c == '"' {
			quote, started = c, true
			continue
		}
		if c == '#' && !started {
			for i < len(command) && command[i] != '\n' {
				i++
			}
			c = '\n'
		}
		if strings.ContainsRune(";|&\n()<>", rune(c)) {
			flush()
			if c == '<' || c == '>' {
				op := string(c)
				if i+1 < len(command) && command[i+1] == c {
					op += string(c)
					i++
				}
				words = append(words, hookShellWord{text: op, op: true})
				continue
			}
			segments = append(segments, words)
			words = nil
			if c == '\n' {
				for _, delimiter := range heredocs {
					for i+1 < len(command) {
						start := i + 1
						end := strings.IndexByte(command[start:], '\n')
						if end < 0 {
							end = len(command)
						} else {
							end += start
						}
						i = end
						if command[start:end] == delimiter {
							break
						}
					}
				}
				heredocs = nil
			}
			continue
		}
		if c == ' ' || c == '\t' || c == '\r' {
			flush()
			continue
		}
		started = true
		word.WriteByte(c)
	}
	flush()
	return append(segments, words)
}

func hookShellArea(words []hookShellWord, cwd, root string) (area string, write bool) {
	command := ""
	var targets []hookShellWord
	var redirects []hookShellWord
	for i := 0; i < len(words); i++ {
		w := words[i]
		if w.op {
			if i+1 < len(words) {
				if w.text != "<<" {
					redirects = append(redirects, hookShellWord{text: words[i+1].text, op: strings.HasPrefix(w.text, ">")})
				}
				i++
			}
			continue
		}
		if command == "" {
			if hookShellAssignment.MatchString(w.text) {
				continue
			}
			base := filepath.Base(w.text)
			if hookShellWrappers[base] {
				// A wrapper runs the command after it: skip the wrapper, its
				// flags and a numeric argument (timeout 5, nice -n 10).
				for i+1 < len(words) && !words[i+1].op && (strings.HasPrefix(words[i+1].text, "-") || isNumber(words[i+1].text)) {
					i++
				}
				continue
			}
			command = base
		} else {
			targets = append(targets, w)
		}
	}
	if command == "prep" || command == "git" {
		return "", false
	}
	writes := hookWrite.MatchString(command)
	if command == "sed" {
		for _, target := range targets {
			if strings.HasPrefix(target.text, "-i") {
				writes = true
			}
		}
	}
	if writes || hookRead.MatchString(command) {
		for _, target := range targets {
			// Quoted prose is a value; a whole literal path still counts.
			if strings.ContainsFunc(target.text, unicode.IsSpace) {
				continue
			}
			if found := hookRecordArea(target.text, cwd, root); found != "" {
				area = found
				write = writes
			}
		}
	}
	for _, target := range redirects {
		if found := hookRecordArea(target.text, cwd, root); found != "" {
			area = found
			write = write || target.op
		}
	}
	return area, write
}
