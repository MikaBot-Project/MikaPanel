package plugin

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var pluginInBufferMap map[string]*bufio.Writer
var pluginInMutexMap map[string]*sync.Mutex

var MessagePluginMap []string
var CmdPluginMap map[string]string
var NoticePluginMap map[string][]string
var pluginOperatorChanMap map[string]chan operator

type intelMessage struct {
	PostType    string `json:"post_type"`
	MessageType string `json:"message_type"`
	RawMessage  string `json:"raw_message"`
	SubType     string `json:"sub_type"`
}

func init() {
	log.SetPrefix("[main] ")
	CmdPluginMap = make(map[string]string)
	NoticePluginMap = make(map[string][]string)
	pluginInBufferMap = make(map[string]*bufio.Writer)
	pluginInMutexMap = make(map[string]*sync.Mutex)
	pluginOperatorChanMap = make(map[string]chan operator)
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
	err = os.MkdirAll("log", 0755)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal)

	// 注册要捕获的信号
	signal.Notify(sigChan,
		syscall.SIGINT,  // Ctrl+C
		syscall.SIGTERM, // 终止信号
		syscall.SIGQUIT, // 退出信号
	)
	go func() {
		<-sigChan
		cancel()
	}()
	for _, file := range files { //启动插件线程
		go func() {
			name := file.Name()
			pluginOperatorChanMap[name] = make(chan operator)
			runPlugin(ctx, name)
		}()
	}
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
	var writer, ok = pluginInBufferMap[name]
	if !ok {
		return
	}
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

func PluginConfigReload(name string) {
	data := intelMessage{
		PostType:    "operator",
		MessageType: "config",
		SubType:     "reload",
	}
	pluginSend(name, data)
}
