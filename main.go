package main

import (
	"MikaPanel/messages"
	"MikaPanel/plugin"
	"MikaPanel/web"
)

func main() {
	go web.Start()
	var data messages.Event
	for {
		data = <-messages.EventChan
		go plugin.RecvEvent(data)
	}
}
