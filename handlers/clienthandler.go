package handlers

import (
	"encoding/json"
	"strings"

	"tiberious/logger"
	"tiberious/types"

	"github.com/gorilla/websocket"
)

var (
	clients = make(map[string]*types.Client)
)

// ClientHandler handles all client interactions
func ClientHandler(conn *websocket.Conn) {
	client := types.NewClient()
	client.Conn = conn
	// Set the UUID and initialize a username of "guest"
	NewUser(client)

	clients[client.ID.String()] = client
	logger.Info("client", client.ID, "connected")

	if err := client.Alert(200, ""); err != nil {
		logger.Error(err)
	}

	/* TODO we may want to remove this later it's just for easy testing.
	 * to allow a client to get their UUID back from the server after
	 * connecting. */
	if err := client.Alert(100, string("Connected with ID "+client.ID.String())); err != nil {
		logger.Error(err)
	}

	// TODO handle authentication for servers with user databases.

	// Never return from this loop unless disconnecting the client...
	for {
		_, p, err := client.Conn.ReadMessage()
		if err != nil {
			logger.Info(err)
			// TODO disconnect the client.
			return
		}

		var message types.MasterObj
		if err := json.Unmarshal(p, &message); err != nil {
			if err := client.Error(400, "invalid object"); err != nil {
				logger.Error(err)
			}
			continue
		}

		if message.Time <= 0 {
			if err := client.Error(400, "missing or invalid time"); err != nil {
				logger.Error(err)
			}
			continue
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
				/* TODO we don't want to stop people from outside of a room from
				 * messaging that room directly unless it's a private channel
				 * (these don't exist yet but should restrict to only members
				 * of said channel) */
				rexists, room := GetRoom(message.To)
				if !rexists {
					if err := client.Error(404, ""); err != nil {
						logger.Error(err)
					}
					continue
				}
				// TODO should this be handled in a channel or goroutine?
				for _, c := range room {
					c.Conn.WriteMessage(websocket.BinaryMessage, p)
				}
			default:
				// Handle 1to1 messaging.

				/* TODO handle server side message logging. handle an error
				 * message for non-existing users (requires user database)
				 * and a separate one for users not being logged on. */

				var relayed = false
				for k, c := range clients {
					if message.To == k {
						c.Conn.WriteMessage(websocket.BinaryMessage, p)
						relayed = true
					}
				}

				if relayed {
					break
				}

				if err := client.Error(404, ""); err != nil {
					logger.Error(err)
				}

				continue
			}

			// Send a response back saying the message was sent.
			if err := client.Alert(200, ""); err != nil {
				logger.Error(err)
			}

			break
		case message.Action == "join":
			// TODO implement private rooms
			var rexists = false
			var room types.Room
			rexists, room = GetRoom(message.Room)
			if !rexists {
				room = GetNewRoom(message.Room)
			}

			room[client.ID.String()] = client
			// Send a response back confirming we joined the room..
			if err := client.Alert(200, ""); err != nil {
				logger.Error(err)
			}

			break
		case message.Action == "leave":
		case message.Action == "part":
			var rexists = false
			var room types.Room
			rexists, room = GetRoom(message.Room)
			if !rexists {
				if err := client.Error(404, ""); err != nil {
					logger.Error(err)
				}
				break
			}

			var ispresent = false
			for k := range room {
				if k == client.ID.String() {
					ispresent = true
					break
				}
			}

			if !ispresent {
				// TODO should this return a different error number?
				if err := client.Error(410, ""); err != nil {
					logger.Error(err)
				}
				break
			}

			delete(room, client.ID.String())

			// Send a response back confirming we left the room..
			if err := client.Alert(200, ""); err != nil {
				logger.Error(err)
			}

			break
		default:
			if err := client.Error(400, ""); err != nil {
				logger.Error(err)
			}
			break
		}
	}
}
