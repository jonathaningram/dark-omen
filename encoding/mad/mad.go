package mad

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/jonathaningram/dark-omen/internal/audio"
)

type Stream struct {
	Blocks []audio.Block

	// Note: Storing these so that a re-encoded stream is correct. sample 99 is
	// always a different value and index 99 is always equal to 99. Not sure if
	// sample99 needs to end up in the decoded stream—currently it does not.

	sample99 int16
	index99  int16
}

func (s *Stream) Channels() int {
	return 1
}

// Decoder reads and decodes a MAD audio stream from an input stream.
type Decoder struct {
	r io.Reader
}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

// Decode reads the encoded MAD audio stream from its input and returns a
// new stream containing decoded blocks.
func (d *Decoder) Decode() (*Stream, error) {
	s := &Stream{
		Blocks: make([]audio.Block, 0),
	}

	for {
		var bs [4]byte
		if _, err := d.r.Read(bs[:]); err != nil {
			// Some MAD streams don't have a trailing PCM block, so if we
			// encounter EOF, we are done and return the decoded stream.
			if err == io.EOF {
				return s, nil
			}
			return nil, err
		}
		sample := int16(binary.LittleEndian.Uint16(bs[0:2]))
		index := int16(binary.LittleEndian.Uint16(bs[2:4]))

		if index == 99 {
			s.sample99 = sample
			s.index99 = index
			break
		}

		const size = 1020
		monoData := make([]byte, size)
		n, err := d.r.Read(monoData)
		if n != size {
			return nil, fmt.Errorf("could not read mono ADPCM data: read %d byte(s), expected %d", n, size)
		}
		if err != nil && err != io.EOF {
			return nil, err
		}

		s.Blocks = append(s.Blocks, audio.NewADPCMBlock(sample, index, monoData))
	}

	// Read remaining bytes.
	buf, err := io.ReadAll(d.r)
	if err != nil {
		return nil, err
	}
	b, err := audio.NewPCM16Block(buf)
	if err != nil {
		return nil, err
	}
	s.Blocks = append(s.Blocks, b)

	return s, nil
}

// Encoder encodes and writes a MAD audio stream to an output stream.
type Encoder struct {
	w io.Writer
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

// Encode writes the encoded MAD audio stream to its output.
func (e *Encoder) Encode(s *Stream) error {
	for i := 0; i < len(s.Blocks); i++ {
		switch b := s.Blocks[i].(type) {
		case *audio.ADPCMBlock:
			if err := e.encodeADPCMBlock(b); err != nil {
				return err
			}
			// Some MAD streams don't have a trailing PCM block, so if we are at
			// the last block and it was an ADPCM block, we are done encoding.
			if i == len(s.Blocks)-1 {
				return nil
			}
		}
	}

	if err := binary.Write(e.w, binary.LittleEndian, s.sample99); err != nil {
		return err
	}
	if err := binary.Write(e.w, binary.LittleEndian, s.index99); err != nil {
		return err
	}

	if err := e.encodePCM16Block(s.Blocks[len(s.Blocks)-1]); err != nil {
		return err
	}

	return nil
}

func (e *Encoder) encodeADPCMBlock(b *audio.ADPCMBlock) error {
	if err := binary.Write(e.w, binary.LittleEndian, b.Sample()); err != nil {
		return err
	}
	if err := binary.Write(e.w, binary.LittleEndian, b.Index()); err != nil {
		return err
	}
	data, err := b.Bytes()
	if err != nil {
		return err
	}
	_, err = e.w.Write(data)
	return err
}

func (e *Encoder) encodePCM16Block(b audio.Block) error {
	bs, err := b.Bytes()
	if err != nil {
		return err
	}
	_, err = e.w.Write(bs)
	return err
}
