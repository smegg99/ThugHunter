// core/screenshot/native.go
package screenshot

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"net"
	"time"

	"smegg.me/thughunter/common/logger"
)

const dialTimeout = 5 * time.Second

// CaptureNative connects to a VNC server, performs the RFB no-auth
// handshake, reads the framebuffer, and returns JPEG bytes.
func CaptureNative(ctx context.Context, ip string, port int) ([]byte, error) {
	conn, reader, err := dialVNC(ctx, ip, port)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Unblock reads immediately on cancellation.
	go func() {
		<-ctx.Done()
		conn.SetDeadline(time.Now())
	}()

	serverInit, err := handshake(reader, conn)
	if err != nil {
		return nil, err
	}

	if err := configurePixelFormat(conn); err != nil {
		return nil, err
	}

	return captureFramebuffer(ctx, reader, conn, ip, port, serverInit)
}

// dialVNC opens a TCP connection with a capped connect timeout and
// propagates the parent context's deadline to the connection.
func dialVNC(ctx context.Context, ip string, port int) (net.Conn, *bufio.Reader, error) {
	addr := net.JoinHostPort(ip, fmt.Sprintf("%d", port))

	dialCtx, dialCancel := context.WithTimeout(ctx, dialTimeout)
	defer dialCancel()

	conn, err := (&net.Dialer{}).DialContext(dialCtx, "tcp", addr)
	if err != nil {
		return nil, nil, fmt.Errorf("connect: %w", err)
	}

	if dl, ok := ctx.Deadline(); ok {
		conn.SetDeadline(dl)
	}

	reader := bufio.NewReaderSize(conn, 64*1024)
	return conn, reader, nil
}

// handshake negotiates RFB version, security (no-auth), and reads ServerInit.
func handshake(reader *bufio.Reader, conn net.Conn) (rfbServerInit, error) {
	vr, err := negotiateVersion(reader, conn)
	if err != nil {
		return rfbServerInit{}, err
	}

	if err := negotiateSecurity(reader, conn, vr); err != nil {
		return rfbServerInit{}, err
	}

	return readServerInit(reader, conn)
}

// configurePixelFormat tells the server to use RGBX pixels and our
// preferred encoding list.
func configurePixelFormat(conn net.Conn) error {
	if err := setPixelFormatRGBX(conn); err != nil {
		return err
	}
	return setEncodings(conn)
}

const (
	// Settle time between incremental framebuffer requests.
	renderSettleDelay = 1500 * time.Millisecond
	// Incremental-update rounds after the initial capture.
	maxRenderRetries = 2
)

// captureFramebuffer reads the full framebuffer, retries with incremental
// updates for a sharper image, and returns the result as JPEG bytes.
func captureFramebuffer(ctx context.Context, reader *bufio.Reader, conn net.Conn, ip string, port int, si rfbServerInit) ([]byte, error) {
	if err := requestFramebuffer(conn, si.Width, si.Height, false); err != nil {
		return nil, err
	}

	img := image.NewRGBA(image.Rect(0, 0, si.Width, si.Height))
	pixels := readFramebufferUpdate(ctx, reader, conn, img, si.Width, si.Height)

	if pixels == 0 {
		return nil, fmt.Errorf("no framebuffer data received")
	}

	pixels += retryIncrementalUpdates(ctx, reader, conn, img, si)

	logger.Debug().
		Str("ip", ip).Int("port", port).
		Int("pixels", pixels).Int("total", si.Width*si.Height).
		Msg("native VNC framebuffer captured")

	return encodeJPEG(img)
}

// retryIncrementalUpdates waits briefly then re-reads the framebuffer
// so the server can finish drawing.
func retryIncrementalUpdates(ctx context.Context, reader *bufio.Reader, conn net.Conn, img *image.RGBA, si rfbServerInit) int {
	total := 0
	for i := 0; i < maxRenderRetries; i++ {
		select {
		case <-time.After(renderSettleDelay):
		case <-ctx.Done():
			return total
		}
		if err := requestFramebuffer(conn, si.Width, si.Height, true); err != nil {
			return total
		}
		total += readFramebufferUpdate(ctx, reader, conn, img, si.Width, si.Height)
	}
	return total
}

// encodeJPEG encodes an RGBA image to JPEG at quality 80.
func encodeJPEG(img *image.RGBA) ([]byte, error) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}); err != nil {
		return nil, fmt.Errorf("encode JPEG: %w", err)
	}
	return buf.Bytes(), nil
}
