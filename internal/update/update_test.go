package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func tarGz(t *testing.T, name string, content []byte) []byte {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	tw := tar.NewWriter(gz)
	for _, f := range []struct {
		name string
		data []byte
	}{{"README.md", []byte("readme")}, {name, content}} {
		if err := tw.WriteHeader(&tar.Header{Name: f.name, Mode: 0o755, Size: int64(len(f.data)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		tw.Write(f.data)
	}
	tw.Close()
	gz.Close()
	return b.Bytes()
}

func zipped(t *testing.T, name string, content []byte) []byte {
	var b bytes.Buffer
	zw := zip.NewWriter(&b)
	w, err := zw.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	w.Write(content)
	zw.Close()
	return b.Bytes()
}

// fakeRelease serves a release with the given archives; corrupt flips the
// checksum of every archive.
func fakeRelease(t *testing.T, tag string, archives map[string][]byte, corrupt bool) *httptest.Server {
	var srv *httptest.Server
	mux := http.NewServeMux()
	mux.HandleFunc("/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		rel := Release{Tag: tag}
		for name := range archives {
			rel.Assets = append(rel.Assets, Asset{Name: name, URL: srv.URL + "/dl/" + name})
		}
		rel.Assets = append(rel.Assets, Asset{Name: "checksums.txt", URL: srv.URL + "/dl/checksums.txt"})
		json.NewEncoder(w).Encode(rel)
	})
	mux.HandleFunc("/dl/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/dl/")
		if name == "checksums.txt" {
			for n, data := range archives {
				sum := sha256.Sum256(data)
				h := hex.EncodeToString(sum[:])
				if corrupt {
					h = strings.Repeat("0", 64)
				}
				fmt.Fprintf(w, "%s  %s\n", h, n)
			}
			return
		}
		w.Write(archives[name])
	})
	srv = httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func updater(t *testing.T, api, goos string) *Updater {
	exe := filepath.Join(t.TempDir(), "prep")
	if err := os.WriteFile(exe, []byte("old binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	return &Updater{API: api, Client: http.DefaultClient, Current: "v0.1.0", Exe: exe, GOOS: goos, GOARCH: "arm64"}
}

func TestInstallReplacesBinary(t *testing.T) {
	srv := fakeRelease(t, "v0.2.0", map[string][]byte{
		"prep_0.2.0_darwin_arm64.tar.gz": tarGz(t, "prep", []byte("new binary")),
		"prep_0.2.0_linux_amd64.tar.gz":  tarGz(t, "prep", []byte("wrong platform")),
	}, false)
	u := updater(t, srv.URL, "darwin")
	r, err := u.Latest()
	if err != nil || r.Tag != "v0.2.0" {
		t.Fatalf("latest: %+v %v", r, err)
	}
	if err := u.Install(r); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(u.Exe)
	if string(b) != "new binary" {
		t.Fatalf("binary = %q", b)
	}
	if fi, _ := os.Stat(u.Exe); fi.Mode().Perm() != 0o755 {
		t.Fatalf("mode = %v", fi.Mode())
	}
}

func TestInstallWindowsZip(t *testing.T) {
	srv := fakeRelease(t, "v0.2.0", map[string][]byte{"prep_0.2.0_windows_arm64.zip": zipped(t, "prep.exe", []byte("new exe"))}, false)
	u := updater(t, srv.URL, "windows")
	r, _ := u.Latest()
	if err := u.Install(r); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(u.Exe); string(b) != "new exe" {
		t.Fatalf("binary = %q", b)
	}
	if b, _ := os.ReadFile(u.Exe + ".old"); string(b) != "old binary" {
		t.Fatalf("old binary not kept aside: %q", b)
	}
}

func TestInstallRefusesBadChecksumAndMissingBuild(t *testing.T) {
	srv := fakeRelease(t, "v0.2.0", map[string][]byte{"prep_0.2.0_darwin_arm64.tar.gz": tarGz(t, "prep", []byte("evil"))}, true)
	u := updater(t, srv.URL, "darwin")
	r, _ := u.Latest()
	if err := u.Install(r); err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("err = %v", err)
	}
	if b, _ := os.ReadFile(u.Exe); string(b) != "old binary" {
		t.Fatal("binary replaced despite a checksum mismatch")
	}
	u.GOOS = "plan9"
	if err := u.Install(r); err == nil || !strings.Contains(err.Error(), "no build for plan9/arm64") {
		t.Fatalf("err = %v", err)
	}
}

func TestNewer(t *testing.T) {
	cases := []struct {
		cur, latest string
		want        bool
	}{
		{"v0.1.0", "v0.2.0", true}, {"v0.2.0", "v0.2.0", false}, {"v0.10.0", "v0.9.0", false},
		{"v1.0.0-rc.1", "v1.0.0", true}, {"1.2.3", "v1.2.4", true},
	}
	for _, c := range cases {
		if got, err := Newer(c.cur, c.latest); err != nil || got != c.want {
			t.Errorf("Newer(%s, %s) = %v %v", c.cur, c.latest, got, err)
		}
	}
	if _, err := Newer("6449764", "v0.1.0"); err == nil {
		t.Error("a commit hash compared as a version")
	}
	if IsRelease("6449764-dirty") || !IsRelease("v0.1.0") {
		t.Error("IsRelease")
	}
}

func TestManaged(t *testing.T) {
	if got := Managed("/opt/homebrew/Cellar/prep/0.1.0/bin/prep"); got != "brew upgrade prep" {
		t.Errorf("homebrew: %q", got)
	}
	gobin := t.TempDir()
	t.Setenv("GOBIN", gobin)
	if got := Managed(filepath.Join(gobin, "prep")); got != GoInstall {
		t.Errorf("go install: %q", got)
	}
	if got := Managed(filepath.Join(t.TempDir(), "prep")); got != "" {
		t.Errorf("plain install: %q", got)
	}
}
