// Package dot implements decoding of Dark Omen's .DOT path files.
package dot

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/jonathaningram/dark-omen/internal/cstringutil"
)

const (
	// format is the format ID used in all .DOT files.
	// "TODW" is probably "WDOT" backwards, which is probably "Warhammer: DOT".
	format = "TODW"

	headerSize = 16
	footerSize = 152
	// footerMapFileOffset is the offset from the start of the footer.
	footerMapFileOffset = 80
)

// Map is made up of a number paths.
type Map struct {
	format string
	Paths  []*Path
	// FileName is the name of the English bitmap file for this map.
	// In Dark Omen, the French and German ENGREL.EXE files refer to their own
	// localized file name so it's likely this is not used.
	FileName string
}

// A Path is made up of a number points at a given x and y coordinate.
type Path struct {
	Points   []Point
	unknown1 uint32 // always 0x05
	unknown2 uint32 // always 0x0A
	unknown3 [36]byte
}

// A Point is an x and y coordinate into a Dark Omen map image.
type Point struct {
	X, Y uint32
}

// Decoder reads and decodes a DOT file from an input stream.
type Decoder struct {
	r io.Reader
}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

// Decode reads the encoded paths information from its input and returns a map
// containing decoded path coordinates.
func (d *Decoder) Decode() (*Map, error) {
	header, err := d.readHeader()
	if err != nil {
		return nil, fmt.Errorf("could not read header: %w", err)
	}

	if f := header.format; f != format {
		return nil, fmt.Errorf("unknown format %q, expected %q", f, format)
	}

	paths, err := d.readPaths(header.pathCount)
	if err != nil {
		return nil, err
	}

	footer, err := d.readFooter()
	if err != nil {
		return nil, fmt.Errorf("could not read footer: %w", err)
	}

	return &Map{
		format:   format,
		Paths:    paths,
		FileName: footer.mapFileName,
	}, nil
}

type header struct {
	format    string
	unknown1  uint32
	unknown2  uint32
	pathCount uint32
}

func (d *Decoder) readHeader() (h *header, err error) {
	var buf [headerSize]byte
	n, err := d.r.Read(buf[:])
	if n != headerSize {
		return nil, fmt.Errorf("read %d byte(s), expected %d", n, headerSize)
	}
	if err != nil && err != io.EOF {
		return nil, err
	}

	return &header{
		format:    string(buf[0:4]),
		unknown1:  binary.LittleEndian.Uint32(buf[4:8]),
		unknown2:  binary.LittleEndian.Uint32(buf[8:12]),
		pathCount: binary.LittleEndian.Uint32(buf[12:16]),
	}, nil
}

func (d *Decoder) readPaths(count uint32) (paths []*Path, err error) {
	paths = make([]*Path, count)

	for i := uint32(0); i < count; i++ {
		paths[i], err = d.readPath()
		if err != nil {
			return nil, fmt.Errorf("could not read path %d: %w", i, err)
		}
	}

	return paths, nil
}

func (d *Decoder) readPath() (path *Path, err error) {
	var pointCount uint32
	if err := binary.Read(d.r, binary.LittleEndian, &pointCount); err != nil {
		return nil, err
	}

	points := make([]Point, pointCount)
	for i := uint32(0); i < pointCount; i++ {
		var x uint32
		if err := binary.Read(d.r, binary.LittleEndian, &x); err != nil {
			return nil, err
		}
		var y uint32
		if err := binary.Read(d.r, binary.LittleEndian, &y); err != nil {
			return nil, err
		}
		var padding [8]byte
		n, err := d.r.Read(padding[:])
		if n != 8 {
			return nil, fmt.Errorf("read %d byte(s), expected %d", n, 8)
		}
		if err != nil {
			return nil, err
		}

		points[i] = Point{X: x, Y: y}
	}

	var unknown1 uint32
	if err := binary.Read(d.r, binary.LittleEndian, &unknown1); err != nil {
		return nil, err
	}
	var unknown2 uint32
	if err := binary.Read(d.r, binary.LittleEndian, &unknown2); err != nil {
		return nil, err
	}
	var unknown3 [36]byte
	n, err := d.r.Read(unknown3[:])
	if n != 36 {
		return nil, fmt.Errorf("read %d byte(s), expected %d", n, 36)
	}
	if err != nil {
		return nil, err
	}

	return &Path{
		Points:   points,
		unknown1: unknown1,
		unknown2: unknown2,
		unknown3: unknown3,
	}, nil
}

type footer struct {
	mapFileName string
}

func (d *Decoder) readFooter() (f *footer, err error) {
	var buf [footerSize]byte
	n, err := d.r.Read(buf[:])
	if n != footerSize {
		return nil, fmt.Errorf("read %d byte(s), expected %d", n, footerSize)
	}
	if err != nil && err != io.EOF {
		return nil, err
	}

	return &footer{
		mapFileName: cstringutil.ToGo(buf[footerMapFileOffset:]),
	}, nil
}
