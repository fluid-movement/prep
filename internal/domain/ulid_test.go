package domain

import (
	"bytes"
	"math/rand/v2"
	"testing"
	"time"
)

func TestULIDEncoding(t *testing.T) {
	now := time.Date(2026, 10, 6, 8, 37, 52, 123e6, time.UTC)
	zero, err := NewULID(now, bytes.NewReader(make([]byte, 10)))
	if err != nil {
		t.Fatal(err)
	}
	if !ValidID(zero) || zero[10:] != "0000000000000000" {
		t.Fatalf("NewULID with zero entropy = %q", zero)
	}
	if got := ULIDTime(zero); !got.Equal(now) {
		t.Fatalf("ULIDTime = %v, want %v", got, now)
	}
	full, _ := NewULID(now, bytes.NewReader(bytes.Repeat([]byte{0xff}, 10)))
	if full[:10] != zero[:10] || full[10:] != "ZZZZZZZZZZZZZZZZ" {
		t.Fatalf("NewULID with full entropy = %q", full)
	}
	if b := decodeULID(full); encodeULID(b) != full {
		t.Fatalf("decode/encode round trip of %q", full)
	}
	if _, err := NewULID(now, bytes.NewReader(nil)); err == nil {
		t.Fatal("NewULID accepted short entropy")
	}
}

func TestValidID(t *testing.T) {
	for id, want := range map[string]bool{
		"01K6WQ0YQV8D3M5A2B7C9E1F4G": true,
		"7ZZZZZZZZZZZZZZZZZZZZZZZZZ": true,
		"8ZZZZZZZZZZZZZZZZZZZZZZZZZ": false, // beyond 128 bits
		"01K6WQ0YQV8D3M5A2B7C9E1F4":  false,
		"01k6wq0yqv8d3m5a2b7c9e1f4g": false, // canonical form is upper case
		"01K6WQ0YQV8D3M5A2B7C9E1F4I": false, // I, L, O and U are not Crockford
		"20261006-083752":            false,
		"export-csv":                 false,
	} {
		if ValidID(id) != want {
			t.Errorf("ValidID(%q) = %v", id, !want)
		}
	}
}

func TestPlanNewULIDsSortByCreation(t *testing.T) {
	tree := NewTree(Project{}, nil, nil, nil)
	r := rand.NewChaCha8([32]byte{1})
	now := time.Date(2026, 10, 6, 9, 0, 0, 0, time.UTC)
	var ids []string
	// Ten issues in one millisecond, then one a millisecond later: each
	// sorts after the ones before, whatever the random bits say.
	for k := 0; k < 11; k++ {
		at := now
		if k == 10 {
			at = now.Add(time.Millisecond)
		}
		c, err := tree.PlanNew(NewIssueInput{Title: "T", Kind: KindCode, Entropy: r}, at)
		if err != nil {
			t.Fatal(err)
		}
		if !ValidID(c.IssueID) {
			t.Fatalf("PlanNew created %q", c.IssueID)
		}
		if len(ids) > 0 && c.IssueID <= ids[len(ids)-1] {
			t.Fatalf("%q does not sort after %q", c.IssueID, ids[len(ids)-1])
		}
		ids = append(ids, c.IssueID)
		tree = tree.Apply(c)
	}
	if !ULIDTime(ids[9]).Equal(now) {
		t.Fatalf("same-millisecond ID moved to %v", ULIDTime(ids[9]))
	}
}

func TestNextULIDCarries(t *testing.T) {
	now := time.Date(2026, 10, 6, 9, 0, 0, 0, time.UTC)
	low, _ := NewULID(now, bytes.NewReader(make([]byte, 10)))
	top := low[:10] + "0000000000ZZZZZZ"
	got, err := nextULID(now, bytes.NewReader(make([]byte, 10)), []string{top})
	if err != nil {
		t.Fatal(err)
	}
	if want := low[:10] + "0000000001000000"; got != want {
		t.Fatalf("increment of %q = %q, want %q", top, got, want)
	}
	last := low[:10] + "ZZZZZZZZZZZZZZZZ"
	got, _ = nextULID(now, bytes.NewReader(make([]byte, 20)), []string{last})
	if !ULIDTime(got).Equal(now.Add(time.Millisecond)) {
		t.Fatalf("exhausted millisecond gave %q", got)
	}
}
