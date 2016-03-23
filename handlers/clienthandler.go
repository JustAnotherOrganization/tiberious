package handlers

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"tiberious/settings"
	"tiberious/types"

	"github.com/gorilla/websocket"
	"github.com/pborman/uuid"
)

var (
	config  types.Config
	clients = make(map[string]*types.Client)
)

func init() {
	config = settings.GetConfig()
}

// ClientHandler handles all client interactions
func ClientHandler(conn *websocket.Conn) {
	client := types.NewClient()
	client.Conn = conn

	/* TODO store UUID in a datastore of some sort (redis would work but
	 * database type should be configurable in datastore handler). */
	client.ID = uuid.NewRandom()

	clients[client.ID.String()] = client
	log.Println("client", client.ID, "connected")

	alert, err := json.Marshal(types.AlertMin{Response: 200, Time: time.Now().Unix()})
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

		var message types.MasterObj
		if err := json.Unmarshal(p, &message); err != nil {
			log.Println(err)
			return
		}

		if message.Time <= 0 {
			errfull, err := json.Marshal(types.ErrorFull{Response: 400, Time: time.Now().Unix(), Error: "missing or invalid time"})
			if err != nil {
				/* TODO implement better internal error handling in case JSON
				 * marshalling fails for some reason. */
				log.Fatalln(err)
			}

			client.Conn.WriteMessage(websocket.BinaryMessage, errfull)
			log.Println("returned", string(errfull), "to client", client.ID.String())
			return
		}

		switch {
		case message.Action == "msg":
			/* TODO parse the destination and if the destination exists
			 * send the message (should work for 1to1 even if the user is
			 * not currently online (with databasing enabled, otherwise should
			 * return an error)); if destination doesn't exist return an
			 * error (for now just return the message itself for testing).
			 */

			switch {
			// All room's start with "#"
			case strings.HasPrefix(message.To, "#"):
				errfull, err := json.Marshal(types.ErrorFull{Response: 400, Time: time.Now().Unix(), Error: "1toMany messaging is not enabled yet."})
				if err != nil {
					// TODO this needs to be replaced with proper logging/handling.
					log.Fatalln(err)
				}

				client.Conn.WriteMessage(websocket.BinaryMessage, errfull)
				log.Println("returned", string(errfull), "to client", client.ID.String())
				break
			default:
				// Handle 1to1 messaging.

				/* TODO handle server side message logging. handle an error
				 * message for non-existing users (requires user database)
				 * and a separate one for users not being logged on. */

				var relayed = false
				for k, c := range clients {
					if message.To == k {
						c.Conn.WriteMessage(websocket.BinaryMessage, p)
						log.Println("relayed message to", k)
						relayed = true
					}
				}

				if !relayed {
					errmin, err := json.Marshal(types.ErrorMin{Response: 404, Time: time.Now().Unix()})
					if err != nil {
						// TODO afforementioned logging/error handling.
						log.Fatalln(err)
					}

					client.Conn.WriteMessage(websocket.BinaryMessage, errmin)
					log.Println("returned", string(errmin), "to client", client.ID.String())
					return
				}

				// Send a response back saying the message was sent.
				alertmin, err := json.Marshal(types.AlertMin{Response: 200, Time: time.Now().Unix()})
				if err != nil {
					// TODO this needs to be replaced with proper logging/handling.
					log.Fatalln(err)
				}

				client.Conn.WriteMessage(websocket.BinaryMessage, alertmin)
			}

			break
		default:
			errmin, err := json.Marshal(types.ErrorMin{Response: 400, Time: time.Now().Unix()})
			if err != nil {
				/* Uncertain if this should be fatal or not, invalid
				 * operation on the server side should definitely cause
				 * some form of error presentation to the administrator
				 * but I'm uncertain about full shutdown. */
				log.Fatalln(err)
			}
			client.Conn.WriteMessage(websocket.BinaryMessage, errmin)
			log.Println("returned", string(errmin), "to client", client.ID.String())
			break
		}
	}
}
