package reader

import (
	"fmt"
	"os"
	"time"
)

type Reader struct {
	bufferSize int
	showDetail bool
}

func NewReader(showDetail bool, bufferSize int) *Reader {

	if bufferSize < defaultReadMinBufferSize {
		bufferSize = defaultReadMinBufferSize
	}
	if bufferSize > defaultReadMaxBufferSize {
		bufferSize = defaultReadMaxBufferSize
	}
	return &Reader{
		bufferSize: bufferSize,
		showDetail: showDetail,
	}
}
func (reader *Reader) Read(path string, handle func(b []byte) error) error {
	start := time.Now()
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	if reader.showDetail {
		defer func(file *os.File, path string, start time.Time) {
			fmt.Printf("%-3v\t%-3s\n", time.Since(start), path)
			defer file.Close()
		}(f, path, start)
	}
	rb := make([]byte, reader.bufferSize)
	os.Stat(path)
	for {
		switch nr, err := f.Read(rb[:]); true {
		case nr < 0:
			return err
		case nr == 0:
			return nil
		case nr > 0:
			if handle != nil {
				handle(rb)
			}
		}
	}
}
