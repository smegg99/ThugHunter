// core/screenshot/rfb_framebuffer.go
package screenshot

import (
	"bufio"
	"context"
	"encoding/binary"
	"image"
	"io"
	"net"

	"smegg.me/thughunter/common/logger"
)

// readFramebufferUpdate reads server messages into img until all pixels
// are filled or ctx is cancelled. Returns the pixel count received.
func readFramebufferUpdate(ctx context.Context, reader *bufio.Reader, conn net.Conn, img *image.RGBA, width, height int) int {
	pix := img.Pix
	stride := img.Stride
	total := width * height
	received := 0

	for received < total {
		if ctx.Err() != nil {
			break
		}

		msgType, err := reader.ReadByte()
		if err != nil {
			break
		}

		switch msgType {
		case 0: // FramebufferUpdate
			n := handleFBUpdateMsg(reader, pix, stride, width, height, received)
			if n < 0 {
				return received
			}
			received += n
			if received < total {
				_ = requestFramebuffer(conn, width, height, true)
			}
		case 1: // SetColourMapEntries
			skipColourMap(reader)
		case 2: // Bell
			continue
		case 3: // ServerCutText
			skipCutText(reader)
		default:
			logger.Debug().Uint8("type", msgType).Msg("unknown RFB message, stopping")
			return received
		}
	}
	return received
}

// handleFBUpdateMsg decodes one FramebufferUpdate (padding + rectangles).
// Returns new pixel count, or -1 on fatal read error.
func handleFBUpdateMsg(reader *bufio.Reader, pix []byte, stride, width, height, already int) int {
	if _, err := reader.ReadByte(); err != nil {
		return -1 // padding
	}
	var numRects uint16
	if err := binary.Read(reader, binary.BigEndian, &numRects); err != nil {
		return -1
	}

	pixels := 0
	for i := 0; i < int(numRects); i++ {
		var rect struct {
			X, Y, W, H uint16
			Encoding   int32
		}
		if err := binary.Read(reader, binary.BigEndian, &rect); err != nil {
			return -1
		}

		rw, rh := int(rect.W), int(rect.H)
		n := -1
		switch rect.Encoding {
		case 0: // Raw
			n = decodeRawRect(reader, pix, stride, width, height, int(rect.X), int(rect.Y), rw, rh)
		case 1: // CopyRect
			n = decodeCopyRect(reader, pix, stride, width, height, int(rect.X), int(rect.Y), rw, rh)
		default:
			logger.Debug().Int32("encoding", rect.Encoding).Msg("unsupported encoding, skipping")
			return pixels
		}

		if n < 0 && (already+pixels) == 0 {
			return -1
		}
		if n < 0 {
			return pixels
		}
		pixels += n
	}
	return pixels
}

// decodeRawRect reads raw BGRX pixel data for a rectangle into the image.
func decodeRawRect(reader *bufio.Reader, pix []byte, stride, imgW, imgH, rx, ry, rw, rh int) int {
	data := make([]byte, rw*rh*4)
	if _, err := io.ReadFull(reader, data); err != nil {
		return -1
	}

	for py := 0; py < rh; py++ {
		iy := ry + py
		if iy >= imgH {
			break
		}
		dst := iy*stride + rx*4
		src := py * rw * 4
		for px := 0; px < rw; px++ {
			if rx+px >= imgW {
				break
			}
			pix[dst] = data[src+2]   // R (server sends B,G,R,X)
			pix[dst+1] = data[src+1] // G
			pix[dst+2] = data[src]   // B
			pix[dst+3] = 255         // A
			dst += 4
			src += 4
		}
	}
	return rw * rh
}

// decodeCopyRect copies a rectangle from another position in the image.
func decodeCopyRect(reader *bufio.Reader, pix []byte, stride, imgW, imgH, dx, dy, rw, rh int) int {
	var srcX, srcY uint16
	if err := binary.Read(reader, binary.BigEndian, &srcX); err != nil {
		return -1
	}
	if err := binary.Read(reader, binary.BigEndian, &srcY); err != nil {
		return -1
	}

	sx, sy := int(srcX), int(srcY)
	for py := 0; py < rh; py++ {
		for px := 0; px < rw; px++ {
			six, siy := sx+px, sy+py
			dix, diy := dx+px, dy+py
			if six < imgW && siy < imgH && dix < imgW && diy < imgH {
				sOff := siy*stride + six*4
				dOff := diy*stride + dix*4
				copy(pix[dOff:dOff+4], pix[sOff:sOff+4])
			}
		}
	}
	return rw * rh
}

// skipColourMap skips a SetColourMapEntries message.
func skipColourMap(reader *bufio.Reader) {
	skip := make([]byte, 5)
	io.ReadFull(reader, skip)
	n := binary.BigEndian.Uint16(skip[3:5])
	io.ReadFull(reader, make([]byte, int(n)*6))
}

// skipCutText skips a ServerCutText message.
func skipCutText(reader *bufio.Reader) {
	skip := make([]byte, 7)
	io.ReadFull(reader, skip)
	textLen := binary.BigEndian.Uint32(skip[3:7])
	if textLen > 1<<20 {
		textLen = 1 << 20
	}
	io.ReadFull(reader, make([]byte, textLen))
}
