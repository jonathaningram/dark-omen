package engrel

import (
	"fmt"
	"io"
)

const (
	magicItemNamesStartOffset = 0xDB374
	magicItemCount            = 64
)

// ReadMagicItemNames reads all magic item names from r and returns a slice that
// can be indexed to return a magic item name at a particular position.
// The given reader should contain the contents of the PRG_ENG/ENGREL.EXE file
// found in Dark Omen.
func ReadMagicItemNames(r io.ByteReader) ([]string, error) {
	return readMagicItemNames(r)
}

// ReadMagicItemName reads the magic item name at the given index from r.
// The given reader should contain the contents of the PRG_ENG/ENGREL.EXE file
// found in Dark Omen.
func ReadMagicItemName(r io.ByteReader, index int64) (string, error) {
	if index > magicItemCount-1 {
		return "", fmt.Errorf("expected index to be less than %d, got %d", magicItemCount, index)
	}
	items, err := readMagicItemNames(r)
	if err != nil {
		return "", err
	}
	return items[index], nil
}

func readMagicItemNames(r io.ByteReader) ([]string, error) {
	names := make([]string, magicItemCount)

	var n int
	var i int
	var bs []byte
	for {
		b, err := r.ReadByte()
		if err != nil {
			if err == io.EOF {
				return nil, io.ErrUnexpectedEOF
			}
			return nil, err
		}
		n++
		// Hacky way of reading and ignoring all of the previous bytes until the
		// offset of interest.
		if n <= magicItemNamesStartOffset {
			continue
		}
		if b == '\x00' {
			s := string(bs)
			if s == "" {
				bs = nil
				continue
			}
			names[len(names)-1-i] = s
			i++
			if i == magicItemCount {
				return names, nil
			}
			bs = nil
			continue
		}
		bs = append(bs, b)
	}
}
