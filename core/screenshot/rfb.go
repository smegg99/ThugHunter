// core/screenshot/rfb.go
package screenshot

// pixelFormat matches the 16-byte RFB PixelFormat structure.
type pixelFormat struct {
	BitsPerPixel uint8
	Depth        uint8
	BigEndian    uint8
	TrueColour   uint8
	RedMax       uint16
	GreenMax     uint16
	BlueMax      uint16
	RedShift     uint8
	GreenShift   uint8
	BlueShift    uint8
	_            [3]byte // padding
}

// rfbServerInit holds width/height from the ServerInit message.
type rfbServerInit struct {
	Width  int
	Height int
}

// rfbVersionResult holds negotiated RFB version flags.
type rfbVersionResult struct {
	IsV38      bool
	ServerIs33 bool
}
