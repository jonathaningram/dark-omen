package m3d

import (
	"testing"
)

func TestModelFlags(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		want     Flags
	}{
		{
			name:     "empty",
			fileName: "",
			want:     0,
		},
		{
			name:     "too short",
			fileName: "A.M3D",
			want:     0,
		},
		{
			name:     "wrong extension",
			fileName: "A.PRJ",
			want:     0,
		},
		{
			name:     "no flags embedded in file name",
			fileName: "KBARREL.M3D",
			want:     0,
		},
		{
			name:     "_0",
			fileName: "_0FILE.M3D",
			want:     0,
		},
		{
			name:     "_4",
			fileName: "_4FILE.M3D",
			want:     0b100,
		},
		{
			name:     "_6",
			fileName: "_6FILE.M3D",
			want:     0b110,
		},
		{
			name:     "_7",
			fileName: "_7FILE.M3D",
			want:     0b111,
		},
		{
			name:     "_K",
			fileName: "_KFILE.M3D",
			want:     0b10100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ModelFlags(tt.fileName); got != tt.want {
				t.Errorf("ModelFlags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFlags_Has(t *testing.T) {
	tests := []struct {
		name string
		f    Flags
		test Flags
		want bool
	}{
		{
			name: "no flags",
			f:    ModelFlags("BASE.M3D"),
			test: UVAnimation | AlphaTransparency,
			want: false,
		},
		{
			name: "_6",
			f:    ModelFlags("_6FILE.M3D"),
			test: UVAnimation | AlphaTransparency,
			want: true,
		},
		{
			name: "_7",
			f:    ModelFlags("_7FILE.M3D"),
			test: Translucency | UVAnimation | AlphaTransparency,
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.Has(tt.test); got != tt.want {
				t.Errorf("Flags.Has() = %v, want %v", got, tt.want)
			}
		})
	}
}
