//go:build windows

package mdstore

// Directory fsync is not supported by the Windows file APIs used here. File
// contents are synced before the rename; the guarantee is process-interruption
// safety, not power-loss durability.
func syncDirectory(string) error { return nil }
