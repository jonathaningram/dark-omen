// Package spr implements decoding of Dark Omen's .SPR sprite files.
//
// The method used in this decoder is based off the method from the Dark Omen
// Wiki at http://wiki.dark-omen.org/do/DO/Updated_Sprite_Format.
package spr

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"io"

	"github.com/disintegration/imaging"
)

const (
	// format is the sprite format ID used in all .SPR files.
	// "WHDO" is an initialism for "Warhammer: Dark Omen".
	format = "WHDO"

	headerSize      = 32
	frameHeaderSize = 32
)

// Decoder reads and decodes a sprite from an input stream.
type Decoder struct {
	r io.ReaderAt
}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.ReaderAt) *Decoder {
	return &Decoder{r: r}
}

// Decode reads the encoded sprite information from its input and returns a new
// Sprite containing decoded information and frames.
func (d *Decoder) Decode() (*Sprite, error) {
	header, err := d.readHeader()
	if err != nil {
		return nil, err
	}

	if f := header.format; f != format {
		return nil, fmt.Errorf("unknown sprite format %q, expected %q", f, format)
	}

	frameHeaders, err := d.readFrameHeaders(header)
	if err != nil {
		return nil, err
	}

	sprite := &Sprite{format: header.format}

	if len(frameHeaders) == 0 {
		return sprite, nil
	}

	colors, err := d.readColorTable(header)
	if err != nil {
		return nil, err
	}

	frames, err := d.readFrameData(header, frameHeaders, colors)
	if err != nil {
		return nil, err
	}

	sprite.Frames = frames

	return sprite, nil
}

type header struct {
	format            string
	fileSize          uint16
	frameHeaderOffset uint16
	frameDataOffset   int64
	colorTableOffset  int64
	colorTableEntries uint16
	paletteCount      uint16
	frameCount        uint16
}

func (d *Decoder) readHeader() (*header, error) {
	buf := make([]byte, headerSize)
	n, err := d.r.ReadAt(buf, 0)
	if n != headerSize {
		return nil, fmt.Errorf("header only read %d byte(s), expected %d", n, headerSize)
	}
	if err != nil && err != io.EOF {
		return nil, err
	}

	return &header{
		format:            string(buf[0:4]),
		fileSize:          binary.LittleEndian.Uint16(buf[4:8]),
		frameHeaderOffset: binary.LittleEndian.Uint16(buf[8:12]),
		frameDataOffset:   int64(binary.LittleEndian.Uint16(buf[12:16])),
		colorTableOffset:  int64(binary.LittleEndian.Uint16(buf[16:20])),
		colorTableEntries: binary.LittleEndian.Uint16(buf[20:24]),
		paletteCount:      binary.LittleEndian.Uint16(buf[24:28]),
		frameCount:        binary.LittleEndian.Uint16(buf[28:32]),
	}, nil
}

type frameHeader struct {
	frameType        FrameType
	compressionType  compressionType
	colorCount       int
	x, y             int
	width, height    int
	dataOffset       int64
	compressedSize   int
	uncompressedSize int
	colorTableOffset int
	// last 4 bytes are not used
}

func (d *Decoder) readFrameHeaders(header *header) ([]*frameHeader, error) {
	headers := make([]*frameHeader, header.frameCount)

	for i := uint16(0); i < header.frameCount; i++ {
		entry := make([]byte, frameHeaderSize)
		_, err := d.r.ReadAt(entry, int64(header.frameHeaderOffset+i*frameHeaderSize))
		if err != nil {
			if err == io.EOF {
				return nil, fmt.Errorf("sprite does not contain enough frame headers, expected to find %d, but got EOF while reading frame at index %d: %w", header.frameCount, i, io.ErrUnexpectedEOF)
			}
			return nil, err
		}

		frameType := FrameType(entry[0])
		compressionType := compressionType(entry[1])
		colorCount := binary.LittleEndian.Uint16(entry[2:4])
		x := binary.LittleEndian.Uint16(entry[4:6])
		y := binary.LittleEndian.Uint16(entry[6:8])
		w := binary.LittleEndian.Uint16(entry[8:10])
		h := binary.LittleEndian.Uint16(entry[10:12])
		dataOffset := binary.LittleEndian.Uint32(entry[12:16])
		compressedSize := binary.LittleEndian.Uint16(entry[16:20])
		uncompressedSize := binary.LittleEndian.Uint16(entry[20:24])
		colorTableOffset := binary.LittleEndian.Uint16(entry[24:28])
		// last 4 bytes are not used

		headers[i] = &frameHeader{
			frameType:        frameType,
			compressionType:  compressionType,
			colorCount:       int(colorCount),
			x:                int(x),
			y:                int(y),
			width:            int(w),
			height:           int(h),
			dataOffset:       int64(dataOffset),
			compressedSize:   int(compressedSize),
			uncompressedSize: int(uncompressedSize),
			colorTableOffset: int(colorTableOffset),
			// last 4 bytes are not used
		}
	}

	return headers, nil
}

