package handlers

import (
	"encoding/json"
	"log"
	"strings"

	"tiberious/logger"
	"tiberious/types"

	"github.com/JustAnotherOrganization/jgordon"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
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

func handlePrivateMessage(client *types.Client, message types.MasterObj, rawmsg []byte) bool {
	// Handle 1to1 messaging.

	/* TODO handle server side message logging. handle an error
	 * message for non-existing users (requires user database)
	 * and a separate one for users not being logged on. */

	relayed := false
	for k, c := range clients {
		if message.To == k {
			c.Conn.WriteMessage(websocket.BinaryMessage, rawmsg)
			relayed = true
		}
	}

	if relayed {
		return true
	}

	if err := client.Error(jgordon.NotFound, ""); err != nil {
		logger.Error(errors.Wrapf(err, "client.Error %s", jgordon.NotFound))
	}

	return false
}

func handleRoomMessage(client *types.Client, message types.MasterObj, rawmsg []byte) bool {
	if !strings.Contains(message.To, "/") {
		str := "room names should be type of 'group/room'"
		if err := client.Error(jgordon.BadRequestOrObject, str); err != nil {
			logger.Error(errors.Wrapf(err, "client.Error %s : %s", jgordon.BadRequestOrObject, str))
		}
		return false
	}
	slice := strings.Split(message.To, "/")
	group := GetGroup(slice[0])
	if group == nil {
		str := "group does not exist"
		if err := client.Error(jgordon.NotFound, str); err != nil {
			logger.Error(errors.Wrapf(err, "client.Error %s : %s", jgordon.NotFound, str))
		}
		return false
	}

	// Block guest connections from messaging outside of group #default.
	if client.User.Type == "guest" && group.Title != "#default" {
		str := "guest account, please authenticate"
		if err := client.Error(jgordon.Forbidden, str); err != nil {
			logger.Error(errors.Wrapf(err, "client.Error %s : %s", jgordon.Forbidden, str))
		}
		return false
	}

	// Block messages from outside a group.
	var member = false
	for _, g := range client.User.Groups {
		if group.Title == g {
			member = true
		}
	}

	if !member {
		if err := client.Error(jgordon.Forbidden, ""); err != nil {
			logger.Error(errors.Wrapf(err, "client.Error %s", jgordon.Forbidden))
		}
		client.RaiseBan(1)
		return false
	}

	room, err := GetRoom(slice[0], slice[1])
	if err != nil {
		logger.Error(errors.Wrapf(err, "GetRoom %s/%s", slice[0], slice[1]))
	}

	if room == nil {
		if err := client.Error(jgordon.NotFound, ""); err != nil {
			logger.Error(errors.Wrapf(err, "client.Error %s", jgordon.NotFound))
		}
		return false
	}

	// Block external messages on private rooms.
	member = false
	for k := range room.Users {
		if client.User.ID.String() == k {
			member = true
		}
	}

	if room.Private && !member {
		if err := client.Error(jgordon.Forbidden, ""); err != nil {
			log.Fatalln(errors.Wrapf(err, "client.Error %s", jgordon.Forbidden))
		}
		client.RaiseBan(1)
		return false
	}

	relayToRoom(room, rawmsg)
	return true
}

// Main functionality of ParseMessage
func parseMessage(client *types.Client, message types.MasterObj, rawmsg []byte) {
	switch {
	case message.Action == "authenticate":
		authenticate(client, message.User) // in clienthandler.go
		return
	case message.Action == "msg":
		/* TODO Fixup message parsing (should work for 1to1 even if the user is
		 * not currently online (with databasing enabled, otherwise should
		 * return an error)); if destination doesn't exist return an error. */

		sent := false
		switch {
		// All room's start with "#"
		case IsRoomName(message.To):
			sent = handleRoomMessage(client, message, rawmsg)
		default:
			sent = handlePrivateMessage(client, message, rawmsg)
		}

		// Send a response back saying the message was sent.
		if sent {
			if err := client.Alert(jgordon.OK, ""); err != nil {
				logger.Error(errors.Wrapf(err, "client.Alert %s", jgordon.OK))
			}
		}

		break
	// Join messages should include both a group and a room name.
	case message.Action == "join":
		if !IsRoomName(message.Room) {
			str := "room names should start with '#'"
			if err := client.Error(jgordon.BadRequestOrObject, str); err != nil {
				logger.Error(errors.Wrapf(err, "client.Error %s : %s", jgordon.BadRequestOrObject, str))
			}
			return
		}

		if !strings.Contains(message.Room, "/") {
			str := "room names should be type of 'group/room'"
			if err := client.Error(jgordon.BadRequestOrObject, str); err != nil {
				logger.Error(errors.Wrapf(err, "client.Error %s : %s", jgordon.BadRequestOrObject, str))
			}
			return
		}
		slice := strings.Split(message.Room, "/")
		if len(slice) != 2 {
			str := "room names should be type of 'group/room'"
			if err := client.Error(jgordon.BadRequestOrObject, str); err != nil {
				logger.Error(errors.Wrapf(err, "client.Error %s : %s", jgordon.BadRequestOrObject, str))
			}
			return
		}
		group := GetGroup(slice[0])
		if group == nil {
			str := "group does not exist"
			if err := client.Error(jgordon.NotFound, str); err != nil {
				logger.Error(errors.Wrapf(err, "client.Error %s : %s", jgordon.NotFound, str))
			}
			return
		}

		// Block guest connections from messaging outside of group #default.
		if client.User.Type == "guest" && group.Title != "#default" {
			str := "guest account, please authenticate"
			if err := client.Error(jgordon.Forbidden, str); err != nil {
				logger.Error(errors.Wrapf(err, "client.Error %s : %s", jgordon.Forbidden, str))
			}
			return
		}

		// Block users from joining if they're outside the group.
		var member = false
		for _, g := range client.User.Groups {
			if group.Title == g {
				member = true
			}
		}

		if !member {
			if err := client.Error(jgordon.Forbidden, ""); err != nil {
				logger.Error(errors.Wrapf(err, "client.Error %s", jgordon.Forbidden))
			}
			client.RaiseBan(1)
			return
		}

		// Use GetNewRoom cause it will just grab the existing one if it does
		// in fact exists.
		// TODO implement private rooms
		room, err := GetNewRoom(slice[0], slice[1])
		if err != nil {
			logger.Error(errors.Wrapf(err, "GetNewRoom %s/%s", slice[0], slice[1]))
			return
		}

		room.Users[client.User.ID.String()] = client.User

		// Update the room data for the database.
		if err := WriteRoomData(room); err != nil {
			logger.Error(errors.Wrapf(err, "WriteRoomData %s", room))
		}

		// Send a response back confirming we joined the room..
		if err := client.Alert(jgordon.OK, ""); err != nil {
			logger.Error(errors.Wrapf(err, "client.Alert %s", jgordon.OK))
		}

		break
	case message.Action == "leave":
	case message.Action == "part":
		if !IsRoomName(message.Room) {
			str := "room names should start with '#'"
			if err := client.Error(jgordon.BadRequestOrObject, str); err != nil {
				logger.Error(errors.Wrapf(err, "client.Error %s : %s", jgordon.BadRequestOrObject))
			}
			return
		}

		if !strings.Contains(message.Room, "/") {
			str := "room names should be type of 'group/room'"
			if err := client.Error(jgordon.BadRequestOrObject, str); err != nil {
				logger.Error(errors.Wrapf(err, "client.Error %s : %s", jgordon.BadRequestOrObject, str))
			}
			return
		}
		slice := strings.Split(message.Room, "/")
		group := GetGroup(slice[0])
		if group == nil {
			str := "group does not exist"
			if err := client.Error(jgordon.NotFound, str); err != nil {
				logger.Error(errors.Wrapf(err, "client.Error %s : %s", jgordon.NotFound, str))
			}
			return
		}

		room, err := GetRoom(slice[0], slice[1])
		if err != nil {
			logger.Error(errors.Wrapf(err, "GetRoom %s/%s", slice[0], slice[1]))
			return
		}
		if room == nil {
			if err := client.Error(jgordon.NotFound, ""); err != nil {
				logger.Error(errors.Wrapf(err, "client.Error %s", jgordon.NotFound))
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
			if err := client.Error(jgordon.Gone, ""); err != nil {
				logger.Error(errors.Wrapf(err, "client.Error %s", jgordon.Gone))
			}
			break
		}

		delete(room.Users, client.User.ID.String())

		// Update the room data for the database.
		if err := WriteRoomData(room); err != nil {
			logger.Error(errors.Wrapf(err, "WriteRoomData %s", room))
		}

		// Send a response back confirming we left the room..
		if err := client.Alert(jgordon.OK, ""); err != nil {
			logger.Error(errors.Wrapf(err, "client.Alert %s", jgordon.OK))
		}

		break
	default:
		if err := client.Error(jgordon.BadRequestOrObject, ""); err != nil {
			logger.Error(errors.Wrapf(err, "client.Error %s", jgordon.BadRequestOrObject))
		}
		break
	}

	return
}

// ParseMessage parses a message object and if neccessary raises the client's
// banScore.
func ParseMessage(client *types.Client, rawmsg []byte) {
	var message types.MasterObj
	if err := json.Unmarshal(rawmsg, &message); err != nil {
		str := "invalid object"
		if err := client.Error(jgordon.BadRequestOrObject, str); err != nil {
			logger.Error(errors.Wrapf(err, "client.Error %s : %s", jgordon.BadRequestOrObject, str))
		}
		return
	}

	if message.Time <= 0 {
		str := "missing or invalid time"
		if err := client.Error(jgordon.BadRequestOrObject, str); err != nil {
			logger.Error(errors.Wrapf(err, "client.Error %s : %s", jgordon.BadRequestOrObject, str))
		}
		return
	}

	if !config.AllowGuests && !client.Authorized {
		if message.Action != "authenticate" {
			if err := client.Error(jgordon.NotAuthorized, ""); err != nil {
				logger.Error(errors.Wrapf(err, "client.Error %s", jgordon.NotAuthorized))
			}
			client.RaiseBan(1)
			return
		}
	}

	parseMessage(client, message, rawmsg)
}
