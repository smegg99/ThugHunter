// patcher_binary.go
package unrevealed

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

func (p *Patcher) extract(data []byte) error {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}

	for _, f := range r.File {
		if f.FileInfo().IsDir() || filepath.Base(f.Name) != p.exeName {
			continue
		}
		return p.extractFile(f)
	}
	return fmt.Errorf("chromedriver binary not found in archive")
}

func (p *Patcher) extractFile(f *zip.File) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	out, err := os.OpenFile(p.DriverPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, rc); err != nil {
		out.Close()
		os.Remove(p.DriverPath)
		return err
	}
	return out.Close()
}

// ErrCDCNotFound is returned when the ChromeDriver binary does not contain
// the cdc_ automation marker. The binary may be already patched or from an
// unsupported ChromeDriver version.
var ErrCDCNotFound = errors.New("cdc pattern not found in chromedriver binary")

func (p *Patcher) patch() error {
	data, err := os.ReadFile(p.DriverPath)
	if err != nil {
		return err
	}

	re := regexp.MustCompile(cdcPattern)
	loc := re.FindIndex(data)
	if loc == nil {
		return ErrCDCNotFound
	}

	replacement := padToLength([]byte(patchReplacement), loc[1]-loc[0])
	copy(data[loc[0]:loc[1]], replacement)

	return os.WriteFile(p.DriverPath, data, 0o755)
}

func padToLength(buf []byte, target int) []byte {
	for len(buf) < target {
		buf = append(buf, ' ')
	}
	return buf
}
