package plugin

import (
	"MikaPanel/util"
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
)

type operator string

const (
	stop    = operator("stop")
	start   = operator("start")
	restart = operator("restart")
)

func runPlugin(ctx context.Context, name string) {
	// 创建可取消的上下文
	ctx, cancel := context.WithCancel(ctx)

	//初始化线程
	logFile, _ := os.OpenFile(fmt.Sprintf("log/%s.log", name), os.O_CREATE|os.O_WRONLY, os.ModePerm)
	inReader, inWriter := io.Pipe()
	outReader, outWriter := io.Pipe()
	logReader, logWriter := io.Pipe()
	logOutBuffer := bufio.NewReader(logReader)
	outBuffer := bufio.NewReader(outReader)
	pluginInBufferMap[name] = bufio.NewWriter(inWriter)
	pluginInMutexMap[name] = new(sync.Mutex)
	logWriters := io.MultiWriter(logFile, logWriter)
	cmd := exec.CommandContext(ctx, "./plugin/"+name, "./config/"+name+"/")
	cmd.Stdout = outWriter
	cmd.Stderr = logWriters
	cmd.Stdin = inReader
	runErr := cmd.Start()
	if runErr != nil {
		log.Println(runErr)
		cancel()
		return
	}

	// 启动监听线程
	go func() { //log线程
		for {
			line, _ := logOutBuffer.ReadString('\n')
			if len(line) == 0 {
				continue
			}
			fmt.Print("[", name, "] ", line)
		}
	}()
	go func() { //读取输出
		for {
			line, _ := outBuffer.ReadString('\n')
			if len(line) == 0 {
				continue
			}
			go pluginRecv(line, name)
		}
	}()
	go func() { //plugin线程指令处理
		for {
			op := <-pluginOperatorChanMap[name]
			switch op {
			case stop:
				mutex := pluginInMutexMap[name]
				mutex.Lock()
				delete(pluginInBufferMap, name)
				unRegister(name)
				cancel()
				mutex.Unlock()
			case start:
				mutex := pluginInMutexMap[name]
				mutex.Lock()
				pluginInBufferMap[name] = bufio.NewWriter(inWriter)
				cmd = exec.CommandContext(ctx, "./plugin/"+name, "./config/"+name+"/")
				cmd.Stdout = outWriter
				cmd.Stderr = logWriters
				cmd.Stdin = inReader
				runErr = cmd.Start()
				if runErr != nil {
					log.Println(runErr)
					cancel()
					return
				}
				mutex.Unlock()
			case restart:
				mutex := pluginInMutexMap[name]
				mutex.Lock()
				unRegister(name)
				cancel()
				time.Sleep(2 * time.Second)
				cmd = exec.CommandContext(ctx, "./plugin/"+name, "./config/"+name+"/")
				cmd.Stdout = outWriter
				cmd.Stderr = logWriters
				cmd.Stdin = inReader
				runErr = cmd.Start()
				mutex.Unlock()
			}
		}
	}()
}

func unRegister(name string) {
	for i, n := range MessagePluginMap {
		if n == name {
			util.ArrayFastDelete(MessagePluginMap, i)
		}
	}
	for _, arr := range NoticePluginMap {
		for i, n := range arr {
			if n == name {
				util.ArrayFastDelete(arr, i)
			}
		}
	}
	for k, n := range CmdPluginMap {
		if n == name {
			delete(CmdPluginMap, k)
		}
	}
}

func StopPlugin(name string) {
	pluginOperatorChanMap[name] <- stop
}

func StartPlugin(name string) {
	pluginOperatorChanMap[name] <- start
}

func RestartPlugin(name string) {
	pluginOperatorChanMap[name] <- restart
}
