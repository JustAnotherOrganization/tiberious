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
		/* Uncertain if this should be fatal or not, invalid
		 * operation on the server side should definitely cause
		 * some form of error presentation to the administrator
		 * but I'm uncertain about full shutdown. */
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

		var message types.MasterObj
		if err := json.Unmarshal(p, &message); err != nil {
			log.Println(err)
			return
		}

		switch {
		case message.Action == "msg":
			/* TODO parse the destination and if the destination exists
			 * send the message (should work for 1to1 even if the user is
			 * not currently online); if destination doesn't exist return an
			 * error (for now just return the message itself for testing).
			 */
			client.Conn.WriteMessage(websocket.BinaryMessage, p)
			break
		default:
			errmin, err := json.Marshal(types.ErrorMin{Response: 400})
			if err != nil {
				/* Uncertain if this should be fatal or not, invalid
				 * operation on the server side should definitely cause
				 * some form of error presentation to the administrator
				 * but I'm uncertain about full shutdown. */
				log.Fatalln(err)
			}
			client.Conn.WriteMessage(websocket.BinaryMessage, errmin)
			break
		}
	}
}
