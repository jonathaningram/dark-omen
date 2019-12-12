package spr

import (
	"bufio"
	"io"
)

func zeroRuns(r io.Reader) ([]byte, error) {
	buf := make([]byte, 128)
	dst := make([]byte, 0, 1024)
	br, ok := r.(byteReader)
	if !ok {
		br = bufio.NewReader(r)
	}

	for {
		b, err := br.ReadByte()
		if err != nil {
			if err == io.EOF {
				return dst, nil
			}
			return nil, err
		}
		code := int(int8(b))
		switch {
		case code >= 0:
			n, err := io.ReadFull(br, buf[:code+1])
			if err != nil {
				return nil, err
			}
			dst = append(dst, buf[:n]...)
		default:
			for j := 0; j < -code; j++ {
				dst = append(dst, byte(0))
			}
		}
	}
}
