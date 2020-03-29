package reader

const (
	defaiultReadMinBufferSize = 4096
	defaiultReadMaxBufferSize = 4096 * 1024
)

type Reader interface {
	Read(path string, flag bool, handle func(b []byte) error) error
}
