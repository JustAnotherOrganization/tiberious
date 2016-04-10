package handlers

import (
	"encoding/json"
	"log"
	"strings"

	"tiberious/logger"
	"tiberious/settings"
	"tiberious/types"

	"github.com/gorilla/websocket"
)

/*ParseMessage parses a message object and returns an int back, with a ban-score
 *if this is greater than 0 it is applied to the clients ban-score. */
func ParseMessage(client *types.Client, rawmsg []byte) int {
	config := settings.GetConfig()

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
			if !strings.Contains(message.To, "/") {
				if err := client.Error(types.BadRequestOrObject, "room names should be type of 'group/room'"); err != nil {
					logger.Error(err)
				}
				return 0
			}
			slice := strings.Split(message.To, "/")
			group := GetGroup(slice[0])
			if group == nil {
				if err := client.Error(types.NotFound, "group does not exist"); err != nil {
					logger.Error(err)
				}
				return 0
			}

			room := GetRoom(slice[0], slice[1])
			if room == nil {
				if err := client.Error(types.NotFound, ""); err != nil {
					logger.Error(err)
				}
				return 0
			}

			// Block external messages on private rooms.
			var member = false
			for k := range room.Users {
				if client.User.ID.String() == k {
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
			for _, u := range room.Users {
				c := GetClientForUser(u)
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
	// Join messages should include both a group and a room name.
	case message.Action == "join":
		if !strings.Contains(message.Room, "/") {
			if err := client.Error(types.BadRequestOrObject, "room names should be type of 'group/room'"); err != nil {
				logger.Error(err)
			}
			return 0
		}
		slice := strings.Split(message.Room, "/")
		group := GetGroup(slice[0])
		if group == nil {
			if err := client.Error(types.NotFound, "group does not exist"); err != nil {
				logger.Error(err)
			}
			return 0
		}

		// TODO implement private rooms
		var room *types.Room
		room = GetRoom(slice[0], slice[1])
		if room == nil {
			room = GetNewRoom(slice[0], slice[1])
		}

		room.Users[client.User.ID.String()] = client.User

		// Update the room data for the database.
		if config.UserDatabase != 0 {
			if err := WriteRoomData(room); err != nil {
				logger.Error(err)
			}
		}

		// Send a response back confirming we joined the room..
		if err := client.Alert(types.OK, ""); err != nil {
			logger.Error(err)
		}

		break
	case message.Action == "leave":
	case message.Action == "part":
		if !strings.Contains(message.Room, "/") {
			if err := client.Error(types.BadRequestOrObject, "room names should be type of 'group/room'"); err != nil {
				logger.Error(err)
			}
			return 0
		}
		slice := strings.Split(message.Room, "/")
		group := GetGroup(slice[0])
		if group == nil {
			if err := client.Error(types.NotFound, "group does not exist"); err != nil {
				logger.Error(err)
			}
			return 0
		}

		var room *types.Room
		room = GetRoom(slice[0], slice[1])
		if room == nil {
			if err := client.Error(types.NotFound, ""); err != nil {
				logger.Error(err)
			}
			break
		}

		var ispresent = false
		for k := range room.Users {
			if k == client.User.ID.String() {
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

		delete(room.Users, client.User.ID.String())

		// Update the room data for the database.
		if config.UserDatabase != 0 {
			if err := WriteRoomData(room); err != nil {
				logger.Error(err)
			}
		}

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
