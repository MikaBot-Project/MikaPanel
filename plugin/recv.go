package plugin

import (
	"MikaPanel/messages"
	"MikaPanel/util"
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

func pluginRecv(recvData string, name string) {
	data := strings.Split(recvData[:len(recvData)-1], ":##:")
	dataLen := len(data)
	switch data[0] {
	case "init":
		log.Println("plugin ", name, "init")
		for _, item := range data[1:] {
			log.Println(fmt.Sprintf("[%s] %s", name, item))
		}
	case "send_msg":
		if dataLen < 5 {
			log.Println(fmt.Sprintf("[%s] send_msg: args number lass than 5", name))
			return
		}
		var marshal = []byte(data[3])
		var err error
		if json.Valid(marshal) {
			var msg []messages.MessageItem
			_ = json.Unmarshal(marshal, &msg)
			marshal, err = json.Marshal(messages.SendMessage(msg, util.StringToInt64(data[1]), util.StringToInt64(data[2])))
		} else {
			marshal, err = json.Marshal(messages.SendMessage(data[3], util.StringToInt64(data[1]), util.StringToInt64(data[2])))
		}
		if err != nil {
			log.Println("json err:", err)
			return
		}
		send := intelMessage{
			PostType:    "return",
			MessageType: "send_msg",
			RawMessage:  string(marshal),
		}
		pluginSend(pluginInBufferMap[name], send)
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
		send := intelMessage{
			PostType:    "return",
			MessageType: "send_api",
			RawMessage:  string(messages.Send([]byte(data[1]), data[2])),
		}
		pluginSend(pluginInBufferMap[name], send)
	}
}
