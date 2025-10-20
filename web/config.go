package web

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

type configHandler struct {
	fileServer http.Handler
}

func (m *configHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/config/" {
		m.fileServer.ServeHTTP(w, r)
		return
	}
	if r.URL.Path[len(r.URL.Path)-1] == '/' {
		data, _ := json.Marshal(getDir(r.URL.Path))
		if len(data) == 4 {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write(data)
		if err != nil {
			log.Println("web config write error:", err)
			return
		}
	} else {
		m.fileServer.ServeHTTP(w, r)
	}
}

func getDir(path string) map[string]any {
	dir, err := os.ReadDir("." + path)
	if err != nil {
		return nil
	}
	result := make(map[string]any)
	for _, entry := range dir {
		dirName := entry.Name()
		if entry.IsDir() {
			result[dirName] = getDir(path + dirName + "/")
		} else {
			fileInfo, _ := entry.Info()
			result[dirName] = fileInfo.Size()
		}
	}
	return result
}

func getConfigHandler() http.Handler {
	return &configHandler{fileServer: http.FileServer(http.Dir("./"))}
}
