package mdstore

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/fluid-movement/prep/internal/domain"
	"gopkg.in/yaml.v3"
)

// splitFrontmatter separates a leading YAML frontmatter block from the body.
// ok is false when the file does not start with a frontmatter fence.
func splitFrontmatter(raw string) (fm, body string, ok bool, err error) {
	raw = strings.TrimPrefix(strings.ReplaceAll(raw, "\r\n", "\n"), "\ufeff") // a BOM some editors add
	if !strings.HasPrefix(raw, "---\n") {
		return "", raw, false, nil
	}
	rest := raw[4:]
	if strings.HasPrefix(rest, "---\n") || rest == "---" {
		return "", strings.TrimPrefix(strings.TrimPrefix(rest, "---"), "\n"), true, nil
	}
	end := strings.Index(rest, "\n---\n")
	if end < 0 {
		if strings.HasSuffix(rest, "\n---") {
			return rest[:len(rest)-4], "", true, nil
		}
		return "", "", true, fmt.Errorf("frontmatter is not closed with ---")
	}
	return rest[:end+1], rest[end+5:], true, nil
}

// decodeStrict decodes YAML into v, rejecting unknown keys.
func decodeStrict(fm string, v any) error {
	if strings.TrimSpace(fm) == "" {
		return nil
	}
	dec := yaml.NewDecoder(strings.NewReader(fm))
	dec.KnownFields(true)
	if err := dec.Decode(v); err != nil {
		return cleanYAMLErr(err)
	}
	return nil
}

func cleanYAMLErr(err error) error {
	return fmt.Errorf("%s", strings.TrimPrefix(err.Error(), "yaml: "))
}

// encodeYAML renders v with two-space indentation; struct field order is the
// canonical key order.
func encodeYAML(v any) string {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(v); err != nil {
		panic(err) // only fixed, well-formed structs are encoded
	}
	_ = enc.Close()
	s := buf.String()
	if s == "{}\n" {
		return ""
	}
	return s
}

// withFrontmatter renders a canonical file: frontmatter, a blank line, body.
func withFrontmatter(fm, body string) string {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString(fm)
	b.WriteString("---\n")
	if body != "" {
		b.WriteString("\n")
		b.WriteString(body)
		b.WriteString("\n")
	}
	return b.String()
}

// normalize is the canonical text form: LF line endings, no trailing
// whitespace, single blank lines outside code fences, no leading or trailing
// blank lines. It never ends with a newline; renderers add one.
func normalize(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	var fence domain.Fence
	blank := false
	for _, l := range lines {
		in, delim := fence.Line(l)
		code := in && !delim
		if !code {
			l = strings.TrimRight(l, " \t")
		}
		if l == "" && !code {
			if blank {
				continue
			}
			blank = true
		} else {
			blank = false
		}
		out = append(out, l)
	}
	return strings.Trim(strings.Join(out, "\n"), "\n")
}

// fileText renders a plain markdown file canonically.
func fileText(body string) string {
	body = normalize(body)
	if body == "" {
		return ""
	}
	return body + "\n"
}

var checkboxRe = regexp.MustCompile(`^(\s*)[-*+]\s+\[([ xX])\]\s+(.*)$`)

// normalizeChecklist canonicalizes checkbox bullets to "- [ ]" / "- [x]".
func normalizeChecklist(s string) string {
	lines := strings.Split(s, "\n")
	for k, l := range lines {
		if m := checkboxRe.FindStringSubmatch(l); m != nil {
			mark := " "
			if m[2] != " " {
				mark = "x"
			}
			lines[k] = m[1] + "- [" + mark + "] " + m[3]
		}
	}
	return strings.Join(lines, "\n")
}
