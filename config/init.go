package config

import (
	"encoding/json"
	"os"
)

var WebPort = "8080"
var MysqlHost = "127.0.0.1:3306"
var NapcatHost = "http://127.0.0.1:8088/"
var CmdChar []string

func init() {
	var config = struct {
		WebPort    string   `json:"webPort"`
		MysqlHost  string   `json:"mysqlHost"`
		NapcatHost string   `json:"napcatHost"`
		CmdChar    []string `json:"cmdChar"`
	}{WebPort: WebPort, MysqlHost: MysqlHost, NapcatHost: NapcatHost}
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
	if len(config.CmdChar) == 0 {
		CmdChar = []string{"#", "＃"}
	} else {
		CmdChar = config.CmdChar
	}
	WebPort = config.WebPort
	MysqlHost = config.MysqlHost
	NapcatHost = config.NapcatHost
}