func (d *Decoder) readColorTable(header *header) ([]color.RGBA, error) {
	colorTable := make([]byte, 4*header.colorTableEntries)
	_, err := d.r.ReadAt(colorTable, header.colorTableOffset)
	if err != nil {
		return nil, err
	}

	colors := make([]color.RGBA, header.colorTableEntries)

	for i := uint16(0); i < header.colorTableEntries; i++ {
		entry := colorTable[4*i : 4*(i+1)]

		// byte 4 (index 3) is not used
		b, g, r, a := entry[0], entry[1], entry[2], uint8(255)

		if b < 8 && g < 8 && r < 8 {
			a = 0
		}

		colors[i] = color.RGBA{
			B: entry[0],
			G: entry[1],
			R: entry[2],
			A: a,
		}
	}

	return colors, nil
}

func (d *Decoder) readFrameData(header *header, frameHeaders []*frameHeader, colors []color.RGBA) ([]*Frame, error) {
	frames := make([]*Frame, len(frameHeaders))

	for i, info := range frameHeaders {
		var raw []byte
		var err error

		switch info.compressionType {
		case compressionTypeNone:
			raw = make([]byte, info.compressedSize)
			_, err := d.r.ReadAt(raw, header.frameDataOffset+info.dataOffset)
			if err != nil {
				return nil, err
			}
		case compressionTypePackbits:
			raw, err = unpackBits(io.NewSectionReader(d.r, header.frameDataOffset+info.dataOffset, int64(info.compressedSize)))
			if err != nil {
				return nil, err
			}
		case compressionTypeZeroRuns:
			raw, err = zeroRuns(io.NewSectionReader(d.r, header.frameDataOffset+info.dataOffset, int64(info.compressedSize)))
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unsupported compression type %d", info.compressionType)
		}

		img := image.NewNRGBA(image.Rect(0, 0, info.width, info.height))

		var x int
		var y int

		for _, b := range raw {
			img.Set(x, y, colors[info.colorTableOffset+int(b)])

			if x == img.Rect.Max.X-1 {
				x = 0
				y++
				continue
			}
			x++
		}

		switch info.frameType {
		case FrameTypeFlipHorizontally:
			img = imaging.FlipH(img)
		case FrameTypeFlipHorizontallyAndVertically:
			img = imaging.FlipH(img)
			img = imaging.FlipV(img)
		case FrameTypeFlipVertically:
			img = imaging.FlipV(img)
		}

		frames[i] = &Frame{
			Type:  info.frameType,
			Image: img,
		}
	}

	return frames, nil
}

// A Sprite is made up of a list of frames.
type Sprite struct {
	format string
	Frames []*Frame
}

// FrameType provides information about how to interpret the frame image.
type FrameType uint8

const (
	// FrameTypeRepeat indicates this frame is a repeat of a previous frame.
	FrameTypeRepeat FrameType = iota
	// FrameTypeFlipHorizontally indicates this frame should be flipped
	// horizontally.
	FrameTypeFlipHorizontally
	// FrameTypeFlipVertically indicates this frame should be flipped
	// vertically.
	FrameTypeFlipVertically
	// FrameTypeFlipHorizontallyAndVertically indicates this frame should be
	// flipped horizontally and vertically.
	FrameTypeFlipHorizontallyAndVertically
	// FrameTypeNormal indicates this is a normal frame.
	FrameTypeNormal
	// FrameTypeEmpty indicates the frame is empty.
	// There is no frame or palette data associated with this frame.
	// The frame's width and height are 0.
	FrameTypeEmpty
)

type compressionType uint8

const (
	compressionTypeNone compressionType = iota
	compressionTypePackbits
	compressionTypeZeroRuns
)

// A Frame contains an in-memory representation of the image.
type Frame struct {
	// Type provides information about how to interpret the frame image.
	Type FrameType
	// Image is the decoded frame data converted into an image.Image.
	Image image.Image
}
