package watch

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// debounce is how long the watcher waits for writes to settle before it
// asks for a reload; a transition writes several files at once.
const debounce = 150 * time.Millisecond

// Change says what changed since the last report: the project (issues,
// knowledge, config) or the per-machine state under the local directory
// (the activity stream), or both.
type Change struct {
	Tree  bool
	Local bool
}

// Local is the directory under the watched one that holds per-machine
// state; changes in it are reported as Local.
const Local = "local"

// Dir reports changes anywhere under dir on the returned channel, at most
// one pending report at a time, merging what changed until it is read.
// fsnotify is not recursive, so every directory is watched and new ones
// are added as they appear.
func Dir(dir string) (<-chan Change, func(), error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, nil, err
	}
	addTree := func(root string) {
		_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err == nil && d.IsDir() {
				_ = w.Add(p)
			}
			return nil
		})
	}
	addTree(dir)

	out := make(chan Change, 1)
	local := filepath.Join(dir, Local)
	go func() {
		var pending Change
		var fire <-chan time.Time // the debounce timer; nil while nothing waits
		timer := time.NewTimer(debounce)
		timer.Stop()
		wait := func() {
			timer.Reset(debounce)
			fire = timer.C
		}
		for {
			select {
			case ev, ok := <-w.Events:
				if !ok {
					return
				}
				if ev.Op&fsnotify.Create != 0 {
					if fi, err := os.Stat(ev.Name); err == nil && fi.IsDir() {
						addTree(ev.Name)
					}
				}
				if ev.Name == local || strings.HasPrefix(ev.Name, local+string(filepath.Separator)) {
					pending.Local = true
				} else {
					pending.Tree = true
				}
				wait()
			case _, ok := <-w.Errors:
				if !ok {
					return
				}
				// An overflow or a failed watch means changes may have gone
				// unseen: report everything, so the reader reloads it all.
				pending = Change{Tree: true, Local: true}
				wait()
			case <-fire:
				fire = nil
				// Only this goroutine sends, so a queued report can be taken
				// back and merged with the new one without blocking.
				c := pending
				pending = Change{}
				select {
				case old := <-out:
					c.Tree, c.Local = c.Tree || old.Tree, c.Local || old.Local
				default:
				}
				out <- c
			}
		}
	}()
	return out, func() { w.Close() }, nil
}
