package domain

import (
	"fmt"
	"strings"
)

// OpConfig writes the project configuration.
const OpConfig Op = "config"

// PlanConfig validates a new project configuration and returns the change
// that writes it: saved views with distinct, non-empty names whose queries
// parse. Views keep the order of ViewOrder.
func (t *Tree) PlanConfig(cfg Config) (*Change, error) {
	var unmet []Unmet
	seen := map[string]bool{}
	for _, name := range cfg.ViewOrder {
		switch {
		case strings.TrimSpace(name) == "":
			unmet = append(unmet, Unmet{GateConfig, "a saved view needs a name"})
		case name != strings.TrimSpace(name):
			unmet = append(unmet, Unmet{GateConfig, fmt.Sprintf("view %q: no spaces around the name", name)})
		case seen[name]:
			unmet = append(unmet, Unmet{GateConfig, fmt.Sprintf("view %q exists twice", name)})
		}
		seen[name] = true
		if _, err := ParseFilter(strings.Fields(cfg.Views[name])); err != nil {
			unmet = append(unmet, Unmet{GateConfig, fmt.Sprintf("view %q: %v", name, err)})
		}
	}
	if len(cfg.Views) != len(cfg.ViewOrder) {
		unmet = append(unmet, Unmet{GateConfig, "every saved view needs a place in the view order"})
	}
	if len(unmet) > 0 {
		return nil, &Error{Code: ErrUsage, Message: "cannot change the settings", Unmet: unmet}
	}
	c := cfg
	c.Views = map[string]string{}
	for _, n := range cfg.ViewOrder {
		c.Views[n] = strings.Join(strings.Fields(cfg.Views[n]), " ")
	}
	c.ViewOrder = append([]string(nil), cfg.ViewOrder...)
	return &Change{Op: OpConfig, Config: &c}, nil
}
