package fixture

import (
	"encoding/json"
	"strings"
	"testing"
)

const twoProjects = `
primary: web
projects:
  api:
    issues:
      - {alias: auth, title: Auth endpoint}
    knowledge:
      - {path: /components/auth.md, type: component, title: Auth, description: Auth service, body: Tokens.}
  web:
    links: [api]
    issues:
      - {alias: login, title: Login page, state: ready, depends_on: [api/auth], body: 'Needs [auth](api:/components/auth.md) and {{api/auth}}.'}
`

func buildWS(t *testing.T, src string) *Workspace {
	t.Helper()
	ws, err := LoadWorkspace(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	w, err := BuildWorkspace(ws, t.TempDir(), Options{Binary: prepBin})
	if err != nil {
		t.Fatal(err)
	}
	return w
}

func TestWorkspaceCrossProject(t *testing.T) {
	w := buildWS(t, twoProjects)
	if w.Primary != "web" || strings.Join(w.Order, ",") != "api,web" {
		t.Fatalf("primary %q order %v", w.Primary, w.Order)
	}
	web, api := w.Projects["web"], w.Projects["api"]
	out, err := web.Prep("list", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var list struct {
		Issues []struct {
			ID        string `json:"id"`
			Blocked   bool   `json:"blocked"`
			CrossDeps []any  `json:"cross_deps"`
		} `json:"issues"`
	}
	if err := json.Unmarshal([]byte(out), &list); err != nil {
		t.Fatalf("%v\n%.300s", err, out)
	}
	if len(list.Issues) != 1 || !list.Issues[0].Blocked || len(list.Issues[0].CrossDeps) != 1 {
		t.Errorf("web list: %.600s", out)
	}
	if out, _ := web.Prep("show", web.IDs["login"]); !strings.Contains(out, api.IDs["auth"]) {
		t.Errorf("cross dependency or expanded alias missing:\n%s", out)
	}
	out, _ = web.Prep("projects")
	if !strings.Contains(out, "api") || !strings.Contains(out, "web") {
		t.Errorf("projects:\n%s", out)
	}
	for _, p := range []*Project{web, api} {
		if out, err := p.Prep("check"); err != nil || !strings.Contains(out, "0 errors") || !strings.Contains(out, "0 warnings") {
			t.Errorf("%s check: %v\n%s", p.Name, err, out)
		}
	}
	// --project reaches the other project through the workspace registry.
	if out, err := web.Prep("--project", "api", "show", api.IDs["auth"]); err != nil || !strings.Contains(out, "Auth endpoint") {
		t.Errorf("--project: %v\n%s", err, out)
	}
}

func TestWorkspaceDeterministicAndIsolated(t *testing.T) {
	a, b := buildWS(t, twoProjects), buildWS(t, twoProjects)
	for _, n := range []string{"api", "web"} {
		if treeHash(t, a.Projects[n].Dir) != treeHash(t, b.Projects[n].Dir) {
			t.Errorf("%s differs between builds", n)
		}
	}
	if !strings.HasPrefix(a.State, a.Dir) {
		t.Errorf("registry %q is outside the workspace %q", a.State, a.Dir)
	}
}

func TestWorkspaceErrors(t *testing.T) {
	for name, src := range map[string]string{
		"cycle":           "projects:\n  a: {links: [b]}\n  b: {links: [a]}\n",
		"unknown link":    "projects:\n  a: {links: [zzz]}\n",
		"unknown project": "projects:\n  a:\n    issues: [{title: x, depends_on: [zzz/y]}]\n",
		"unknown alias":   "projects:\n  a:\n    issues: [{title: x, depends_on: [b/y]}]\n  b: {}\n",
		"bad primary":     "primary: q\nprojects:\n  a: {}\n",
	} {
		if _, err := LoadWorkspace(strings.NewReader(src)); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
}
