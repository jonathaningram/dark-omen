package spr

import (
	"bufio"
	"errors"
	"io"
)

type byteReader interface {
	io.Reader
	io.ByteReader
}

// unpackBits decodes the PackBits-compressed data in r and returns the
// uncompressed data.
// Copied off https://github.com/golang/image/blob/da761ea9ff43b0defcf66e8784f2aa4faa517dde/tiff/compress.go#L22
func unpackBits(r io.Reader) ([]byte, error) {
	buf := make([]byte, 128)
	dst := make([]byte, 0, 1024)
	br, ok := r.(byteReader)
	if !ok {
		br = bufio.NewReader(r)
	}

	for {
		b, err := br.ReadByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
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
		case code == -128:
			// No-op.
		default:
			if b, err = br.ReadByte(); err != nil {
				return nil, err
			}
			for j := 0; j < 1-code; j++ {
				buf[j] = b
			}
			dst = append(dst, buf[:1-code]...)
		}
	}
}
