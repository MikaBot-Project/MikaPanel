package config

import (
	"encoding/json"
	"os"
	"sort"
)

type pluginPolicy struct {
	Type      string `json:"type"`
	GroupOnly bool   `json:"group_only"`
	Groups    []int  `json:"groups"`
}

var Host = "127.0.0.1:8080"
var MysqlHost = "127.0.0.1:3306"
var Policies map[string]pluginPolicy

func init() {
	var config = struct {
		Host      string                  `json:"host"`
		MysqlHost string                  `json:"mysqlHost"`
		Policies  map[string]pluginPolicy `json:"policies"`
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
	Policies = config.Policies
	for _, policy := range config.Policies {
		sort.Ints(policy.Groups)
	}
}
