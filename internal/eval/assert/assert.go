// Package assert is the core of the checks a test case makes: the
// Assertion interface, the registry the case loader decodes kinds from, and
// the input an assertion sees. The kinds live in assert/tools and
// assert/outcome and register themselves.
package assert

import (
	"fmt"
	"sort"
	"sync"

	"gopkg.in/yaml.v3"

	"github.com/fluid-movement/prep/internal/eval/fixture"
	"github.com/fluid-movement/prep/internal/eval/transcript"
)

// Group is where in a case an assertion is written.
type Group string

const (
	// Tools assertions read the transcript.
	Tools Group = "tools"
	// Outcome assertions read the workspace the agent left.
	Outcome Group = "outcome"
)

// Result is what one assertion found.
type Result struct {
	Kind string `json:"kind"`
	// Desc is the assertion in words, for reports.
	Desc   string `json:"desc"`
	Pass   bool   `json:"pass"`
	Reason string `json:"reason"`
	// Step is the transcript step that decided the result, or -1.
	Step int `json:"step"`
}

// Input is everything an assertion may look at.
type Input struct {
	Transcript *transcript.Transcript
	Workspace  *fixture.Workspace
	// Project is the primary project, the one the agent started in.
	Project *fixture.Project
	// Before is the snapshot taken after the fixture was built and before
	// the agent ran; After is the state when it finished.
	Before, After *Snapshot
}

// Assertion is one check.
type Assertion interface {
	// Kind is the name used in case files.
	Kind() string
	// Describe says what is expected, in words.
	Describe() string
	Check(in *Input) Result
}

// Pass and Fail build results for an assertion.
func Pass(a Assertion, reason string, step int) Result {
	return Result{Kind: a.Kind(), Desc: a.Describe(), Pass: true, Reason: reason, Step: step}
}

func Fail(a Assertion, reason string, step int) Result {
	return Result{Kind: a.Kind(), Desc: a.Describe(), Pass: false, Reason: reason, Step: step}
}

// Decoder builds an assertion from the value of its kind's key in a case.
type Decoder func(n *yaml.Node) (Assertion, error)

var (
	mu       sync.RWMutex
	registry = map[Group]map[string]Decoder{}
)

// Register adds a kind to a group; registering twice is a programming error.
func Register(g Group, kind string, d Decoder) {
	mu.Lock()
	defer mu.Unlock()
	if registry[g] == nil {
		registry[g] = map[string]Decoder{}
	}
	if _, dup := registry[g][kind]; dup {
		panic(fmt.Sprintf("assert: %s assertion %q registered twice", g, kind))
	}
	registry[g][kind] = d
}

// Kinds lists the registered kinds of a group, sorted.
func Kinds(g Group) []string {
	mu.RLock()
	defer mu.RUnlock()
	var out []string
	for k := range registry[g] {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// Decode builds an assertion from a mapping node with exactly one key, the
// kind: `- ran: guide` or `- not_ran: {tool: Read, path: ".prep/**"}`.
func Decode(g Group, n *yaml.Node) (Assertion, error) {
	if n.Kind != yaml.MappingNode || len(n.Content) != 2 {
		return nil, fmt.Errorf("line %d: an assertion is a mapping with one key, its kind", n.Line)
	}
	kind := n.Content[0].Value
	mu.RLock()
	d, ok := registry[g][kind]
	mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("line %d: unknown %s assertion %q (known: %v)", n.Content[0].Line, g, kind, Kinds(g))
	}
	a, err := d(n.Content[1])
	if err != nil {
		return nil, fmt.Errorf("line %d: %s: %w", n.Content[0].Line, kind, err)
	}
	return a, nil
}

// Snapshot is the state of a workspace at one moment; see snapshot.go.
type Snapshot struct{}
