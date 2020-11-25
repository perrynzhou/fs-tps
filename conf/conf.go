package conf

import (
	"encoding/json"
	"io/ioutil"
)

type Conf struct {
	IndexName      string `json:"index_name"`
	IndexPath      string `json:"index_path"`
	Count          uint64 `json:"count"`
	ReadBufferSize int    `json:"read_buffer_size"`
	ShowDetail     bool   `json:"show_detail"`
	Suffix         string `json:"suffix"`
	Ticker         int    `json:"output_summary_ticker"`
}

func NewConf(path string) (*Conf, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	conf := &Conf{}
	if err = json.Unmarshal(b, conf); err != nil {
		return nil, err
	}
	return conf, nil
}
