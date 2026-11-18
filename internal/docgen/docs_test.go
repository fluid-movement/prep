package docgen

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

func TestCommandsCurrent(t *testing.T) {
	// Help generation must not depend on the caller's selected project.
	t.Setenv("PREP_PROJECT", "")
	want, err := Commands()
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile("../../docs/commands.md")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatal("docs/commands.md is out of date; run just docs-commands")
	}
}

func markdown(data []byte) ast.Node {
	return goldmark.New(goldmark.WithParserOptions(parser.WithAutoHeadingID())).
		Parser().Parse(text.NewReader(data))
}

func checkLink(source, destination string) error {
	u, err := url.Parse(destination)
	if err != nil {
		return err
	}
	if u.IsAbs() || u.Host != "" {
		return nil
	}
	if strings.HasPrefix(u.Path, "/") {
		return fmt.Errorf("link must be relative: %s", destination)
	}
	target := source
	if u.Path != "" {
		target = filepath.Join(filepath.Dir(source), filepath.FromSlash(u.Path))
	}
	data, err := os.ReadFile(target)
	if err != nil {
		return err
	}
	if u.Fragment == "" {
		return nil
	}
	found := false
	err = ast.Walk(markdown(data), func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && n.Kind() == ast.KindHeading {
			if id, ok := n.AttributeString("id"); ok && string(id.([]byte)) == u.Fragment {
				found = true
			}
		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("missing heading #%s in %s", u.Fragment, target)
	}
	return nil
}

func TestDocumentationLinks(t *testing.T) {
	pages, err := filepath.Glob("../../docs/*.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) == 0 {
		t.Fatal("no documentation pages found")
	}
	pages = append(pages, "../../README.md")
	for _, page := range pages {
		t.Run(filepath.Base(page), func(t *testing.T) {
			data, err := os.ReadFile(page)
			if err != nil {
				t.Fatal(err)
			}
			err = ast.Walk(markdown(data), func(n ast.Node, entering bool) (ast.WalkStatus, error) {
				if link, ok := n.(*ast.Link); ok && entering {
					if err := checkLink(page, string(link.Destination)); err != nil {
						t.Errorf("%s: %v", link.Destination, err)
					}
				}
				return ast.WalkContinue, nil
			})
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCheckLink(t *testing.T) {
	dir := t.TempDir()
	page := filepath.Join(dir, "page.md")
	if err := os.WriteFile(page, []byte("# A heading\n"), 0644); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		link string
		ok   bool
	}{
		{"page.md", true},
		{"page.md#a-heading", true},
		{"#a-heading", true},
		{"https://example.com/missing", true},
		{"missing.md", false},
		{"page.md#missing", false},
		{"/page.md", false},
	} {
		t.Run(tt.link, func(t *testing.T) {
			if err := checkLink(page, tt.link); (err == nil) != tt.ok {
				t.Fatalf("checkLink(%q) = %v; want success %v", tt.link, err, tt.ok)
			}
		})
	}
}
