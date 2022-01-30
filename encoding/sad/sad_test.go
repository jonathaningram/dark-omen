package sad

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"
)

var stream *Stream

func BenchmarkDecode(b *testing.B) {
	tests := []string{
		"1BOUN001.SAD",
		"1CHAS001.SAD",
	}
	for _, tt := range tests {
		b.Run(tt, func(b *testing.B) {
			b.ReportAllocs()
			bs, err := ioutil.ReadFile(path.Join("testdata", tt))
			if err != nil {
				b.Fatal(err)
			}
			r := bytes.NewReader(bs)
			var s *Stream
			for n := 0; n < b.N; n++ {
				b.StopTimer()
				r.Reset(bs)
				d := NewDecoder(r)
				b.StartTimer()
				var err error
				s, err = d.Decode()
				if err != nil {
					b.Fatalf("Decode() error = %v, want nil", err)
				}
			}
			stream = s
		})
	}
}

func TestRoundTrip(t *testing.T) {
	tests := []string{
		"1BOUN001.SAD",
		"1CHAS001.SAD",
	}
	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			f, err := os.Open(path.Join("testdata", tt))
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			want := &bytes.Buffer{}
			stream, err := NewDecoder(io.TeeReader(f, want)).Decode()
			if err != nil {
				t.Fatalf("Decode() error = %v, want nil", err)
			}

			got := &bytes.Buffer{}
			err = NewEncoder(got).Encode(stream)
			if err != nil {
				t.Fatalf("Encode() error = %v, want nil", err)
			}
			if !reflect.DeepEqual(got.Bytes(), want.Bytes()) {
				t.Errorf("got encoded bytes = %v [output truncated], want %v [output truncated]", truncateBytes(got.Bytes(), 10), truncateBytes(want.Bytes(), 10))
			}
		})
	}
}

func BenchmarkEncodeToWAV(b *testing.B) {
	tests := []string{
		"1BOUN001.SAD",
		"1CHAS001.SAD",
	}
	for _, tt := range tests {
		b.Run(tt, func(b *testing.B) {
			b.ReportAllocs()
			bs, err := ioutil.ReadFile(path.Join("testdata", tt))
			if err != nil {
				b.Fatal(err)
			}
			r := bytes.NewReader(bs)
			s, err := NewDecoder(r).Decode()
			if err != nil {
				b.Fatal(err)
			}
			for n := 0; n < b.N; n++ {
				b.StopTimer()
				tmp, err := ioutil.TempFile("", "wav")
				if err != nil {
					b.Fatal(err)
				}
				defer os.Remove(tmp.Name())
				b.StartTimer()
				if err := s.EncodeToWAV(tmp); err != nil {
					b.Fatalf("EncodeToWAV() error = %v, want nil", err)
				}
			}
		})
	}
}

func TestEncodeToWAV(t *testing.T) {
	tests := []string{
		"1BOUN001.SAD",
		"1CHAS001.SAD",
		"09EERIE.SAD",
	}
	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			f, err := os.Open(path.Join("testdata", tt))
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			stream, err := NewDecoder(f).Decode()
			if err != nil {
				t.Fatalf("Decode() error = %v, want nil", err)
			}

			tmp, err := ioutil.TempFile("", "wav")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmp.Name())

			if err := stream.EncodeToWAV(tmp); err != nil {
				t.Fatalf("EncodeToWAV() error = %v, want nil", err)
			}

			got, err := ioutil.ReadFile(tmp.Name())
			if err != nil {
				t.Fatal(err)
			}
			want, err := ioutil.ReadFile(path.Join("testdata", strings.ReplaceAll(tt, path.Ext(tt), ".WAV")))
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(got, want) {
				t.Errorf("got WAV bytes = %v [output truncated], want %v [output truncated]", truncateBytes(got, 10), truncateBytes(want, 10))
			}
		})
	}
}

func truncateBytes(bs []byte, size int) []byte {
	if len(bs) > size {
		return bs[:size]
	}
	return bs
}
