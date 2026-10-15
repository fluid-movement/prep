package mdstore

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/fluid-movement/prep/internal/activity"
	"github.com/fluid-movement/prep/internal/domain"
)

const transactionFile = Dir + "/.write-transaction.json"
const lockFile = Dir + "/local/write.lock"

const absentDirectoryManifest = "<absent>"

type directoryEntry struct {
	Name string `json:"name"`
	Mode uint32 `json:"mode"`
}

func (s *Store) recordDirectoryEntries(rel string, entries []os.DirEntry, include func(string) bool) {
	manifest := make([]directoryEntry, 0, len(entries))
	for _, entry := range entries {
		if include != nil && !include(entry.Name()) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			// ReadDir returned this entry while the checkout lock was held. If
			// metadata is unavailable, preserve its name with an unknown mode;
			// the next validation will conservatively conflict.
			manifest = append(manifest, directoryEntry{Name: entry.Name()})
			continue
		}
		manifest = append(manifest, directoryEntry{Name: entry.Name(), Mode: uint32(info.Mode())})
	}
	b, _ := json.Marshal(manifest)
	s.loadedDirs[rel] = hashOf(b)
}

func (s *Store) directoryManifest(rel string) (string, error) {
	entries, err := os.ReadDir(s.abs(rel))
	if errors.Is(err, fs.ErrNotExist) {
		return absentDirectoryManifest, nil
	}
	if err != nil {
		return "", err
	}
	include := func(string) bool { return true }
	switch rel {
	case Dir:
		include = func(name string) bool {
			return name != transactionFile[len(Dir)+1:] && name != activity.Dir && !strings.HasPrefix(name, tempPrefix)
		}
	case Dir + "/issues":
		include = func(name string) bool {
			return !strings.HasPrefix(name, tempPrefix) && !strings.HasPrefix(name, ".")
		}
	default:
		if strings.HasPrefix(rel, Dir+"/issues/") && !strings.HasSuffix(rel, "/baselines") {
			include = func(name string) bool { return !strings.HasPrefix(name, tempPrefix) }
		}
	}
	manifest := make([]directoryEntry, 0, len(entries))
	for _, entry := range entries {
		if !include(entry.Name()) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return "", err
		}
		manifest = append(manifest, directoryEntry{Name: entry.Name(), Mode: uint32(info.Mode())})
	}
	b, err := json.Marshal(manifest)
	if err != nil {
		return "", err
	}
	return hashOf(b), nil
}

func (s *Store) validateDirectorySnapshot() error {
	for rel, want := range s.loadedDirs {
		got, err := s.directoryManifest(rel)
		if err != nil {
			return err
		}
		if got != want {
			return &domain.Error{Code: domain.ErrConflict, Message: fmt.Sprintf("%s directory contents changed since it was read; re-run the command", rel)}
		}
	}
	return nil
}

func (s *Store) refreshDirectorySnapshot(touched []string) error {
	dirs := make(map[string]bool, len(s.loadedDirs))
	for rel := range s.loadedDirs {
		dirs[rel] = true
	}
	for _, rel := range touched {
		if !strings.HasPrefix(rel, Dir+"/") {
			continue
		}
		parent := filepath.ToSlash(filepath.Dir(rel))
		for strings.HasPrefix(parent, Dir) && parent != "." {
			dirs[parent] = true
			if parent == Dir {
				break
			}
			parent = filepath.ToSlash(filepath.Dir(parent))
		}
	}
	for rel := range dirs {
		manifest, err := s.directoryManifest(rel)
		if errors.Is(err, fs.ErrNotExist) {
			s.loadedDirs[rel] = absentDirectoryManifest
			continue
		}
		if err != nil {
			return err
		}
		s.loadedDirs[rel] = manifest
	}
	return nil
}

type transactionEntry struct {
	Path   string `json:"path"`
	Temp   string `json:"temp,omitempty"`
	Before string `json:"before"`
	After  string `json:"after"`
	Remove bool   `json:"remove,omitempty"`
}

type writeTransaction struct {
	Version int                `json:"version"`
	Entries []transactionEntry `json:"entries"`
}

// lock serializes cooperating mdstore readers and writers in this checkout.
// A missing .prep directory stays missing for read-only probes.
func (s *Store) lock(create bool) (func(), error) {
	prep := s.abs(Dir)
	if _, err := os.Stat(prep); errors.Is(err, fs.ErrNotExist) && !create {
		return nil, nil
	} else if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	lockPath := s.abs(lockFile)
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o700); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	unlock, err := lockFileHandle(f)
	if err != nil {
		f.Close()
		return nil, err
	}
	return func() { _ = unlock(); _ = f.Close() }, nil
}

