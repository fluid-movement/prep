//go:build windows

package mdstore

// Directory fsync is not supported by the Windows file APIs used here. File
// contents are synced before journaling; the guarantee is process-interruption
// recovery, not power-loss durability.
func syncDirectory(string) error { return nil }
