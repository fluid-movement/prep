//go:build unix

package mdstore

import (
	"os"

	"golang.org/x/sys/unix"
)

func lockFileHandle(f *os.File) (func() error, error) {
	if err := unix.Flock(int(f.Fd()), unix.LOCK_EX); err != nil {
		return nil, err
	}
	return func() error { return unix.Flock(int(f.Fd()), unix.LOCK_UN) }, nil
}
