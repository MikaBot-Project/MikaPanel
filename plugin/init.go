package plugin

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
)

var pluginLogBufferMap map[string]*bufio.Reader
var pluginOutBufferMap map[string]*bufio.Reader
var pluginInBufferMap map[string]*bufio.Writer
var pluginInMutexMap map[string]*sync.Mutex

var MessagePluginMap []string
var CmdPluginMap map[string]string
var NoticePluginMap map[string][]string

type intelMessage struct {
	PostType    string `json:"post_type"`
	MessageType string `json:"message_type"`
	RawMessage  string `json:"raw_message"`
}

func init() {
	log.SetPrefix("[main] ")
	CmdPluginMap = make(map[string]string)
	NoticePluginMap = make(map[string][]string)
	pluginLogBufferMap = make(map[string]*bufio.Reader)
	pluginOutBufferMap = make(map[string]*bufio.Reader)
	pluginInBufferMap = make(map[string]*bufio.Writer)
	pluginInMutexMap = make(map[string]*sync.Mutex)
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
			pluginInMutexMap[file.Name()] = new(sync.Mutex)
			logWriters := io.MultiWriter(logFile, logWriter)
			// 创建可取消的上下文
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			cmd := exec.CommandContext(ctx, "./plugin/"+file.Name(), "./config/"+file.Name()+"/")
			cmd.Stdout = outWriter
			cmd.Stderr = logWriters
			cmd.Stdin = inReader
			runErr := cmd.Start()
			if runErr != nil {
				log.Println(runErr)
				return
			}
			// 等待命令完成
			if err = cmd.Wait(); err != nil {
				// 如果是因为上下文取消而退出，这是预期的
				if ctx.Err() != nil && errors.Is(ctx.Err(), context.Canceled) {
					return
				}
				log.Println(file.Name(), err)
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

func SendEcho(name, text string) {
	data := intelMessage{
		PostType:    "echo",
		MessageType: "echo",
		RawMessage:  text,
	}
	pluginSend(name, data)
}

func pluginSend(name string, data interface{}) {
	marshal, _ := json.Marshal(data)
	mutex := pluginInMutexMap[name]
	mutex.Lock()
	defer mutex.Unlock()
	var writer = pluginInBufferMap[name]
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
