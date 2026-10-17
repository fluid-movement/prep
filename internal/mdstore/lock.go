package mdstore

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

const lockFile = Dir + "/local/write.lock"

// lock serializes cooperating mdstore readers and writers in this checkout.
// A missing .prep directory stays missing for read-only probes. While the
// store holds the lock for a whole write command (Hold), lock does nothing:
// flock on a second descriptor of the same file blocks even in one process.
func (s *Store) lock(create bool) (func(), error) {
	if s.held {
		return func() {}, nil
	}
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

// Hold takes the exclusive checkout lock for a whole write command: load,
// plan and write. Until the returned function releases it, Load, LoadConfig
// and Apply on this store skip locking. Holding twice is a no-op.
func (s *Store) Hold() (func(), error) {
	if s.held {
		return func() {}, nil
	}
	unlock, err := s.lock(true)
	if err != nil {
		return nil, err
	}
	s.held = true
	return func() { s.held = false; unlock() }, nil
}

// Locked runs fn while holding the checkout lock.
func (s *Store) Locked(fn func() error) error {
	release, err := s.Hold()
	if err != nil {
		return err
	}
	defer release()
	return fn()
}
