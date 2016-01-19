package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

func newConnection(w http.ResponseWriter, r *http.Request) {
	_, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("client connected")
}

func main() {
	http.HandleFunc("/", http.NotFound)
	if err := http.ListenAndServe(":4002", http.HandlerFunc(newConnection)); err != nil {
		log.Fatalln(err)
	}
}
