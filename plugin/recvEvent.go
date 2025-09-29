package plugin

import (
	"MikaPanel/config"
	"MikaPanel/messages"
	"MikaPanel/util"
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
					log.Println("get cmd " + cmd + " from plugin")
					data.PostType = "command"
					data.CommandArgs = args
					if pluginPolicyCheck(name, int(data.GroupId)) {
						pluginSend(pluginInBufferMap[name], data)
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
					pluginSend(pluginInBufferMap[name], data)
				}
			}
		}
	case "notice":
		for _, name := range NoticePluginMap[data.NoticeType] {
			log.Println("notice", data.NoticeType, data.SubType, name)
			if pluginPolicyCheck(name, int(data.GroupId)) {
				pluginSend(pluginInBufferMap[name], data)
			}
		}
	case "request":
		log.Println("get request")
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