func (s *Store) recoverTransaction() error {
	b, err := os.ReadFile(s.abs(transactionFile))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	var tx writeTransaction
	if err := json.Unmarshal(b, &tx); err != nil || tx.Version != 1 || len(tx.Entries) == 0 {
		return recoveryError("invalid transaction journal")
	}
	seen := map[string]bool{}
	temps := map[string]bool{}
	for _, entry := range tx.Entries {
		if !validTransactionEntry(entry) || seen[entry.Path] {
			return recoveryError("invalid or duplicate transaction path")
		}
		seen[entry.Path] = true
		if entry.Temp != "" {
			tempPath := filepath.Join(filepath.Dir(entry.Path), entry.Temp)
			if temps[tempPath] {
				return recoveryError("duplicate staged path")
			}
			temps[tempPath] = true
		}
	}
	// Validate every target and staged file before mutating any target. A bad
	// later entry therefore cannot leave an earlier entry half-installed.
	for _, entry := range tx.Entries {
		if err := s.validateRecoveryEntry(entry); err != nil {
			return err
		}
	}
	for i, entry := range tx.Entries {
		if err := s.applyTransactionEntry(entry); err != nil {
			return err
		}
		if s.txHook != nil {
			if err := s.txHook("after-install", i); err != nil {
				return err
			}
		}
	}
	if err := os.Remove(s.abs(transactionFile)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return recoveryError("could not remove completed transaction journal: " + err.Error())
	}
	return syncDirectory(filepath.Dir(s.abs(transactionFile)))
}

func validTransactionEntry(e transactionEntry) bool {
	clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(e.Path)))
	if clean != e.Path || !transactionPathAllowed(clean) {
		return false
	}
	for _, part := range strings.Split(clean, "/") {
		if strings.HasPrefix(part, tempPrefix) {
			return false
		}
	}
	if len(e.Before) != 0 && !validHash(e.Before) || len(e.After) != 0 && !validHash(e.After) {
		return false
	}
	if e.Remove {
		return e.Temp == "" && e.After == ""
	}
	return e.Temp != "" && filepath.Base(e.Temp) == e.Temp && strings.HasPrefix(e.Temp, tempPrefix) && e.After != ""
}

func transactionPathAllowed(path string) bool {
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[0] != Dir {
		return false
	}
	switch parts[1] {
	case "project.md", ConfigOwn, ConfigDist, ".gitignore":
		return len(parts) == 2
	case "issues":
		if len(parts) == 3 {
			return parts[2] == ".gitkeep"
		}
		if len(parts) < 4 || !domain.ValidID(parts[2]) {
			return false
		}
		if len(parts) == 5 {
			return parts[3] == "baselines" && strings.HasSuffix(parts[4], ".md") && domain.ValidStamp(strings.TrimSuffix(parts[4], ".md"))
		}
		return len(parts) == 4 && issueFiles[parts[3]]
	case "knowledge":
		return (len(parts) == 3 && parts[2] == ".gitkeep") || (len(parts) >= 3 && strings.HasSuffix(parts[len(parts)-1], ".md"))
	default:
		return false
	}
}

func (s *Store) validateRecoveryEntry(e transactionEntry) error {
	path := s.abs(e.Path)
	if err := s.validateNoSymlinkParents(path); err != nil {
		return err
	}
	info, err := os.Lstat(path)
	if err == nil && info.Mode()&os.ModeSymlink != 0 {
		return recoveryError(fmt.Sprintf("%s is a symlink; journal retained", e.Path))
	}
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return recoveryError(fmt.Sprintf("cannot inspect %s: %v", e.Path, err))
	}
	b, err := os.ReadFile(path)
	current := ""
	if err == nil {
		current = hashString(b)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return recoveryError(fmt.Sprintf("cannot inspect %s: %v", e.Path, err))
	}
	if e.Remove {
		if current != "" && current != e.Before {
			return recoveryError(fmt.Sprintf("%s changed during recovery; journal retained", e.Path))
		}
		return nil
	}
	if current == e.After {
		return nil
	}
	if current != e.Before {
		return recoveryError(fmt.Sprintf("%s changed during recovery; journal retained", e.Path))
	}
	temp := filepath.Join(filepath.Dir(path), e.Temp)
	if err := s.validateNoSymlinkParents(temp); err != nil {
		return err
	}
	if info, err := os.Lstat(temp); err != nil || info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return recoveryError(fmt.Sprintf("staged content for %s is missing or unsafe; journal retained", e.Path))
	}
	tb, err := os.ReadFile(temp)
	if err != nil || hashString(tb) != e.After {
		return recoveryError(fmt.Sprintf("staged content for %s is missing or damaged; journal retained", e.Path))
	}
	return nil
}

