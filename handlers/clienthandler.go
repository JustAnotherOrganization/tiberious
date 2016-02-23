package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"tiberious/types"

	"github.com/gorilla/websocket"
	"github.com/pborman/uuid"
)

// ClientHandler handles all client interactions
func ClientHandler(conn *websocket.Conn) {
	client := types.NewClient()
	client.Conn = conn

	/* TODO store UUID in a datastore of some sort (redis would work but
	 * database type should be configurable in datastore handler). */
	client.ID = uuid.NewRandom()

	log.Println("client", client.ID, "connected")

	alert, err := json.Marshal(types.AlertMin{Response: 200})
	if err != nil {
		log.Fatalln(err)
	}

	client.Conn.WriteMessage(websocket.BinaryMessage, alert)

	for {
		_, p, err := client.Conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		// TODO proper message handling
		fmt.Printf("%s\n", p)
	}
}
