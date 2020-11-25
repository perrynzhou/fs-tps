package reader

import (
	"bufio"
	"encoding/json"
	"fmt"
	"glusterfs-tps/conf"
	"glusterfs-tps/format"
	"glusterfs-tps/metric"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultFilecount         = 2
	defaultIndexMetaFileName = "index.meta"
)

type Meta struct {
	Index        uint64
	Length       uint64
	Count        uint64
	MilliSeconds uint64
	Start        time.Time
	End          time.Time
}
type Job struct {
	Header   string
	JobsInfo []string
}
type Fetcher struct {
	root        string
	name        string
	indexPath   string
	index       uint64
	count       uint64
	length      uint64
	writer      *bufio.Writer
	metrics     []*metric.Metric
	indexFile   *os.File
	done        chan struct{}
	stop        chan struct{}
	in          chan *Meta
	out         chan *Meta
	wg          *sync.WaitGroup
	reader      IReader
	suffix      string
	ticker      time.Duration
	saveTpsInfo *format.FileFormat
}

func NewFetcher(cf *conf.Conf, root string) (*Fetcher, error) {
	if cf == nil {
		return nil, fmt.Errorf("conf is nil")
	}
	filecount := uint64(cf.Count)
	if filecount < defaultFilecount {
		filecount = defaultFilecount
	}
	fetcher := &Fetcher{
		root:      root,
		name:      fmt.Sprintf("%s/%s", cf.IndexPath, cf.IndexName),
		index:     0,
		in:        make(chan *Meta),
		out:       make(chan *Meta),
		count:     filecount,
		stop:      make(chan struct{}),
		done:      make(chan struct{}),
		metrics:   make([]*metric.Metric, 0),
		wg:        &sync.WaitGroup{},
		suffix:    cf.Suffix,
		indexPath: cf.IndexPath,
		ticker:    time.Second * time.Duration(cf.Ticker),
	}
	fetcher.reader = NewReader(cf.OpType,cf.ShowDetail,cf.ReadBufferSize)
	return fetcher, nil
}
func (fetcher *Fetcher) initIndexFile() error {
	if fetcher.writer != nil && fetcher.indexFile != nil {
		fetcher.writer.Flush()
		fetcher.indexFile.Close()
	}
	indexFileName := fmt.Sprintf("%s.%d", fetcher.name, fetcher.index)
	indexFile, err := os.OpenFile(indexFileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0775)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fetcher.indexFile = indexFile
	fetcher.writer = bufio.NewWriter(fetcher.indexFile)
	fmt.Printf("new file %s success\n", indexFileName)
	return nil
}
func (fetcher *Fetcher) flush(path string) error {
	if _, err := fetcher.writer.WriteString(fmt.Sprintf("%s\n", path)); err != nil {
		return err
	}
	defer fetcher.writer.Flush()
	return nil
}
func (fetcher *Fetcher) loadIndex() bool {
	indexMetaPath := fmt.Sprintf("%s/%s", fetcher.indexPath, defaultIndexMetaFileName)
	b, err := ioutil.ReadFile(indexMetaPath)
	if err != nil {
		fmt.Println(err)
		return false
	}
	metas := make([]*Meta, 0)

	if err = json.Unmarshal(b, &metas); err != nil {
		fmt.Println(err)
		return false
	}
	atomic.AddUint64(&fetcher.index, uint64(len(metas)-1))
	for _, meta := range metas {
		atomic.SwapUint64(&meta.MilliSeconds, 0)
		fetcher.in <- meta
	}
	return true
}
func (fetcher *Fetcher) startIndexJob() error {
	defer fetcher.wg.Done()
	var err error
	if !fetcher.loadIndex() {
		err = fetcher.createIndex()
	}
	return err
}
func (fetcher *Fetcher) createIndex() error {
	defer fetcher.indexFile.Close()
	err := fetcher.initIndexFile()
	if err != nil {
		return err
	}
	metas := make([]*Meta, 0)
	meta := &Meta{
		Count:  0,
		Length: 0,
		Index:  fetcher.index,
	}
	defer fmt.Println("...exit fetch index service", fetcher.root)
	err = filepath.Walk(fetcher.root,
		func(path string, f os.FileInfo, err error) error {
			if f == nil {
				return err
			}
			if f.IsDir() {
				return nil
			}
			if len(fetcher.suffix)==0 ||(len(fetcher.suffix) > 0 && strings.HasSuffix(path, fetcher.suffix)) {
				atomic.AddUint64(&meta.Count, 1)
				atomic.AddUint64(&meta.Length, uint64(f.Size()))
				fetcher.flush(path)
				if meta.Count >= fetcher.count {
					fmt.Printf("finish append index to  %s.%d,count:%d\n", fetcher.name, fetcher.index, meta.Count)
					metas = metas[:0]
					metas = append(metas, meta)
					fetcher.in <- meta
					atomic.AddUint64(&fetcher.index, 1)
					meta = &Meta{
						Count:  0,
						Length: 0,
						Index:  fetcher.index,
					}
					if err = fetcher.initIndexFile(); err != nil {
						fmt.Println(err)
						return err
					}
				}
			}
				return nil

		})
	if err != nil {
		return err
	}
	if meta.Count < fetcher.count {
		atomic.SwapUint64(&meta.Index, fetcher.index)
		metas = append(metas, meta)
		fetcher.in <- meta
	}
	if len(metas) > 0 {
		indexMetaPath := fmt.Sprintf("%s/%s", fetcher.indexPath, defaultIndexMetaFileName)
		os.Truncate(indexMetaPath, 0)
		if b, err := json.MarshalIndent(metas, " ", " "); err == nil {
			if err = ioutil.WriteFile(indexMetaPath, b, os.ModePerm); err == nil {
				fmt.Printf(".....new %s success\n", indexMetaPath)
			}
		}
	}
	fmt.Printf("finish  append index to  %s.%d,count:%d\n", fetcher.name, fetcher.index, meta.Count)
	return nil

}
func (fetcher *Fetcher) startMetricJob() {
	defer fetcher.wg.Done()
	var metricLen int
	go func(fetcher *Fetcher) {
		ticker := time.NewTicker(fetcher.ticker)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				metricLen = fetcher.PrintMetric("Running Jobs Info")
				if uint64(metricLen) == (fetcher.index + 1) {
					fetcher.PrintMetric("Finish All Jobs")
					fmt.Printf("###################################%s %v#################################\n", "All Jobs Has Already Finish ", time.Now().Format("2006-01-02 15:04:05"))
					return
				}
			}
		}
	}(fetcher)
