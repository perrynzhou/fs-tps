package conf

import (
	"encoding/json"
	"io/ioutil"
)

type Conf struct {
	Address    string `json:"addr"`
	Port       int    `json:"port"`
	Volume     string `json:"volume"`
	IndexName  string `json:"index_name"`
	IndexPath  string `json:"index_path"`
	Count      uint64 `json:"count"`
	OutputFlag bool   `json:"output_flag"`
	ApiEnable  bool   `json:"api_enable"`
	BufferSize int    `json:"buffer_size"`
	Suffix     string `json:"suffix"`
	Ticker     int    `json:"ticker"`
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
