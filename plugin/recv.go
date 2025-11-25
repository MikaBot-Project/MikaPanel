package plugin

import (
	"MikaPanel/messages"
	"encoding/json"
	"log"
)

type dataType struct {
	Action       string   `json:"action"`
	Echo         []byte   `json:"echo"`
	UserId       int64    `json:"user_id"`
	SelfId       int64    `json:"self_id"`
	GroupId      int64    `json:"group_id"`
	ApiName      string   `json:"api_name"`
	Data         []byte   `json:"data"`
	RegisterType string   `json:"register_type"`
	SubType      string   `json:"sub_type"`
	Arguments    []string `json:"arguments"`
}

func pluginRecv(recvData []byte, name string) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()
	var data dataType
	err := json.Unmarshal(recvData, &data)
	if err != nil {
		log.Println(err.Error())
		return
	}
	if data.SelfId == 0 {
		data.SelfId = selfId
	}
	switch data.Action {
	case "send_msg": //send_msg <userId> <groupId> <data> <sub_type> <echo>
		var marshal []byte
		if data.SubType == "array" {
			var msg []messages.MessageItem
			err = json.Unmarshal(data.Data, &msg)
			if err != nil {
				marshal, err = json.Marshal(messages.SendMessage(string(data.Data), data.UserId, data.GroupId, data.SelfId))
			} else {
				marshal, err = json.Marshal(messages.SendMessage(msg, data.UserId, data.GroupId, data.SelfId))
			}
		} else {
			marshal, err = json.Marshal(messages.SendMessage(string(data.Data), data.UserId, data.GroupId, data.SelfId))
		}
		if err != nil {
			log.Println("json err:", err)
			return
		}
		log.Println("plugin", name, "send msg:", string(data.Data))
		sendPluginResp(name, string(marshal), string(data.Echo))
	case "send_poke": //send_poke <userId> <groupId>
		log.Println("plugin", name, "send poke:", data.UserId, data.GroupId)
		messages.SendPoke(data.UserId, data.GroupId, data.SelfId)
	case "send_api": //send_api <api_name> <data> <echo>
		sendPluginResp(name, string(messages.SendData(data.Data, data.ApiName, data.Echo, data.SelfId)), string(data.Echo))
	case "register": //register <type> <sub_type>
		switch data.RegisterType {
		case "message":
			log.Println(name, "register message")
			MessagePluginMap = append(MessagePluginMap, name)
		case "command":
			log.Println(name, "register cmd", data.SubType)
			CmdPluginMap.Set(data.SubType, name)
		case "notice":
			log.Println(name, "register notice", data.SubType)
			NoticePluginMap[data.SubType] = append(NoticePluginMap[data.SubType], name)
		}
	case "operator": //operator <api_name> <sub_type> <arguments...>
		if data.ApiName == "panel" {
			pluginSend(name, panelOperator(data.SubType, data.Arguments))
		} else {
			send := struct {
				PostType    string   `json:"post_type"`
				MessageType string   `json:"message_type"`
				SubType     string   `json:"sub_type"`
				CommandArgs []string `json:"command_args"`
			}{
				PostType:    "operator",
				MessageType: data.SubType,
				SubType:     name,
				CommandArgs: data.Arguments,
			}
			pluginSend(data.ApiName, send)
		}
	}
}

func sendPluginResp(name, data, echo string) {
	send := intelMessage{
		PostType:    "return",
		MessageType: echo,
		RawMessage:  data,
	}
	pluginSend(name, send)
}