Loop:
	for {
		select {
		case meta := <-fetcher.out:
			indexFile := fmt.Sprintf("%s.%d", fetcher.name, meta.Index)
			metric := metric.NewMetric(indexFile, meta.Start, meta.End)
			metric.Compute(meta.Count, meta.Length, meta.MilliSeconds)
			fetcher.metrics = append(fetcher.metrics, metric)
			sort.Slice(fetcher.metrics, func(i, j int) bool {
				return fetcher.metrics[i].Milliseconds < fetcher.metrics[j].Milliseconds
			})
			format := format.NewFormat()
			fmt.Printf("###########################################Job %d###################################\n", meta.Index)
			fmt.Fprintln(format, "finish-jobs\tindexfile\tfiles\tlength\tseconds\t\tstart\t\tend")
			fmt.Fprintf(format, "%d\t%s\t%d\t%.2f\t%v\t\t%v\t\t%v\n", len(fetcher.metrics), indexFile, meta.Count, float64(meta.Length)/1024/1024, float64(meta.MilliSeconds)/1000, meta.Start.Format("2006-01-02 15:04:05"), meta.End.Format("2006-01-02 15:04:05"))
			format.Flush()
		case <-fetcher.stop:
			fmt.Println("....exit fetch index meta service")
			break Loop
		}
	}
}
func (fetcher *Fetcher) Run() {
	fetcher.wg.Add(3)
	go fetcher.startIndexJob()
	go fetcher.startMetricJob()
	go fetcher.startReadingJobs()
}
func (fetcher *Fetcher) Stop() {
	fetcher.stop <- struct{}{}
	fetcher.done <- struct{}{}
	defer close(fetcher.in)
	defer close(fetcher.out)
	defer close(fetcher.done)
	defer close(fetcher.stop)
	defer fetcher.wg.Wait()
}
func (fetcher *Fetcher) PrintMetric(msg string) int {
	var FileCount, FileLength uint64
	var metricLen int
	if fetcher.metrics != nil && len(fetcher.metrics) > 0 {
		sort.Slice(fetcher.metrics, func(i, j int) bool {
			return fetcher.metrics[i].Milliseconds < fetcher.metrics[j].Milliseconds
		})
		metricLen = len(fetcher.metrics)
		fmt.Printf("###################################%v-%s#################################\n", time.Now().Format("2006-01-02 15:04:05"), msg)
		format := format.NewFormat()
		defer format.Flush()
		fmt.Fprintln(format, "index\ttps\tfiles\tlength(mb)\tseconds\t\tstart\t\tend")
		var milliseconds uint64
		for _, m := range fetcher.metrics {
			if milliseconds < m.Milliseconds {
				milliseconds = m.Milliseconds
			}
			seconds := float64(m.Milliseconds) / 1000
			tps := float64(m.FileCount) / seconds
			atomic.AddUint64(&FileCount, m.FileCount)
			atomic.AddUint64(&FileLength, m.FileLength)
			fmt.Fprintf(format, "%s\t%.2f\t%d\t%.2f\t%.5f\t\t%v\t\t%v\n", m.Name, tps, m.FileCount, float64(m.FileLength)/1024/1024, seconds, m.Start.Format("2006-01-02 15:04:05"), m.End.Format("2006-01-02 15:04:05"))
		}
		seconds := float64(milliseconds) / 1000
		fmt.Fprintln(format, "\ntotal files\t\ttotal length(mb)\t\ttotal tps\t\tfinish jobs\t\ttotal jobs")
		fmt.Fprintf(format, "%d\t\t%.2f\t\t%.2f\t\t%d\t\t%d\n", FileCount, float64(FileLength)/1024/1024, float64(FileCount)/seconds, len(fetcher.metrics), fetcher.index+1)
	}
	return metricLen
}
func (fetcher *Fetcher) handleIndexFile(meta *Meta /*,wg *sync.WaitGroup */) error {
	indexFilePath := fmt.Sprintf("%s.%d", fetcher.name, meta.Index)
	fmt.Printf("....start  goroutine to  %s....\n", indexFilePath)
	indexFile, err := os.OpenFile(indexFilePath, os.O_RDONLY, 775)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer indexFile.Close()
	scanner := bufio.NewScanner(indexFile)
	meta.Start = time.Now()
	for scanner.Scan() {
		filePath := scanner.Text()
		if err = fetcher.reader.Read(filePath, nil); err != nil {
			atomic.SwapUint64(&meta.Count, meta.Count-1)
			continue
		}
	}
	meta.MilliSeconds = uint64(time.Since(meta.Start).Milliseconds())
	meta.End = time.Now()
	fetcher.out <- meta
	return nil
}
func (fetcher *Fetcher) startReadingJobs() error {
	defer fetcher.wg.Done()
Loop:
	for {
		select {
		case meta, ok := <-fetcher.in:
			if !ok {
				break Loop
			}
			fmt.Printf("%s.%d contains files :%d\n", fetcher.name, meta.Index, meta.Count)
			go fetcher.handleIndexFile(meta)
		case <-fetcher.done:
			break Loop
		}
	}
	return nil
}
