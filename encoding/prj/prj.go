// Package prj implements decoding of Dark Omen's .PRJ project files.
package prj

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/jonathaningram/dark-omen/internal/cstringutil"
)

const (
	// format is the format ID used in all .PRJ files (trailing spaces intended).
	format = "Dark Omen Battle file 1.10      "

	headerSize               = 32
	blockHeaderSize          = 8
	furnitureBlockHeaderSize = blockHeaderSize + 4
	instancesBlockHeaderSize = blockHeaderSize + 4 + 4
	terrainBlockHeaderSize   = blockHeaderSize + 20

	baseID       = "BASE"
	waterID      = "WATR"
	furnitureID  = "FURN"
	instancesID  = "INST"
	terrainID    = "TERR"
	attributesID = "ATTR"
)

// A Project is made up of 10 specialized blocks of information that describe
// a battle.
type Project struct {
	format     string
	Base       *Base
	Water      *Water
	Furniture  *Furniture
	Instances  []*Instance
	Terrain    *Terrain
	Attributes *Attributes
}

type Base struct {
	ModelFileName string
}

type Water struct {
	ModelFileName string
}

type Furniture struct {
	FileNames []string
}

// A Vector in 3-dimensional space.
type Vector struct {
	X, Y, Z float32
}

type Instance struct {
	prev                     int32
	next                     int32
	Selected                 int32
	ExcludeFromTerrain       int32
	Position                 Vector
	Rotation                 Vector
	Min                      Vector
	Max                      Vector
	MeshSlot                 int32
	MeshID                   int32
	Attackable               int32
	Toughness                int32
	Wounds                   int32
	unknown1                 int32
	OwnerUnitIndex           int32
	Burnable                 int32
	SFXCode                  int32
	GFXCode                  int32
	Locked                   int32
	ExcludeFromTerrainShadow int32
	ExcludeFromWalk          int32
	MagicItemCode            int32
	ParticleEffectCode       int32
	DeadMeshSlot             int32
	DeadMeshID               int32
	Light                    int32
	LightRadius              int32
	LightAmbient             int32
	unknown2                 int32
	unknown3                 int32
}

type Terrain struct {
	Width  uint32
	Height uint32
	// Heightmap1Blocks is a list of large blocks for the first heightmap.
	Heightmap1Blocks []*TerrainBlock
	// Heightmap2Blocks is a list of large blocks for the second heightmap.
	Heightmap2Blocks []*TerrainBlock
	// Offsets is a list of offsets for 8x8 block. Height offset for each block
	// based on minimum height.
	Offsets [][]byte
}

type TerrainBlock struct {
	Minimum     uint32
	OffsetIndex uint32
}

type Attributes struct {
	MapWidth  uint32
	MapHeight uint32
}

// Decoder reads and decodes a project from an input stream.
type Decoder struct {
	r io.Reader
}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

// Decode reads the encoded project information from its input and returns a
// new project containing decoded blocks.
func (d *Decoder) Decode() (*Project, error) {
	header, err := d.readHeader()
	if err != nil {
		return nil, fmt.Errorf("could not read header: %w", err)
	}

	if f := header.format; f != format {
		return nil, fmt.Errorf("unknown format %q, expected %q", f, format)
	}

	base, err := d.readBase()
	if err != nil {
		return nil, fmt.Errorf("could not read base: %w", err)
	}
	water, err := d.readWater()
	if err != nil {
		return nil, fmt.Errorf("could not read water: %w", err)
	}
	furniture, err := d.readFurniture()
	if err != nil {
		return nil, fmt.Errorf("could not read furniture: %w", err)
	}
	instances, err := d.readInstances(furniture)
	if err != nil {
		return nil, fmt.Errorf("could not read instances: %w", err)
	}
	terrain, err := d.readTerrain()
	if err != nil {
		return nil, fmt.Errorf("could not read terrain: %w", err)
	}
	attributes, err := d.readAttributes()
	if err != nil {
		return nil, fmt.Errorf("could not read attributes: %w", err)
	}

	return &Project{
		format:     format,
		Base:       base,
		Water:      water,
		Furniture:  furniture,
		Instances:  instances,
		Terrain:    terrain,
		Attributes: attributes,
	}, nil
}

