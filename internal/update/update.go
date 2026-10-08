// Package update replaces the running prep binary with the latest GitHub
// release after verifying its checksum. It uses the standard library only.
package update

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Repo is the GitHub repository releases come from.
const Repo = "fluid-movement/prep"

// GoInstall is the command that updates a go install build.
const GoInstall = "go install github.com/" + Repo + "/cmd/prep@latest"

// Updater finds and installs releases. Fields default to the real values;
// tests point them at a fake server and a temporary binary.
type Updater struct {
	API     string // https://api.github.com/repos/<Repo>
	Client  *http.Client
	Current string // version of the running binary
	Exe     string // binary to replace
	GOOS    string
	GOARCH  string
}

// New returns an updater for the running binary.
func New(current string) (*Updater, error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, err
	}
	if p, err := filepath.EvalSymlinks(exe); err == nil {
		exe = p
	}
	return &Updater{API: "https://api.github.com/repos/" + Repo, Client: &http.Client{Timeout: 60 * time.Second},
		Current: current, Exe: exe, GOOS: runtime.GOOS, GOARCH: runtime.GOARCH}, nil
}

// Release is the part of a GitHub release prep needs.
type Release struct {
	Tag    string  `json:"tag_name"`
	Assets []Asset `json:"assets"`
}

// Asset is one downloadable file of a release.
type Asset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

