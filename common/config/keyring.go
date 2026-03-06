// common/config/keyring.go
package config

import (
	"fmt"

	"github.com/zalando/go-keyring"
)

// SetKeyringValue stores a value in the OS keyring.
func SetKeyringValue(key, value string) error {
	if err := keyring.Set(keyringService, key, value); err != nil {
		return fmt.Errorf("keyring set %q: %w", key, err)
	}
	return nil
}

// GetKeyringValue retrieves a value from the OS keyring.
func GetKeyringValue(key string) (string, error) {
	v, err := keyring.Get(keyringService, key)
	if err != nil {
		return "", fmt.Errorf("keyring get %q: %w", key, err)
	}
	return v, nil
}

// DeleteKeyringValue removes a value from the OS keyring.
func DeleteKeyringValue(key string) error {
	if err := keyring.Delete(keyringService, key); err != nil {
		return fmt.Errorf("keyring delete %q: %w", key, err)
	}
	return nil
}
