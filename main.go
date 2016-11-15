package main

import (
	"net/http"

	"tiberious/handlers"
	"tiberious/logger"
	"tiberious/settings"

	"github.com/gorilla/websocket"
)

var config settings.Config

// TODO move this into a separate handler of some form.
func newConnection(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  config.ReadBufferSize,
		WriteBufferSize: config.WriteBufferSize,
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); ok {
			w.WriteHeader(404)
			w.Write([]byte("Invalid websocket handshake"))
			return
		}
		logger.Error(err) // TODO Don't call logger.Error here?
		return
	}

	go handlers.ClientHandler(conn)
}

func init() {
	config = settings.GetConfig()
}

func main() {
	http.HandleFunc("/", http.NotFound)
	http.HandleFunc("/ws", newConnection)
	logger.Info("Starting Tiberious on", config.Port)
	if err := http.ListenAndServe(config.Port, nil); err != nil {
		logger.Error(err)
	}
}
