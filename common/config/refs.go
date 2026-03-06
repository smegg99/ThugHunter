// common/config/refs.go
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/zalando/go-keyring"
)

// Reference source prefixes inside @{source:key}. Public means that it will end up exported in the public config file, private means it won't. The "?" suffix makes resolution optional, so that failures return "" instead of an error.
const (
	srcEnv     = "env"     // private: environment variable
	srcFile    = "file"    // private: mounted file
	srcKeyring = "keyring" // private: OS keyring / credential store
	srcPubEnv  = "pubenv"  // public: environment variable
	srcPubFile = "pubfile" // public: mounted file
	srcCfgDir  = "cfgdir"  // public: path relative to the config directory
	srcDataDir = "datadir" // public: path relative to the app data directory
)

// keyringService is the service name used for all keyring operations.
const keyringService = appName

// resolvedCfgDir and resolvedDataDir are set during Initialize, before
// ref resolution runs, so that cfgdir / datadir refs can expand paths.
var (
	resolvedCfgDir  string
	resolvedDataDir string
)

// Matches @{source:key} or @{source?:key} references.
var refPattern = regexp.MustCompile(`^@\{(\w+)(\??):(.+)\}$`)

type ref struct {
	Source   string
	Key      string
	Optional bool // when true, resolution failures return "" instead of an error
}

func parseRef(s string) (ref, bool) {
	m := refPattern.FindStringSubmatch(s)
	if m == nil {
		return ref{}, false
	}
	return ref{Source: m[1], Key: m[3], Optional: m[2] == "?"}, true
}

func (r ref) isPrivate() bool {
	switch r.Source {
	case srcEnv, srcFile, srcKeyring:
		return true
	}
	return false
}

func (r ref) resolve() (string, error) {
	v, err := r.resolveStrict()
	if err != nil && r.Optional {
		return "", nil
	}
	return v, err
}

func (r ref) resolveStrict() (string, error) {
	switch r.Source {
	case srcEnv, srcPubEnv:
		return resolveEnvRef(r.Key)
	case srcFile, srcPubFile:
		return resolveFileRef(r.Key)
	case srcKeyring:
		return resolveKeyringRef(r.Key)
	case srcCfgDir:
		return resolveDirRef(resolvedCfgDir, r.Key, "cfgdir")
	case srcDataDir:
		return resolveDirRef(resolvedDataDir, r.Key, "datadir")
	default:
		return "", fmt.Errorf("unknown reference source %q", r.Source)
	}
}

func resolveEnvRef(key string) (string, error) {
	v, ok := os.LookupEnv(key)
	if !ok {
		return "", fmt.Errorf("environment variable %q not set", key)
	}
	return v, nil
}

func resolveFileRef(key string) (string, error) {
	b, err := os.ReadFile(key)
	if err != nil {
		return "", fmt.Errorf("read %q: %w", key, err)
	}
	return strings.TrimSpace(string(b)), nil
}

func resolveKeyringRef(key string) (string, error) {
	v, err := keyring.Get(keyringService, key)
	if err != nil {
		return "", fmt.Errorf("keyring lookup %q: %w", key, err)
	}
	return v, nil
}

// resolveDirRef joins a relative path with a base directory, creating the
// directory tree if it doesn't exist yet.
func resolveDirRef(base, rel, label string) (string, error) {
	if base == "" {
		return "", fmt.Errorf("%s: base directory not set", label)
	}
	full := filepath.Join(base, rel)
	if dir := filepath.Dir(full); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", fmt.Errorf("%s: create %q: %w", label, dir, err)
		}
	}
	return full, nil
}
