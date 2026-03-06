//go:build !(linux || windows)

package config

// No-op on unsupported platforms, so the IDE shuts up.
func ShowDefaultGeneratedDialog(_ *ErrDefaultGenerated) {}
