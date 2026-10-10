package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/fluid-movement/prep/internal/mdstore"
	"github.com/fluid-movement/prep/internal/watch"
)

// watchContext ends prep watch; tests replace it to stop the stream.
var watchContext = defaultWatchContext

// startParent is the parent process at start, read before any setup so a
// parent that exits while prep watch starts up is still noticed.
var startParent = os.Getppid()

func defaultWatchContext() (context.Context, context.CancelFunc) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	go endWithParent(ctx, cancel, startParent, time.Second)
	return ctx, cancel
}

// endWithParent cancels when the process that started prep watch is gone
// (on Unix the orphan is re-parented, so its parent ID changes). A client
// that crashes or is killed never gets to stop its watcher, and an idle
// watcher would otherwise notice the closed output only on the next change.
func endWithParent(ctx context.Context, cancel context.CancelFunc, parent int, every time.Duration) {
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if os.Getppid() != parent {
				cancel()
				return
			}
		}
	}
}

// cmdWatch prints one line per debounced change under .prep until it is
// killed or its output is closed, so harness integrations can refresh.
func cmdWatch(a *app, args []string) error {
	fs := flag.NewFlagSet("watch", flag.ContinueOnError)
	if _, err := parse(fs, args); err != nil {
		return err
	}
	if err := a.open(); err != nil {
		return err
	}
	changes, stop, err := watch.Dir(filepath.Join(a.store.Root, mdstore.Dir))
	if err != nil {
		return err
	}
	defer stop()
	ctx, cancel := watchContext()
	defer cancel()
	line := "changed\n"
	if a.json {
		line = "{\"event\":\"changed\"}\n"
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case c := <-changes:
			if !c.Tree {
				continue // the activity stream is not the project
			}
			// A closed reader (the harness went away) ends the stream.
			if _, err := fmt.Fprint(a.out, line); err != nil {
				return nil
			}
		}
	}
}
