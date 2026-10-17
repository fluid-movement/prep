package fixture

import (
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
    issues:
      - {alias: login, title: Login page, state: ready, body: 'Build the login page.'}
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

func TestWorkspaceProjectSelection(t *testing.T) {
	w := buildWS(t, twoProjects)
	if w.Primary != "web" || strings.Join(w.Order, ",") != "api,web" {
		t.Fatalf("primary %q order %v", w.Primary, w.Order)
	}
	web, api := w.Projects["web"], w.Projects["api"]
	out, _ := web.Prep("projects")
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
		"removed links field": "projects:\n  a: {links: [zzz]}\n",
		"unknown project":     "projects:\n  a:\n    issues: [{title: x, depends_on: [zzz/y]}]\n",
		"unknown alias":       "projects:\n  a:\n    issues: [{title: x, depends_on: [b/y]}]\n  b: {}\n",
		"bad primary":         "primary: q\nprojects:\n  a: {}\n",
	} {
		if _, err := LoadWorkspace(strings.NewReader(src)); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
}
