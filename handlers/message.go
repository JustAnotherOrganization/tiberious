package handlers

import (
	"encoding/json"
	"log"
	"strings"

	"tiberious/logger"
	"tiberious/types"

	"github.com/gorilla/websocket"
)

/*ParseMessage parses a message object and returns an int back, with a ban-score
 *if this is greater than 0 it is applied to the clients ban-score. */
func ParseMessage(client *types.Client, rawmsg []byte) int {
	var message types.MasterObj
	if err := json.Unmarshal(rawmsg, &message); err != nil {
		if err := client.Error(types.BadRequestOrObject, "invalid object"); err != nil {
			logger.Error(err)
		}
		return 0
	}

	if message.Time <= 0 {
		if err := client.Error(types.BadRequestOrObject, "missing or invalid time"); err != nil {
			logger.Error(err)
		}
		return 0
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
			rexists, room := GetRoom(message.To)
			if !rexists {
				if err := client.Error(types.NotFound, ""); err != nil {
					logger.Error(err)
				}
				return 0
			}

			// Block external messages on private rooms.
			var member = false
			for k := range room.List {
				if client.ID.String() == k {
					member = true
				}
			}

			if room.Private && !member {
				if err := client.Error(types.Forbidden, ""); err != nil {
					log.Fatalln(err)
				}
				return 1
			}

			// TODO should this be handled in a channel or goroutine?
			for _, c := range room.List {
				c.Conn.WriteMessage(websocket.BinaryMessage, rawmsg)
			}
		default:
			// Handle 1to1 messaging.

			/* TODO handle server side message logging. handle an error
			 * message for non-existing users (requires user database)
			 * and a separate one for users not being logged on. */

			var relayed = false
			for k, c := range clients {
				if message.To == k {
					c.Conn.WriteMessage(websocket.BinaryMessage, rawmsg)
					relayed = true
				}
			}

			if relayed {
				break
			}

			if err := client.Error(types.NotFound, ""); err != nil {
				logger.Error(err)
			}

			return 0
		}

		// Send a response back saying the message was sent.
		if err := client.Alert(types.OK, ""); err != nil {
			logger.Error(err)
		}

		break
	case message.Action == "join":
		// TODO implement private rooms
		var rexists = false
		var room *types.Room
		rexists, room = GetRoom(message.Room)
		if !rexists {
			room = GetNewRoom(message.Room)
		}

		room.List[client.ID.String()] = client
		// Send a response back confirming we joined the room..
		if err := client.Alert(types.OK, ""); err != nil {
			logger.Error(err)
		}

		break
	case message.Action == "leave":
	case message.Action == "part":
		var rexists = false
		var room *types.Room
		rexists, room = GetRoom(message.Room)
		if !rexists {
			if err := client.Error(types.NotFound, ""); err != nil {
				logger.Error(err)
			}
			break
		}

		var ispresent = false
		for k := range room.List {
			if k == client.ID.String() {
				ispresent = true
				break
			}
		}

		if !ispresent {
			// TODO should this return a different error number?
			if err := client.Error(types.Gone, ""); err != nil {
				logger.Error(err)
			}
			break
		}

		delete(room.List, client.ID.String())

		// Send a response back confirming we left the room..
		if err := client.Alert(types.OK, ""); err != nil {
			logger.Error(err)
		}

		break
	default:
		if err := client.Error(types.BadRequestOrObject, ""); err != nil {
			logger.Error(err)
		}
		break
	}

	return 0
}