type header struct {
	format string
}

func (d *Decoder) readHeader() (h *header, err error) {
	buf := make([]byte, headerSize)
	n, err := d.r.Read(buf)
	if n != headerSize {
		return nil, fmt.Errorf("read %d byte(s), expected %d", n, headerSize)
	}
	if err != nil {
		return nil, err
	}

	return &header{
		format: string(buf[0:]),
	}, nil
}

func (d *Decoder) readBlock(id string) (data []byte, err error) {
	buf := make([]byte, blockHeaderSize)
	n, err := d.r.Read(buf)
	if n != blockHeaderSize {
		return nil, fmt.Errorf("header read %d byte(s), expected %d", n, blockHeaderSize)
	}
	if err != nil {
		return nil, err
	}
	if f := string(buf[0:4]); f != id {
		return nil, fmt.Errorf("unexpected ID %q, expected %q", f, id)
	}
	size := int(binary.LittleEndian.Uint32(buf[4:]))
	buf = make([]byte, size)
	n, err = d.r.Read(buf)
	if n != size {
		return nil, fmt.Errorf("data read %d byte(s), expected %d", n, size)
	}
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func (d *Decoder) readBase() (block *Base, err error) {
	data, err := d.readBlock(baseID)
	if err != nil {
		return nil, fmt.Errorf("could not read base block: %w", err)
	}
	return &Base{
		ModelFileName: cstringutil.ToGo(data),
	}, nil
}

func (d *Decoder) readWater() (block *Water, err error) {
	data, err := d.readBlock(waterID)
	if err != nil {
		return nil, fmt.Errorf("could not read water block: %w", err)
	}
	return &Water{
		ModelFileName: cstringutil.ToGo(data),
	}, nil
}

func (d *Decoder) readFurniture() (block *Furniture, err error) {
	buf := make([]byte, furnitureBlockHeaderSize)
	n, err := d.r.Read(buf)
	if n != furnitureBlockHeaderSize {
		return nil, fmt.Errorf("header read %d byte(s), expected %d", n, furnitureBlockHeaderSize)
	}
	if err != nil {
		return nil, err
	}
	if f := string(buf[0:4]); f != furnitureID {
		return nil, fmt.Errorf("unexpected ID %q, expected %q", f, furnitureID)
	}
	count := int(binary.LittleEndian.Uint32(buf[8:]))
	size := 4*count + int(binary.LittleEndian.Uint32(buf[4:8])) - 4
	buf = make([]byte, size)
	n, err = d.r.Read(buf)
	if n != size {
		return nil, fmt.Errorf("data read %d byte(s), expected %d", n, size)
	}
	if err != nil {
		return nil, err
	}

	var pos int
	fileNames := make([]string, count)
	for i := 0; i < count; i++ {
		size := int(binary.LittleEndian.Uint32(buf[pos : pos+4]))
		fileNames[i] = cstringutil.ToGo(buf[pos+4 : pos+4+size])
		pos += 4 + size
	}

	return &Furniture{FileNames: fileNames}, nil
}

func (d *Decoder) readInstances(furniture *Furniture) (instances []*Instance, err error) {
	buf := make([]byte, instancesBlockHeaderSize)
	n, err := d.r.Read(buf)
	if n != instancesBlockHeaderSize {
		return nil, fmt.Errorf("header read %d byte(s), expected %d", n, instancesBlockHeaderSize)
	}
	if err != nil {
		return nil, err
	}
	if f := string(buf[0:4]); f != instancesID {
		return nil, fmt.Errorf("unexpected ID %q, expected %q", f, instancesID)
	}
	size := int(binary.LittleEndian.Uint32(buf[4:8]))
	count := int(binary.LittleEndian.Uint32(buf[8:12]))
	instanceSize := int(binary.LittleEndian.Uint32(buf[12:]))
	buf = make([]byte, size)
	n, err = d.r.Read(buf)
	if n != size {
		return nil, fmt.Errorf("data read %d byte(s), expected %d", n, size)
	}
	if err != nil {
		return nil, err
	}

	instances = make([]*Instance, count)
	for i := 0; i < count; i++ {
		b := buf[i*instanceSize : (i+1)*instanceSize]

		instance := &Instance{
			prev:               int32(binary.LittleEndian.Uint32(b[0:4])),
			next:               int32(binary.LittleEndian.Uint32(b[4:8])),
			Selected:           int32(binary.LittleEndian.Uint32(b[8:12])),
			ExcludeFromTerrain: int32(binary.LittleEndian.Uint32(b[12:16])),
			Position: Vector{
				X: float32(int32(binary.LittleEndian.Uint32(b[16:20]))) / 1024,
				Y: float32(int32(binary.LittleEndian.Uint32(b[20:24]))) / 1024,
				Z: float32(int32(binary.LittleEndian.Uint32(b[24:28]))) / 1024,
			},
			Rotation: Vector{
				X: float32(int32(binary.LittleEndian.Uint32(b[28:32]))) / 4096,
				Y: float32(int32(binary.LittleEndian.Uint32(b[32:36]))) / 4096,
				Z: float32(int32(binary.LittleEndian.Uint32(b[36:40]))) / 4096,
			},
			Min: Vector{
				X: float32(int32(binary.LittleEndian.Uint32(b[40:44]))) / 1024,
				Y: float32(int32(binary.LittleEndian.Uint32(b[44:48]))) / 1024,
				Z: float32(int32(binary.LittleEndian.Uint32(b[48:52]))) / 1024,
			},
			Max: Vector{
				X: float32(int32(binary.LittleEndian.Uint32(b[52:56]))) / 1024,
				Y: float32(int32(binary.LittleEndian.Uint32(b[56:60]))) / 1024,
				Z: float32(int32(binary.LittleEndian.Uint32(b[60:64]))) / 1024,
			},
			MeshSlot:                 int32(binary.LittleEndian.Uint32(b[64:68])),
			MeshID:                   int32(binary.LittleEndian.Uint32(b[68:72])),
			Attackable:               int32(binary.LittleEndian.Uint32(b[72:76])),
			Toughness:                int32(binary.LittleEndian.Uint32(b[76:80])),
			Wounds:                   int32(binary.LittleEndian.Uint32(b[80:84])),
			unknown1:                 int32(binary.LittleEndian.Uint32(b[84:88])),
			OwnerUnitIndex:           int32(binary.LittleEndian.Uint32(b[88:92])),
			Burnable:                 int32(binary.LittleEndian.Uint32(b[92:96])),
			SFXCode:                  int32(binary.LittleEndian.Uint32(b[96:100])),
			GFXCode:                  int32(binary.LittleEndian.Uint32(b[100:104])),
			Locked:                   int32(binary.LittleEndian.Uint32(b[104:108])),
			ExcludeFromTerrainShadow: int32(binary.LittleEndian.Uint32(b[108:112])),
			ExcludeFromWalk:          int32(binary.LittleEndian.Uint32(b[112:116])),
			MagicItemCode:            int32(binary.LittleEndian.Uint32(b[116:120])),
			ParticleEffectCode:       int32(binary.LittleEndian.Uint32(b[120:124])),
			DeadMeshSlot:             int32(binary.LittleEndian.Uint32(b[124:128])),
			DeadMeshID:               int32(binary.LittleEndian.Uint32(b[128:132])),
			Light:                    int32(binary.LittleEndian.Uint32(b[132:136])),
			LightRadius:              int32(binary.LittleEndian.Uint32(b[136:140])),
			LightAmbient:             int32(binary.LittleEndian.Uint32(b[140:144])),
			unknown2:                 int32(binary.LittleEndian.Uint32(b[144:148])),
			unknown3:                 int32(binary.LittleEndian.Uint32(b[148:152])),
		}

		instances[i] = instance
	}

	return instances, nil
}

func (d *Decoder) readTerrain() (block *Terrain, err error) {
	buf := make([]byte, terrainBlockHeaderSize)
	n, err := d.r.Read(buf)
	if n != terrainBlockHeaderSize {
		return nil, fmt.Errorf("header read %d byte(s), expected %d", n, terrainBlockHeaderSize)
	}
	if err != nil {
		return nil, err
	}
	if f := string(buf[0:4]); f != terrainID {
		return nil, fmt.Errorf("unexpected ID %q, expected %q", f, terrainID)
	}

	_ = binary.LittleEndian.Uint32(buf[4:8]) // size, not used
	width := binary.LittleEndian.Uint32(buf[8:12])
	height := binary.LittleEndian.Uint32(buf[12:16])
	compressedBlockCount := binary.LittleEndian.Uint32(buf[16:20])
	uncompressedBlockCount := binary.LittleEndian.Uint32(buf[20:24])
	mapBlockSize := binary.LittleEndian.Uint32(buf[24:28])

	// First heightmap.
	buf = make([]byte, mapBlockSize/2)
	n, err = d.r.Read(buf)
	if n != int(mapBlockSize)/2 {
		return nil, fmt.Errorf("heightmap 1 data read %d byte(s), expected %d", n, mapBlockSize/2)
	}
	if err != nil {
		return nil, err
	}
	heightmap1Blocks := make([]*TerrainBlock, uncompressedBlockCount)
	for i := uint32(0); i < uncompressedBlockCount; i++ {
		minimum := binary.LittleEndian.Uint32(buf[0:4])
		offsetIndex := binary.LittleEndian.Uint32(buf[4:8])
		if offsetIndex&64 != 0 {
			return nil, fmt.Errorf("heightmap 1: offset index is not a multiple of 64, got %v", offsetIndex)
		}
		offsetIndex /= 64
		heightmap1Blocks[i] = &TerrainBlock{Minimum: minimum, OffsetIndex: offsetIndex}
	}

	// Second heightmap.
	buf = make([]byte, mapBlockSize/2)
	n, err = d.r.Read(buf)
	if n != int(mapBlockSize)/2 {
		return nil, fmt.Errorf("heightmap 2 data read %d byte(s), expected %d", n, mapBlockSize/2)
	}
	if err != nil {
		return nil, err
	}
	heightmap2Blocks := make([]*TerrainBlock, uncompressedBlockCount)
	for i := uint32(0); i < uncompressedBlockCount; i++ {
		minimum := binary.LittleEndian.Uint32(buf[0:4])
		offsetIndex := binary.LittleEndian.Uint32(buf[4:8])
		if offsetIndex&64 != 0 {
			return nil, fmt.Errorf("heightmap 2: offset index is not a multiple of 64, got %v", offsetIndex)
		}
		offsetIndex /= 64
		heightmap2Blocks[i] = &TerrainBlock{Minimum: minimum, OffsetIndex: offsetIndex}
	}

	// Read offsets.
	buf = make([]byte, 4)
	n, err = d.r.Read(buf)
	if n != 4 {
		return nil, fmt.Errorf("offset count read %d byte(s), expected %d", n, 4)
	}
	if err != nil {
		return nil, err
	}
	offsetCount := binary.LittleEndian.Uint32(buf[:])

	if compressedBlockCount*64 != offsetCount {
		return nil, fmt.Errorf("compressed block count and offset count mismatch: got %v, %v", compressedBlockCount, offsetCount)
	}

	buf = make([]byte, offsetCount)
	n, err = d.r.Read(buf)
	if n != int(offsetCount) {
		return nil, fmt.Errorf("offset data read %d byte(s), expected %d", n, offsetCount)
	}
	if err != nil {
		return nil, err
	}

	offsets := make([][]byte, compressedBlockCount)
	for i := uint32(0); i < compressedBlockCount; i++ {
		offsets[i] = buf[i*64 : (i+1)*64]
	}

	return &Terrain{
		Width:            width,
		Height:           height,
		Heightmap1Blocks: heightmap1Blocks,
		Heightmap2Blocks: heightmap2Blocks,
		Offsets:          offsets,
	}, nil
}

func (d *Decoder) readAttributes() (block *Attributes, err error) {
	data, err := d.readBlock(attributesID)
	if err != nil {
		return nil, err
	}
	mapWidth := binary.LittleEndian.Uint32(data[0:4])
	mapHeight := binary.LittleEndian.Uint32(data[4:8])

	return &Attributes{
		MapWidth:  mapWidth,
		MapHeight: mapHeight,
	}, nil
}
