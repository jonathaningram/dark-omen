package audio

import "testing"

// See https://www.cs.columbia.edu/~hgs/audio/dvi/IMA_ADPCM.pdf at page 32 for
// the algorithm and for example input. Note: The example input does appear to
// be wrong though because `if (0x8763 > 32767) == FALSE` is actually true.
func Test_blockDecoder_decode(t *testing.T) {
	d := &blockDecoder{
		sample: 0x8700,
		index:  24,
	}
	var originalSample byte = 0x3

	t.Run("returns new sample", func(t *testing.T) {
		got := d.decode(originalSample)
		var want int16 = 0x7FFF
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("save new sample on decoder", func(t *testing.T) {
		got := d.sample
		var want int = 0x7FFF
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("save new index on decoder", func(t *testing.T) {
		got := d.index
		var want int16 = 23
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}
