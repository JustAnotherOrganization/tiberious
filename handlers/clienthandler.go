package handlers

import (
	"strconv"
	"strings"
	"tiberious/db"
	"tiberious/logger"
	"tiberious/types"

	"github.com/JustAnotherOrganization/jgordon"
	"github.com/gorilla/websocket"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
)

var (
	clients = make(map[string]*types.Client)
)

// This internal function is used in message.go
func authenticate(client *types.Client, token types.AuthToken) {
	keys, err := db.GetKeySet("user-*-" + token.AccountName + "-*")
	if err != nil {
		logger.Error(errors.Wrapf(err, "db.GetKeySet %s", token.AccountName))
	}

	if len(keys) == 0 {
		if err = client.Error(jgordon.IncorrectCredentials, ""); err != nil {
			logger.Error(errors.Wrap(err, "client.Error"))
		}

		client.RaiseBan(1)
		return
	}

	slice := strings.Split(keys[0], "-")
	user, err := db.GetUserData(strings.Join(slice[3:], "-"))
	if err != nil {
		if err == db.ErrNotInDB {
			if err := client.Error(jgordon.IncorrectCredentials, ""); err != nil {
				logger.Error(errors.Wrap(err, "client.Error"))
			}
		} else {
			logger.Error(err)
		}
	}

	if token.Password != user.Password {
		if err := client.Error(jgordon.IncorrectCredentials, ""); err != nil {
			logger.Error(errors.Wrap(err, "client.Error"))
		}

		client.RaiseBan(1)
		return
	}

	if client.User.Type == "guest" {
		if err := db.DeleteUser(client.User); err != nil {
			logger.Error(errors.Wrapf(err, "db.DeleteUser %s", client.User.ID.String()))
		}
	}

	delete(clients, client.User.ID.String())

	/* TODO add any rooms that exist in the current user to the new user
	 * before discarding it. */
	client.User = user
	client.Authorized = true
	client.User.Connected = true
	if err := db.WriteUserData(client.User); err != nil {
		logger.Error(errors.Wrapf(err, "db.WriteUserData %s", client.User.ID.String()))
	}

	clients[client.User.ID.String()] = client

	if err := client.Alert(jgordon.OK, ""); err != nil {
		logger.Error(errors.Wrap(err, "client.Alert"))
	}
}

/* Always make sure a new ID is unique...
 * the probability of a UUID collision is somewhere around 1% in 100 million
 * UUIDs but we'll be overly cautious and check anyway. */
func getUniqueID() uuid.UUID {
	var id uuid.UUID
	for {
		id = uuid.NewRandom()
		exists, err := db.UserExists(id.String())
		if err != nil {
			logger.Error(errors.Wrapf(err, "db.UserExists %s", id.String()))
		}
		if exists {
			continue
		}
		break
	}

	return id
}

// ClientHandler handles all client interactions
func ClientHandler(conn *websocket.Conn) {
	client := types.NewClient()
	client.Conn = conn
	client.User = new(types.User)
	// Set the UUID and initialize a username of "guest"
	client.User.ID = getUniqueID()

	str := "user-guest-*-*"
	guests, err := db.GetKeySet(str)
	if err != nil {
		logger.Error(errors.Wrapf(err, "db.GetKeySet %s", str))
	}

	// TODO give guests a numeric suffix, allow disabling guest connections.
	client.User.Username = "guest" + strconv.Itoa(len(guests)+1)
	client.User.LoginName = client.User.Username
	client.User.Type = "guest"
	client.User.Connected = true

	clients[client.User.ID.String()] = client

	if config.AllowGuests {
		defgroup := GetGroup("#default")
		defgroup.Users[client.User.ID.String()] = client.User
		client.User.Groups = append(client.User.Groups, "#default")
		room, err := GetRoom("#default", "#general")
		if err != nil {
			logger.Error(errors.Wrap(err, "GetRoom"))
		}
		client.User.Rooms = append(client.User.Rooms, "#default/#general")
		room.Users[client.User.ID.String()] = client.User

		if err := db.WriteUserData(client.User); err != nil {
			logger.Error(errors.Wrapf(err, "db.WriteUserData %s", client.User.ID.String()))
		}
		if err := db.WriteGroupData(defgroup); err != nil {
			logger.Error(errors.Wrapf(err, "db.WriteGroupData %s", defgroup.Title))
		}
		if err := db.WriteRoomData(room); err != nil {
			logger.Error(errors.Wrapf(err, "db.WriteRoomData %s", room))
		}

		logger.Info("guest", client.User.ID.String(), "connected")
	} else {
		logger.Info("new client connected")
	}

	if err := client.Alert(jgordon.OK, ""); err != nil {
		logger.Error(errors.Wrap(err, "client.Alert"))
	}

	/* TODO we may want to remove this later it's just for easy testing.
	 * to allow a client to get their UUID back from the server after
	 * connecting. */
	if config.AllowGuests {
		if err := client.Alert(jgordon.GeneralNotice, string("Connected as guest with ID "+client.User.ID.String())); err != nil {
			logger.Error(errors.Wrap(err, "client.Alert"))
		}
	} else {
		if err := client.Alert(jgordon.ImportantNotice, "No Guests Allowed : send authentication token to continue"); err != nil {
			logger.Error(errors.Wrap(err, "client.Alert"))
		}
	}

	/* Never return from this loop!
	 * Never break from this loop unless intending to disconnect the client. */
	for {
		_, rawmsg, err := client.Conn.ReadMessage()
		if err != nil {
			switch {
			case websocket.IsCloseError(err, websocket.CloseNormalClosure):
				if client.User != nil {
					logger.Info(client.User.Type, client.User.ID.String(), "disconnected")
				} else {
					logger.Info("client disconnected")
				}
				break
			// TODO handle these different cases appropriately.
			case websocket.IsCloseError(err, websocket.CloseGoingAway):
			case websocket.IsCloseError(err, websocket.CloseProtocolError, websocket.CloseUnsupportedData):
				// This should utilize the ban-score to combat possible spammers
			case websocket.IsCloseError(err, websocket.ClosePolicyViolation, websocket.CloseMessageTooBig):
				// These should also utilize the ban-score but with a higher ban
			default:
				logger.Info(err)
			}
			break
		}

		go ParseMessage(client, rawmsg)
		// TODO handle ban-score
	}

	// We broke out of the loop so disconnect the client.
	client.Conn.Close()
	if client.User != nil {
		if client.User.Type == "guest" {
			if err := db.DeleteUser(client.User); err != nil {
				logger.Error(errors.Wrapf(err, "db.DeleteUser %s", client.User.ID.String()))
			}
		} else {
			client.User.Connected = false
			db.WriteUserData(client.User)
		}
	}

	delete(clients, client.User.ID.String())
}

// GetClientForUser returns a client object that houses a given user
func GetClientForUser(user *types.User) *types.Client {
	for _, c := range clients {
		if c.User.ID.String() == user.ID.String() {
			return c
		}
	}

	return nil
}

/*GetClientsForUsers returns a slice of clients that contain the users in a
 * given slice. */
func GetClientsForUsers(users []*types.User) []*types.Client {
	var ret []*types.Client
	for _, u := range users {
		for _, c := range clients {
			if c.User.ID.String() == u.ID.String() {
				ret = append(ret, c)
				break
			}
		}
	}

	return ret
}
