package web

import (
	"MikaPanel/messages"
	"log"
	"time"

	"github.com/lxzan/gws"
)

const (
	PingInterval = 5 * time.Second
	PingWait     = 10 * time.Second
)

type SocketHandler struct{}

func (c *SocketHandler) OnOpen(socket *gws.Conn) {
	_ = socket.SetDeadline(time.Now().Add(PingInterval + PingWait))
}

func (c *SocketHandler) OnClose(socket *gws.Conn, err error) {}

func (c *SocketHandler) OnPing(socket *gws.Conn, payload []byte) {
	_ = socket.SetDeadline(time.Now().Add(PingInterval + PingWait))
	_ = socket.WritePong(nil)
}

func (c *SocketHandler) OnPong(socket *gws.Conn, payload []byte) {}

func (c *SocketHandler) OnMessage(socket *gws.Conn, message *gws.Message) {
	defer func(message *gws.Message) {
		err := message.Close()
		if err != nil {
			log.Println(err)
		}
	}(message)
	if message.Opcode == gws.OpcodeText {
		messages.RecvChan <- message.Bytes()
	}
}
