package main

import (
	"MikaPanel/messages"
	"MikaPanel/plugin"
	"MikaPanel/web"
	"log"
	"time"
)

func main() {
	web.Start()
	var data messages.Event
	go func() {
		for {
			time.Sleep(1 * time.Second)
			log.Println("send echo")
			plugin.SendEcho("test.exe", "test input")
		}
	}()
	for {
		data = <-messages.EventChan
		plugin.RecvEvent(data)
	}
}
