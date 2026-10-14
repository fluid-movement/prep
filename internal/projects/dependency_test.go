package projects

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fluid-movement/prep/internal/domain"
)

func TestResolveExternalTreatsMalformedTargetAsUnknown(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	const sourceID = "01M4CCCCCCCCCCCCCCCCCCCCCC"
	const targetID = "01M4BBBBBBBBBBBBBBBBBBBBBB"
	const dependentID = "01M4AAAAAAAAAAAAAAAAAAAAAA"
	source := identified(t, sourceID, "api")
	issueDir := filepath.Join(source, ".prep", "issues", targetID)
	if err := os.MkdirAll(issueDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(issueDir, "issue.md"), []byte("---\ntitle: [invalid\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Touch(source, time.Now()); err != nil {
		t.Fatal(err)
	}
	ref := "api:" + targetID
	i := &domain.Issue{ID: dependentID, Title: "Depends", Kind: domain.KindManual, DependsOn: []string{ref}}
	tree := domain.NewTree(domain.Project{Links: []domain.ProjectLink{{Name: "api", ID: sourceID}}}, []*domain.Issue{i}, nil, nil)
	ResolveExternal(tree)
	x := tree.XIssueOf(ref)
	if x.Outcome != domain.DependencyUnknown || x.Status != domain.XUnknown || !strings.Contains(x.Uncertainty, "invalid") {
		t.Fatalf("malformed source observation should be unknown: %#v", x)
	}
	if !tree.Blocked(dependentID) {
		t.Fatal("malformed external target must block claim")
	}
	change, err := tree.PlanDependencyOverride(dependentID, domain.DependencyOverrideInput{Ref: ref, Actor: "human:alex", Reason: "Proceed while target is malformed", Now: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	withOverride := tree.Apply(change)
	if withOverride.DependencyOverrideStatus(dependentID, ref) != "current" {
		t.Fatal("override should match malformed source observation")
	}
	if err := os.WriteFile(filepath.Join(issueDir, "issue.md"), []byte("---\ntitle: [still invalid but changed\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ResolveExternal(withOverride)
	if withOverride.DependencyOverrideStatus(dependentID, ref) != "stale" {
		t.Fatal("uncommitted target record change must stale override")
	}
}

func TestResolveExternalChangedProjectIdentityDoesNotSatisfyDependency(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	const linkedID = "01M4CCCCCCCCCCCCCCCCCCCCCC"
	const actualID = "01M4DDDDDDDDDDDDDDDDDDDDDD"
	const targetID = "01M4BBBBBBBBBBBBBBBBBBBBBB"
	const dependentID = "01M4AAAAAAAAAAAAAAAAAAAAAA"
	source := identified(t, linkedID, "api")
	if err := Touch(source, time.Now()); err != nil {
		t.Fatal(err)
	}
	projectFile := filepath.Join(source, ".prep", "project.md")
	b, err := os.ReadFile(projectFile)
	if err != nil {
		t.Fatal(err)
	}
	updated := strings.Replace(string(b), "id: "+linkedID, "id: "+actualID, 1)
	if err := os.WriteFile(projectFile, []byte(updated), 0o644); err != nil {
		t.Fatal(err)
	}
	ref := "api:" + targetID
	i := &domain.Issue{ID: dependentID, Title: "Depends", Kind: domain.KindManual, DependsOn: []string{ref}}
	tree := domain.NewTree(domain.Project{Links: []domain.ProjectLink{{Name: "api", ID: linkedID}}}, []*domain.Issue{i}, nil, nil)
	ResolveExternal(tree)
	x := tree.XIssueOf(ref)
	if x.Outcome != domain.DependencyUnknown || !strings.Contains(x.Uncertainty, "not checked out") {
		t.Fatalf("replaced project checkout should stay unknown: %#v", x)
	}
}

func TestResolveExternalClearsStaleObservationsWhenNoRefsRemain(t *testing.T) {
	tree := domain.NewTree(domain.Project{}, nil, nil, nil)
	tree.External = map[string]domain.XIssue{"old:01M4BBBBBBBBBBBBBBBBBBBBBB": {Outcome: domain.DependencySatisfied}}
	ResolveExternal(tree)
	if len(tree.External) != 0 {
		t.Fatalf("stale observations survived refresh: %#v", tree.External)
	}
}
