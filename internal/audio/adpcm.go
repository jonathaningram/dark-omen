package audio

import (
	"math"
)

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
	d := &blockDecoder{sample: int(b.sample), index: b.index}

	var data []int16

	for _, byt := range b.data {
		data = append(data, d.decode(byt&0x0f))
		data = append(data, d.decode(byt>>4))
	}

	return &PCM16Block{Data: data}
}

type blockDecoder struct {
	sample int
	index  int16
}

// originalSample is a 4-bit ADPCM sample.
//
// The return value newSample is the resulting 16-bit two's complement variable.
//
// See https://www.cs.columbia.edu/~hgs/audio/dvi/IMA_ADPCM.pdf at page 32 for
// the algorithm and for example input. Note: The example input does appear to
// be wrong though because `if (0x8763 > 32767) == FALSE` is actually true.
func (d *blockDecoder) decode(originalSample byte) int16 {
	// Find quantizer step size.
	stepSize := int(stepTable[d.index])

	// Calculate difference:
	//
	//   diff = (originalSample + 1/2) * stepSize/4
	//
	// Perform multiplication through repetitive addition.
	var diff int
	if originalSample&4 != 0 {
		diff += stepSize
	}
	if originalSample&2 != 0 {
		diff += stepSize >> 1
	}
	if originalSample&1 != 0 {
		diff += stepSize >> 2
	}
	diff += stepSize >> 3

	// Account for sign bit.
	if originalSample&8 != 0 {
		diff = -diff
	}

	// Adjust predicted sample based on calculated difference.
	newSample := d.sample + diff
	if newSample > math.MaxInt16 { // check for overflow
		newSample = math.MaxInt16
	} else if newSample < math.MinInt16 {
		newSample = math.MinInt16
	}

	// Store 16-bit new sample.
	d.sample = newSample

	// Adjust index into step size lookup table using original sample.
	index := d.index + indexTable[originalSample]
	if index < 0 { // check for index underflow
		index = 0
	} else if index > 88 { // check for index overflow
		index = 88
	}
	d.index = index

	// Value has been clamped, can now convert to int16.
	return int16(newSample)
}
