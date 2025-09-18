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

func init() {
	EventChan = make(chan Event, 10)
}

func MessageHandler(data []byte) error {
	var event Event
	err := json.Unmarshal(data, &event)
	if err != nil {
		log.Println("json:", err)
		return err
	}
	EventChan <- event
	return nil
}

type jsonWriter struct {
	data *[]byte
}

func (p jsonWriter) Write(b []byte) (n int, err error) {
	*p.data = append(*p.data, b...)
	return len(b), nil
}

func SendPoke(userId, groupId int64) {
	if groupId == 0 {
		data := struct {
			UserId int64 `json:"user_id"`
		}{userId}
		send, _ := json.Marshal(data)
		SendPost(send, "friend_poke")
	} else {
		data := struct {
			GroupId int64 `json:"group_id"`
			UserId  int64 `json:"user_id"`
		}{groupId, userId}
		send, _ := json.Marshal(data)
		SendPost(send, "group_poke")
	}
}

func GetMsg(MsgId int64) Message {
	data := struct {
		MessageId int64 `json:"message_id"`
	}{MsgId}
	send, _ := json.Marshal(data)
	rev := SendPost(send, "get_msg")
	var respDataStruct sendMessageResponse
	err := json.Unmarshal(rev, &respDataStruct)
	if err != nil {
		log.Println("GetMsg json err", err)
		log.Println(string(rev))
		return Message{}
	}
	return respDataStruct.Data
}
