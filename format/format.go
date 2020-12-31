package format

import (
	"fmt"
	"os"
	"text/tabwriter"
)

type FileFormat struct {
	Writer *tabwriter.Writer
}

func NewFormat() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 15, 0, 1, ' ', tabwriter.AlignRight)
}
func NewFileFormat(path string) *tabwriter.Writer {
	var file *os.File
	if _, err := os.Stat(path); err != nil {
		file, err = os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_APPEND, os.ModePerm)
		if err != nil {
			fmt.Println(err)
			return nil
		}
	}
	file.Truncate(0)
	return tabwriter.NewWriter(file, 15, 0, 1, ' ', tabwriter.AlignRight)
}
