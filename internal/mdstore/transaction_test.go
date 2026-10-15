package mdstore

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fluid-movement/prep/internal/domain"
)

func transactionIssue(t *testing.T, s *Store) string {
	t.Helper()
	if _, err := s.Init("01M4AAAAAAAAAAAAAAAAAAAAAA", "test"); err != nil {
		t.Fatal(err)
	}
	c, err := domain.NewTree(domain.Project{}, nil, nil, nil).PlanNew(
		domain.NewIssueInput{Title: "Write transaction", Kind: domain.KindCode, Body: "Exercise recovery."},
		time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Apply(c); err != nil {
		t.Fatal(err)
	}
	return c.IssueID
}

func TestApplySerializesStaleWritersAcrossPreflightAndRename(t *testing.T) {
	root := t.TempDir()
	id := transactionIssue(t, Open(root))
	a, b := Open(root), Open(root)
	for _, s := range []*Store{a, b} {
		if _, _, _, err := s.Load(); err != nil {
			t.Fatal(err)
		}
	}
	rel := IssueDir(id) + "/history.md"
	entered, release := make(chan struct{}), make(chan struct{})
	a.txHook = func(stage string, _ int) error {
		if stage == "before-journal" {
			close(entered)
			<-release
		}
		return nil
	}
	first := make(chan error, 1)
	go func() { first <- a.write(rel, "writer A\n") }()
	<-entered // A has checked the loaded hash and still holds the checkout lock.
	second := make(chan error, 1)
	go func() { second <- b.write(rel, "writer B\n") }()
	close(release)
	if err := <-first; err != nil {
		t.Fatal(err)
	}
	if err := <-second; err == nil {
		t.Fatal("stale writer unexpectedly succeeded")
	} else {
		var de *domain.Error
		if !errors.As(err, &de) || de.Code != domain.ErrConflict {
			t.Fatalf("stale writer error = %v", err)
		}
	}
	got, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil || string(got) != "writer A\n" {
		t.Fatalf("stored value = %q, %v", got, err)
	}
}

func TestApplyRecoversInterruptedMultiFileTransitionBeforeLoad(t *testing.T) {
	root := t.TempDir()
	s := Open(root)
	if _, err := s.Init("01M4AAAAAAAAAAAAAAAAAAAAAA", "test"); err != nil {
		t.Fatal(err)
	}
	c, err := domain.NewTree(domain.Project{}, nil, nil, nil).PlanNew(
		domain.NewIssueInput{Title: "Recover me", Kind: domain.KindCode, Body: "A multi-file create."},
		time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatal(err)
	}
	injected := errors.New("injected process interruption")
	s.txHook = func(stage string, index int) error {
		if stage == "after-install" && index == 0 {
			return injected
		}
		return nil
	}
	if _, err := s.Apply(c); !errors.Is(err, injected) {
		t.Fatalf("Apply error = %v", err)
	}
	if _, err := os.Stat(s.abs(transactionFile)); err != nil {
		t.Fatalf("journal not retained: %v", err)
	}
	// A restart has no failure hook. Load recovers before observing records.
	restarted := Open(root)
	p, issues, _, err := restarted.Load()
	if err != nil {
		t.Fatal(err)
	}
	if p.ID == "" || len(issues) != 1 || issues[0].ID != c.IssueID {
		t.Fatalf("incomplete recovered tree: project=%q issues=%v", p.ID, issues)
	}
	if _, err := os.Stat(restarted.abs(transactionFile)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("journal remains after recovery: %v", err)
	}
}

func TestApplyRecoversJournaledWriteBeforeFirstInstall(t *testing.T) {
	root := t.TempDir()
	s := Open(root)
	if _, err := s.Init("01M4AAAAAAAAAAAAAAAAAAAAAA", "test"); err != nil {
		t.Fatal(err)
	}
	c, err := domain.NewTree(domain.Project{}, nil, nil, nil).PlanNew(
		domain.NewIssueInput{Title: "Recover staged change", Kind: domain.KindCode, Body: "The journal precedes installs."},
		time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatal(err)
	}
	injected := errors.New("interrupted after journal sync")
	s.txHook = func(stage string, _ int) error {
		if stage == "after-journal" {
			return injected
		}
		return nil
	}
	if _, err := s.Apply(c); !errors.Is(err, injected) {
		t.Fatalf("Apply error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, IssueDir(c.IssueID), "issue.md")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("write installed before injected interruption: %v", err)
	}
	restarted := Open(root)
	if _, issues, _, err := restarted.Load(); err != nil || len(issues) != 1 || issues[0].ID != c.IssueID {
		t.Fatalf("Load did not roll forward staged transaction: issues=%v err=%v", issues, err)
	}
}

func TestRecoveryRefusesToOverwriteUnexpectedDirectEdit(t *testing.T) {
	root := t.TempDir()
	s := Open(root)
	if _, err := s.Init("01M4AAAAAAAAAAAAAAAAAAAAAA", "test"); err != nil {
		t.Fatal(err)
	}
	c, err := domain.NewTree(domain.Project{}, nil, nil, nil).PlanNew(
		domain.NewIssueInput{Title: "Recover conflict", Kind: domain.KindCode, Body: "A multi-file create."},
		time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatal(err)
	}
	s.txHook = func(stage string, index int) error {
		if stage == "after-install" && index == 0 {
			return errors.New("interrupt")
		}
		return nil
	}
	if _, err := s.Apply(c); err == nil {
		t.Fatal("expected injected interruption")
	}
	s.txHook = nil
	first := filepath.Join(root, IssueDir(c.IssueID), "issue.md")
	if err := os.WriteFile(first, []byte("external edit\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, _, err = Open(root).Load()
	var de *domain.Error
	if !errors.As(err, &de) || de.Code != domain.ErrConflict {
		t.Fatalf("recovery error = %v", err)
	}
	got, err := os.ReadFile(first)
	if err != nil || string(got) != "external edit\n" {
		t.Fatalf("external edit overwritten: %q, %v", got, err)
	}
	if _, err := os.Stat(filepath.Join(root, transactionFile)); err != nil {
		t.Fatalf("conflicting journal was not retained: %v", err)
	}
}

func TestRecoveryValidatesWholeManifestBeforeAnyInstall(t *testing.T) {
	root := t.TempDir()
	s := Open(root)
	if _, err := s.Init("01M4AAAAAAAAAAAAAAAAAAAAAA", "test"); err != nil {
		t.Fatal(err)
	}
	projectPath := filepath.Join(root, Dir, "project.md")
	old, err := os.ReadFile(projectPath)
	if err != nil {
		t.Fatal(err)
	}
	newContent := []byte("new project identity\n")
	tempName := tempPrefix + "staged-project"
	if err := os.WriteFile(filepath.Join(root, Dir, tempName), newContent, 0o644); err != nil {
		t.Fatal(err)
	}
	tx := writeTransaction{Version: 1, Entries: []transactionEntry{
		{Path: Dir + "/project.md", Temp: tempName, Before: hashString(old), After: hashString(newContent)},
		{Path: lockFile, Remove: true}, // forbidden metadata target follows a valid record.
	}}
	b, err := json.Marshal(tx)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(s.abs(transactionFile), b, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := Open(root).Load(); err == nil {
		t.Fatal("malformed manifest unexpectedly recovered")
	}
	got, err := os.ReadFile(projectPath)
	if err != nil || string(got) != string(old) {
		t.Fatalf("earlier valid target was partially installed: %q, %v", got, err)
	}
}

func TestLoadOutsideProjectDoesNotCreatePrepOrLock(t *testing.T) {
	root := t.TempDir()
	if _, _, _, err := Open(root).Load(); err == nil {
		t.Fatal("Load without .prep unexpectedly succeeded")
	}
	if _, err := os.Stat(filepath.Join(root, Dir)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("read created .prep: %v", err)
	}
}

func TestStaleClaimCannotIgnoreAppendedFailureEvidence(t *testing.T) {
	root := t.TempDir()
	id := transactionIssue(t, Open(root))
	rel := IssueDir(id) + "/verification.md"
	now := time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)
	pass := domain.VerificationEvidence{Contract: "contract", Requirement: "check", Result: "pass", Method: "manual",
		Observation: "Earlier check passed", Provenance: "reported", By: "agent/test", At: now, Scopes: []string{"src"}, Snapshot: strings.Repeat("a", 64)}
	pass.ID = pass.ContentID()
	if err := os.WriteFile(filepath.Join(root, rel), []byte(renderVerification([]domain.VerificationEvidence{pass})), 0o644); err != nil {
		t.Fatal(err)
	}
	stale := Open(root)
	if _, _, _, err := stale.Load(); err != nil {
		t.Fatal(err)
	}
	fail := pass
	fail.Previous, fail.Result, fail.Observation = pass.ID, "fail", "A later check failed"
	fail.ID = fail.ContentID()
	if err := os.WriteFile(filepath.Join(root, rel), []byte(renderVerification([]domain.VerificationEvidence{pass, fail})), 0o644); err != nil {
		t.Fatal(err)
	}
	claim := &domain.Change{Op: domain.OpClaim, IssueID: id, Claim: &domain.Claim{By: "agent/test", At: now}}
	if _, err := stale.Apply(claim); err == nil {
		t.Fatal("stale claim succeeded after failure evidence was appended")
	} else {
		var de *domain.Error
		if !errors.As(err, &de) || de.Code != domain.ErrConflict {
			t.Fatalf("stale claim error = %v", err)
		}
	}
	if _, err := os.Stat(filepath.Join(root, IssueDir(id), "claim.md")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("stale claim record was written: %v", err)
	}
	current, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		t.Fatal(err)
	}
	if parsed, err := parseVerification(string(current)); err != nil || len(parsed) != 2 || parsed[1].Result != "fail" {
		t.Fatalf("failure evidence was not preserved: %#v, %v", parsed, err)
	}
}

func TestApplyAllowsBaselinePath(t *testing.T) {
	root := t.TempDir()
	s := Open(root)
	id := transactionIssue(t, s)
	project, issues, diagnostics, err := Open(root).Load()
	if err != nil {
		t.Fatal(err)
	}
	tree := domain.NewTree(project, issues, nil, diagnostics)
	c, err := tree.Plan(id, domain.OpDefine, domain.Input{Actor: "human:test", Now: time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Open(root).Apply(c); err != nil {
		t.Fatalf("baseline transaction: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, IssueDir(id), "baselines", c.Baseline.Name+".md")); err != nil {
		t.Fatalf("baseline was not written: %v", err)
	}
}

func TestEmptyDirectoryManifestMatchesLoadAndApply(t *testing.T) {
	root := t.TempDir()
	rel := IssueDir("01M4AAAAAAAAAAAAAAAAAAAAAA") + "/baselines"
	if err := os.MkdirAll(filepath.Join(root, rel), 0o755); err != nil {
		t.Fatal(err)
	}
	s := Open(root)
	entries, err := os.ReadDir(filepath.Join(root, rel))
	if err != nil {
		t.Fatal(err)
	}
	s.loadedDirs = make(map[string]string)
	s.recordDirectoryEntries(rel, entries, nil)
	got, err := s.directoryManifest(rel)
	if err != nil {
		t.Fatal(err)
	}
	if want := s.loadedDirs[rel]; got != want {
		t.Fatalf("empty directory manifest differs: load=%q apply=%q", want, got)
	}
}

func TestStalePlanCannotMissNewBaselineInExistingIssueDirectory(t *testing.T) {
	root := t.TempDir()
	id := transactionIssue(t, Open(root))
	loadPlan := func(now time.Time) (*Store, *domain.Change) {
		t.Helper()
		s := Open(root)
		project, issues, diagnostics, err := s.Load()
		if err != nil {
			t.Fatal(err)
		}
		c, err := domain.NewTree(project, issues, nil, diagnostics).Plan(id, domain.OpDefine, domain.Input{Actor: "human:test", Now: now})
		if err != nil {
			t.Fatal(err)
		}
		return s, c
	}
	staleStore, stalePlan := loadPlan(time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC))
	freshStore, freshPlan := loadPlan(time.Date(2026, 10, 14, 12, 1, 0, 0, time.UTC))
	if _, err := freshStore.Apply(freshPlan); err != nil {
		t.Fatalf("fresh baseline Apply: %v", err)
	}
	if err := freshStore.validateDirectorySnapshot(); err != nil {
		t.Fatalf("successful Apply did not refresh directory manifests: %v", err)
	}
	if _, err := staleStore.Apply(stalePlan); err == nil {
		t.Fatal("stale plan unexpectedly succeeded after a baseline entry was added")
	} else {
		var de *domain.Error
		if !errors.As(err, &de) || de.Code != domain.ErrConflict {
			t.Fatalf("stale plan error = %v", err)
		}
		if !strings.Contains(de.Message, "directory contents changed") {
			t.Fatalf("stale plan did not fail on the directory manifest: %v", err)
		}
	}
	_, issues, _, err := Open(root).Load()
	if err != nil {
		t.Fatal(err)
	}
	var issue *domain.Issue
	for _, candidate := range issues {
		if candidate.ID == id {
			issue = candidate
		}
	}
	if issue == nil || len(issue.Baselines) != 1 {
		t.Fatalf("fresh baseline missing or duplicate: issue=%#v", issue)
	}
}

func TestStaleClaimCannotMissNewChildIssueDirectory(t *testing.T) {
	root := t.TempDir()
	setup := Open(root)
	if _, err := setup.Init("01M4AAAAAAAAAAAAAAAAAAAAAA", "test"); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)
	loadTree := func(s *Store) *domain.Tree {
		t.Helper()
		project, issues, diagnostics, err := s.Load()
		if err != nil {
			t.Fatal(err)
		}
		return domain.NewTree(project, issues, nil, diagnostics)
	}
	newParent, err := domain.NewTree(domain.Project{}, nil, nil, nil).PlanNew(domain.NewIssueInput{
		Title: "Parent", Kind: domain.KindManual, Body: "A ready parent before a child exists.",
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := setup.Apply(newParent); err != nil {
		t.Fatal(err)
	}
	parentID := newParent.IssueID
	policy, err := loadTree(setup).PlanPolicy(domain.CompletionPolicy{Verification: "Inspect", Documentation: "Review", By: "human:test", At: now})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := setup.Apply(policy); err != nil {
		t.Fatal(err)
	}
	criteria, err := loadTree(setup).PlanRecord(parentID, domain.OpCriterion, domain.RecordInput{
		Actor: "human:test", Now: now, Acceptance: []domain.AcceptanceOp{{Op: "add", Text: "Parent behavior is sound"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := setup.Apply(criteria); err != nil {
		t.Fatal(err)
	}
	defined, err := loadTree(setup).Plan(parentID, domain.OpDefine, domain.Input{Actor: "human:test", Now: now})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := setup.Apply(defined); err != nil {
		t.Fatal(err)
	}
	ready, err := loadTree(setup).Plan(parentID, domain.OpReady, domain.Input{Actor: "human:test", Now: now})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := setup.Apply(ready); err != nil {
		t.Fatal(err)
	}
	stale := Open(root)
	claim, err := loadTree(stale).Plan(parentID, domain.OpClaim, domain.Input{Actor: "agent:test", Now: now})
	if err != nil {
		t.Fatal(err)
	}
	childWriter := Open(root)
	child, err := loadTree(childWriter).PlanNew(domain.NewIssueInput{Title: "New child", Kind: domain.KindManual, Parent: parentID, Body: "A child changes the parent's leaf gate."}, now.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := childWriter.Apply(child); err != nil {
		t.Fatal(err)
	}
	if _, err := stale.Apply(claim); err == nil {
		t.Fatal("stale parent claim succeeded after a child issue was added")
	} else {
		var de *domain.Error
		if !errors.As(err, &de) || de.Code != domain.ErrConflict {
			t.Fatalf("stale claim error = %v", err)
		}
	}
	if _, err := os.Stat(filepath.Join(root, IssueDir(parentID), "claim.md")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("stale parent claim record was written: %v", err)
	}
}
