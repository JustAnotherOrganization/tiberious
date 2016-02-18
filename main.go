package main

import (
	"log"
	"net/http"

	"tiberious/handlers"

	"github.com/gorilla/websocket"
)

// TODO benchmark and tune buffer-sizes
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// TODO move this into a separate handler of some form.
func newConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	go handlers.ClientHandler(conn)
}

func main() {
	http.HandleFunc("/", http.NotFound)
	http.HandleFunc("/ws", newConnection)
	var port = ":4002"
	log.Println("Starting Tiberious on", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalln(err)
	}
}
