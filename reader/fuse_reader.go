package reader

import (
	"fmt"
	"os"
	"time"
)

type FuseReader struct {
	bufferSize int
}

func NewFuseReader(bufferSize int) *FuseReader {

	if bufferSize < defaiultReadMinBufferSize {
		bufferSize = defaiultReadMinBufferSize
	}
	if bufferSize > defaiultReadMaxBufferSize {
		bufferSize = defaiultReadMaxBufferSize
	}
	return &FuseReader{
		bufferSize: bufferSize,
	}
}
func (fuseReader *FuseReader) Read(path string, flag bool, handle func(b []byte) error) error {
	start := time.Now()
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func(path string, start time.Time) {
		if flag {
			fmt.Printf("%-3v\t%-3s\n", time.Since(start), path)
		}
	}(path, start)
	defer f.Close()
	rb := make([]byte, 4096)
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
