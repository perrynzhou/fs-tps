package writer

import (
	"bufio"
	"fmt"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
)

const (
	TemplateFilePath = "./template_file"
)

type Writer struct {
	templateFile   *os.File
	fileCount      uint64
	wg             *sync.WaitGroup
	goroutineCount int
	root           string
	CurFileCount   uint64
}

func (w *Writer) Copy(src, dst string) error {
	originFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer originFile.Close()
	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	r := bufio.NewReader(originFile)
	buf := make([]byte, 1024*1024)
	for {
		n, err := r.Read(buf)
		buf = buf[:n]
		if n > 0 {
			dstFile.Write(buf)
		}
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {

			if err == io.EOF {
				break
			}
		}
	}
	return nil
}
func NewWriter(k int, fileCount uint64, wg *sync.WaitGroup, goroutineCount int, root string) (*Writer, error) {
	n := 1024 * k
	buf := make([]byte, n)
	ascii := "ABCEFGHIJKLMNOPQRSTUVWXYZ0123456789abcefghijklmnopqrstuvwxyz"
	asciiLen := len(ascii)
	for i := 0; i < n; i++ {
		buf[i] = ascii[rand.Intn(asciiLen-1)]
	}
	templateFile, err := os.OpenFile(TemplateFilePath, os.O_CREATE|os.O_TRUNC|os.O_APPEND|os.O_RDWR, os.ModePerm)
	if err != nil {
		return nil, err
	}
	if err = ioutil.WriteFile(TemplateFilePath, buf, os.ModePerm); err != nil {
		return nil, err
	}
	log.Println("start writer on ", root)
	return &Writer{
		templateFile:   templateFile,
		fileCount:      fileCount,
		wg:             wg,
		goroutineCount: goroutineCount,
		root:           root,
		CurFileCount:   uint64(0),
	}, nil
}
func (w *Writer) start(n int) {
	w.wg.Done()
	for i := uint64(0); i < w.fileCount; i++ {
		dstFilePath := fmt.Sprintf("%s/%s.dat", w.root, uuid.NewV4())
		if err := w.Copy(TemplateFilePath, dstFilePath); err != nil {
			log.Errorln(err)
			break
		}
		atomic.AddUint64(&w.CurFileCount, 1)
	}
	fmt.Printf("goroutine finish %d files\n", w.CurFileCount)
}
func (w *Writer) Run() {
	for i := 0; i < w.goroutineCount; i++ {
		w.start(i)
	}
}
