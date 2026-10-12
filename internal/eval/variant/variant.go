// Package variant lays out the harness configuration under test: the
// plugin (skill text, hooks) a session runs with. A variant is a directory
// with a variant.yaml naming a base and what it changes.
package variant

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/fluid-movement/prep/internal/eval/run"
)

// Built-in bases.
const (
	// Current is the plugin as shipped.
	Current = "current"
	// None is no prep integration at all.
	None = "none"
)

// Variant is one variant.yaml.
type Variant struct {
	Name        string `yaml:"-"`
	Description string `yaml:"description"`
	// Base is "current", "none" or another variant's name.
	Base string `yaml:"base"`
	// Replace maps a plugin-relative path (which must exist in the base) to
	// a file in the variant directory.
	Replace map[string]string `yaml:"replace"`
	// Add maps a new plugin-relative path (which must not exist) to a file in
	// the variant directory.
	Add map[string]string `yaml:"add"`
	// Remove lists plugin-relative paths to delete.
	Remove []string `yaml:"remove"`
	// AppendSystemPrompt is added to the session's system prompt.
	AppendSystemPrompt string `yaml:"append_system_prompt"`

	dir string
}

// Resolver finds variants.
type Resolver struct {
	// Root is eval/variants: <Root>/<name>/variant.yaml.
	Root string
	// PluginDir is the shipped plugin, plugins/claude-code.
	PluginDir string
}

// Load reads a variant by name. The built-in names have no file.
func (r Resolver) Load(name string) (*Variant, error) {
	switch name {
	case Current, None:
		return &Variant{Name: name, Description: map[string]string{Current: "the plugin as shipped", None: "no prep integration"}[name]}, nil
	}
	dir := filepath.Join(r.Root, filepath.FromSlash(name))
	f, err := os.Open(filepath.Join(dir, "variant.yaml"))
	if err != nil {
		return nil, fmt.Errorf("variant %q: %w", name, err)
	}
	defer f.Close()
	var v Variant
	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)
	if err := dec.Decode(&v); err != nil {
		return nil, fmt.Errorf("variant %q: %w", name, err)
	}
	if v.Base == "" {
		return nil, fmt.Errorf("variant %q: no base", name)
	}
	v.Name, v.dir = name, dir
	return &v, nil
}

// Materialize writes the variant's plugin under dst (created) and returns
// what the runner needs. The hash covers every file of the plugin and the
// appended prompt, so a result says exactly what ran.
func (r Resolver) Materialize(name, dst string) (run.Variant, error) {
	chain, err := r.chain(name)
	if err != nil {
		return run.Variant{}, err
	}
	// chain[0] is the built-in root, the rest are layers, outermost last.
	root := chain[0].Name
	plugin := ""
	if root == Current {
		if r.PluginDir == "" {
			return run.Variant{}, fmt.Errorf("variant %q: the shipped plugin directory is not configured", name)
		}
		if err := copyTree(r.PluginDir, dst); err != nil {
			return run.Variant{}, fmt.Errorf("variant %q: copy plugin: %w", name, err)
		}
		plugin = dst
	} else if err := os.MkdirAll(dst, 0o755); err != nil {
		return run.Variant{}, err
	}
	var prompts []string
	for _, v := range chain[1:] {
		if err := v.apply(dst); err != nil {
			return run.Variant{}, err
		}
		if len(v.Replace)+len(v.Add) > 0 {
			plugin = dst
		}
		if v.AppendSystemPrompt != "" {
			prompts = append(prompts, v.AppendSystemPrompt)
		}
	}
	if root == None && plugin == "" {
		os.RemoveAll(dst)
	}
	out := run.Variant{Name: name, PluginDir: plugin, AppendSystemPrompt: strings.Join(prompts, "\n\n")}
	h, err := hashTree(plugin, out.AppendSystemPrompt)
	if err != nil {
		return run.Variant{}, err
	}
	out.Hash = h
	return out, nil
}

