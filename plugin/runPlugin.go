package plugin

import (
	"MikaPanel/util"
	"bufio"
	"context"
	"errors"
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

func RunPlugin(ctx context.Context, name string) {
	// 创建可取消的上下文
	ctxCmd, cancel := context.WithCancel(ctx)
	pluginOperatorChanMap[name] = make(chan operator)

	//初始化线程
	logFile, _ := os.OpenFile(fmt.Sprintf("log/%s.log", name), os.O_CREATE|os.O_WRONLY, os.ModePerm)
	inReader, inWriter := io.Pipe()
	outReader, outWriter := io.Pipe()
	logReader, logWriter := io.Pipe()
	logOutBuffer := bufio.NewReader(logReader)
	outBuffer := bufio.NewReader(outReader)
	pluginInBufferMap.Set(name, bufio.NewWriter(inWriter))
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
			line, _ := outBuffer.ReadBytes('\n')
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
				if StatusMap[name] == "stopped" {
					log.Println("plugin is already stopped")
					break
				}
				log.Println("plugin stoping")
				mutex := pluginInMutexMap[name]
				mutex.Lock()
				pluginInBufferMap.Delete(name)
				unRegister(name)
				cancel()
				StatusMap[name] = "stopped"
				mutex.Unlock()
				log.Println("plugin stopped")
			case start:
				if StatusMap[name] == "running" {
					log.Println("plugin is already started")
					break
				}
				log.Println("plugin starting")
				mutex := pluginInMutexMap[name]
				mutex.Lock()
				inReader, inWriter = io.Pipe()
				pluginInBufferMap.Set(name, bufio.NewWriter(inWriter))
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
				go cmdErrListener(name, cmd, ctxCmd)
				StatusMap[name] = "running"
				mutex.Unlock()
				log.Println("plugin started")
			case restart:
				log.Println("plugin restarting")
				mutex := pluginInMutexMap[name]
				mutex.Lock()
				StatusMap[name] = "restarting"
				unRegister(name)
				cancel()
				time.Sleep(2 * time.Second)
				ctxCmd, cancel = context.WithCancel(ctx)
				inReader, inWriter = io.Pipe()
				pluginInBufferMap.Set(name, bufio.NewWriter(inWriter))
				cmd = exec.CommandContext(ctxCmd, cmdArgs[0], cmdArgs[1:]...)
				cmd.Stdout = outWriter
				cmd.Stderr = logWriters
				cmd.Stdin = inReader
				runErr = cmd.Start()
				if runErr != nil {
					log.Println(runErr)
					cancel()
				}
				go cmdErrListener(name, cmd, ctxCmd)
				StatusMap[name] = "running"
				mutex.Unlock()
				log.Println("plugin restarted")
			}
		}
	}()
	go cmdErrListener(name, cmd, ctxCmd)
	StatusMap[name] = "running"
}

func cmdErrListener(name string, cmd *exec.Cmd, cmdCtx context.Context) {
	if err := cmd.Wait(); err != nil {
		if !errors.Is(cmdCtx.Err(), context.Canceled) {
			log.Println("plugin exited by err:", err.Error())
		}
	}
	mutex := pluginInMutexMap[name]
	mutex.Lock()
	pluginInBufferMap.Delete(name)
	unRegister(name)
	StatusMap[name] = "stopped"
	mutex.Unlock()
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
	CmdPluginMap.Range(func(k string, n string) bool {
		if n == name {
			CmdPluginMap.Delete(k)
		}
		return true
	})
}

func StopPlugin(name string) {
	_, ok := StatusMap[name]
	if !ok {
		return
	}
	pluginOperatorChanMap[name] <- stop
}

func StartPlugin(name string) {
	_, ok := StatusMap[name]
	if !ok {
		RunPlugin(ctx, name)
		return
	}
	pluginOperatorChanMap[name] <- start
}

func RestartPlugin(name string) {
	_, ok := StatusMap[name]
	if !ok {
		RunPlugin(ctx, name)
		return
	}
	pluginOperatorChanMap[name] <- restart
}
