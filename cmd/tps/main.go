package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"glusterfs-tps/conf"
	"glusterfs-tps/reader"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
)

const (
	defaultConfTemplate = "./conf.json"
)

var (
	defaultConfPath = flag.String("c", "./conf.json", "default config")
	defaultDataPath = flag.String("d", "/home/perrynzhou/", "default data path")
)

func genConfTemplate() error {
	_, err := os.Stat(defaultConfTemplate)
	if err == nil {
		return nil
	}
	if os.IsNotExist(err) {
		cf := &conf.Conf{
			IndexName:      "index",
			IndexPath:      "/tmp",
			Count:          4,
			ReadBufferSize: 8192,
			ShowDetail:     false,
			Suffix:         "",
			Ticker:         4,
		}
		b, err := json.MarshalIndent(cf, " ", " ")
		if err != nil {
			return err
		}
		return ioutil.WriteFile(defaultConfTemplate, b, os.ModePerm)
	}
	return nil
}
func handlerErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
func main() {
	flag.Parse()
	handlerErr(genConfTemplate())
	conf, err := conf.NewConf(*defaultConfPath)
	handlerErr(err)
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
