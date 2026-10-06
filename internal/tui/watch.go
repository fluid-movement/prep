package tui

import (
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// debounce is how long the watcher waits for writes to settle before it
// asks for a reload; a transition writes several files at once.
const debounce = 150 * time.Millisecond

// watch reports changes anywhere under dir on the returned channel, at most
// one pending signal at a time. fsnotify is not recursive, so every
// directory is watched and new ones are added as they appear.
func watch(dir string) (<-chan struct{}, func(), error) {
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

	out := make(chan struct{}, 1)
	var mu sync.Mutex
	var timer *time.Timer
	signal := func() {
		select {
		case out <- struct{}{}:
		default:
		}
	}
	go func() {
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
				mu.Lock()
				if timer == nil {
					timer = time.AfterFunc(debounce, signal)
				} else {
					timer.Reset(debounce)
				}
				mu.Unlock()
			case _, ok := <-w.Errors:
				if !ok {
					return
				}
			}
		}
	}()
	return out, func() { w.Close() }, nil
}
