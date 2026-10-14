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
