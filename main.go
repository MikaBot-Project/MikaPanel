package main

import (
	"MikaPanel/messages"
	"MikaPanel/plugin"
	"MikaPanel/web"
	"log"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()
	go web.Start()
	var data messages.Event
	for {
		data = <-messages.EventChan
		go plugin.RecvEvent(data)
	}
}
