package fnt

import (
	"bytes"
	"image"
	"os"
	"path"
	"testing"
)

var font *Font

func BenchmarkDecode(b *testing.B) {
	tests := []string{
		"F_MENBG.FNT",
		"F_HELP.FNT",
	}
	for _, tt := range tests {
		b.Run(tt, func(b *testing.B) {
			b.ReportAllocs()
			bs, err := os.ReadFile(path.Join("testdata", tt))
			if err != nil {
				b.Fatal(err)
			}
			r := bytes.NewReader(bs)
			var f *Font
			for n := 0; n < b.N; n++ {
				b.StopTimer()
				r.Reset(bs)
				d := NewDecoder(r)
				b.StartTimer()
				var err error
				f, err = d.Decode()
				if err != nil {
					b.Fatalf("Decode() error = %v, want nil", err)
				}
			}
			font = f
		})
	}
}

func Test_xy(t *testing.T) {
	type args struct {
		img *image.NRGBA
		x   int
		y   int
	}
	tests := []struct {
		name  string
		args  args
		wantX int
		wantY int
	}{
		{
			name: "2x2 image first row",
			args: args{
				image.NewNRGBA(image.Rect(0, 0, 2, 2)),
				0, 0,
			},
			wantX: 1,
			wantY: 0,
		},
		{
			name: "2x2 image second row",
			args: args{
				image.NewNRGBA(image.Rect(0, 0, 2, 2)),
				1, 0,
			},
			wantX: 0,
			wantY: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotX, gotY := xy(tt.args.img, tt.args.x, tt.args.y)
			if !(gotX == tt.wantX && gotY == tt.wantY) {
				t.Errorf("xy() = (%v,%v), want (%v,%v)", gotX, gotY, tt.wantX, tt.wantY)
			}
		})
	}
}
