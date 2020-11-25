package reader

const (
	defaultReadMinBufferSize = 4096
	defaultReadMaxBufferSize = 8192 * 1024
)

type IReader interface {
	Read(path string,  handle func(b []byte,) error) error
}
