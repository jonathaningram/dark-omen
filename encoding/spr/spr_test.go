package spr

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"
)

func runIfDarkOmenPathSet(t *testing.T) string {
	t.Helper()

	const darkOmenPathEnv = "DARK_OMEN_PATH"

	v := os.Getenv(darkOmenPathEnv)
	if v == "" {
		t.Skipf("skipping test when %s environment variable is not set", darkOmenPathEnv)
	}
	return v
}

func TestDecoder_Decode(t *testing.T) {
	tests := []struct {
		name   string
		r      func() io.ReaderAt
		sprite *Sprite
		err    func(err error) (wantDesc string, pass bool)
	}{
		{
			name: "header too short",
			r: func() io.ReaderAt {
				buf := &bytes.Buffer{}
				buf.WriteString("WHDO")                 // write the format
				buf.WriteString(strings.Repeat("x", 4)) // write junk for 4 bytes
				return bytes.NewReader(buf.Bytes())
			},
			err: func(err error) (string, bool) {
				partial := "header only read 8 byte(s)"
				return partial, err != nil && strings.Contains(err.Error(), partial)
			},
		},
		{
			name: "unknown sprite format",
			r: func() io.ReaderAt {
				buf := &bytes.Buffer{}
				buf.WriteString("ABCD")                  // write junk format
				buf.WriteString(strings.Repeat("x", 24)) // write junk for 24 bytes
				frameCount := make([]byte, 4)
				binary.LittleEndian.PutUint16(frameCount, 0) // write a frame count of 0
				buf.Write(frameCount)
				return bytes.NewReader(buf.Bytes())
			},
			err: func(err error) (string, bool) {
				partial := "unknown sprite format"
				return partial, err != nil && strings.Contains(err.Error(), partial)
			},
		},
		{
			name: "only contains header data",
			r: func() io.ReaderAt {
				buf := &bytes.Buffer{}
				buf.WriteString("WHDO")                  // write the format
				buf.WriteString(strings.Repeat("x", 24)) // write junk for 24 bytes
				frameCount := make([]byte, 4)
				binary.LittleEndian.PutUint16(frameCount, 0) // write a frame count of 0
				buf.Write(frameCount)
				return bytes.NewReader(buf.Bytes())
			},
			sprite: &Sprite{format: "WHDO"},
			err: func(err error) (string, bool) {
				return "nil", err == nil
			},
		},
		{
			name: "not enough frame headers",
			r: func() io.ReaderAt {
				buf := &bytes.Buffer{}
				buf.WriteString("WHDO")                  // write the format
				buf.WriteString(strings.Repeat("x", 24)) // write junk for 24 bytes
				frameCount := make([]byte, 4)
				binary.LittleEndian.PutUint16(frameCount, 50) // write a frame count of 50
				buf.Write(frameCount)
				return bytes.NewReader(buf.Bytes())
			},
			sprite: nil,
			err: func(err error) (string, bool) {
				partial := "sprite does not contain enough frame headers, expected to find 50"
				return partial, err != nil && strings.Contains(err.Error(), partial) && errors.Is(err, io.ErrUnexpectedEOF)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDecoder(tt.r())
			got, err := d.Decode()
			if wantDesc, ok := tt.err(err); !ok {
				t.Errorf("Decoder.Decode() error = %v, want %v", err, wantDesc)
				return
			}
			if !reflect.DeepEqual(got, tt.sprite) {
				t.Errorf("Decoder.Decode() sprite = %v, want %v", got, tt.sprite)
			}
		})
	}
}

func TestDecoder_DecodeReal(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		sprite     *Sprite
		frameCount int
		err        func(err error) (wantDesc string, pass bool)
	}{
		{
			name:       "SPRITES/BERNHD.SPR",
			path:       path.Join("DARKOMEN", "DARKOMEN", "GRAPHICS", "SPRITES", "BERNHD.SPR"),
			frameCount: 104,
			err: func(err error) (string, bool) {
				return "nil", err == nil
			},
		},
		{
			name:       "BANNERS/HBGRUCAV.SPR",
			path:       path.Join("DARKOMEN", "DARKOMEN", "GRAPHICS", "BANNERS", "HBGRUCAV.SPR"),
			frameCount: 2,
			err: func(err error) (string, bool) {
				return "nil", err == nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			darkOmenPath := runIfDarkOmenPathSet(t)

			f, err := os.Open(path.Join(darkOmenPath, tt.path))
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			d := NewDecoder(f)
			got, err := d.Decode()
			if wantDesc, ok := tt.err(err); !ok {
				t.Errorf("Decoder.Decode() error = %v, want %v", err, wantDesc)
				return
			}
			if len(got.Frames) != tt.frameCount {
				t.Errorf("Decoder.Decode() sprite frame count = %v, want %v", len(got.Frames), tt.frameCount)
			}
		})
	}
}
