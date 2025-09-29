package messages

import (
	"encoding/json"
	"log"
)

type MessageItem struct {
	Type string         `json:"type"`
	Data map[string]any `json:"data"`
}

func (item *MessageItem) GetString(name string) string {
	if val, ok := item.Data[name]; ok {
		return val.(string)
	}
	return ""
}

func (item *MessageItem) GetNumber(name string) int {
	if val, ok := item.Data[name]; ok {
		return val.(int)
	}
	return 0
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
	MessageArray  []MessageItem `json:"message"`
	RawMessage    string        `json:"raw_message"`
	NoticeType    string        `json:"notice_type"`
	TargetId      int64         `json:"target_id"`
	MetaEventType string        `json:"meta_event_type"`
	AtMe          bool          `json:"at_me"`
	CommandArgs   []string      `json:"command_args"`
}

type Message struct {
	Time         int64         `json:"time"`
	SelfId       int64         `json:"self_id"`
	UserId       int64         `json:"user_id"`
	GroupId      int64         `json:"group_id"`
	MessageType  string        `json:"message_type"`
	SubType      string        `json:"sub_type"`
	MessageId    int64         `json:"message_id"`
	MessageArray []MessageItem `json:"message"`
	RawMessage   string        `json:"raw_message"`
}

type sendMessageResponse struct {
	Status  string `json:"status"`
	Retcode int    `json:"retcode"`
	Data    struct {
		MessageId int64 `json:"message_id"`
	} `json:"data"`
}

var EventChan chan Event
var SendChan chan []byte
var RecvChan chan []byte
var sendRecvMap map[string][]byte

func init() {
	EventChan = make(chan Event, 10)
	SendChan = make(chan []byte, 10)
	RecvChan = make(chan []byte, 10)
	sendRecvMap = make(map[string][]byte)
	go func() {
		var data []byte
		recv := struct {
			Status any    `json:"status"`
			Echo   string `json:"echo"`
		}{
			Status: "event",
			Echo:   "",
		}
		for {
			data = <-RecvChan
			recv.Status = "event"
			err := json.Unmarshal(data, &recv)
			if err != nil {
				log.Println("data recv:", err)
				continue
			}
			switch recv.Status.(type) {
			case string:
				switch recv.Status.(string) {
				case "event":
					var event Event
					err = json.Unmarshal(data, &event)
					if err != nil {
						log.Println("data recv:", err)
						continue
					}
					EventChan <- event
				case "ok":
					sendRecvMap[recv.Echo] = data
				default:
					returnMsg := struct {
						Message string `json:"status"`
					}{}
					err = json.Unmarshal(data, &returnMsg)
					if err != nil {
						log.Println("data recv:", err)
						return
					}
					log.Println("return Status:", recv.Status)
					log.Println("return Msg:", returnMsg.Message)
					sendRecvMap[recv.Echo] = data
				}
			default:
				var event Event
				err = json.Unmarshal(data, &event)
				if err != nil {
					log.Println("data recv:", err)
					continue
				}
				EventChan <- event
			}
		}
	}()
}
