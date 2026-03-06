// launch_chromedriver.go
package unrevealed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/go-rod/rod"
)

type sessionCapabilities struct {
	DebuggerAddr string `json:"goog:chromeOptions.debuggerAddress"`
}

type sessionValue struct {
	Capabilities sessionCapabilities `json:"capabilities"`
}

type sessionResponse struct {
	Value sessionValue `json:"value"`
}

func newViaChromeDriver(ctx context.Context, chromePath string, cfg Config) (*Browser, error) {
	driverPath := cfg.ChromeDriverPath
	var patcher *Patcher

	if driverPath == "" {
		major, err := ChromeMajorVersion(chromePath)
		if err != nil {
			return nil, fmt.Errorf("detect chrome version: %w", err)
		}

		patcher = NewPatcher(major)
		driverPath, err = patcher.Run(ctx)
		if err != nil {
			patcher.Cleanup()
			return nil, fmt.Errorf("patch chromedriver: %w", err)
		}
	}

	cdPort, cdListener, err := freePort()
	if err != nil {
		if patcher != nil {
			patcher.Cleanup()
		}
		return nil, fmt.Errorf("free port: %w", err)
	}

	userDataDir, tmpDir, err := setupDataDir(cfg.UserDataDir)
	if err != nil {
		cdListener.Close()
		if patcher != nil {
			patcher.Cleanup()
		}
		return nil, fmt.Errorf("data dir: %w", err)
	}

	chrArgs := chromeLaunchArgs(userDataDir, cfg)

	cmd := exec.Command(driverPath, fmt.Sprintf("--port=%d", cdPort))

	cdListener.Close()
	if err := cmd.Start(); err != nil {
		cleanupTmpDir(tmpDir)
		if patcher != nil {
			patcher.Cleanup()
		}
		return nil, fmt.Errorf("start chromedriver: %w", err)
	}

	b := &Browser{cmd: cmd, tmpDir: tmpDir, patcher: patcher}

	wsURL, err := createChromeDriverSession(ctx, cdPort, chromePath, chrArgs, cfg.ConnectTimeout)
	if err != nil {
		b.Close()
		return nil, fmt.Errorf("chromedriver session: %w", err)
	}

	browser := rod.New().ControlURL(wsURL).NoDefaultDevice()
	if err := browser.Connect(); err != nil {
		b.Close()
		return nil, fmt.Errorf("connect rod: %w", err)
	}

	b.Browser = browser
	return b, nil
}

func createChromeDriverSession(ctx context.Context, port int, chromeBinary string, chromeArgs []string, timeout time.Duration) (string, error) {
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	client := &http.Client{Timeout: 30 * time.Second}

	deadline := time.Now().Add(timeout)
	ready := false
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/status", nil)
		if resp, err := client.Do(req); err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				ready = true
				break
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !ready {
		return "", fmt.Errorf("timeout waiting for chromedriver on port %d", port)
	}

	caps := map[string]any{
		"capabilities": map[string]any{
			"alwaysMatch": map[string]any{
				"goog:chromeOptions": map[string]any{
					"binary":          chromeBinary,
					"args":            chromeArgs,
					"excludeSwitches": DeleteFlags(),
				},
			},
		},
	}

	body, err := json.Marshal(caps)
	if err != nil {
		return "", fmt.Errorf("marshal capabilities: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/session", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("create session: HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var session sessionResponse
	if err := json.Unmarshal(respBody, &session); err != nil {
		return "", fmt.Errorf("decode session: %w", err)
	}

	debuggerAddr := session.Value.Capabilities.DebuggerAddr
	if debuggerAddr == "" {
		debuggerAddr = extractDebuggerAddr(respBody)
	}
	if debuggerAddr == "" {
		return "", fmt.Errorf("debugger address not found in session response")
	}

	if !strings.Contains(debuggerAddr, "://") {
		debuggerAddr = "http://" + debuggerAddr
	}

	return resolveWSURL(ctx, 0, timeout, debuggerAddr)
}

func extractDebuggerAddr(data []byte) string {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return ""
	}
	var value map[string]json.RawMessage
	if err := json.Unmarshal(raw["value"], &value); err != nil {
		return ""
	}
	var caps map[string]json.RawMessage
	if err := json.Unmarshal(value["capabilities"], &caps); err != nil {
		return ""
	}
	if optRaw, ok := caps["goog:chromeOptions"]; ok {
		var opts map[string]any
		if err := json.Unmarshal(optRaw, &opts); err == nil {
			if addr, ok := opts["debuggerAddress"].(string); ok && addr != "" {
				return addr
			}
		}
	}
	return ""
}
