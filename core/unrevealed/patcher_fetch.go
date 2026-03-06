// patcher_fetch.go
package unrevealed

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var httpClient = &http.Client{Timeout: 30 * time.Second}

func (p *Patcher) fetchVersion(ctx context.Context) (string, error) {
	if p.MajorVersion <= legacyMaxVersion {
		return p.fetchLegacyVersion(ctx)
	}
	return p.fetchModernVersion(ctx)
}

func (p *Patcher) fetchLegacyVersion(ctx context.Context) (string, error) {
	url := fmt.Sprintf(legacyVersionURL, p.MajorVersion)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch version: HTTP %d from %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 256))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(body)), nil
}

func (p *Patcher) fetchModernVersion(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, modernVersionURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch version: HTTP %d from %s", resp.StatusCode, modernVersionURL)
	}

	var result milestonesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	ms, ok := result.Milestones[fmt.Sprintf("%d", p.MajorVersion)]
	if !ok {
		return "", fmt.Errorf("chromedriver milestone %d not found", p.MajorVersion)
	}
	return ms.Version, nil
}

func (p *Patcher) download(ctx context.Context, version string) ([]byte, error) {
	url := p.buildDownloadURL(version)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed: HTTP %d from %s", resp.StatusCode, url)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, p.MaxDownloadSize+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > p.MaxDownloadSize {
		return nil, fmt.Errorf("download exceeds maximum size of %d bytes", p.MaxDownloadSize)
	}

	hash := sha256.Sum256(data)
	p.DownloadSHA256 = hex.EncodeToString(hash[:])

	if p.ExpectedSHA256 != "" && p.DownloadSHA256 != p.ExpectedSHA256 {
		return nil, fmt.Errorf("SHA256 mismatch: got %s, want %s", p.DownloadSHA256, p.ExpectedSHA256)
	}

	return data, nil
}

func (p *Patcher) buildDownloadURL(version string) string {
	if p.MajorVersion <= legacyMaxVersion {
		return fmt.Sprintf(legacyDownloadURL, version, p.platform)
	}
	return fmt.Sprintf(modernDownloadURL, version, p.platform, p.platform)
}
