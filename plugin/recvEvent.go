package plugin

import (
	"MikaPanel/config"
	"MikaPanel/messages"
	"MikaPanel/util"
	"encoding/json"
	"log"
	"sort"
	"strings"
)

func RecvEvent(data messages.Event) {
	switch data.PostType {
	case "message":
		data.AtMe = false
		var isCmd bool
		for _, msg := range data.MessageArray {
			switch msg.Type {
			case "text":
				var text = msg.GetString("text")
				args := strings.Split(text, " ")
				var cmd string
				for key, arg := range args {
					if arg == "" {
						continue
					} else {
						args = args[key:]
						cmd = arg
						break
					}
				}
				name, ok := CmdPluginMap[cmd]
				if ok {
					data.PostType = "command"
					data.CommandArgs = args
					if pluginPolicyCheck(name, int(data.GroupId)) {
						pluginSend(name, data)
					}
					isCmd = true
				}
			case "at":
				var at = msg.GetString("qq")
				if data.SelfId == util.StringToInt64(at) {
					data.AtMe = true
				}
			}
		}
		if !isCmd {
			for _, name := range MessagePluginMap {
				if pluginPolicyCheck(name, int(data.GroupId)) {
					pluginSend(name, data)
				}
			}
		}
	case "notice":
		for _, name := range NoticePluginMap[data.NoticeType] {
			if pluginPolicyCheck(name, int(data.GroupId)) {
				pluginSend(name, data)
			}
		}
	case "request":
		log.Println("get request")
		switch data.RequestType {
		case "friend":
			if data.UserId == config.AdminId {
				send := struct {
					Flag    string `json:"flag"`
					Approve bool   `json:"approve"`
					Remark  string `json:"remark"`
				}{Flag: data.Flag, Approve: true, Remark: ""}
				bytes, _ := json.Marshal(send)
				messages.Send(bytes, "set_friend_add_request")
			} else {
				send := struct {
					UserId int64 `json:"user_id"`
				}{UserId: data.UserId}
				bytes, _ := json.Marshal(send)
				bytes = messages.Send(bytes, "get_stranger_info")
				recv := struct {
					Data struct {
						Nickname string `json:"nickname"`
					} `json:"data"`
				}{}
				recv.Data.Nickname = ""
				err := json.Unmarshal(bytes, &recv)
				if err != nil {
					return
				}
				messages.SendMessage(recv.Data.Nickname+" "+util.Int64ToString(data.UserId)+"请求添加为好友", config.AdminId, 0)
			}
		case "group":
			switch data.SubType {
			case "invite":
				if data.UserId == config.AdminId {
					send := struct {
						Flag    string `json:"flag"`
						Approve bool   `json:"approve"`
						Reason  string `json:"reason"`
					}{Flag: data.Flag, Approve: true, Reason: ""}
					bytes, _ := json.Marshal(send)
					messages.Send(bytes, "set_group_add_request")
				} else {
					send := struct {
						UserId int64 `json:"user_id"`
					}{UserId: data.UserId}
					bytes, _ := json.Marshal(send)
					bytes = messages.Send(bytes, "get_stranger_info")
					recv := struct {
						Data struct {
							Nickname string `json:"nickname"`
						} `json:"data"`
					}{}
					recv.Data.Nickname = ""
					err := json.Unmarshal(bytes, &recv)
					if err != nil {
						return
					}
					messages.SendMessage(recv.Data.Nickname+" "+util.Int64ToString(data.UserId)+"请求拉入群聊"+util.Int64ToString(data.GroupId),
						config.AdminId, 0)
				}
			case "add":
				for _, name := range NoticePluginMap["group_add"] {
					data.PostType = "notice"
					data.NoticeType = "group_add"
					if pluginPolicyCheck(name, int(data.GroupId)) {
						pluginSend(name, data)
					}
				}
			}
		}
	case "meta_event":
		switch data.MetaEventType {
		case "lifecycle":
			log.Println("bot连接成功 ", data.SubType)
		case "heartbeat":

		}
	default:
		log.Println(data)
	}
}

func pluginPolicyCheck(name string, groupId int) bool {
	policy, ok := config.Policies[name]
	if !ok {
		return true
	}
	if groupId == 0 {
		return !policy.GroupOnly
	}
	if policy.Type == "black" {
		index := sort.SearchInts(policy.Groups, groupId)
		return !(index < len(policy.Groups) && policy.Groups[index] == groupId)
	}
	if policy.Type == "white" {
		index := sort.SearchInts(policy.Groups, groupId)
		return index < len(policy.Groups) && policy.Groups[index] == groupId
	}
	return true
}
