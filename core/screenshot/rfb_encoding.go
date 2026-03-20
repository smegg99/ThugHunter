// core/screenshot/rfb_encoding.go
//
// RFB pixel format configuration, encoding selection, and framebuffer requests.
package screenshot

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

// setPixelFormatRGBX tells the server to use 32-bit RGBX pixels.
func setPixelFormatRGBX(conn net.Conn) error {
	pf := pixelFormat{
		BitsPerPixel: 32, Depth: 24,
		BigEndian: 0, TrueColour: 1,
		RedMax: 255, GreenMax: 255, BlueMax: 255,
		RedShift: 16, GreenShift: 8, BlueShift: 0,
	}
	buf := &bytes.Buffer{}
	buf.WriteByte(0)           // message type
	buf.Write([]byte{0, 0, 0}) // padding
	binary.Write(buf, binary.BigEndian, pf)
	_, err := conn.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf("set pixel format: %w", err)
	}
	return nil
}

// setEncodings tells the server to prefer CopyRect with Raw fallback.
func setEncodings(conn net.Conn) error {
	buf := &bytes.Buffer{}
	buf.WriteByte(2)                               // message type
	buf.WriteByte(0)                               // padding
	binary.Write(buf, binary.BigEndian, uint16(2)) // count
	binary.Write(buf, binary.BigEndian, int32(1))  // CopyRect
	binary.Write(buf, binary.BigEndian, int32(0))  // Raw
	_, err := conn.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf("set encodings: %w", err)
	}
	return nil
}

// requestFramebuffer sends a FramebufferUpdateRequest.
func requestFramebuffer(conn net.Conn, width, height int, incremental bool) error {
	buf := &bytes.Buffer{}
	buf.WriteByte(3) // message type
	if incremental {
		buf.WriteByte(1)
	} else {
		buf.WriteByte(0)
	}
	binary.Write(buf, binary.BigEndian, uint16(0))      // x
	binary.Write(buf, binary.BigEndian, uint16(0))      // y
	binary.Write(buf, binary.BigEndian, uint16(width))  // width
	binary.Write(buf, binary.BigEndian, uint16(height)) // height
	_, err := conn.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf("request framebuffer: %w", err)
	}
	return nil
}
