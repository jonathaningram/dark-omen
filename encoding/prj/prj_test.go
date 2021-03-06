package prj

import (
	"os"
	"path"
	"testing"

	"github.com/google/go-cmp/cmp"
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

func TestDecoder_DecodeReal(t *testing.T) {
	type terrain struct {
		Width                int
		Height               int
		Heightmap1BlockCount int
		Heightmap2BlockCount int
		OffsetCount          int
	}
	tests := []struct {
		name    string
		path    string
		project *Project
		terrain terrain
		err     func(err error) (wantDesc string, pass bool)
	}{
		{
			name: "B1_01",
			path: path.Join("DARKOMEN", "DARKOMEN", "GAMEDATA", "1PBAT", "B1_01", "B1_01.PRJ"),
			terrain: terrain{
				Width:                184,
				Height:               200,
				Heightmap1BlockCount: 575,
				Heightmap2BlockCount: 575,
				OffsetCount:          473,
			},
			err: func(err error) (string, bool) {
				return "nil", err == nil
			},
		},
		{
			name: "B3_01",
			path: path.Join("DARKOMEN", "DARKOMEN", "GAMEDATA", "1PBAT", "B3_01", "B3_01.PRJ"),
			terrain: terrain{
				Width:                240,
				Height:               240,
				Heightmap1BlockCount: 900,
				Heightmap2BlockCount: 900,
				OffsetCount:          304,
			},
			err: func(err error) (string, bool) {
				return "nil", err == nil
			},
		},
		{
			name: "B4_01",
			path: path.Join("DARKOMEN", "DARKOMEN", "GAMEDATA", "1PBAT", "B4_01", "B4_01.PRJ"),
			terrain: terrain{
				Width:                220,
				Height:               320,
				Heightmap1BlockCount: 1120,
				Heightmap2BlockCount: 1120,
				OffsetCount:          503,
			},
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
			diff := cmp.Diff(tt.terrain, terrain{
				Width:                int(got.Terrain.Width),
				Height:               int(got.Terrain.Height),
				Heightmap1BlockCount: len(got.Terrain.Heightmap1Blocks),
				Heightmap2BlockCount: len(got.Terrain.Heightmap2Blocks),
				OffsetCount:          len(got.Terrain.Offsets),
			})
			if diff != "" {
				t.Errorf("terrain mismatch (-want +got):\n%v", diff)
			}
		})
	}
}
