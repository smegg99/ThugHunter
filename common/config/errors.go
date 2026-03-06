package config

import (
	"errors"
	"fmt"
)

// ErrDefaultGenerated is returned when a default config was written and
// failOnDefaultConfigGenerate is true.
type ErrDefaultGenerated struct {
	Path string
}

func (e *ErrDefaultGenerated) Error() string {
	return fmt.Sprintf("config not found, a default has been written to %s", e.Path)
}

// IsDefaultGenerated checks whether err wraps an ErrDefaultGenerated.
func IsDefaultGenerated(err error) (*ErrDefaultGenerated, bool) {
	var e *ErrDefaultGenerated
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}
