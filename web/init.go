package web

import (
	"MikaPanel/config"
	"MikaPanel/messages"
	"log"
	"net/http"
)

var Mux *http.ServeMux

func init() {
	Mux = http.NewServeMux()
	Mux.HandleFunc("/onebot/v11/", func(writer http.ResponseWriter, request *http.Request) {
		bodyLen := request.ContentLength
		body := make([]byte, bodyLen)
		_, err := request.Body.Read(body)
		if err != nil {
			log.Println("read post data err:", err)
			return
		}
		err = messages.MessageHandler(body)
		if err != nil {
			return
		}
		writer.WriteHeader(http.StatusOK)
	})
	Mux.Handle("/", http.FileServer(http.Dir("./web")))
}

func Start() {
	server := &http.Server{
		Addr:    ":" + config.WebPort,
		Handler: Mux,
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Println("web server start error", err)
		return
	}
}
