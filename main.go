package main

import (
	"net/http"

	"tiberious/handlers/client"
	"tiberious/logger"
	"tiberious/settings"
	"tiberious/types"

	"github.com/gorilla/websocket"
)

var (
	config        settings.Config
	clientHandler client.Handler
)

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

	go clientHandler.HandleConnection(conn)
}

func main() {
	config = settings.GetConfig()
	var err error
	clientHandler, err = client.NewHandler(config, make(map[string]*types.Client))
	if err != nil {
		logger.Error(err)
	}

	http.HandleFunc("/", http.NotFound)
	http.HandleFunc("/ws", newConnection)
	logger.Info("Starting Tiberious on", config.Port)
	if err := http.ListenAndServe(config.Port, nil); err != nil {
		logger.Error(err)
	}
}
