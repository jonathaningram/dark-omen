package audio

type Block interface {
	Bytes() ([]byte, error)
	AsPCM16Block() *PCM16Block
}
