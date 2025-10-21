package config

import (
	"encoding/json"
	"log"
	"os"
	"sort"
)

type pluginPolicy struct {
	Type      string `json:"type"`
	GroupOnly bool   `json:"group_only"`
	Groups    []int  `json:"groups"`
	AdminId   int64  `json:"admin_id"`
}

var Host = "127.0.0.1:8080"
var MysqlHost = "127.0.0.1:3306"
var Policies = make(map[string]pluginPolicy)
var WebHost = "127.0.0.1:8080"
var AdminId = int64(0)
var config = struct {
	Host      string                  `json:"host"`
	MysqlHost string                  `json:"mysqlHost"`
	Policies  map[string]pluginPolicy `json:"policies"`
	WebHost   string                  `json:"webHost"` // bot前端(napcat)可访问地址
	AdminId   int64                   `json:"adminId"`
}{Host: Host,
	MysqlHost: MysqlHost,
	Policies:  Policies,
	WebHost:   WebHost,
	AdminId:   AdminId,
}

func init() {
	LoadConfig()
}

func LoadConfig() {
	file, err := os.OpenFile("./config/config.json", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(file)
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
	Host = config.Host
	MysqlHost = config.MysqlHost
	WebHost = config.WebHost
	AdminId = config.AdminId
	for _, policy := range config.Policies {
		sort.Ints(policy.Groups)
	}
}

func SaveConfig() {
	config.Policies = Policies
	file, err := os.OpenFile("./config/config.json", os.O_RDWR|os.O_TRUNC, 0755)
	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
		}
	}(file)
	var bytes []byte
	bytes, err = json.Marshal(config)
	if err != nil {
		return
	}
	_, err = file.Write(bytes)
	_, err = file.Write([]byte("\n"))
	if err != nil {
		return
	}
}
