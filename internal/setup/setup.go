// Package setup installs, refreshes and removes harness integrations: the
// skills, hooks and commands that make an agent harness understand prep.
// Every harness implements Harness; prep setup drives them through Apply
// and Refresh and records the user's choice in the user configuration.
package setup

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Harness is one agent harness prep can integrate with.
type Harness interface {
	// Name is the stable identifier used in flags and the config file.
	Name() string
	// Title is the human name, such as "Claude Code".
	Title() string
	// Detect reports whether the harness is present on this machine.
	Detect() bool
	// Installed reports the installed integration's version, if any.
	Installed() (version string, ok bool, err error)
	// Install installs the integration for version, or updates it to it.
	// Warnings report what went wrong around it without failing it.
	Install(version string) (warnings []string, err error)
	// Remove removes the integration.
	Remove() error
}

var (
	mu       sync.Mutex
	registry = map[string]Harness{}
)

// Register adds a harness. Implementations register themselves from init;
// tests register fakes.
func Register(h Harness) {
	mu.Lock()
	defer mu.Unlock()
	registry[h.Name()] = h
}

// Unregister removes a harness, for tests.
func Unregister(name string) {
	mu.Lock()
	defer mu.Unlock()
	delete(registry, name)
}

// All returns the registered harnesses sorted by name.
func All() []Harness {
	mu.Lock()
	defer mu.Unlock()
	out := make([]Harness, 0, len(registry))
	for _, h := range registry {
		out = append(out, h)
	}
	sort.Slice(out, func(a, b int) bool { return out[a].Name() < out[b].Name() })
	return out
}

// Get returns a registered harness by name.
func Get(name string) (Harness, bool) {
	mu.Lock()
	defer mu.Unlock()
	h, ok := registry[name]
	return h, ok
}

// Names lists the registered harness names.
func Names() []string {
	var out []string
	for _, h := range All() {
		out = append(out, h.Name())
	}
	return out
}

// Status is what is known about one harness on this machine.
type Status struct {
	Name      string `json:"name"`
	Title     string `json:"title"`
	Detected  bool   `json:"detected"`
	Installed string `json:"installed,omitempty"` // version; empty when not installed
	Chosen    bool   `json:"chosen"`              // listed in the user configuration
	Err       string `json:"error,omitempty"`
}

// Survey reports the status of every registered harness.
func Survey(chosen []string) []Status {
	var out []Status
	for _, h := range All() {
		s := Status{Name: h.Name(), Title: h.Title(), Detected: h.Detect(), Chosen: contains(chosen, h.Name())}
		if v, ok, err := h.Installed(); err != nil {
			s.Err = err.Error()
		} else if ok {
			s.Installed = v
		}
		out = append(out, s)
	}
	return out
}

// Result is what happened to one harness.
type Result struct {
	Name    string `json:"name"`
	Action  string `json:"action"` // installed, updated, removed, up to date, skipped, failed
	Version string `json:"version,omitempty"`
	Err     string `json:"error,omitempty"`
	// Warnings from a successful install, such as other plugins it could not keep.
	Warnings []string `json:"warnings,omitempty"`
}

// Validate checks that every name is a registered harness.
func Validate(names []string) error {
	var unknown []string
	for _, n := range names {
		if _, ok := Get(n); !ok {
			unknown = append(unknown, n)
		}
	}
	if len(unknown) > 0 {
		avail := strings.Join(Names(), ", ")
		if avail == "" {
			avail = "none in this version"
		}
		return fmt.Errorf("unknown harness %s; available: %s", strings.Join(unknown, ", "), avail)
	}
	return nil
}

// Apply makes the selection true: every selected harness gets the
// integration at version, every other harness that has one installed loses
// it. It returns one result per touched harness and keeps going after a
// failure, so one broken harness does not block the others.
func Apply(selected []string, version string) []Result {
	var out []Result
	for _, h := range All() {
		v, installed, err := h.Installed()
		switch {
		case contains(selected, h.Name()):
			out = append(out, install(h, v, installed, err, version))
		case err == nil && installed:
			r := Result{Name: h.Name(), Action: "removed"}
			if err := h.Remove(); err != nil {
				r.Action, r.Err = "failed", err.Error()
			}
			out = append(out, r)
		}
	}
	return out
}

// Refresh brings the chosen harnesses to version without changing the
// selection; harnesses that are no longer on the machine are skipped.
func Refresh(chosen []string, version string) []Result {
	var out []Result
	for _, name := range chosen {
		h, ok := Get(name)
		if !ok {
			out = append(out, Result{Name: name, Action: "skipped", Err: "not supported by this version of prep"})
			continue
		}
		if !h.Detect() {
			out = append(out, Result{Name: name, Action: "skipped", Err: h.Title() + " is not on this machine"})
			continue
		}
		v, installed, err := h.Installed()
		out = append(out, install(h, v, installed, err, version))
	}
	return out
}

func install(h Harness, current string, installed bool, statusErr error, version string) Result {
	r := Result{Name: h.Name(), Version: version}
	switch {
	case statusErr == nil && installed && current == version:
		r.Action = "up to date"
		return r
	case statusErr == nil && installed:
		r.Action = "updated"
	default:
		r.Action = "installed"
	}
	warnings, err := h.Install(version)
	r.Warnings = warnings
	if err != nil {
		r.Action, r.Err = "failed", err.Error()
	}
	return r
}

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}
