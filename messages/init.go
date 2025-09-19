package messages

import (
	"MikaPanel/config"
	"encoding/json"
	"log"
)

type MessageItem struct {
	Type string         `json:"type"`
	Data map[string]any `json:"data"`
}

func (item *MessageItem) Get(name string) string {
	if val, ok := item.Data[name]; ok {
		return val.(string)
	}
	return ""
}

func (item *MessageItem) Set(name string, value any) {
	if item.Data == nil {
		item.Data = make(map[string]any)
	}
	item.Data[name] = value
}

type Event struct {
	Time          int64         `json:"time"`
	SelfId        int64         `json:"self_id"`
	PostType      string        `json:"post_type"`
	UserId        int64         `json:"user_id"`
	GroupId       int64         `json:"group_id"`
	MessageType   string        `json:"message_type"`
	SubType       string        `json:"sub_type"`
	MessageId     int64         `json:"message_id"`
	MessageArray  []MessageItem `json:"messages"`
	RawMessage    string        `json:"raw_message"`
	NoticeType    string        `json:"notice_type"`
	TargetId      int64         `json:"target_id"`
	MetaEventType string        `json:"meta_event_type"`
}

type Message struct {
	Time         int64         `json:"time"`
	SelfId       int64         `json:"self_id"`
	UserId       int64         `json:"user_id"`
	GroupId      int64         `json:"group_id"`
	MessageType  string        `json:"message_type"`
	SubType      string        `json:"sub_type"`
	MessageId    int64         `json:"message_id"`
	MessageArray []MessageItem `json:"messages"`
	RawMessage   string        `json:"raw_message"`
}

type sendMessageResponse struct {
	Status  string  `json:"status"`
	Retcode int     `json:"retcode"`
	Data    Message `json:"data"`
}

var httpUrl = config.NapcatHost

var EventChan chan Event
var SendChan chan []byte
var RecvChan chan []byte
var SendRecvMap map[string][]byte

func init() {
	EventChan = make(chan Event, 10)
	SendChan = make(chan []byte, 10)
	RecvChan = make(chan []byte, 10)
	go func() {
		var data []byte
		recv := struct {
			Status string `json:"status"`
			Echo   string `json:"echo"`
		}{
			Status: "event",
			Echo:   "",
		}
		for {
			data = <-RecvChan
			err := json.Unmarshal(data, &recv)
			if err != nil {
				log.Println("json err:", err)
				return
			}
			if recv.Status == "ok" {
				SendRecvMap[recv.Echo] = data
			} else {
				var event Event
				err = json.Unmarshal(data, &event)
				if err != nil {
					log.Println("json:", err)
					return
				}
				EventChan <- event
			}
		}
	}()
}

func GetMsg(MsgId int64) Message {
	data := struct {
		MessageId int64 `json:"message_id"`
	}{MsgId}
	send, _ := json.Marshal(data)
	rev := Send(send, "get_msg")
	var respDataStruct sendMessageResponse
	err := json.Unmarshal(rev, &respDataStruct)
	if err != nil {
		log.Println("GetMsg json err", err)
		log.Println(string(rev))
		return Message{}
	}
	return respDataStruct.Data
}
