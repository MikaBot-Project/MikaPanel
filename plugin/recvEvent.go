package plugin

import (
	"MikaPanel/messages"
	"log"
	"strings"
)

func RecvEvent(data messages.Event) {
	switch data.PostType {
	case "message":
		var isCmd bool
		for _, msg := range data.MessageArray {
			if msg.Type == "text" {
				var text = msg.Get("text")
				args := strings.Split(text, " ")
				var cmd string
				for _, arg := range args {
					if arg == "" {
						continue
					} else {
						cmd = arg
						break
					}
				}
				name, ok := CmdPluginMap[cmd]
				if ok {
					log.Println("get cmd " + cmd + " from plugin")
					data.PostType = "command"
					data.MetaEventType = cmd
					pluginSend(pluginInBufferMap[name], data)
					isCmd = true
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
			log.Println("notice", data.NoticeType, data.SubType, name)
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
