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
	ctxCmd, cancel := context.WithCancel(ctx)

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
	cmdArgs := []string{"./plugin/" + name, "./config/" + name + "/", "./data/" + name + "/"}
	cmd := exec.CommandContext(ctxCmd, cmdArgs[0], cmdArgs[1:]...)
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
				log.Println("plugin stoping")
				mutex := pluginInMutexMap[name]
				mutex.Lock()
				delete(pluginInBufferMap, name)
				unRegister(name)
				cancel()
				PluginMap[name] = "stopped"
				mutex.Unlock()
				log.Println("plugin stopped")
			case start:
				log.Println("plugin starting")
				mutex := pluginInMutexMap[name]
				mutex.Lock()
				pluginInBufferMap[name] = bufio.NewWriter(inWriter)
				ctxCmd, cancel = context.WithCancel(ctx)
				cmd = exec.CommandContext(ctxCmd, cmdArgs[0], cmdArgs[1:]...)
				cmd.Stdout = outWriter
				cmd.Stderr = logWriters
				cmd.Stdin = inReader
				runErr = cmd.Start()
				if runErr != nil {
					log.Println(runErr)
					cancel()
				}
				PluginMap[name] = "running"
				mutex.Unlock()
				log.Println("plugin started")
			case restart:
				log.Println("plugin restarting")
				mutex := pluginInMutexMap[name]
				mutex.Lock()
				PluginMap[name] = "restarting"
				unRegister(name)
				cancel()
				time.Sleep(2 * time.Second)
				ctxCmd, cancel = context.WithCancel(ctx)
				cmd = exec.CommandContext(ctxCmd, cmdArgs[0], cmdArgs[1:]...)
				cmd.Stdout = outWriter
				cmd.Stderr = logWriters
				cmd.Stdin = inReader
				runErr = cmd.Start()
				if runErr != nil {
					log.Println(runErr)
					cancel()
				}
				PluginMap[name] = "running"
				mutex.Unlock()
				log.Println("plugin restarted")
			}
		}
	}()
}

func unRegister(name string) {
	for i, n := range MessagePluginMap {
		if n == name {
			MessagePluginMap = util.ArrayFastDelete(MessagePluginMap, i)
			break
		}
	}
	for key, arr := range NoticePluginMap {
		for i, n := range arr {
			if n == name {
				NoticePluginMap[key] = util.ArrayFastDelete(arr, i)
				break
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
