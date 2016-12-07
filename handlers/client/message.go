package client

import (
	"encoding/json"
	"strings"
	"tiberious/types"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

func (h *handler) relayToRoom(room *types.Room, rawmsg []byte) {
	for _, u := range room.Users {
		c := h.GetClientForUser(u)
		if c != nil {
			c.Conn.WriteMessage(websocket.BinaryMessage, rawmsg)
		}
	}
}

func (h *handler) relayToGroup(group *types.Group, rawmsg []byte) {
	for _, u := range group.Users {
		c := h.GetClientForUser(u)
		if c != nil {
			c.Conn.WriteMessage(websocket.BinaryMessage, rawmsg)
		}
	}
}

// parseMessage parses a message object and returns an int back, with a ban-score
// if this is greater than 0 it is applied to the clients ban-score. */
func (h *handler) parseMessage(client *types.Client, rawmsg []byte) (banScore int, err error) {
	banScore = 0

	var message types.MasterObj
	if err = json.Unmarshal(rawmsg, &message); err != nil {
		if err2 := client.Error(types.BadRequestOrObject, "invalid object"); err2 != nil {
			err = errors.Wrap(err2, "client.Error")
		}
		return
	}

	if message.Time <= 0 {
		if err = client.Error(types.BadRequestOrObject, "missing or invalid time"); err != nil {
			err = errors.Wrap(err, "client.Error")
		}
		return
	}

	if !h.config.AllowGuests && !client.Authorized {
		if message.Action != "authenticate" {
			banScore = 1
			if err = client.Error(types.NotAuthorized, ""); err != nil {
				err = errors.Wrap(err, "client.Error")
			}
			return
		}
	}

	switch {
	case message.Action == "authenticate":
		banScore, err = h.authenticate(client, message.User)
		if err != nil {
			err = errors.Wrap(err, "authenticate")
		}
		return
	case message.Action == "msg":
		/* TODO Fixup message parsing (should work for 1to1 even if the user is
		 * not currently online (with databasing enabled, otherwise should
		 * return an error)); if destination doesn't exist return an error. */

		switch {
		// All room's start with "#"
		case h.groupHandler.IsRoomName(message.To):
			var (
				group *types.Group
				room  *types.Room
			)

			if !strings.Contains(message.To, "/") {
				if err = client.Error(types.BadRequestOrObject, "room names should be type of 'group/room'"); err != nil {
					err = errors.Wrap(err, "client.Error")
				}
				return
			}
			slice := strings.Split(message.To, "/")
			group, err = h.groupHandler.GetGroup(slice[0])
			if err != nil {
				err = errors.Wrap(err, "GetGroup")
				return
			}

			if group == nil {
				if err = client.Error(types.NotFound, "group does not exist"); err != nil {
					err = errors.Wrap(err, "client.Error")
				}
				return
			}

			// Block guest connections from messaging outside of group #default.
			if client.User.Type == "guest" && group.Title != "#default" {
				if err = client.Error(types.Forbidden, "guest account, please authenticate"); err != nil {
					err = errors.Wrap(err, "client.Error")
				}
				return
			}

			// Block messages from outside a group.
			var member = false
			for _, g := range client.User.Groups {
				if group.Title == g {
					member = true
				}
			}

			if !member {
				banScore = 1
				if err = client.Error(types.Forbidden, ""); err != nil {
					err = errors.Wrap(err, "client.Error")
				}
				return
			}

			room, err = h.groupHandler.GetRoom(slice[0], slice[1])
			if err != nil {
				err = errors.Wrap(err, "GetRoom")
				return
			}
			if room == nil {
				if err = client.Error(types.NotFound, ""); err != nil {
					err = errors.Wrap(err, "client.Error")
				}
				return
			}

			// Block external messages on private rooms.
			member = false
			for k := range room.Users {
				if client.User.ID.String() == k {
					member = true
				}
			}

			if room.Private && !member {
				banScore = 1
				if err = client.Error(types.Forbidden, ""); err != nil {
					err = errors.Wrap(err, "client.Error")
				}
				return
			}
			go h.relayToRoom(room, rawmsg)
			break
		default:
			// Handle 1to1 messaging.

			/* TODO handle server side message logging. handle an error
			 * message for non-existing users (requires user database)
			 * and a separate one for users not being logged on. */

			var relayed = false
			for k, c := range h.clients {
				if message.To == k {
					c.Conn.WriteMessage(websocket.BinaryMessage, rawmsg)
					relayed = true
				}
			}

			if relayed {
				break
			}

			if err = client.Error(types.NotFound, ""); err != nil {
				err = errors.Wrap(err, "client.Error")
			}

			return
		}

		// Send a response back saying the message was sent.
		if err = client.Alert(types.OK, ""); err != nil {
			err = errors.Wrap(err, "client.Alert")
		}

		break
	// Join messages should include both a group and a room name.
	case message.Action == "join":
		var (
			group *types.Group
			room  *types.Room
		)

		if !h.groupHandler.IsRoomName(message.Room) {
			if err = client.Error(types.BadRequestOrObject, "room names should start with '#'"); err != nil {
				err = errors.Wrap(err, "client.Error")
			}
			return
		}

		if !strings.Contains(message.Room, "/") {
			if err = client.Error(types.BadRequestOrObject, "room names should be type of 'group/room'"); err != nil {
				err = errors.Wrap(err, "client.Error")
			}
			return
		}
		slice := strings.Split(message.Room, "/")
		if len(slice) != 2 {
			if err = client.Error(types.BadRequestOrObject, "room names should be type of 'group/room'"); err != nil {
				err = errors.Wrap(err, "client.Error")
			}
			return
		}
		group, err = h.groupHandler.GetGroup(slice[0])
		if err != nil {
			err = errors.Wrap(err, "GetGroup")
			return
		}
		if group == nil {
			if err = client.Error(types.NotFound, "group does not exist"); err != nil {
				err = errors.Wrap(err, "client.Error")
			}
			return
		}

		// Block guest connections from messaging outside of group #default.
		if client.User.Type == "guest" && group.Title != "#default" {
			banScore = 1
			if err = client.Error(types.Forbidden, "guest account, please authenticate"); err != nil {
				err = errors.Wrap(err, "client.Error")
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
			banScore = 1
			if err = client.Error(types.Forbidden, ""); err != nil {
				err = errors.Wrap(err, "client.Error")
			}
			return
		}

		// TODO implement private rooms
		room, err = h.groupHandler.GetRoom(slice[0], slice[1])
		if err != nil {
			err = errors.Wrap(err, "GetRoom")
			return
		}
		if room == nil {
			room, err = h.groupHandler.GetNewRoom(slice[0], slice[1])
			if err != nil {
				err = errors.Wrap(err, "GetNewRoom")
				return
			}
			room.Users = make(map[string]*types.User)
		}

		room.Users[client.User.ID.String()] = client.User

		// Update the room data for the database.
		if err = h.groupHandler.WriteRoomData(room); err != nil {
			err = errors.Wrap(err, "WriteRoomData")
		}

		// Send a response back confirming we joined the room..
		if err = client.Alert(types.OK, ""); err != nil {
			err = errors.Wrap(err, "client.Alert")
		}

		break
	case message.Action == "leave":
	case message.Action == "part":
		var (
			group *types.Group
			room  *types.Room
		)

		if !h.groupHandler.IsRoomName(message.Room) {
			if err = client.Error(types.BadRequestOrObject, "room names should start with '#'"); err != nil {
				err = errors.Wrap(err, "client.Error")
			}
			return
		}

		if !strings.Contains(message.Room, "/") {
			if err = client.Error(types.BadRequestOrObject, "room names should be type of 'group/room'"); err != nil {
				err = errors.Wrap(err, "client.Error")
			}
			return
		}
		slice := strings.Split(message.Room, "/")
		group, err = h.groupHandler.GetGroup(slice[0])
		if err != nil {
			err = errors.Wrap(err, "GetGroup")
			return
		}
		if group == nil {
			if err = client.Error(types.NotFound, "group does not exist"); err != nil {
				err = errors.Wrap(err, "client.Error")
			}
			return
		}

		room, err = h.groupHandler.GetRoom(slice[0], slice[1])
		if err != nil {
			err = errors.Wrap(err, "GetRoom")
			return
		}
		if room == nil {
			if err = client.Error(types.NotFound, ""); err != nil {
				err = errors.Wrap(err, "client.Error")
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
			if err = client.Error(types.Gone, ""); err != nil {
				err = errors.Wrap(err, "client.Error")
			}
			break
		}

		delete(room.Users, client.User.ID.String())

		// Update the room data for the database.
		if err = h.groupHandler.WriteRoomData(room); err != nil {
			err = errors.Wrap(err, "WriteRoomData")
		}

		// Send a response back confirming we left the room..
		if err = client.Alert(types.OK, ""); err != nil {
			err = errors.Wrap(err, "client.Alert")
		}

		break
	default:
		if err = client.Error(types.BadRequestOrObject, ""); err != nil {
			err = errors.Wrap(err, "client.Error")
		}
		break
	}

	return
}
