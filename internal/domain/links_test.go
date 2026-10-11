package domain

import (
	"sort"
	"strings"
	"testing"
)

const (
	selfID = "01M4AAAAAAAAAAAAAAAAAAAAAA"
	apiID  = "01M4BBBBBBBBBBBBBBBBBBBBBB"
	webID  = "01M4CCCCCCCCCCCCCCCCCCCCCC"
	dep    = "01M4DDDDDDDDDDDDDDDDDDDDDD"
)

func TestRenameLinkInText(t *testing.T) {
	in := "---\ntitle: X\ndepends_on:\n  - api:" + dep + "\n  - 'api:" + dep + "'\n  - " + dep + "\n---\n\nSee [the API](api:/components/x.md#auth) and [ours](/a.md); apis:/x and api:/plain stay.\n"
	got := RenameLinkInText(in, "api", "backend")
	want := "---\ntitle: X\ndepends_on:\n  - backend:" + dep + "\n  - 'backend:" + dep + "'\n  - " + dep + "\n---\n\nSee [the API](backend:/components/x.md#auth) and [ours](/a.md); apis:/x and api:/plain stay.\n"
	if got != want {
		t.Fatalf("got:\n%s", got)
	}
}

func linkTree(links []ProjectLink, issues ...*Issue) *Tree {
	for _, i := range issues {
		if i.Kind == "" {
			i.Kind = KindCode
		}
	}
	return NewTree(Project{Schema: SchemaVersion, ID: selfID, Links: links}, issues, []*Entry{{Path: "/a.md", Body: "Uses [web](web:/b.md)."}}, nil)
}

func xcodes(ds []Diagnostic) string {
	var out []string
	for _, d := range ds {
		if d.Code[0] == 'X' {
			out = append(out, d.Code)
		}
	}
	sort.Strings(out)
	return strings.Join(out, " ")
}

func TestValidateLinks(t *testing.T) {
	i := &Issue{ID: dep, Title: "T", Context: "See [x](api:/x.md).", DependsOn: []string{"nope:" + apiID}}
	tr := linkTree([]ProjectLink{{Name: "api", ID: apiID}, {Name: "api2", ID: apiID}, {Name: "me", ID: selfID}, {Name: "Bad", ID: webID}}, i)
	if got := xcodes(Validate(tr)); got != "X001 X001 X002 X002 X003" {
		t.Fatalf("got %q", got)
	}
	refs := tr.LinkRefs()
	if len(refs) != 3 || !refs[0].Dep || refs[1].File != "context.md" || refs[2].Entry != "/a.md" {
		t.Fatalf("refs: %+v", refs)
	}
}

func TestPlanLinks(t *testing.T) {
	tr := linkTree([]ProjectLink{{Name: "api", ID: apiID}, {Name: "web", ID: webID}},
		&Issue{ID: dep, Title: "T", Context: "See [x](api:/x.md)."})
	for _, bad := range [][]ProjectLink{
		{{Name: "x", ID: selfID}},
		{{Name: "x", ID: "nope"}},
		{{Name: "X", ID: apiID}},
		{{Name: "a", ID: apiID}, {Name: "b", ID: apiID}},
		{{Name: "a", ID: apiID}, {Name: "a", ID: webID}},
	} {
		if _, err := tr.PlanLinks(bad); err == nil {
			t.Errorf("accepted %+v", bad)
		}
	}
	if _, err := tr.PlanLinkRemove("api"); err == nil || !strings.Contains(err.Error(), "still referred to") {
		t.Fatalf("remove referenced: %v", err)
	}
	if _, err := tr.PlanLinkRemove("web"); err == nil {
		t.Fatal("removed web, which /a.md refers to")
	}
	c, err := tr.PlanLinkRename("api", "backend")
	if err != nil {
		t.Fatal(err)
	}
	after := tr.Apply(c)
	if _, ok := after.Project.Link("backend"); !ok || after.Issues[dep].Context != "See [x](backend:/x.md)." {
		t.Fatalf("after: %+v %q", after.Project.Links, after.Issues[dep].Context)
	}
	if err := tr.CheckWrite(c); err != nil {
		t.Fatal(err)
	}
	if _, err := tr.PlanLinkRename("api", "web"); err == nil {
		t.Fatal("renamed onto an existing name")
	}
}
