// Package m3d implements decoding of Dark Omen's .M3D 3D model files.
package m3d

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/jonathaningram/dark-omen/internal/cstringutil"
)

const (
	// format is the format ID used in all .M3D files.
	// "PD3M" is probably "M3DP" backwards, which is probably "Model 3D
	// <something>".
	format = "PD3M"

	headerSize       = 24
	textureSize      = 96
	vectorSize       = 12
	objectHeaderSize = 52 + vectorSize
	objectFaceSize   = 16 + vectorSize
	objectVertexSize = (2 * vectorSize) + 20
)

// A Model is made up of a list of textures and a list of objects.
// In Dark Omen, the format is always "PD3M".
type Model struct {
	format   string
	Textures []*Texture
	Objects  []*Object
}

// A Texture contains information about texturing a 3D surface.
type Texture struct {
	// Path appears to be a directory on the original Dark Omen developer's
	// machine. It does not seem to be used for anything useful and might best
	// be treated as an Easter egg.
	Path string
	// FileName is the name of the texture image file.
	FileName string
}

// A Vector in 3-dimensional space.
type Vector struct {
	X, Y, Z float32
}

type Object struct {
	Name        string
	ParentIndex int16
	padding     int16
	Pivot       Vector
	Flags       uint32
	unknown1    uint32
	unknown2    uint32
	Faces       []*Face
	Vertexes    []*Vertex
}

type Face struct {
	Indexes      [3]uint16
	TextureIndex uint16
	Normal       Vector
	unknown1     uint32
	unknown2     uint32
}

type Color struct {
	R, G, B, A uint8
}

type Vertex struct {
	Position Vector
	Normal   Vector
	Color    Color
	U, V     float32
	Index    uint32
	unknown1 uint32
}

// Decoder reads and decodes a 3D model from an input stream.
type Decoder struct {
	r io.ReaderAt
}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.ReaderAt) *Decoder {
	return &Decoder{r: r}
}

// Decode reads the encoded 3D model information from its input and returns a
// new model containing decoded textures and objects.
func (d *Decoder) Decode() (*Model, error) {
	header, pos, err := d.readHeader()
	if err != nil {
		return nil, fmt.Errorf("could not read header: %w", err)
	}

	if f := header.format; f != format {
		return nil, fmt.Errorf("unknown format %q, expected %q", f, format)
	}

	textures, pos, err := d.readTextures(header.textureCount, pos)
	if err != nil {
		return nil, err
	}

	objects, err := d.readObjects(header.objectCount, pos)
	if err != nil {
		return nil, err
	}

	return &Model{
		format:   format,
		Textures: textures,
		Objects:  objects,
	}, nil
}

type header struct {
	format       string
	magic        uint32
	version      uint32
	crc          uint32
	notCRC       uint32
	textureCount uint16
	objectCount  uint16
}

func (d *Decoder) readHeader() (h *header, pos int64, err error) {
	buf := make([]byte, headerSize)
	n, err := d.r.ReadAt(buf, 0)
	pos = int64(n)
	if n != headerSize {
		return nil, pos, fmt.Errorf("read %d byte(s), expected %d", n, headerSize)
	}
	if err != nil && err != io.EOF {
		return nil, pos, err
	}

	return &header{
		format:       string(buf[0:4]),
		magic:        binary.LittleEndian.Uint32(buf[4:8]),
		version:      binary.LittleEndian.Uint32(buf[8:12]),
		crc:          binary.LittleEndian.Uint32(buf[12:16]),
		notCRC:       binary.LittleEndian.Uint32(buf[16:20]),
		textureCount: binary.LittleEndian.Uint16(buf[20:22]),
		objectCount:  binary.LittleEndian.Uint16(buf[22:24]),
	}, pos, nil
}

func (d *Decoder) readTextures(count uint16, startPos int64) (textures []*Texture, pos int64, err error) {
	textures = make([]*Texture, count)
	pos = startPos

	for i := uint16(0); i < count; i++ {
		textures[i], pos, err = d.readTexture(pos)
		if err != nil {
			return nil, pos, fmt.Errorf("could not read texture %d: %w", i, err)
		}
	}

	return textures, pos, nil
}

func (d *Decoder) readTexture(startPos int64) (texture *Texture, pos int64, err error) {
	pos = startPos

	buf := make([]byte, textureSize)
	n, err := d.r.ReadAt(buf, pos)
	pos += int64(n)
	if n != textureSize {
		return nil, pos, fmt.Errorf("read %d byte(s), expected %d", n, textureSize)
	}
	if err != nil && err != io.EOF {
		return nil, pos, err
	}
	return &Texture{
		Path:     cstringutil.ToGo(buf[:64]),
		FileName: cstringutil.ToGo(buf[64:]),
	}, pos, nil
}

func (d *Decoder) readObjects(count uint16, startPos int64) (objects []*Object, err error) {
	objects = make([]*Object, count)
	pos := startPos

	for i := uint16(0); i < count; i++ {
		objects[i], pos, err = d.readObject(pos)
		if err != nil {
			return nil, fmt.Errorf("could not read object %d: %w", i, err)
		}
	}

	return objects, nil
}

