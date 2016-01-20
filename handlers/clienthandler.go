package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"tiberious/types"

	"github.com/gorilla/websocket"
	"gopkg.in/redis.v3"
)

const (
	rdishost = "localhost:6379"
	rdispass = ""
)

// ClientHandler handles all client interactions
func ClientHandler(conn *websocket.Conn) {
	client := types.NewClient()
	client.Conn = conn

	rdis := redis.NewClient(&redis.Options{
		Addr:     rdishost,
		Password: rdispass,
		DB:       0,
	})

	/* Set a client ID based on the number of connected guests; authorized users
	 * will have an ID assigned by their number in the database upon creation.
	 * We assign a guest ID here for all connections because authentication is
	 * handled in a message. */
	rdis.Incr("guests")
	var err error
	client.ID, err = rdis.Get("guests").Int64()
	if err != nil {
		log.Fatalln(err)
	}

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
