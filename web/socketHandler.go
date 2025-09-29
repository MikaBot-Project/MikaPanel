package web

import (
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
}

func (c *SocketHandler) OnOpen(socket *gws.Conn) {
	_ = socket.SetDeadline(time.Now().Add(PingInterval + PingWait))
	go func() { //发送数据
		var data []byte
		for {
			data = <-messages.SendChan
			err := socket.WriteMessage(gws.OpcodeText, data)
			if err != nil {
				log.Println(err)
				messages.SendChan <- data
				_ = socket.WriteClose(1000, nil)
				return
			}
		}
	}()
	go func() {
		for {
			time.Sleep(10 * time.Second)
			err := socket.WritePing([]byte(util.RandomString(8)))
			if err != nil {
				return
			}
		}
	}()
}

func (c *SocketHandler) OnClose(socket *gws.Conn, err error) {
	log.Println("websocket close")
	log.Println(err)
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
