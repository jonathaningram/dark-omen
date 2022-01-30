package sad

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/jonathaningram/dark-omen/internal/audio"
)

type Stream struct {
	LeftBlocks  []audio.Block
	RightBlocks []audio.Block

	// Note: storing these so that a re-encoded stream is correct. sample 99 is
	// always a different value and index 99 is always equal to 99. Not sure if
	// sample 99 needs to end up in the decoded stream—currently it does not.

	leftSample99  int16
	leftIndex99   int16
	rightSample99 int16
	rightIndex99  int16
}

func (s *Stream) Channels() int {
	return 2
}

// Decoder reads and decodes a SAD audio stream from an input stream.
type Decoder struct {
	r io.Reader
}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

// Decode reads the encoded SAD audio stream from its input and returns a
// new stream containing decoded left and right blocks.
func (d *Decoder) Decode() (*Stream, error) {
	s := &Stream{
		LeftBlocks:  make([]audio.Block, 0),
		RightBlocks: make([]audio.Block, 0),
	}

	for {
		var bs [8]byte
		n, err := d.r.Read(bs[:])
		if n != 8 {
			return nil, fmt.Errorf("could not read stereo sample and index data: read %d byte(s), expected %d", n, 8)
		}
		if err != nil {
			return nil, err
		}
		leftSample := int16(binary.LittleEndian.Uint16(bs[0:2]))
		leftIndex := int16(binary.LittleEndian.Uint16(bs[2:4]))
		rightSample := int16(binary.LittleEndian.Uint16(bs[4:6]))
		rightIndex := int16(binary.LittleEndian.Uint16(bs[6:8]))

		if leftIndex == 99 && rightIndex == 99 {
			s.leftSample99 = leftSample
			s.leftIndex99 = leftIndex
			s.rightSample99 = rightSample
			s.rightIndex99 = rightIndex
			break
		}

		const size = 1016
		buf := make([]byte, size)
		n, err = d.r.Read(buf)
		if n != size {
			return nil, fmt.Errorf("could not read stereo ADPCM data: read %d byte(s), expected %d", n, size)
		}
		if err != nil {
			return nil, err
		}

		leftData := make([]byte, size/2)
		rightData := make([]byte, size/2)

		for i := 0; i < size/8; i++ {
			for j := 0; j < 4; j++ {
				leftData[i*4+j] = buf[i*8+j]
			}
			for j := 4; j < 8; j++ {
				rightData[i*4+j-4] = buf[i*8+j]
			}
		}

		s.LeftBlocks = append(s.LeftBlocks, audio.NewADPCMBlock(leftSample, leftIndex, leftData))
		s.RightBlocks = append(s.RightBlocks, audio.NewADPCMBlock(rightSample, rightIndex, rightData))
	}

	// Read remaining bytes.
	buf, err := io.ReadAll(d.r)
	if err != nil {
		return nil, err
	}

	leftBuf := make([]int16, len(buf)/4)
	rightBuf := make([]int16, len(buf)/4)

	for i := 0; i < len(buf)/4; i++ {
		leftSample := int16(binary.LittleEndian.Uint16([]byte{buf[i*4], buf[i*4+1]}))
		rightSample := int16(binary.LittleEndian.Uint16([]byte{buf[i*4+2], buf[i*4+3]}))
		leftBuf[i] = leftSample
		rightBuf[i] = rightSample
	}

	s.LeftBlocks = append(s.LeftBlocks, audio.NewPCM16BlockFromInt16Slice(leftBuf))
	s.RightBlocks = append(s.RightBlocks, audio.NewPCM16BlockFromInt16Slice(rightBuf))

	return s, nil
}

// Encoder encodes and writes a SAD audio stream to an output stream.
type Encoder struct {
	w io.Writer
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

// Encode writes the encoded SAD audio stream to its output.
func (e *Encoder) Encode(s *Stream) error {
	for i := 0; i < len(s.LeftBlocks)-1; i++ {
		leftBlock, ok := s.LeftBlocks[i].(*audio.ADPCMBlock)
		if !ok {
			return fmt.Errorf("left block at position %d is not an ADPCM block", i)
		}
		if err := binary.Write(e.w, binary.LittleEndian, leftBlock.Sample()); err != nil {
			return err
		}
		if err := binary.Write(e.w, binary.LittleEndian, leftBlock.Index()); err != nil {
			return err
		}
		rightBlock, ok := s.RightBlocks[i].(*audio.ADPCMBlock)
		if !ok {
			return fmt.Errorf("right block at position %d is not an ADPCM block", i)
		}
		if err := binary.Write(e.w, binary.LittleEndian, rightBlock.Sample()); err != nil {
			return err
		}
		if err := binary.Write(e.w, binary.LittleEndian, rightBlock.Index()); err != nil {
			return err
		}

		leftData, err := leftBlock.Bytes()
		if err != nil {
			return err
		}
		rightData, err := rightBlock.Bytes()
		if err != nil {
			return err
		}

		for j := 0; j < len(leftData); j += 4 {
			for k := 0; k < 4; k++ {
				n, err := e.w.Write([]byte{leftData[j+k]})
				if n != 1 {
					return fmt.Errorf("wrote %d byte(s), expected %d", n, 1)
				}
				if err != nil {
					return err
				}
			}
			for k := 0; k < 4; k++ {
				n, err := e.w.Write([]byte{rightData[j+k]})
				if n != 1 {
					return fmt.Errorf("wrote %d byte(s), expected %d", n, 1)
				}
				if err != nil {
					return err
				}
			}
		}
	}

	if err := binary.Write(e.w, binary.LittleEndian, s.leftSample99); err != nil {
		return err
	}
	if err := binary.Write(e.w, binary.LittleEndian, s.leftIndex99); err != nil {
		return err
	}
	if err := binary.Write(e.w, binary.LittleEndian, s.rightSample99); err != nil {
		return err
	}
	if err := binary.Write(e.w, binary.LittleEndian, s.rightIndex99); err != nil {
		return err
	}

	leftBytes, err := s.LeftBlocks[len(s.LeftBlocks)-1].Bytes()
	if err != nil {
		return err
	}
	rightBytes, err := s.RightBlocks[len(s.LeftBlocks)-1].Bytes()
	if err != nil {
		return err
	}
	for i := 0; i < len(leftBytes); i += 2 {
		if err := binary.Write(e.w, binary.LittleEndian, []byte{leftBytes[i], leftBytes[i+1]}); err != nil {
			return err
		}
		if err := binary.Write(e.w, binary.LittleEndian, []byte{rightBytes[i], rightBytes[i+1]}); err != nil {
			return err
		}
	}

	return nil
}
