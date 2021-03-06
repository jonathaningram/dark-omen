package fnt

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"io"
)

const (
	// format is the format ID used in all .FNT files.
	format = "FONT"

	headerSize           = 16
	colorTableColorCount = 16
	colorTableSize       = 4 * colorTableColorCount
	glyphCount           = 256
	glyphHeaderSize      = 16
)

type Font struct {
	format      string
	colorTable1 []color.RGBA
	colorTable2 []color.RGBA

	BaseAdvanceWidth uint16
	// LineHeight is the total height of each line in the font. This can be used
	// for rendering text over multiple lines.
	// For example, F_BOOKS.FNT has a line height of 14 pixels.
	LineHeight uint32
	// The BaseAdvanceHeight is the amount of pixels to advance when rendering a
	// new line of text. If this is set to 0 then the bottoms of glyphs will
	// touch the tops of glyphs from the following line.
	BaseAdvanceHeight uint16

	// TODO
	Height2  uint16
	Unknown1 uint16

	Glyphs []*Glyph
}

// GlyphType provides information about how to interpret the glyph image.
type GlyphType uint8

const (
	// GlyphTypeNormal indicates this is a normal glyph.
	GlyphTypeNormal GlyphType = iota
	// GlyphTypeEmpty indicates the glyph is empty.
	// There is no glyph or palette data associated with this glyph.
	// The glyph's width and height are 0.
	GlyphTypeEmpty
)

type Glyph struct {
	// Type provides information about how to interpret the glyph image.
	Type GlyphType
	// Image is the decoded glyph data converted into an image.Image.
	Image image.Image
	// The AdvanceWidth of the glyph. After the glyph is rendered, the start of
	// the next glyph is offset on the x-axis from the current glyph origin by
	// this amount, plus the base advance width found in Font.BaseAdvanceWidth.
	// This is always less than or equal to the width of the glyph. It is the
	// offset on the x-axis to move the current point to the start of the next
	// glyph.
	AdvanceWidth uint16

	// TODO: Both of these affect vertical/y-axis positioning when rendering.
	Unknown1 uint8
	Unknown2 uint8
}

// Decoder reads and decodes a font from an input stream.
type Decoder struct {
	r io.ReaderAt
}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.ReaderAt) *Decoder {
	return &Decoder{r: r}
}

// Decode reads the encoded 3D model information from its input and returns a
// new model containing decoded textures and objects.
func (d *Decoder) Decode() (*Font, error) {
	header, pos, err := d.readHeader()
	if err != nil {
		return nil, fmt.Errorf("could not read header: %w", err)
	}
	if f := header.format; f != format {
		return nil, fmt.Errorf("unknown format %q, expected %q", f, format)
	}

	colorTable1, pos, err := d.readColorTable(pos)
	if err != nil {
		return nil, fmt.Errorf("could not read color table 1: %w", err)
	}
	colorTable2, pos, err := d.readColorTable(pos)
	if err != nil {
		return nil, fmt.Errorf("could not read color table 2: %w", err)
	}

	glyphHeaders, err := d.readGlyphHeaders(header, pos)
	if err != nil {
		return nil, fmt.Errorf("could not read glyph headers: %w", err)
	}

	f := &Font{
		format:            format,
		BaseAdvanceWidth:  header.baseAdvanceWidth,
		LineHeight:        header.lineHeight,
		BaseAdvanceHeight: header.baseAdvanceHeight,

		// TODO
		Height2:  header.height2,
		Unknown1: header.unknown1,
	}

	if len(glyphHeaders) == 0 {
		return f, nil
	}

	glyphs, err := d.readGlyphData(header, glyphHeaders, colorTable1, colorTable2)
	if err != nil {
		return nil, err
	}

	f.Glyphs = glyphs

	return f, nil
}

type header struct {
	format            string
	baseAdvanceWidth  uint16
	lineHeight        uint32
	baseAdvanceHeight uint16

	// TODO
	height2  uint16
	unknown1 uint16

	glyphDataOffset uint16
}

func (d *Decoder) readHeader() (h *header, pos int64, err error) {
	buf := make([]byte, headerSize)
	n, err := d.r.ReadAt(buf, 0)
	pos = int64(n)
	if n != headerSize {
		return nil, pos, fmt.Errorf("read %d byte(s), expected %d", n, headerSize)
	}
	if err != nil {
		return nil, pos, err
	}
	h = &header{
		format:            string(buf[0:4]),
		baseAdvanceWidth:  binary.LittleEndian.Uint16(buf[4:6]),
		lineHeight:        uint32(binary.LittleEndian.Uint16(buf[6:8]) + binary.LittleEndian.Uint16(buf[8:10])),
		baseAdvanceHeight: binary.LittleEndian.Uint16(buf[6:8]),

		// TODO
		height2:  binary.LittleEndian.Uint16(buf[8:10]),
		unknown1: binary.LittleEndian.Uint16(buf[10:12]),

		glyphDataOffset: binary.LittleEndian.Uint16(buf[12:14]),
		// buf[14:16] is always 0x0000
	}
	return h, pos, nil
}

