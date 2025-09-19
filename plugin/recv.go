package plugin

import (
	"MikaPanel/messages"
	"MikaPanel/util"
	"fmt"
	"log"
	"strings"
)

func pluginRecv(recvData string, name string) {
	data := strings.Split(recvData, ":##:")
	dataLen := len(data)
	switch data[0] {
	case "init":
		log.Println("plugin ", name, "init")
		for _, item := range data[1:] {
			log.Println(fmt.Sprintf("[%s] %s", name, item))
		}
	case "send_msg":
		if dataLen < 4 {
			log.Println(fmt.Sprintf("[%s] send_msg: args number lass than 4", name))
			return
		}
		messages.SendMessage(data[3], util.StringToInt64(data[1]), util.StringToInt64(data[2]))
	case "send_poke":
		if dataLen < 3 {
			log.Println(fmt.Sprintf("[%s] send_poke: args number lass than 3", name))
			return
		}
		messages.SendPoke(util.StringToInt64(data[1]), util.StringToInt64(data[2]))
	case "send_api":
		if dataLen < 3 {
			log.Println(fmt.Sprintf("[%s] send_api: args number lass than 3", name))
			return
		}
		messages.Send([]byte(data[1]), data[2])
	}
}
