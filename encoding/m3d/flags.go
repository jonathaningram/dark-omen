package m3d

import "strings"

// Flags for an .M3D model.
type Flags uint8

const (
	Translucency Flags = 1 << iota
	UVAnimation
	AlphaTransparency
	_
	ColorKeying
)

// Has returns whether or not the flags contains the provided test.
func (f Flags) Has(test Flags) bool {
	return f&test != 0
}

// ModelFlags returns the flags associated with the given model file name.
// This function is required because flags are embedded in the model's file
// name. If the first character is '_', then the second character contains the
// flags. Otherwise the model has no flags.
// In Dark Omen, only the following flags are used:
// - "_4": 00000100
// - "_6": 00000110
// - "_7": 00000111
// - "_K": 00010100
func ModelFlags(fileName string) Flags {
	// Minimum possible file name is technically _0.M3D, though probably
	// non-existent because it is missing a model name in the file name.
	if len(fileName) < len("_0.M3D") {
		return 0
	}
	// Must be a .M3D file.
	if !strings.HasSuffix(strings.ToUpper(fileName), ".M3D") {
		return 0
	}
	// Models starting without an underscore do not have any flags.
	if fileName[0] != '_' {
		return 0
	}
	c := fileName[1]
	switch {
	case c >= '0' && c <= '9':
		return Flags(c) - 48
	case c >= 'A' && c <= 'Z':
		return Flags(c) - 55
	case c >= 'a' && c <= 'z':
		return Flags(c) - 87
	}
	return 0
}
