// Package engrel provides functions that can access information found in Dark
// Omen's PRG_ENG/ENGREL.exe executable file.
package engrel

import (
	"fmt"
	"io"

	"github.com/jonathaningram/dark-omen/internal/cstringutils"
)

const (
	spriteNamesStartOffset = 0x000CCB54
	spriteCount            = 252
	spriteNameSize         = 44
)

// ReadSpriteNames reads all sprite names from r and returns a slice that can
// be indexed to return a sprite name at a particular position.
// The given reader should contain the contents of the PRG_ENG/ENGREL.EXE file
// found in Dark Omen.
func ReadSpriteNames(r io.ReaderAt) ([]string, error) {
	names := make([]string, spriteCount)

	for index := int64(0); index < spriteCount; index++ {
		name, err := readSpriteName(r, index)
		if err != nil {
			return nil, fmt.Errorf("could not read sprite name at index %d: %w", index, err)
		}
		names[index] = name
	}

	return names, nil
}

// ReadSpriteName reads the sprite name at the given index from r.
// The given reader should contain the contents of the PRG_ENG/ENGREL.EXE file
// found in Dark Omen.
func ReadSpriteName(r io.ReaderAt, index int64) (string, error) {
	return readSpriteName(r, index)
}

func readSpriteName(r io.ReaderAt, index int64) (string, error) {
	if index > spriteCount-1 {
		return "", fmt.Errorf("expected sprite name index to be less than %d, got %d", spriteCount, index)
	}
	buf := make([]byte, spriteNameSize)
	n, err := r.ReadAt(buf, spriteNamesStartOffset+spriteNameSize*index)
	if n != spriteNameSize {
		return "", fmt.Errorf("read %d byte(s), expected to read %d", n, spriteNameSize)
	}
	if err != nil {
		if err == io.EOF {
			return "", io.ErrUnexpectedEOF
		}
		return "", err
	}
	return cstringutils.ToGo(buf), nil
}
