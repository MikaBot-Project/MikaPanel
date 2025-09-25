package messages

import (
	"MikaPanel/util"
	"bytes"
	"encoding/json"
	"time"
)

func Send(sendParams []byte, api string) []byte {
	send := struct {
		Action string `json:"action"`
		Params string `json:"params"`
		Echo   string `json:"echo"`
	}{
		Action: api,
		Params: "p",
		Echo:   util.RandomString(64),
	}
	data, _ := json.Marshal(send)
	data = bytes.Replace(data, []byte("\"p\""), sendParams, 1)
	SendChan <- data
	defer delete(sendRecvMap, send.Echo)
	var exists = false
	for !exists {
		time.Sleep(1 * time.Second)
		_, exists = sendRecvMap[send.Echo]
	}
	return sendRecvMap[send.Echo]
}

func sendMsg(data any, api string) (messageId int32) {
	var send []byte
	send, err := json.Marshal(data)
	if err != nil {
		return 0
	}
	respData := Send(send, api)
	var respDataStruct sendMessageResponse
	err = json.Unmarshal(respData, &respDataStruct)
	if err != nil {
		return 0
	}
	return int32(respDataStruct.Data.MessageId)
}

func SendMessage[T string | []MessageItem](msg T, userId int64, groupId int64) (messageId []int32) {
	message := any(msg)
	res := make([]int32, 0)
	switch message.(type) {
	case []MessageItem:
		length := len(message.([]MessageItem))
		start := 0
		for i := 0; i < length; i++ {
			switch message.([]MessageItem)[i].Type {
			case "record":
				res = append(res, sendMessage(message.([]MessageItem)[start:i], userId, groupId))
				start = i + 1
				res = append(res, sendMessage(message.([]MessageItem)[i:start], userId, groupId))
			}
		}
		if start != length {
			res = append(res, sendMessage(message.([]MessageItem)[start:], userId, groupId))
		}
	case string:
		res = append(res, sendMessage(msg, userId, groupId))
	}
	return res
}

func sendMessage[T string | []MessageItem](msg T, userId int64, groupId int64) (messageId int32) {
	if groupId == 0 {
		return SendPrivateMessage(msg, userId)
	} else {
		return SendGroupMessage(msg, groupId)
	}
}

func SendPrivateMessage[T string | []MessageItem](msg T, userId int64) (messageId int32) {
	data := struct {
		UserId  int64 `json:"user_id"`
		Message T     `json:"message"`
	}{userId, msg}
	return sendMsg(data, "send_private_msg")
}

func SendGroupMessage[T string | []MessageItem](msg T, groupId int64) (messageId int32) {
	data := struct {
		GroupId int64 `json:"group_id"`
		Message T     `json:"message"`
	}{groupId, msg}
	return sendMsg(data, "send_group_msg")
}

func SendPoke(userId, groupId string) {
	if groupId == "0" {
		data := struct {
			UserId   string `json:"user_id"`
			TargetId string `json:"target_id"`
		}{userId, userId}
		send, _ := json.Marshal(data)
		Send(send, "friend_poke")
	} else {
		data := struct {
			GroupId string `json:"group_id"`
			UserId  string `json:"user_id"`
		}{groupId, userId}
		send, _ := json.Marshal(data)
		Send(send, "group_poke")
	}
}
