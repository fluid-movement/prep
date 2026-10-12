package transcript

import (
	"reflect"
	"testing"
)

func TestParseShell(t *testing.T) {
	cases := []struct {
		line string
		want []Command
	}{
		{"prep list", []Command{{Name: "list", Raw: "prep list"}}},
		{"cd /x && prep guide 01ABC && echo done", []Command{{Name: "guide", Args: []string{"01ABC"}, Raw: "prep guide 01ABC"}}},
		{"prep --project api --by me list --state open", []Command{{Name: "list", Args: []string{"--state", "open"}, Project: "api", By: "me", Raw: "prep --project api --by me list --state open"}}},
		{"prep list --project=api", []Command{{Name: "list", Args: []string{"--project=api"}, Project: "api", Raw: "prep list --project=api"}}},
		{"/usr/local/bin/prep show 01A | head -5", []Command{{Name: "show", Args: []string{"01A"}, Raw: "/usr/local/bin/prep show 01A"}}},
		{"PREP_ACTOR=x prep next", []Command{{Name: "next", Raw: "PREP_ACTOR=x prep next"}}},
		{`prep new --title "A b; c" --kind code`, []Command{{Name: "new", Args: []string{"--title", "A b; c", "--kind", "code"}, Raw: `prep new --title "A b; c" --kind code`}}},
		{"prep edit 01A --body-file - <<'EOF'\nprep list\nEOF", []Command{{Name: "edit", Args: []string{"01A", "--body-file", "-"}, Raw: "prep edit 01A --body-file -"}}},
		{"cat <<'EOF' | prep edit 01A --body-file - --by me\nprep drop inside\nEOF\nprep define 01A", []Command{
			{Name: "edit", Args: []string{"01A", "--body-file", "-", "--by", "me"}, By: "me", Raw: "prep edit 01A --body-file - --by me"},
			{Name: "define", Args: []string{"01A"}, Raw: "prep define 01A"}}},
		{"prep new --title x --body-file - <<-EOT\n\tprep drop inside\n\tEOT", []Command{{Name: "new", Args: []string{"--title", "x", "--body-file", "-"}, Raw: "prep new --title x --body-file -"}}},
		{"echo prep list", nil},
		{`echo "prep list"`, nil},
		{"x=$(prep next --json)", []Command{{Name: "next", Args: []string{"--json"}, Raw: "prep next --json"}}},
		{"prep list > out.txt 2>&1", []Command{{Name: "list", Raw: "prep list > out.txt 2>&1"}}},
		{"prep", nil},
		{"git status; prep check", []Command{{Name: "check", Raw: "prep check"}}},
	}
	for _, c := range cases {
		got := ParseShell(c.line)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("ParseShell(%q)\n got  %+v\n want %+v", c.line, got, c.want)
		}
	}
}

func TestCommandIsAndFlag(t *testing.T) {
	c := ParseShell("prep knowledge find --max 3 auth tokens")[0]
	if !c.Is("knowledge") || !c.Is("knowledge find") || c.Is("knowledge show") || c.Is("list") {
		t.Errorf("Is: %+v", c)
	}
	if v, ok := c.Flag("max"); !ok || v != "3" {
		t.Errorf("Flag max = %q %v", v, ok)
	}
	if _, ok := c.Flag("json"); ok {
		t.Error("json flag found")
	}
	d := ParseShell("prep new --title find --kind code")[0]
	if d.Is("find") || !d.Is("new") {
		t.Errorf("flag value taken as a word: %+v", d)
	}
}

func TestClip(t *testing.T) {
	long := make([]byte, ClipLen+50)
	for i := range long {
		long[i] = 'é'
	}
	got := Clip(string(long))
	if len(got) > ClipLen+20 || got == string(long) {
		t.Errorf("not clipped: %d", len(got))
	}
	if Clip("short") != "short" {
		t.Error("short text changed")
	}
}
