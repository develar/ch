package proto

// Compression status.
type Compression byte

// Compression satuses.
const (
	CompressionDisabled Compression = 0
	CompressionEnabled  Compression = 1
)

// Encode to buffer.
func (c Compression) Encode(b *Buffer) {
	b.PutUVarInt(uint64(c))
}