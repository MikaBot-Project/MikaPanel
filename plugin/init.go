package plugin

import (
	"MikaPanel/messages"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
)

var pluginLogBufferMap map[string]*bytes.Buffer
var pluginOutBufferMap map[string]*bytes.Buffer
var pluginInBufferMap map[string]*bytes.Buffer

func init() {
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
	messages.MessagePluginFunc = recvMsg
	messages.CmdPluginFunc = recvCmd
	messages.NoticePluginFunc = recvNotice
}

func recvMsg(data []byte) {
	for _, name := range messages.MessagePluginMap {
		pluginSend(pluginInBufferMap[name], data)
	}
}

func recvCmd(plugin string, data []byte) {
	pluginSend(pluginInBufferMap[plugin], data)
}

func recvNotice(plugins []string, data []byte) {
	for _, name := range plugins {
		pluginSend(pluginOutBufferMap[name], data)
	}
}

func pluginSend(writer io.Writer, data []byte) {
	_, err := writer.Write(data)
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
