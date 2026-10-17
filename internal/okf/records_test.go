package okf

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fluid-movement/prep/internal/domain"
)

func TestRecordNeverOverwrites(t *testing.T) {
	s := &Store{Root: t.TempDir()}
	id := "01M45YYRG0AP2P2A41FRCHGR0Q"
	r := &domain.KnowledgeRecord{ID: id, Path: "/decisions/" + id + ".md", Type: "decision", Title: "Choice", Description: "Why", At: time.Now(), Body: "Body"}
	if _, err := s.Render(r); err != nil {
		t.Fatal(err)
	}
	// A competing writer creates a file after rendering, including unreadable
	// content that would otherwise be absent from the loaded tree.
	p := s.file(r.Path)
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		t.Fatal(err)
	}
	const original = "someone else's record\n"
	if err := os.WriteFile(p, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Apply(r); err == nil {
		t.Fatal("overwrote concurrent record")
	} else {
		var de *domain.Error
		if !errors.As(err, &de) || de.Code != domain.ErrConflict {
			t.Fatalf("error: %v", err)
		}
	}
	b, _ := os.ReadFile(p)
	if string(b) != original {
		t.Fatal("original changed")
	}
	if _, err := s.Render(r); err == nil {
		t.Fatal("render accepted existing unreadable file")
	}
	if err := os.Remove(p); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Render(r); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Apply(r); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Apply(r); err == nil {
		t.Fatal("second apply overwrote record")
	}
}