var versionRe = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z.-]+))?$`)

// devRe matches the pre-release parts of development builds, which are
// not releases: Go pseudo-versions (v0.0.0-20261006090120-93aa83c32597),
// git describe after a tag (v0.1.0-14-gaa2b6fc), and dirty work trees.
var devRe = regexp.MustCompile(`(^|\.)\d{14}-[0-9a-f]{12}$|(^|-)\d+-g[0-9a-f]+(-dirty)?$|(^|-)dirty$`)

// IsRelease reports whether v is a release version such as v1.2.3 or
// v1.2.3-rc.1; development builds are not.
func IsRelease(v string) bool {
	m := versionRe.FindStringSubmatch(v)
	return m != nil && !devRe.MatchString(m[4])
}

// Newer reports whether latest is newer than current. A pre-release is
// older than the release with the same numbers.
func Newer(current, latest string) (bool, error) {
	c, l := versionRe.FindStringSubmatch(current), versionRe.FindStringSubmatch(latest)
	if c == nil || l == nil {
		return false, fmt.Errorf("cannot compare versions %q and %q", current, latest)
	}
	for k := 1; k <= 3; k++ {
		a, _ := strconv.Atoi(c[k])
		b, _ := strconv.Atoi(l[k])
		if a != b {
			return b > a, nil
		}
	}
	return c[4] != "" && l[4] == "", nil
}

// Managed returns the command that updates prep when another tool installed
// it (Homebrew or go install), or "" when prep update may replace it.
func Managed(exe string) string {
	p := filepath.ToSlash(exe)
	for _, m := range []string{"/Cellar/", "/homebrew/", "/linuxbrew/"} {
		if strings.Contains(p, m) {
			return "brew upgrade prep"
		}
	}
	dir := filepath.Dir(exe)
	var gobins []string
	if b := os.Getenv("GOBIN"); b != "" {
		gobins = append(gobins, b)
	}
	if gp := os.Getenv("GOPATH"); gp != "" {
		for _, p := range filepath.SplitList(gp) {
			gobins = append(gobins, filepath.Join(p, "bin"))
		}
	} else if home, err := os.UserHomeDir(); err == nil {
		gobins = append(gobins, filepath.Join(home, "go", "bin"))
	}
	for _, b := range gobins {
		if filepath.Clean(b) == filepath.Clean(dir) {
			return GoInstall
		}
	}
	return ""
}

// Latest fetches the latest release.
func (u *Updater) Latest() (*Release, error) {
	req, err := http.NewRequest("GET", u.API+"/releases/latest", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := u.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("checking for a new release: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("checking for a new release: GitHub answered %s", resp.Status)
	}
	var r Release
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("reading the release: %v", err)
	}
	return &r, nil
}

// ArchiveName is the release archive for a version and platform, matching
// .goreleaser.yaml.
func ArchiveName(tag, goos, goarch string) string {
	ext := "tar.gz"
	if goos == "windows" {
		ext = "zip"
	}
	return fmt.Sprintf("prep_%s_%s_%s.%s", strings.TrimPrefix(tag, "v"), goos, goarch, ext)
}

// Install downloads the release archive for this platform, verifies it
// against checksums.txt, and atomically replaces the binary.
func (u *Updater) Install(r *Release) error {
	name := ArchiveName(r.Tag, u.GOOS, u.GOARCH)
	var archive, sums string
	for _, a := range r.Assets {
		switch a.Name {
		case name:
			archive = a.URL
		case "checksums.txt":
			sums = a.URL
		}
	}
	if archive == "" {
		return fmt.Errorf("release %s has no build for %s/%s (%s)", r.Tag, u.GOOS, u.GOARCH, name)
	}
	if sums == "" {
		return fmt.Errorf("release %s has no checksums.txt; refusing to install unverified", r.Tag)
	}
	sumData, err := u.get(sums)
	if err != nil {
		return err
	}
	want, err := checksum(sumData, name)
	if err != nil {
		return err
	}
	data, err := u.get(archive)
	if err != nil {
		return err
	}
	got := sha256.Sum256(data)
	if hex.EncodeToString(got[:]) != want {
		return fmt.Errorf("checksum mismatch for %s: the download is corrupt or tampered with; nothing was replaced", name)
	}
	bin := "prep"
	if u.GOOS == "windows" {
		bin = "prep.exe"
	}
	exe, err := extract(data, name, bin)
	if err != nil {
		return err
	}
	return replace(u.Exe, exe, u.GOOS == "windows")
}

func (u *Updater) get(url string) ([]byte, error) {
	resp, err := u.Client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("downloading %s: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("downloading %s: %s", url, resp.Status)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 256<<20))
}

func checksum(sums []byte, name string) (string, error) {
	sc := bufio.NewScanner(bytes.NewReader(sums))
	for sc.Scan() {
		f := strings.Fields(sc.Text())
		if len(f) == 2 && strings.TrimPrefix(f[1], "*") == name {
			return strings.ToLower(f[0]), nil
		}
	}
	return "", fmt.Errorf("checksums.txt has no entry for %s", name)
}

func extract(data []byte, name, bin string) ([]byte, error) {
	if strings.HasSuffix(name, ".zip") {
		zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		if err != nil {
			return nil, err
		}
		for _, f := range zr.File {
			if filepath.Base(f.Name) == bin {
				rc, err := f.Open()
				if err != nil {
					return nil, err
				}
				defer rc.Close()
				return io.ReadAll(rc)
			}
		}
		return nil, fmt.Errorf("%s does not contain %s", name, bin)
	}
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("%s does not contain %s", name, bin)
		}
		if err != nil {
			return nil, err
		}
		if h.Typeflag == tar.TypeReg && filepath.Base(h.Name) == bin {
			return io.ReadAll(tr)
		}
	}
}

// replace writes the new binary next to the old one and renames it into
// place, so the path never holds a partial file.
func replace(exe string, data []byte, windows bool) error {
	dir := filepath.Dir(exe)
	tmp, err := os.CreateTemp(dir, ".prep-update-*")
	if err != nil {
		return fmt.Errorf("cannot write next to %s: %v", exe, err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o755); err != nil {
		return err
	}
	if windows {
		// A running executable cannot be overwritten on Windows, but it can
		// be renamed aside.
		old := exe + ".old"
		os.Remove(old)
		if err := os.Rename(exe, old); err != nil {
			return err
		}
		if err := os.Rename(tmp.Name(), exe); err != nil {
			if rerr := os.Rename(old, exe); rerr != nil {
				return fmt.Errorf("%v; the previous binary is at %s", err, old)
			}
			return err
		}
		return nil
	}
	return os.Rename(tmp.Name(), exe)
}
