package plugin

import (
	"MikaPanel/config"
	"MikaPanel/messages"
	"log"
	"strings"
)

func RecvEvent(data messages.Event) {
	var cmd []string
	var isCmd bool
	switch data.PostType {
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
