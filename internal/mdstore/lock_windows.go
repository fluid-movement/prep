//go:build windows

package mdstore

import (
	"os"

	"golang.org/x/sys/windows"
)

func lockFileHandle(f *os.File) (func() error, error) {
	var overlapped windows.Overlapped
	err := windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK, 0, 1, 0, &overlapped)
	if err != nil {
		return nil, err
	}
	return func() error { return windows.UnlockFileEx(windows.Handle(f.Fd()), 0, 1, 0, &overlapped) }, nil
}
