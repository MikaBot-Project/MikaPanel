package plugin

import (
	"MikaPanel/config"
	"MikaPanel/messages"
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

var pluginLogBufferMap map[string]*bufio.Reader
var pluginOutBufferMap map[string]*bufio.Reader
var pluginInBufferMap map[string]*bufio.Writer

var MessagePluginMap []string
var CmdPluginMap map[string]string
var NoticePluginMap map[string][]string

type intelMessage struct {
	MessageType string `json:"message_type"`
	SubType     string `json:"sub_type"`
	RawMessage  string `json:"raw_message"`
}

func init() {
	log.SetPrefix("[main] ")
	CmdPluginMap = make(map[string]string)
	NoticePluginMap = make(map[string][]string)
	pluginLogBufferMap = make(map[string]*bufio.Reader)
	pluginOutBufferMap = make(map[string]*bufio.Reader)
	pluginInBufferMap = make(map[string]*bufio.Writer)
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
	for _, file := range files { //启动插件线程
		go func() {
			logFile, _ := os.OpenFile(fmt.Sprintf("log/%s.log", file.Name()), os.O_CREATE|os.O_WRONLY, os.ModePerm)
			inReader, inWriter := io.Pipe()
			outReader, outWriter := io.Pipe()
			logReader, logWriter := io.Pipe()
			pluginLogBufferMap[file.Name()] = bufio.NewReader(logReader)
			pluginOutBufferMap[file.Name()] = bufio.NewReader(outReader)
			pluginInBufferMap[file.Name()] = bufio.NewWriter(inWriter)
			logWriters := io.MultiWriter(logFile, logWriter)
			// 创建可取消的上下文
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			cmd := exec.CommandContext(ctx, "./plugin/"+file.Name(), "config/"+file.Name())
			cmd.Stdout = outWriter
			cmd.Stderr = logWriters
			cmd.Stdin = inReader
			runErr := cmd.Start()
			if runErr != nil {
				log.Println(runErr)
				return
			}
			// 等待命令完成
			if err := cmd.Wait(); err != nil {
				// 如果是因为上下文取消而退出，这是预期的
				if ctx.Err() != nil && errors.Is(ctx.Err(), context.Canceled) {
					return
				}
				panic(err)
			}
		}()
	}
	go func() { //log线程
		for {
			for name, buf := range pluginLogBufferMap {
				line, _ := buf.ReadString('\n')
				if len(line) == 0 {
					continue
				}
				fmt.Print("[", name, "] ", line)
			}
		}
	}()
	go func() { //读取输出
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
			pluginSend(pluginInBufferMap[name], data)
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

func SendEcho(name, text string) {
	data := intelMessage{
		MessageType: "echo",
		SubType:     "echo",
		RawMessage:  text,
	}
	pluginSend(pluginInBufferMap[name], data)
}

func pluginSend(writer *bufio.Writer, data interface{}) {
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
	err = writer.Flush()
	if err != nil {
		return
	}
}
