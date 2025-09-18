package plugin

import (
	"MikaPanel/config"
	"MikaPanel/messages"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

var pluginLogBufferMap map[string]*bytes.Buffer
var pluginOutBufferMap map[string]*bytes.Buffer
var pluginInBufferMap map[string]*bytes.Buffer

var MessagePluginMap []string
var CmdPluginMap map[string]string
var NoticePluginMap map[string][]string

func init() {
	CmdPluginMap = make(map[string]string)
	NoticePluginMap = make(map[string][]string)
	dirInfo, err := os.Stat("plugin")
	if err != nil {
		log.Println("读取插件文件夹路径信息失败")
		return
	}
	if !dirInfo.IsDir() {
		log.Println("./plugin 非文件夹路径")
		return
	}
	dir, err := os.Open("plugin")
	if err != nil {
		log.Fatal(err)
	}
	files, err := dir.Readdir(-1)
	err = dir.Close()
	if err != nil {
		return
	}
	go func() {
		for {
			for name, buf := range pluginLogBufferMap {
				line, _ := buf.ReadString('\n')
				if len(line) == 0 {
					continue
				}
				fmt.Print('[', name, ']', line)
			}
		}
	}()
	go func() {
		for {
			for name, buf := range pluginOutBufferMap {
				line, _ := buf.ReadString('\n')
				if len(line) == 0 {
					continue
				}
				go pluginRecv(line, name)
			}
		}
	}()
	for _, file := range files {
		go func() {
			logFile, _ := os.OpenFile(file.Name()+".log", os.O_CREATE|os.O_WRONLY, os.ModePerm)
			pluginLogBufferMap[file.Name()] = bytes.NewBuffer(nil)
			pluginOutBufferMap[file.Name()] = bytes.NewBuffer(nil)
			pluginInBufferMap[file.Name()] = bytes.NewBuffer(nil)
			mw := io.MultiWriter(logFile, pluginLogBufferMap[file.Name()])
			cmd := exec.Command("./plugin/"+file.Name(), "./config/"+file.Name())
			cmd.Stdout = pluginOutBufferMap[file.Name()]
			cmd.Stderr = mw
			cmd.Stdin = pluginInBufferMap[file.Name()]
			runErr := cmd.Run()
			if runErr != nil {
				log.Println(runErr)
				return
			}
		}()
	}
}

func RecvEvent(data messages.Event) {
	var cmd []string
	var isCmd bool
	switch data.MessageType {
	case "messages":
		for _, msg := range data.MessageArray {
			if msg.Type == "text" {
				var text string
				for _, item := range config.CmdChar {
					text = msg.Get("text")
					for len(text) != 0 {
						if text[0] != ' ' {
							break
						}
						text = text[1:]
					}
					if len(text) == 0 {
						break
					}
					cmd = strings.Split(text, item)
					if cmd[0] == "" {
						name, ok := CmdPluginMap[cmd[1]]
						if ok {
							pluginSend(pluginInBufferMap[name], data)
							isCmd = true
						}
					}
				}
				break
			}
		}
		if !isCmd {
			for _, name := range MessagePluginMap {
				pluginSend(pluginInBufferMap[name], data)
			}
		}
	case "notice":
		for _, name := range NoticePluginMap[data.NoticeType] {
			pluginSend(pluginOutBufferMap[name], data)
		}
	case "request":
		log.Println("get request")
	case "meta_event":
		switch data.MetaEventType {
		case "lifecycle":
			log.Println("bot连接成功 ", data.SubType)
		}
	default:
		log.Println(data)
	}
}

func pluginSend(writer io.Writer, data messages.Event) {
	marshal, _ := json.Marshal(data)
	_, err := writer.Write(marshal)
	if err != nil {
		log.Println(err)
		return
	}
	_, err = writer.Write([]byte("\n"))
	if err != nil {
		log.Println(err)
		return
	}
}