// chain returns the built-in root followed by the layers from the one
// nearest the root to the named variant; a cycle is an error.
func (r Resolver) chain(name string) ([]*Variant, error) {
	var layers []*Variant
	seen := map[string]bool{}
	for {
		if seen[name] {
			return nil, fmt.Errorf("variants form a cycle at %q", name)
		}
		seen[name] = true
		v, err := r.Load(name)
		if err != nil {
			return nil, err
		}
		if name == Current || name == None {
			out := []*Variant{v}
			for i := len(layers) - 1; i >= 0; i-- {
				out = append(out, layers[i])
			}
			return out, nil
		}
		layers = append(layers, v)
		name = v.Base
	}
}

// apply changes the plugin under dst.
func (v *Variant) apply(dst string) error {
	for _, rel := range v.Remove {
		p, err := within(dst, rel)
		if err != nil {
			return fmt.Errorf("variant %q: remove %q: %w", v.Name, rel, err)
		}
		if _, err := os.Stat(p); err != nil {
			return fmt.Errorf("variant %q: remove %q: no such file in the plugin", v.Name, rel)
		}
		if err := os.RemoveAll(p); err != nil {
			return err
		}
	}
	put := func(m map[string]string, mustExist bool) error {
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, rel := range keys {
			p, err := within(dst, rel)
			if err != nil {
				return fmt.Errorf("variant %q: %q: %w", v.Name, rel, err)
			}
			_, statErr := os.Stat(p)
			if mustExist && statErr != nil {
				return fmt.Errorf("variant %q: replace %q: no such file in the base plugin", v.Name, rel)
			}
			if !mustExist && statErr == nil {
				return fmt.Errorf("variant %q: add %q: the plugin already has it (use replace)", v.Name, rel)
			}
			src, err := within(v.dir, m[rel])
			if err != nil {
				return fmt.Errorf("variant %q: %q: %w", v.Name, m[rel], err)
			}
			data, err := os.ReadFile(src)
			if err != nil {
				return fmt.Errorf("variant %q: %w", v.Name, err)
			}
			if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(p, data, 0o644); err != nil {
				return err
			}
		}
		return nil
	}
	if err := put(v.Replace, true); err != nil {
		return err
	}
	return put(v.Add, false)
}

// within joins rel to root and refuses paths that leave it.
func within(root, rel string) (string, error) {
	if rel == "" || filepath.IsAbs(rel) {
		return "", fmt.Errorf("path must be relative")
	}
	p := filepath.Join(root, filepath.FromSlash(rel))
	if r, err := filepath.Rel(root, p); err != nil || strings.HasPrefix(r, "..") {
		return "", fmt.Errorf("path leaves the directory")
	}
	return p, nil
}

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		if d.IsDir() && (d.Name() == "node_modules" || d.Name() == ".git") {
			return fs.SkipDir
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		in, err := os.Open(p)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.Create(target)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, in); err != nil {
			out.Close()
			return err
		}
		return out.Close()
	})
}

// hashTree hashes the relative paths and contents under dir (none if dir is
// empty) and the appended prompt.
func hashTree(dir, prompt string) (string, error) {
	h := sha256.New()
	if dir != "" {
		var paths []string
		err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() {
				rel, _ := filepath.Rel(dir, p)
				paths = append(paths, rel)
			}
			return nil
		})
		if err != nil {
			return "", err
		}
		sort.Strings(paths)
		for _, rel := range paths {
			data, err := os.ReadFile(filepath.Join(dir, rel))
			if err != nil {
				return "", err
			}
			fmt.Fprintf(h, "%s\x00%d\x00", filepath.ToSlash(rel), len(data))
			h.Write(data)
		}
	}
	fmt.Fprintf(h, "prompt\x00%s", prompt)
	return fmt.Sprintf("%x", h.Sum(nil))[:16], nil
}
