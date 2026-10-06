package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/fluid-movement/prep/internal/mdstore"
	"github.com/fluid-movement/prep/internal/watch"
)

// watchContext ends prep watch; tests replace it to stop the stream.
var watchContext = defaultWatchContext

func defaultWatchContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
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
		case <-changes:
			// A closed reader (the harness went away) ends the stream.
			if _, err := fmt.Fprint(a.out, line); err != nil {
				return nil
			}
		}
	}
}
