package main

import (
	"log"
	"net/http"

	"tiberious/handlers"
	"tiberious/types"

	"github.com/gorilla/websocket"
)

var config types.Config

// TODO move this into a separate handler of some form.
func newConnection(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  config.ReadBufferSize,
		WriteBufferSize: config.WriteBufferSize,
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	go handlers.ClientHandler(conn)
}

func main() {
	handlers.InitConfig(&config)

	http.HandleFunc("/", http.NotFound)
	http.HandleFunc("/ws", newConnection)
	log.Println("Starting Tiberious on", config.Port)
	if err := http.ListenAndServe(config.Port, nil); err != nil {
		log.Fatalln(err)
	}
}
