// Package cstringutil contains utility functions for working with C strings.
package cstringutil

// ToGo converts the C string represented by the provided bytes to a Go string.
func ToGo(n []byte) string {
	return string(n[:clen(n)])
}

// clen returns the index of the first NULL byte in n or len(n) if n contains
// no NULL byte.
func clen(n []byte) int {
	for i := 0; i < len(n); i++ {
		if n[i] == 0 {
			return i
		}
	}
	return len(n)
}
