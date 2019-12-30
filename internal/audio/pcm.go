package audio

import (
	"bytes"
	"encoding/binary"
)

type PCM16Block struct {
	Data []int16
}

var _ Block = (*PCM16Block)(nil)

func NewPCM16Block(bs []byte) (*PCM16Block, error) {
	data := make([]int16, len(bs)/2)

	buf := bytes.NewBuffer(bs)

	for i := 0; i < len(data); i++ {
		var sample int16
		if err := binary.Read(buf, binary.LittleEndian, &sample); err != nil {
			return nil, err
		}
		data[i] = sample
	}

	return &PCM16Block{Data: data}, nil
}

func NewPCM16BlockFromInt16Slice(data []int16) *PCM16Block {
	return &PCM16Block{Data: data}
}

func (b *PCM16Block) Bytes() ([]byte, error) {
	buf := &bytes.Buffer{}
	for _, v := range b.Data {
		if err := binary.Write(buf, binary.LittleEndian, v); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func (b *PCM16Block) AsPCM16Block() *PCM16Block { return b }
