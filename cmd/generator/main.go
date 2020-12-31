package main

import (
	"flag"
	"fmt"
	"glusterfs-tps/writer"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	rootPath       = flag.String("p", "/", "default path")
	goroutineCount = flag.Int("g", 4, "goroutine count")
	kBytes         = flag.Int("k", 1, "default is 1k")
	fileCount      = flag.Uint64("n", 1024, "default is 1024")
	tickerSeconds  = flag.Int("t", 1, "default ticker seconds time")
)

func main() {
	flag.Parse()
	wg := &sync.WaitGroup{}
	wg.Add(*goroutineCount)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	w, err := writer.NewWriter(*kBytes, *fileCount, wg, *goroutineCount, *rootPath)
	if err != nil {
		log.Fatalln(err)
	}
	w.Run()
	ticker := time.NewTicker(time.Duration(*tickerSeconds) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-sigs:
			if err = os.Remove(writer.TemplateFilePath); err != nil {
				log.Errorln(err)
			}
			fmt.Printf("Total File Count:%d\n", w.CurFileCount)
			return
		case <-ticker.C:
			fmt.Printf("Total File Count:%d\n", w.CurFileCount)
			break
		}
	}
}
