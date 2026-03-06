package unrevealed

import (
	"fmt"
	"net"
	"os/exec"
	"sync"
	"time"
)

// Xvfb manages a virtual X11 framebuffer so Chrome can run in full headful
// mode (passing some bot-detection checks) while remaining invisible to the
// user. The display is destroyed when Close is called.
type Xvfb struct {
	Display string // e.g. ":99"
	cmd     *exec.Cmd
	mu      sync.Mutex
	closed  bool
}

// StartXvfb launches Xvfb on a free display number with the given resolution.
// The caller must call Close when finished.
func StartXvfb(width, height int) (*Xvfb, error) {
	display, err := freeDisplay()
	if err != nil {
		return nil, fmt.Errorf("xvfb: find free display: %w", err)
	}

	screen := fmt.Sprintf("%dx%dx24", width, height)
	cmd := exec.Command("Xvfb", display, "-screen", "0", screen, "-nolisten", "tcp", "-ac")
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("xvfb: start: %w", err)
	}

	// Give Xvfb a moment to initialise the display socket.
	time.Sleep(500 * time.Millisecond)

	return &Xvfb{Display: display, cmd: cmd}, nil
}

// Close kills the Xvfb process. Safe to call multiple times.
func (x *Xvfb) Close() error {
	x.mu.Lock()
	defer x.mu.Unlock()
	if x.closed {
		return nil
	}
	x.closed = true

	if x.cmd != nil && x.cmd.Process != nil {
		_ = x.cmd.Process.Kill()
		_, _ = x.cmd.Process.Wait()
	}
	return nil
}

func freeDisplay() (string, error) {
	for n := 99; n < 200; n++ {
		addr := fmt.Sprintf("/tmp/.X11-unix/X%d", n)
		conn, err := net.DialTimeout("unix", addr, 200*time.Millisecond)
		if err != nil {
			return fmt.Sprintf(":%d", n), nil
		}
		conn.Close()
	}
	return "", fmt.Errorf("no free X11 display found in range :99-:199")
}
