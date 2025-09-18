package main

import (
	"MikaPanel/messages"
	"MikaPanel/plugin"
	"MikaPanel/web"
)

func main() {
	web.Start()
	go func() {
		var data messages.Event
		for {
			data = <-messages.EventChan
			plugin.RecvEvent(data)
		}
	}()
}
