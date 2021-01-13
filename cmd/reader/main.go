package main

import (
	"flag"
	"fmt"
	"glusterfs-tps/conf"
	"glusterfs-tps/reader"
	"os"
	"os/signal"
	"syscall"
)

var (
	defaultGoRoutine  = flag.Int("g", 4, "default goroutine number")
	defaultDataPath   = flag.String("d", "/home/perrynzhou/", "default data path")
	defaultOpType     = flag.String("o", "data", "default is meta,you can set  data")
	defaultFileSuffix = flag.String("s", "", "default file suffix")
	defaultTicker = flag.Int("t", 4, "default ticker  value")
	defauleReadBuf  = flag.Int("b",4,"default is 4k")
)

func handlerErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
func main() {
	flag.Parse()
	conf := &conf.Conf{
		IndexName:      "index",
		IndexPath:      "/tmp",
		Count:          *defaultGoRoutine,
		ReadBufferSize: *defauleReadBuf,
		Suffix:         *defaultFileSuffix,
		Ticker:         *defaultTicker,
		OpType:         *defaultOpType,
	}

	fetcher, err := reader.NewFetcher(conf, *defaultDataPath)
	handlerErr(err)
	fetcher.Run()
	defer fetcher.PrintMetric("All Jobs Info")
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	fmt.Printf("##########main pid:%d##########\n", os.Getpid())
	for {
		select {
		case <-sigs:
			fetcher.Stop()
			return
		}
	}
}