func (d *Decoder) readColorTable(startPos int64) (colors []color.RGBA, pos int64, err error) {
	pos = startPos
	table := make([]byte, colorTableSize)
	n, err := d.r.ReadAt(table, pos)
	pos += int64(n)
	if n != colorTableSize {
		return nil, pos, fmt.Errorf("read %d byte(s), expected %d", n, colorTableSize)
	}
	if err != nil {
		return nil, pos, err
	}

	colors = make([]color.RGBA, colorTableColorCount)

	for i := uint16(0); i < colorTableColorCount; i++ {
		entry := table[4*i : 4*(i+1)]

		// byte 4 (index 3) is not used
		colors[i] = color.RGBA{
			B: entry[0],
			G: entry[1],
			R: entry[2],
			A: 255,
		}
	}

	return colors, pos, nil
}

type glyphHeader struct {
	typ GlyphType

	// TODO: Both of these affect vertical/y-axis positioning when rendering.
	unknown1 uint8
	unknown2 uint8

	width        uint16
	advanceWidth uint16
	height       uint16
	dataOffset   uint16
}

func (d *Decoder) readGlyphHeaders(header *header, startPos int64) ([]*glyphHeader, error) {
	pos := startPos

	headers := make([]*glyphHeader, glyphCount)

	for i := 0; i < glyphCount; i++ {
		buf := make([]byte, glyphHeaderSize)
		n, err := d.r.ReadAt(buf, pos)
		pos += int64(n)
		if n != glyphHeaderSize {
			return nil, fmt.Errorf("read %d byte(s), expected %d", n, glyphHeaderSize)
		}
		if err != nil {
			if err == io.EOF {
				return nil, fmt.Errorf("not enough glyph headers, expected to find %d, but got EOF while reading index %d: %w", glyphCount, i, io.ErrUnexpectedEOF)
			}
			return nil, err
		}

		h := &glyphHeader{
			// buf[0] is either decimal 0, 1 or 2. In F_BOOKS.FNT, the `,`
			// character is 1 but in F_TITLE.FNT, it is 2.
			unknown1: buf[2],
			unknown2: buf[3],

			width:        binary.LittleEndian.Uint16(buf[4:6]),
			advanceWidth: binary.LittleEndian.Uint16(buf[6:8]),
			height:       binary.LittleEndian.Uint16(buf[8:10]),
			// buf[10:12] is always 0x0000
			dataOffset: binary.LittleEndian.Uint16(buf[12:14]),
		}

		switch {
		case h.width == 0 && h.height == 0:
			h.typ = GlyphTypeEmpty
		default:
			h.typ = GlyphTypeNormal
		}

		headers[i] = h
	}

	return headers, nil
}

func (d *Decoder) readGlyphData(header *header, glyphHeaders []*glyphHeader, colors1, colors2 []color.RGBA) ([]*Glyph, error) {
	glyphs := make([]*Glyph, len(glyphHeaders))

	for i, info := range glyphHeaders {
		g := &Glyph{
			Type:         info.typ,
			AdvanceWidth: info.advanceWidth,

			// TODO
			Unknown1: info.unknown1,
			Unknown2: info.unknown2,
		}

		// Empty glyphs have no data but we still need a glyph at this index.
		if info.typ == GlyphTypeEmpty {
			glyphs[i] = g
			continue
		}

		size := info.width * info.height / 2
		raw := make([]byte, size)
		n, err := d.r.ReadAt(raw, int64(header.glyphDataOffset+info.dataOffset))
		if n != int(size) {
			return nil, fmt.Errorf("read %d byte(s), expected %d", n, size)
		}
		if err != nil {
			return nil, err
		}

		img := image.NewNRGBA(image.Rect(0, 0, int(info.width), int(info.height)))

		var x int
		var y int

		for _, b := range raw {
			lo := b & 0x0F
			hi := b >> 4

			img.Set(x, y, colors1[lo])
			x, y = xy(img, x, y)
			img.Set(x, y, colors1[hi])
			x, y = xy(img, x, y)
		}

		g.Image = img

		glyphs[i] = g
	}

	return glyphs, nil
}

// xy returns the next x, y coordinates for the given image keeping within its
// dimensions.
func xy(img *image.NRGBA, x, y int) (int, int) {
	if x == img.Rect.Max.X-1 {
		return 0, y + 1
	}
	return x + 1, y
}
