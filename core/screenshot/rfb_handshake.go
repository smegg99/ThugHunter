// core/screenshot/rfb_handshake.go
package screenshot

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
)

// negotiateVersion reads the server banner and agrees on an RFB version.
func negotiateVersion(reader *bufio.Reader, conn net.Conn) (rfbVersionResult, error) {
	banner := make([]byte, 12)
	if _, err := io.ReadFull(reader, banner); err != nil {
		return rfbVersionResult{}, fmt.Errorf("read RFB banner: %w", err)
	}
	version := strings.TrimSpace(string(banner))
	if !strings.HasPrefix(version, "RFB ") {
		return rfbVersionResult{}, fmt.Errorf("not an RFB server: %q", version)
	}

	vr := rfbVersionResult{
		ServerIs33: strings.HasPrefix(version, "RFB 003.003"),
		IsV38:      strings.Compare(version, "RFB 003.008") >= 0,
	}

	reply := banner // echo back by default
	switch {
	case vr.ServerIs33:
		reply = []byte("RFB 003.003\n")
		vr.IsV38 = false
	case vr.IsV38:
		reply = []byte("RFB 003.008\n")
	}

	if _, err := conn.Write(reply); err != nil {
		return rfbVersionResult{}, fmt.Errorf("write version: %w", err)
	}
	return vr, nil
}

// negotiateSecurity selects no-auth (type 1). RFB 3.3 sends a single
// uint32 type; 3.7+ sends a count byte + type list with client response.
func negotiateSecurity(reader *bufio.Reader, conn net.Conn, vr rfbVersionResult) error {
	if vr.ServerIs33 {
		return negotiateSecurity33(reader)
	}
	return negotiateSecurity37(reader, conn, vr.IsV38)
}

// negotiateSecurity33 handles the RFB 3.3 single-type security.
func negotiateSecurity33(reader *bufio.Reader) error {
	var secType uint32
	if err := binary.Read(reader, binary.BigEndian, &secType); err != nil {
		return fmt.Errorf("read RFB 3.3 security type: %w", err)
	}
	if secType == 0 {
		return readServerRejection(reader)
	}
	if secType != 1 {
		return fmt.Errorf("no-auth not offered (RFB 3.3 type: %d)", secType)
	}
	return nil
}

// negotiateSecurity37 handles the RFB 3.7/3.8 count+list security.
func negotiateSecurity37(reader *bufio.Reader, conn net.Conn, isV38 bool) error {
	countByte, err := reader.ReadByte()
	if err != nil {
		return fmt.Errorf("read security count: %w", err)
	}
	if countByte == 0 {
		return readServerRejection(reader)
	}

	types := make([]byte, int(countByte))
	if _, err := io.ReadFull(reader, types); err != nil {
		return fmt.Errorf("read security types: %w", err)
	}

	if !byteSliceContains(types, 1) {
		return fmt.Errorf("no-auth not offered (types: %v)", types)
	}

	if _, err := conn.Write([]byte{1}); err != nil {
		return fmt.Errorf("select no-auth: %w", err)
	}

	if isV38 {
		var secResult uint32
		if err := binary.Read(reader, binary.BigEndian, &secResult); err != nil {
			if err != io.EOF && err != io.ErrUnexpectedEOF {
				return fmt.Errorf("read security result: %w", err)
			}
		} else if secResult != 0 {
			return fmt.Errorf("security handshake failed (result=%d)", secResult)
		}
	}
	return nil
}

// readServerRejection reads the server's reason-string after a zero count.
func readServerRejection(reader *bufio.Reader) error {
	var reasonLen uint32
	if err := binary.Read(reader, binary.BigEndian, &reasonLen); err != nil {
		return fmt.Errorf("server sent 0 security types")
	}
	if reasonLen > 4096 {
		reasonLen = 4096
	}
	reason := make([]byte, reasonLen)
	if _, err := io.ReadFull(reader, reason); err != nil {
		return fmt.Errorf("server sent 0 security types")
	}
	return fmt.Errorf("server rejected connection: %s", string(reason))
}

// byteSliceContains returns true if b contains v.
func byteSliceContains(b []byte, v byte) bool {
	for _, x := range b {
		if x == v {
			return true
		}
	}
	return false
}

// readServerInit sends ClientInit (shared=1) and reads the ServerInit
// response: framebuffer dimensions, pixel format, and desktop name.
func readServerInit(reader *bufio.Reader, conn net.Conn) (rfbServerInit, error) {
	if _, err := conn.Write([]byte{1}); err != nil {
		return rfbServerInit{}, fmt.Errorf("write ClientInit: %w", err)
	}

	var dims struct {
		Width  uint16
		Height uint16
	}
	if err := binary.Read(reader, binary.BigEndian, &dims); err != nil {
		return rfbServerInit{}, fmt.Errorf("read ServerInit dimensions: %w", err)
	}

	width, height := int(dims.Width), int(dims.Height)
	if width == 0 || height == 0 || width > 8192 || height > 8192 {
		return rfbServerInit{}, fmt.Errorf("invalid framebuffer size: %dx%d", width, height)
	}

	if err := skipPixelFormatAndName(reader); err != nil {
		return rfbServerInit{}, err
	}
	return rfbServerInit{Width: width, Height: height}, nil
}

// skipPixelFormatAndName reads and discards the 16-byte pixel format
// and variable-length desktop name from a ServerInit message.
func skipPixelFormatAndName(reader *bufio.Reader) error {
	var pf pixelFormat
	if err := binary.Read(reader, binary.BigEndian, &pf); err != nil {
		return fmt.Errorf("read pixel format: %w", err)
	}
	var nameLen uint32
	if err := binary.Read(reader, binary.BigEndian, &nameLen); err != nil {
		return fmt.Errorf("read name length: %w", err)
	}
	if nameLen > 4096 {
		nameLen = 4096
	}
	if nameLen > 0 {
		if _, err := io.ReadFull(reader, make([]byte, nameLen)); err != nil {
			return fmt.Errorf("read name: %w", err)
		}
	}
	return nil
}