func (s *Store) validateNoSymlinkParents(path string) error {
	root, err := filepath.Abs(s.Root)
	if err != nil {
		return err
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return recoveryError("journal path escapes checkout")
	}
	parts := strings.Split(filepath.Clean(rel), string(filepath.Separator))
	current := root
	for _, part := range parts[:len(parts)-1] {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		if err != nil {
			return recoveryError(fmt.Sprintf("cannot inspect journal path parent: %v", err))
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return recoveryError("journal path has an unsafe parent")
		}
	}
	return nil
}

func validHash(s string) bool {
	if len(s) != 64 {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}

func hashString(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func recoveryError(message string) error {
	return &domain.Error{Code: domain.ErrConflict, Message: "unfinished .prep write transaction: " + message + "; preserve the journal and resolve the conflict before retrying"}
}

func writeJournal(path string, tx writeTransaction) (bool, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false, err
	}
	b, err := json.Marshal(tx)
	if err != nil {
		return false, err
	}
	f, err := os.CreateTemp(filepath.Dir(path), tempPrefix+"journal-*")
	if err != nil {
		return false, err
	}
	defer os.Remove(f.Name())
	if _, err := f.Write(b); err != nil {
		_ = f.Close()
		return false, err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return false, err
	}
	if err := f.Close(); err != nil {
		return false, err
	}
	if err := os.Chmod(f.Name(), 0o600); err != nil {
		return false, err
	}
	if err := os.Rename(f.Name(), path); err != nil {
		return false, err
	}
	return true, syncDirectory(filepath.Dir(path))
}

func (s *Store) applyTransactionEntry(e transactionEntry) error {
	path := s.abs(e.Path)
	b, err := os.ReadFile(path)
	current := ""
	if err == nil {
		current = hashString(b)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return recoveryError(fmt.Sprintf("cannot inspect %s: %v", e.Path, err))
	}
	if e.Remove {
		if current == "" {
			return nil
		}
		if current != e.Before {
			return recoveryError(fmt.Sprintf("%s changed during recovery; journal retained", e.Path))
		}
		if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return recoveryError(fmt.Sprintf("cannot remove %s: %v", e.Path, err))
		}
		return syncDirectory(filepath.Dir(path))
	}
	if current == e.After {
		return nil
	}
	if current != e.Before {
		return recoveryError(fmt.Sprintf("%s changed during recovery; journal retained", e.Path))
	}
	temp := filepath.Join(filepath.Dir(path), e.Temp)
	tb, err := os.ReadFile(temp)
	if err != nil || hashString(tb) != e.After {
		return recoveryError(fmt.Sprintf("staged content for %s is missing or damaged; journal retained", e.Path))
	}
	if err := os.Rename(temp, path); err != nil {
		return recoveryError(fmt.Sprintf("cannot install %s: %v", e.Path, err))
	}
	return syncDirectory(filepath.Dir(path))
}

func (s *Store) applyTransaction(entries []transactionEntry, contents map[string][]byte) error {
	if len(entries) == 0 {
		return nil
	}
	tx := writeTransaction{Version: 1, Entries: entries}
	cleanTemps := func() {
		for _, e := range entries {
			if e.Temp != "" {
				_ = os.Remove(filepath.Join(filepath.Dir(s.abs(e.Path)), e.Temp))
			}
		}
	}
	for i := range tx.Entries {
		e := &tx.Entries[i]
		if err := os.MkdirAll(filepath.Dir(s.abs(e.Path)), 0o755); err != nil {
			cleanTemps()
			return err
		}
		if e.Remove {
			continue
		}
		f, err := os.CreateTemp(filepath.Dir(s.abs(e.Path)), tempPrefix+"*")
		if err != nil {
			cleanTemps()
			return err
		}
		e.Temp = filepath.Base(f.Name())
		content := contents[e.Path]
		if _, err := f.Write(content); err != nil {
			_ = f.Close()
			cleanTemps()
			return err
		}
		if err := f.Sync(); err != nil {
			_ = f.Close()
			cleanTemps()
			return err
		}
		if err := f.Close(); err != nil {
			cleanTemps()
			return err
		}
		if err := os.Chmod(filepath.Join(filepath.Dir(s.abs(e.Path)), e.Temp), 0o644); err != nil {
			cleanTemps()
			return err
		}
		if err := syncDirectory(filepath.Dir(s.abs(e.Path))); err != nil {
			cleanTemps()
			return err
		}
	}
	if s.txHook != nil {
		if err := s.txHook("before-journal", -1); err != nil {
			cleanTemps()
			return err
		}
	}
	// Journal content is durable before the first target rename. A later error
	// leaves it in place so the next locked read/write can roll forward.
	installed, err := writeJournal(s.abs(transactionFile), tx)
	if err != nil {
		if !installed {
			cleanTemps()
		}
		return err
	}
	if s.txHook != nil {
		if err := s.txHook("after-journal", -1); err != nil {
			return err
		}
	}
	return s.recoverTransaction()
}
