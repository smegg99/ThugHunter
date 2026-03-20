// core/scanner/probe_vnc.go
package scanner

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/core/models"
)

// VNCProbe implements the Probe interface for VNC (RFB) services.
type VNCProbe struct{}

func (p *VNCProbe) Name() string         { return "VNC" }
func (p *VNCProbe) ServiceLabel() string { return "VNC" }

func (p *VNCProbe) Check(ctx context.Context, ip string, port int, sc scanConfig) ServiceResult {
	addr := net.JoinHostPort(ip, fmt.Sprintf("%d", port))

	logger.Debug().Str("ip", ip).Int("port", port).Msg("VNC probe: connecting")

	dialer := net.Dialer{Timeout: time.Duration(sc.connectTimeout) * time.Second}
	start := time.Now()
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	latency := time.Since(start)

	if err != nil {
		status := StatusError
		if isTimeout(err) {
			status = StatusTimeout
		}
		return ServiceResult{
			Service: models.ServiceTypeVNC,
			Port:    port,
			Open:    false,
			Latency: latency,
			Status:  status,
			Error:   err.Error(),
		}
	}
	defer conn.Close()

	// Set deadline for the banner/handshake exchange.
	conn.SetDeadline(time.Now().Add(time.Duration(sc.bannerTimeout) * time.Second))

	detail, err := readRFBHandshake(conn)
	if err != nil {
		logger.Debug().Str("ip", ip).Int("port", port).Err(err).Msg("VNC handshake failed")
		return ServiceResult{
			Service: models.ServiceTypeVNC,
			Port:    port,
			Open:    true,
			Latency: latency,
			Status:  StatusError,
			Error:   err.Error(),
		}
	}

	logger.Debug().Str("ip", ip).Int("port", port).
		Str("rfb_version", detail.RFBVersion).
		Str("auth", detail.AuthType.String()).
		Bool("no_auth", detail.NoAuth).
		Msg("VNC probe: handshake complete")

	return ServiceResult{
		Service: models.ServiceTypeVNC,
		Port:    port,
		Open:    true,
		Latency: latency,
		Status:  StatusOK,
		Detail:  detail,
	}
}

// readRFBHandshake reads the server's RFB version banner and security types.
func readRFBHandshake(conn net.Conn) (*VNCDetail, error) {
	reader := bufio.NewReader(conn)

	banner := make([]byte, 12)
	if _, err := io.ReadFull(reader, banner); err != nil {
		return nil, fmt.Errorf("read RFB banner: %w", err)
	}

	version := strings.TrimSpace(string(banner))
	if !strings.HasPrefix(version, "RFB ") {
		return nil, fmt.Errorf("not an RFB server: %q", version)
	}

	if _, err := conn.Write(banner); err != nil {
		return nil, fmt.Errorf("write RFB version: %w", err)
	}

	detail := &VNCDetail{RFBVersion: version, AuthType: models.VNCAuthUnknown}

	if strings.HasPrefix(version, "RFB 003.003") {
		return readRFB33Security(reader, detail)
	}
	return readRFB37Security(reader, detail)
}

// readRFB33Security handles the RFB 3.3 security negotiation (single uint32).
func readRFB33Security(reader *bufio.Reader, detail *VNCDetail) (*VNCDetail, error) {
	var secType uint32
	if err := binary.Read(reader, binary.BigEndian, &secType); err != nil {
		return nil, fmt.Errorf("read RFB 3.3 security type: %w", err)
	}

	detail.AuthType = models.VNCAuthType(secType)
	detail.NoAuth = secType == uint32(models.VNCAuthNone)
	return detail, nil
}

// readRFB37Security handles RFB 3.7+ security negotiation (count + list).
func readRFB37Security(reader *bufio.Reader, detail *VNCDetail) (*VNCDetail, error) {
	countByte, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("read security type count: %w", err)
	}
	count := int(countByte)

	if count == 0 {
		// Server sent 0 types - read the failure reason string.
		var reasonLen uint32
		if err := binary.Read(reader, binary.BigEndian, &reasonLen); err != nil {
			return nil, fmt.Errorf("read failure reason length: %w", err)
		}
		if reasonLen > 4096 {
			reasonLen = 4096
		}
		reason := make([]byte, reasonLen)
		if _, err := io.ReadFull(reader, reason); err != nil {
			return nil, fmt.Errorf("read failure reason: %w", err)
		}
		return nil, fmt.Errorf("server rejected connection: %s", string(reason))
	}

	types := make([]byte, count)
	if _, err := io.ReadFull(reader, types); err != nil {
		return nil, fmt.Errorf("read security types: %w", err)
	}

	// Check for no-auth (type 1). Otherwise take the first type.
	for _, t := range types {
		if models.VNCAuthType(t) == models.VNCAuthNone {
			detail.AuthType = models.VNCAuthNone
			detail.NoAuth = true
			return detail, nil
		}
	}

	detail.AuthType = models.VNCAuthType(types[0])
	detail.NoAuth = false
	return detail, nil
}
