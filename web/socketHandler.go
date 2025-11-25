package web

import (
	"MikaPanel/config"
	"MikaPanel/messages"
	"MikaPanel/util"
	"log"
	"time"

	"github.com/lxzan/gws"
)

const (
	PingInterval = 5 * time.Second
	PingWait     = 100 * time.Second
)

type SocketHandler struct {
	selfId int64
}

func (c *SocketHandler) OnOpen(socket *gws.Conn) {
	_ = socket.SetDeadline(time.Now().Add(PingInterval + PingWait))
	messages.SendChan[c.selfId] = make(chan []byte, 5)
	var i = 0
	for i < config.SendThreadCount {
		go func() { //发送数据
			var data []byte
			for {
				data = <-messages.SendChan[c.selfId]
				err := socket.WriteMessage(gws.OpcodeText, data)
				if err != nil {
					log.Println("Send data err:", err)
					messages.SendChan[c.selfId] <- data
					_ = socket.WriteClose(1000, nil)
					return
				}
			}
		}()
		i++
	}
	go func() {
		for {
			time.Sleep(10 * time.Second)
			err := socket.WritePing([]byte(util.RandomString(8)))
			if err != nil {
				log.Println("Send ping err:", err)
				return
			}
		}
	}()
}

func (c *SocketHandler) OnClose(socket *gws.Conn, err error) {
	delete(messages.SendChan, c.selfId)
	log.Println("websocket close with err:", err)
}

func (c *SocketHandler) OnPing(socket *gws.Conn, payload []byte) {
	log.Println("websocket ping")
	_ = socket.SetDeadline(time.Now().Add(PingInterval + PingWait))
	_ = socket.WritePong(payload)
}

func (c *SocketHandler) OnPong(socket *gws.Conn, payload []byte) {
	_ = socket.SetDeadline(time.Now().Add(PingInterval + PingWait))
}

func (c *SocketHandler) OnMessage(socket *gws.Conn, message *gws.Message) {
	defer func(message *gws.Message) {
		err := message.Close()
		if err != nil {
			log.Println(err)
		}
	}(message)
	switch message.Opcode {
	case gws.OpcodeText:
		messages.RecvChan <- message.Bytes()
	case gws.OpcodePing:
		c.OnPing(socket, message.Bytes())
	}
}
