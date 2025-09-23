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
		if data[1] != "v0.1.0" {
			log.Println("plugin ", name, "Mismatch of library version")
			return
		}
		for _, item := range data[1:] {
			log.Println(fmt.Sprintf("[%s] %s", name, item))
		}
	case "send_msg": //send_msg <userId> <groupId> <message> <echo>
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
			MessageType: data[4],
			RawMessage:  string(marshal),
		}
		log.Println("plugin", name, "send msg:", data[3])
		pluginSend(pluginInBufferMap[name], send)
	case "send_poke": //send_poke <userId> <groupId>
		if dataLen < 3 {
			log.Println(fmt.Sprintf("[%s] send_poke: args number lass than 3", name))
			return
		}
		messages.SendPoke(util.StringToInt64(data[1]), util.StringToInt64(data[2]))
	case "send_api": //send_api <api_name> <data> <echo>
		if dataLen < 4 {
			log.Println(fmt.Sprintf("[%s] send_api: args number lass than 4", name))
			return
		}
		send := intelMessage{
			PostType:    "return",
			MessageType: data[3],
			RawMessage:  string(messages.Send([]byte(data[2]), data[1])),
		}
		pluginSend(pluginInBufferMap[name], send)
	case "register": //register <type> <args>
		switch data[1] {
		case "message":
			log.Println(name, "register message")
			MessagePluginMap = append(MessagePluginMap, name)
		case "cmd":
			log.Println(name, "register cmd", data[2])
			CmdPluginMap[data[2]] = name
		case "notice":
			log.Println(name, "register notice", data[2])
			NoticePluginMap[data[2]] = append(NoticePluginMap[data[2]], name)
		}
	}
}
