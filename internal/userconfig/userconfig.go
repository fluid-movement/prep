// Package userconfig reads and writes the user-level prep configuration in
// $XDG_CONFIG_HOME/prep/config.yaml (default ~/.config/prep/config.yaml).
// It records only the user's choices; anything that can be asked from the
// system, such as what a harness has installed, is never stored here.
package userconfig

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config is the user's choices.
type Config struct {
	// Harnesses lists the harnesses prep integrates with, by name.
	Harnesses []string `yaml:"harnesses"`
	// TUI holds the terminal UI's choices.
	TUI TUI `yaml:"tui,omitempty"`
}

// TUI is the terminal UI's part of the user configuration.
type TUI struct {
	// Mouse false turns mouse capture off, so the terminal selects text
	// with a plain drag; unset means on.
	Mouse *bool `yaml:"mouse,omitempty"`
}

// Path returns the config file location.
func Path() (string, error) {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "prep", "config.yaml"), nil
}

// Load reads the config; a missing file is an empty config.
func Load() (Config, error) {
	var c Config
	p, err := Path()
	if err != nil {
		return c, err
	}
	b, err := os.ReadFile(p)
	if errors.Is(err, fs.ErrNotExist) {
		return c, nil
	}
	if err != nil {
		return c, err
	}
	dec := yaml.NewDecoder(bytes.NewReader(b))
	dec.KnownFields(true)
	if err := dec.Decode(&c); err != nil && !errors.Is(err, io.EOF) {
		return c, fmt.Errorf("%s: %v", p, err)
	}
	return c, nil
}

// Save writes the config atomically, creating its directory.
func Save(c Config) (string, error) {
	p, err := Path()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return "", err
	}
	var b bytes.Buffer
	b.WriteString("# prep user configuration: your choices. See prep setup.\n")
	enc := yaml.NewEncoder(&b)
	enc.SetIndent(2)
	if err := enc.Encode(c); err != nil {
		return "", err
	}
	enc.Close()
	tmp, err := os.CreateTemp(filepath.Dir(p), ".config-*")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(b.Bytes()); err != nil {
		tmp.Close()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}
	return p, os.Rename(tmp.Name(), p)
}