func (d *Decoder) readObject(startPos int64) (object *Object, pos int64, err error) {
	pos = startPos

	buf := make([]byte, objectHeaderSize)
	n, err := d.r.ReadAt(buf, pos)
	pos += int64(n)
	if n != objectHeaderSize {
		return nil, pos, fmt.Errorf("read %d byte(s), expected %d", n, objectHeaderSize)
	}
	if err != nil && err != io.EOF {
		return nil, pos, err
	}

	pivot, err := d.readVector(buf[36:48])
	if err != nil {
		return nil, pos, fmt.Errorf("could not read pivot vector: %w", err)
	}

	vertextCount := binary.LittleEndian.Uint16(buf[48:50])
	faceCount := binary.LittleEndian.Uint16(buf[50:52])

	faces := make([]*Face, faceCount)
	for i := uint16(0); i < faceCount; i++ {
		faces[i], pos, err = d.readFace(pos)
		if err != nil {
			return nil, pos, fmt.Errorf("could not read face %d: %w", i, err)
		}
	}

	vertexes := make([]*Vertex, vertextCount)
	for i := uint16(0); i < vertextCount; i++ {
		vertexes[i], pos, err = d.readVertex(pos)
		if err != nil {
			return nil, pos, fmt.Errorf("could not read vertex %d: %w", i, err)
		}
	}

	return &Object{
		Name:        cstringutil.ToGo(buf[:32]),
		ParentIndex: int16(binary.LittleEndian.Uint16(buf[32:34])),
		padding:     int16(binary.LittleEndian.Uint16(buf[34:36])),
		Pivot:       pivot,
		Flags:       binary.LittleEndian.Uint32(buf[52:56]),
		unknown1:    binary.LittleEndian.Uint32(buf[56:60]),
		unknown2:    binary.LittleEndian.Uint32(buf[60:64]),
		Faces:       faces,
		Vertexes:    vertexes,
	}, pos, nil
}

func (d *Decoder) readFace(startPos int64) (face *Face, pos int64, err error) {
	pos = startPos

	buf := make([]byte, objectFaceSize)
	n, err := d.r.ReadAt(buf, pos)
	pos += int64(n)
	if n != objectFaceSize {
		return nil, pos, fmt.Errorf("read %d byte(s), expected %d", n, objectFaceSize)
	}
	if err != nil && err != io.EOF {
		return nil, pos, err
	}

	normal, err := d.readVector(buf[8:20])
	if err != nil {
		return nil, pos, fmt.Errorf("could not read normal vector: %w", err)
	}

	return &Face{
		Indexes: [3]uint16{
			binary.LittleEndian.Uint16(buf[0:2]),
			binary.LittleEndian.Uint16(buf[2:4]),
			binary.LittleEndian.Uint16(buf[4:6]),
		},
		TextureIndex: binary.LittleEndian.Uint16(buf[6:8]),
		Normal:       normal,
		unknown1:     binary.LittleEndian.Uint32(buf[20:24]),
		unknown2:     binary.LittleEndian.Uint32(buf[24:28]),
	}, pos, nil
}

func (d *Decoder) readVertex(startPos int64) (vertex *Vertex, pos int64, err error) {
	pos = startPos

	buf := make([]byte, objectVertexSize)
	n, err := d.r.ReadAt(buf, pos)
	pos += int64(n)
	if n != objectVertexSize {
		return nil, pos, fmt.Errorf("read %d byte(s), expected %d", n, objectVertexSize)
	}
	if err != nil && err != io.EOF {
		return nil, pos, err
	}

	position, err := d.readVector(buf[0:12])
	if err != nil {
		return nil, pos, fmt.Errorf("could not read position vector: %w", err)
	}

	normal, err := d.readVector(buf[12:24])
	if err != nil {
		return nil, pos, fmt.Errorf("could not read normal vector: %w", err)
	}

	var u float32
	if err := binary.Read(bytes.NewReader(buf[28:32]), binary.LittleEndian, &u); err != nil {
		return nil, pos, fmt.Errorf("could not read u: %w", err)
	}
	var v float32
	if err := binary.Read(bytes.NewReader(buf[32:36]), binary.LittleEndian, &v); err != nil {
		return nil, pos, fmt.Errorf("could not read v: %w", err)
	}

	return &Vertex{
		Position: position,
		Normal:   normal,
		Color: Color{
			R: buf[24],
			G: buf[25],
			B: buf[26],
			A: buf[27],
		},
		U:        u,
		V:        v,
		Index:    binary.LittleEndian.Uint32(buf[36:40]),
		unknown1: binary.LittleEndian.Uint32(buf[40:44]),
	}, pos, nil
}

func (d *Decoder) readVector(buf []byte) (Vector, error) {
	var v Vector
	if err := binary.Read(bytes.NewReader(buf[0:4]), binary.LittleEndian, &v.X); err != nil {
		return v, fmt.Errorf("could not read x: %w", err)
	}
	if err := binary.Read(bytes.NewReader(buf[4:8]), binary.LittleEndian, &v.Y); err != nil {
		return v, fmt.Errorf("could not read y: %w", err)
	}
	if err := binary.Read(bytes.NewReader(buf[8:12]), binary.LittleEndian, &v.Z); err != nil {
		return v, fmt.Errorf("could not read z: %w", err)
	}
	return v, nil
}
