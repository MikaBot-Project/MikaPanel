package web

import (
	"MikaPanel/config"
	"MikaPanel/messages"
	"log"
	"net/http"

	"github.com/lxzan/gws"
)

var Mux *http.ServeMux
var upgrader *gws.Upgrader

func init() {
	Mux = http.NewServeMux()
	upgrader = gws.NewUpgrader(&SocketHandler{}, &gws.ServerOption{
		ParallelEnabled: true,
		Recovery:        gws.Recovery,
	})
	Mux.HandleFunc("/onebot/v11", func(writer http.ResponseWriter, request *http.Request) {
		conn, err := upgrader.Upgrade(writer, request)
		if err != nil {
			log.Println(err)
			return
		}
		go func() { // 接收数据
			conn.ReadLoop()
		}()
		go func() { //发送数据
			var data []byte
			for {
				data = <-messages.SendChan
				_ = conn.WriteMessage(gws.OpcodeText, data)
			}
		}()
		log.Println("websocket upgrade success")
	})
	Mux.Handle("/", http.FileServer(http.Dir("./web")))
}

func Start() {
	server := &http.Server{
		Addr:    config.Host,
		Handler: Mux,
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Println("web server start error", err)
		return
	}
}
