package handlers

import (
	"encoding/json"
	"log"
	"strings"

	"tiberious/logger"
	"tiberious/types"

	"github.com/gorilla/websocket"
)

func relayToRoom(room *types.Room, rawmsg []byte) {
	for _, u := range room.Users {
		c := GetClientForUser(u) // In clienthandler.go
		if c != nil {
			c.Conn.WriteMessage(websocket.BinaryMessage, rawmsg)
		}
	}
}

func relayToGroup(group *types.Group, rawmsg []byte) {
	for _, u := range group.Users {
		c := GetClientForUser(u) // In clienthandler.go
		if c != nil {
			c.Conn.WriteMessage(websocket.BinaryMessage, rawmsg)
		}
	}
}

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

	if !config.AllowGuests && !client.Authorized {
		if message.Action != "authenticate" {
			if err := client.Error(types.NotAuthorized, ""); err != nil {
				logger.Error(err)
			}
			return 1
		}
	}

	switch {
	case message.Action == "authenticate":
		return authenticate(client, message.User)
	case message.Action == "msg":
		/* TODO Fixup message parsing (should work for 1to1 even if the user is
		 * not currently online (with databasing enabled, otherwise should
		 * return an error)); if destination doesn't exist return an error. */

		switch {
		// All room's start with "#"
		case IsRoomName(message.To):
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

			// Block guest connections from messaging outside of group #default.
			if client.User.Type == "guest" && group.Title != "#default" {
				if err := client.Error(types.Forbidden, "guest account, please authenticate"); err != nil {
					logger.Error(err)
				}
				return 0
			}

			// Block messages from outside a group.
			var member = false
			for _, g := range client.User.Groups {
				if group.Title == g {
					member = true
				}
			}

			if !member {
				if err := client.Error(types.Forbidden, ""); err != nil {
					logger.Error(err)
				}
				return 1
			}

			room := GetRoom(slice[0], slice[1])
			if room == nil {
				if err := client.Error(types.NotFound, ""); err != nil {
					logger.Error(err)
				}
				return 0
			}

			// Block external messages on private rooms.
			member = false
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
			go relayToRoom(room, rawmsg)
			break
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
		if !IsRoomName(message.Room) {
			if err := client.Error(types.BadRequestOrObject, "room names should start with '#'"); err != nil {
				logger.Error(err)
			}
			return 0
		}

		if !strings.Contains(message.Room, "/") {
			if err := client.Error(types.BadRequestOrObject, "room names should be type of 'group/room'"); err != nil {
				logger.Error(err)
			}
			return 0
		}
		slice := strings.Split(message.Room, "/")
		if len(slice) != 2 {
			if err := client.Error(types.BadRequestOrObject, "room names should be type of 'group/room'"); err != nil {
				logger.Error(err)
			}
			return 0
		}
		group := GetGroup(slice[0])
		if group == nil {
			if err := client.Error(types.NotFound, "group does not exist"); err != nil {
				logger.Error(err)
			}
			return 0
		}

		// Block guest connections from messaging outside of group #default.
		if client.User.Type == "guest" && group.Title != "#default" {
			if err := client.Error(types.Forbidden, "guest account, please authenticate"); err != nil {
				logger.Error(err)
			}
			return 0
		}

		// Block users from joining if they're outside the group.
		var member = false
		for _, g := range client.User.Groups {
			if group.Title == g {
				member = true
			}
		}

		if !member {
			if err := client.Error(types.Forbidden, ""); err != nil {
				logger.Error(err)
			}
			return 1
		}

		// TODO implement private rooms
		room := GetRoom(slice[0], slice[1])
		if room == nil {
			room = GetNewRoom(slice[0], slice[1])
			room.Users = make(map[string]*types.User)
		}

		room.Users[client.User.ID.String()] = client.User

		// Update the room data for the database.
		if err := WriteRoomData(room); err != nil {
			logger.Error(err)
		}

		// Send a response back confirming we joined the room..
		if err := client.Alert(types.OK, ""); err != nil {
			logger.Error(err)
		}

		break
	case message.Action == "leave":
	case message.Action == "part":
		if !IsRoomName(message.Room) {
			if err := client.Error(types.BadRequestOrObject, "room names should start with '#'"); err != nil {
				logger.Error(err)
			}
			return 0
		}

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
		if err := WriteRoomData(room); err != nil {
			logger.Error(err)
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
