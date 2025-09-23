package config

import (
	"encoding/json"
	"os"
)

type pluginTactics struct {
	Type   string   `json:"type"`
	Groups []string `json:"groups"`
}

var Host = "127.0.0.1:8080"
var MysqlHost = "127.0.0.1:3306"
var Tactics map[string]pluginTactics

func init() {
	var config = struct {
		Host      string                   `json:"host"`
		MysqlHost string                   `json:"mysqlHost"`
		Tactics   map[string]pluginTactics `json:"tactics"`
	}{Host: Host, MysqlHost: MysqlHost}
	file, err := os.OpenFile("./config/config.json", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}
	fileInfo, err := file.Stat()
	if err != nil {
		panic(err)
	}
	var bytes []byte
	if fileInfo.Size() < 2 {
		bytes, err = json.Marshal(config)
		if err != nil {
			return
		}
		_, err = file.Write(bytes)
		_, err = file.Write([]byte("\n"))
		if err != nil {
			return
		}
	} else {
		bytes = make([]byte, fileInfo.Size())
		_, err = file.Read(bytes)
		if err != nil {
			return
		}
		err = json.Unmarshal(bytes, &config)
		if err != nil {
			return
		}
	}
	err = file.Close()
	if err != nil {
		panic(err)
	}
	Host = config.Host
	MysqlHost = config.MysqlHost
	Tactics = config.Tactics
}
