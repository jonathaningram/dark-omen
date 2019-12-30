package sad

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
	for i := 0; i < len(s.LeftBlocks); i++ {
		leftPCM16Block := s.LeftBlocks[i].AsPCM16Block()
		rightPCM16Block := s.RightBlocks[i].AsPCM16Block()

		for j := 0; j < len(leftPCM16Block.Data); j++ {
			buf.Data = append(buf.Data, int(leftPCM16Block.Data[j]), int(rightPCM16Block.Data[j]))
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
