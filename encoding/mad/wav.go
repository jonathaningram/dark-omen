package mad

import (
	"io"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

func (s *Stream) EncodeToWAV(w io.WriteSeeker) error {
	const sampleRate = 22050

	buf := &audio.IntBuffer{
		Format: &audio.Format{
			NumChannels: s.Channels(),
			SampleRate:  sampleRate,
		},
	}
	for _, block := range s.Blocks {
		for _, v := range block.AsPCM16Block().Data {
			buf.Data = append(buf.Data, int(v))
		}
	}
	enc := wav.NewEncoder(
		w,
		sampleRate,
		16, // bit depth
		s.Channels(),
		1, // audio format
	)
	if err := enc.Write(buf); err != nil {
		return err
	}
	if err := enc.Close(); err != nil {
		return err
	}
	return nil
}
