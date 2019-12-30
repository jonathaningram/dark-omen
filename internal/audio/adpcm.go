package audio

import "math"

var indexTable = [16]int16{
	-1, -1, -1, -1, 2, 4, 6, 8,
	-1, -1, -1, -1, 2, 4, 6, 8,
}

var stepTable = [89]int16{
	7, 8, 9, 10, 11, 12, 13, 14, 16, 17,
	19, 21, 23, 25, 28, 31, 34, 37, 41, 45,
	50, 55, 60, 66, 73, 80, 88, 97, 107, 118,
	130, 143, 157, 173, 190, 209, 230, 253, 279, 307,
	337, 371, 408, 449, 494, 544, 598, 658, 724, 796,
	876, 963, 1060, 1166, 1282, 1411, 1552, 1707, 1878, 2066,
	2272, 2499, 2749, 3024, 3327, 3660, 4026, 4428, 4871, 5358,
	5894, 6484, 7132, 7845, 8630, 9493, 10442, 11487, 12635, 13899,
	15289, 16818, 18500, 20350, 22385, 24623, 27086, 29794, 32767,
}

type ADPCMBlock struct {
	sample int16
	index  int16
	data   []byte
}

var _ Block = (*ADPCMBlock)(nil)

func NewADPCMBlock(sample, index int16, data []byte) *ADPCMBlock {
	return &ADPCMBlock{
		sample: sample,
		index:  index,
		data:   data,
	}
}

func (b *ADPCMBlock) Sample() int16          { return b.sample }
func (b *ADPCMBlock) Index() int16           { return b.index }
func (b *ADPCMBlock) Bytes() ([]byte, error) { return b.data, nil }

func (b *ADPCMBlock) AsPCM16Block() *PCM16Block {
	d := &blockDecoder{sample: b.sample, index: b.index}

	var data []int16

	for _, byt := range b.data {
		data = append(data, d.decode(byt&0x0f))
		data = append(data, d.decode(byt>>4))
	}

	return &PCM16Block{Data: data}
}

type blockDecoder struct {
	sample int16
	index  int16
}

func (d *blockDecoder) decode(nibble byte) int16 {
	step := stepTable[d.index]

	diff := int16(0)
	if nibble&4 != 0 {
		diff += step
	}
	if nibble&2 != 0 {
		diff += step >> 1
	}
	if nibble&1 != 0 {
		diff += step >> 2
	}
	diff += step >> 3

	if nibble&8 != 0 {
		diff = -diff
	}

	newSample := d.sample + diff
	if newSample > math.MaxInt16 {
		newSample = math.MaxInt16
	} else if newSample < math.MinInt16 {
		newSample = math.MinInt16
	}
	d.sample = newSample

	index := d.index + indexTable[nibble]
	if index < 0 {
		index = 0
	} else if index >= int16(len(stepTable)) {
		index = int16(len(stepTable)) - 1
	}
	d.index = index

	return newSample
}
