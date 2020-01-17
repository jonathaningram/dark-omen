package arm

import (
	"bytes"
	"os"
	"path"
	"testing"
)

var army *Army

func BenchmarkDecode(b *testing.B) {
	tests := []string{
		"B410NME.ARM",
	}
	for _, tt := range tests {
		b.Run(tt, func(b *testing.B) {
			b.ReportAllocs()
			bs, err := os.ReadFile(path.Join("testdata", tt))
			if err != nil {
				b.Fatal(err)
			}
			r := bytes.NewReader(bs)
			var a *Army
			for n := 0; n < b.N; n++ {
				b.StopTimer()
				r.Reset(bs)
				d := NewDecoder(r)
				b.StartTimer()
				var err error
				a, err = d.Decode()
				if err != nil {
					b.Fatalf("Decode() error = %v, want nil", err)
				}
			}
			army = a
		})
	}
}
