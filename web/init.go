package web

import (
	"MikaPanel/config"
	"MikaPanel/util"
	"log"
	"net/http"

	"github.com/lxzan/gws"
)

var Mux *http.ServeMux

func init() {
	Mux = http.NewServeMux()
	Mux.HandleFunc("/onebot/v11", func(writer http.ResponseWriter, request *http.Request) {
		upgrader := gws.NewUpgrader(&SocketHandler{
			selfId: util.StringToInt64(request.Header.Get("X-Self-ID")),
		}, &gws.ServerOption{
			ParallelEnabled: true,
			Recovery:        gws.Recovery,
		})
		conn, err := upgrader.Upgrade(writer, request)
		if err != nil {
			log.Println(err)
			return
		}
		go func() { // 接收数据
			conn.ReadLoop()
		}()
		log.Println("websocket upgrade success")
	})
	Mux.Handle("/api/config/", http.StripPrefix("/api", getConfigHandler()))
	Mux.Handle("/api/", http.StripPrefix("/api/", &apiHandler{}))
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
