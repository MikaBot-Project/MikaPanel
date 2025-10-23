package web

import (
	"MikaPanel/config"
	"MikaPanel/plugin"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type apiHandler struct{}

func (m *apiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" && r.URL.Path == "plugins" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		data := struct {
			Message []string            `json:"message"`
			Command map[string]string   `json:"command"`
			Notice  map[string][]string `json:"notice"`
			Plugins map[string]string   `json:"plugins"`
		}{
			Message: plugin.MessagePluginMap,
			Command: plugin.CmdPluginMap,
			Notice:  plugin.NoticePluginMap,
			Plugins: plugin.StatusMap,
		}
		marshal, err := json.Marshal(data)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			_, err = w.Write([]byte(err.Error()))
			if err != nil {
			}
			return
		}
		_, err = w.Write(marshal)
		if err != nil {
			log.Println(err)
			return
		}
		return
	}
	urlArgs := strings.Split(r.URL.Path, "/")
	if r.Method == "POST" && urlArgs[0] == "upload" {
		switch urlArgs[1] {
		case "config":
			if len(urlArgs) < 3 {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"message":"plugin name is required"}`))
				return
			}
			configPath := "./" + strings.Join(urlArgs[1:], "/")
			file, err := os.OpenFile(configPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, err = w.Write([]byte(err.Error()))
				if err != nil {
					return
				}
				return
			}
			defer func(file *os.File) {
				err = file.Close()
				if err != nil {
				}
			}(file)
			_, err = io.Copy(file, r.Body)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				_, err = w.Write([]byte(err.Error()))
				return
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("{\"message\":\"ok\"}"))
		case "plugin":
			pluginPath := "./" + strings.Join(urlArgs[1:], "/")
			status, ok := plugin.StatusMap[urlArgs[2]]
			if ok && status != "stopped" {
				plugin.StopPlugin(urlArgs[2])
				plugin.LockPluginMutex(urlArgs[2])
				defer plugin.UnlockPluginMutex(urlArgs[2])
			}
			file, err := os.OpenFile(pluginPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, err = w.Write([]byte(err.Error()))
			}
			defer func(file *os.File) {
				err = file.Close()
				if err != nil {
				}
			}(file)
			_, err = io.Copy(file, r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, err = w.Write([]byte(err.Error()))
				return
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("{\"message\":\"ok\"}"))
		default:
			http.NotFound(w, r)
		}
		return
	}
	if urlArgs[0] == "plugin" {
		switch urlArgs[1] {
		case "stop":
			plugin.StopPlugin(urlArgs[2])
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("{\"message\":\"ok\"}"))
		case "start":
			plugin.StartPlugin(urlArgs[2])
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("{\"message\":\"ok\"}"))
		case "reload":
			plugin.ConfigReload(urlArgs[2])
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("{\"message\":\"ok\"}"))
		case "restart":
			plugin.RestartPlugin(urlArgs[2])
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("{\"message\":\"ok\"}"))
		default:
			http.NotFound(w, r)
		}
		return
	}
	if urlArgs[0] == "panel" {
		switch urlArgs[1] {
		case "reload":
			log.Println("Reload config")
			config.LoadConfig()
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("{\"message\":\"ok\"}"))
		case "policies":
			if len(urlArgs) < 3 {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"message":"plugin name is required"}`))
				return
			}
			policy := config.PluginPolicy{
				Type:      "",
				GroupOnly: false,
			}
			var bytes []byte
			_, err := io.ReadFull(r.Body, bytes)
			if err != nil {
				return
			}
			err = json.Unmarshal(bytes, &policy)
			config.Policies[urlArgs[2]] = policy
			config.SaveConfig()
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("{\"message\":\"ok\"}"))
		default:
			http.NotFound(w, r)
		}
		return
	}
	http.NotFound(w, r)
}
