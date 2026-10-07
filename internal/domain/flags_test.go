package domain

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func TestQueryFlagsMatchParseFilter(t *testing.T) {
	parent := &Issue{ID: "01M40000000000000000000001", Title: "Release", Kind: KindManual, Tags: []string{"release"}}
	child := &Issue{ID: "01M40000000000000000000002", Title: "Fix", Kind: KindCode, Parent: parent.ID, Tags: []string{"cli", "release"}, Priority: PriorityHigh}
	tree := NewTree(Project{}, []*Issue{parent, child}, nil, nil)
	flags := tree.QueryFlags()

	byName := map[string]FlagInfo{}
	for _, f := range flags {
		byName[f.Name] = f
		// Every described flag and each listed value parses.
		args := []string{"--" + f.Name}
		switch f.Kind {
		case FlagText:
			args = append(args, "word")
		case FlagList:
			if len(f.Values) == 0 {
				t.Fatalf("--%s lists no values", f.Name)
			}
			for _, v := range f.Values {
				if _, err := ParseFilter([]string{"--" + f.Name, v.Value}); err != nil {
					t.Errorf("--%s %s: %v", f.Name, v.Value, err)
				}
			}
			args = append(args, f.Values[0].Value)
		}
		if _, err := ParseFilter(args); err != nil {
			t.Errorf("%v: %v", args, err)
		}
	}
	// Every flag ParseFilter's switch handles is described.
	src, err := os.ReadFile("query.go")
	if err != nil {
		t.Fatal(err)
	}
	body := string(src)
	body = body[strings.Index(body, "func ParseFilter"):]
	body = body[:strings.Index(body, "\n}\n")]
	body = body[strings.Index(body, "switch name {"):]
	cases := regexp.MustCompile(`case "([a-z]+)":`).FindAllStringSubmatch(body, -1)
	if len(cases) < 10 {
		t.Fatalf("found %d flags in ParseFilter", len(cases))
	}
	for _, c := range cases {
		if _, ok := byName[c[1]]; !ok {
			t.Errorf("ParseFilter accepts --%s, QueryFlags does not describe it", c[1])
		}
	}

	values := func(name string) string {
		var out []string
		for _, v := range byName[name].Values {
			out = append(out, v.Value+"="+string(rune('0'+v.Count)))
		}
		return strings.Join(out, " ")
	}
	if got := values("tag"); got != "cli=1 release=2" {
		t.Errorf("tags: %s", got)
	}
	if got := values("priority"); got != "critical=0 high=1 medium=1 low=0" {
		t.Errorf("priorities: %s", got)
	}
	if u := byName["under"].Values; len(u) != 1 || u[0].Value != parent.ID || u[0].Label != "Release" || u[0].Count != 1 {
		t.Errorf("under: %+v", u)
	}
	if c := byName["leaf"].Count; c == nil || *c != 1 {
		t.Errorf("leaf count: %v", c)
	}
}

func TestNegatedFlagValues(t *testing.T) {
	parent := &Issue{ID: "01M40000000000000000000001", Title: "Release", Kind: KindManual, Tags: []string{"release"}}
	child := &Issue{ID: "01M40000000000000000000002", Title: "Fix", Kind: KindCode, Parent: parent.ID, Tags: []string{"cli"}, Priority: PriorityHigh}
	other := &Issue{ID: "01M40000000000000000000003", Title: "Docs", Kind: KindCode}
	tree := NewTree(Project{}, []*Issue{parent, child, other}, nil, nil)
	query := func(args ...string) string {
		t.Helper()
		f, err := ParseFilter(args)
		if err != nil {
			t.Fatalf("%v: %v", args, err)
		}
		ids, err := tree.Query(f)
		if err != nil {
			t.Fatal(err)
		}
		var titles []string
		for _, id := range ids {
			titles = append(titles, tree.Issues[id].Title)
		}
		sort.Strings(titles)
		return strings.Join(titles, " ")
	}
	for _, c := range []struct {
		args []string
		want string
	}{
		{[]string{"--kind", "!manual"}, "Docs Fix"},
		{[]string{"--kind", "!manual,!code"}, ""},
		{[]string{"--kind", "code,!manual"}, "Docs Fix"},
		{[]string{"--tag", "!cli"}, "Docs Release"},
		{[]string{"--tag", "release,!cli"}, "Release"},
		{[]string{"--priority", "!high"}, "Docs Release"},
		{[]string{"--under", "!01"}, "Docs Release"},
		{[]string{"--state", "!open"}, ""},
		{[]string{"--state", "!done", "--kind", "!code"}, "Release"},
		{[]string{"--kind", "not-manual"}, "Docs Fix"},
		{[]string{"--tag", "not-release,not-cli"}, "Docs"},
		{[]string{"--state", "not-open,not-done"}, ""},
	} {
		if got := query(c.args...); got != c.want {
			t.Errorf("%v = %q, want %q", c.args, got, c.want)
		}
	}
	for _, bad := range [][]string{{"--state", "!nope"}, {"--kind", "!"}, {"--priority", "not-urgent"}, {"--state", "in_progress"}} {
		if _, err := ParseFilter(bad); err == nil {
			t.Errorf("%v accepted", bad)
		}
	}
}

func TestTagsCannotStartWithNot(t *testing.T) {
	if _, err := NormalizeTags([]string{"not-ready"}); err == nil || !strings.Contains(err.Error(), "not-") {
		t.Fatalf("not-ready accepted: %v", err)
	}
	if got, err := NormalizeTags([]string{"notes", "knot-x"}); err != nil || len(got) != 2 {
		t.Fatalf("tags merely containing not: %v %v", got, err)
	}
}
