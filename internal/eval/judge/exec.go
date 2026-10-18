package judge

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/fluid-movement/prep/internal/eval/proc"
)

// runCmd runs a prepared command, killing its process group when ctx ends,
// and returns its stdout.
func runCmd(ctx context.Context, cmd *exec.Cmd) ([]byte, error) {
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	proc.Group(cmd)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		if err != nil {
			return nil, fmt.Errorf("judge: %v: %s", err, strings.TrimSpace(errb.String()+out.String()))
		}
		return out.Bytes(), nil
	case <-ctx.Done():
		proc.Kill(cmd)
		select {
		case <-done:
		case <-time.After(5 * time.Second):
		}
		return nil, ctx.Err()
	}
}
