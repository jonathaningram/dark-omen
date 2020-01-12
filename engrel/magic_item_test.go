package engrel

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
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

func TestReadMagicItemNamesReal(t *testing.T) {
	darkOmenPath := runIfDarkOmenPathSet(t)

	bs, err := ioutil.ReadFile(path.Join(darkOmenPath, "DARKOMEN", "DARKOMEN", "PRG_ENG", "ENGREL.EXE"))
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		wantLen   int
		wantNames map[int]string
		wantErr   bool
	}{
		{
			name:    "correct length and names",
			wantLen: 64,
			wantNames: map[int]string{
				0:  "Not used.",
				1:  "Grudgebringer Sword",
				42: "Brain Bursta",
				63: "SingleSanguine",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(bs)

			got, err := ReadMagicItemNames(r)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadMagicItemNames() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotLen := len(got); gotLen != tt.wantLen {
				t.Errorf("len(ReadMagicItemNames()) = %v, want %v", gotLen, tt.wantLen)
			}
			for index, wantName := range tt.wantNames {
				if gotName := got[index]; gotName != wantName {
					t.Errorf("ReadMagicItemNames()[%d] = %v, want %v", index, gotName, wantName)
				}
			}
		})
	}
}

func TestReadMagicItemNameReal(t *testing.T) {
	darkOmenPath := runIfDarkOmenPathSet(t)

	bs, err := ioutil.ReadFile(path.Join(darkOmenPath, "DARKOMEN", "DARKOMEN", "PRG_ENG", "ENGREL.EXE"))
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		index   int64
		wantErr bool
	}{
		{
			name:  "Grudgebringer Sword",
			index: 1,
		},
		{
			name:  "Enchanted Shield",
			index: 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(bs)

			got, err := ReadMagicItemName(r, tt.index)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadMagicItemName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.name {
				t.Errorf("ReadMagicItemName() = %v, want %v", got, tt.name)
			}
		})
	}
}
