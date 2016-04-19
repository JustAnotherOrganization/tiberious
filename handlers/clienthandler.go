package handlers

import (
	"strconv"
	"strings"
	"tiberious/db"
	"tiberious/logger"
	"tiberious/types"

	"github.com/gorilla/websocket"
	"github.com/pborman/uuid"
)

var (
	clients = make(map[string]*types.Client)
)

func authenticate(client *types.Client, token types.AuthToken) int {
	keys, err := db.GetKeySet("user-*-" + token.AccountName + "-*")
	if err != nil {
		logger.Error(err)
	}

	if len(keys) == 0 {
		if err := client.Error(types.IncorrectCredentials, ""); err != nil {
			logger.Error(err)
		}

		return 1
	}

	slice := strings.Split(keys[0], "-")
	user, err := db.GetUserData(strings.Join(slice[3:], "-"))
	if err != nil {
		if err == types.NotInDB {
			if err := client.Error(types.IncorrectCredentials, ""); err != nil {
				logger.Error(err)
			}
		} else {
			logger.Error(err)
		}
	}

	if token.Password != user.Password {
		if err := client.Error(types.IncorrectCredentials, ""); err != nil {
			logger.Error(err)
		}

		return 1
	}

	if client.User.Type == "guest" {
		if err := db.DeleteUser(client.User); err != nil {
			logger.Error(err)
		}
	}

	delete(clients, client.User.ID.String())

	/* TODO add any rooms that exist in the current user to the new user
	 * before discarding it. */
	client.User = user
	client.Authorized = true
	client.User.Connected = true
	if err := db.WriteUserData(client.User); err != nil {
		logger.Error(err)
	}

	clients[client.User.ID.String()] = client

	if err := client.Alert(types.OK, ""); err != nil {
		logger.Error(err)
	}

	return 0
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
			logger.Error(err)
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

	guests, err := db.GetKeySet("user-guest-*-*")
	if err != nil {
		logger.Error(err)
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
		room := GetRoom("#default", "#general")
		client.User.Rooms = append(client.User.Rooms, "#default/#general")
		room.Users[client.User.ID.String()] = client.User

		db.WriteUserData(client.User)
		db.WriteGroupData(defgroup)
		db.WriteRoomData(room)

		logger.Info("guest", client.User.ID.String(), "connected")
	} else {
		logger.Info("new client connected")
	}

	if err := client.Alert(types.OK, ""); err != nil {
		logger.Error(err)
	}

	/* TODO we may want to remove this later it's just for easy testing.
	 * to allow a client to get their UUID back from the server after
	 * connecting. */
	if config.AllowGuests {
		if err := client.Alert(types.GeneralNotice, string("Connected as guest with ID "+client.User.ID.String())); err != nil {
			logger.Error(err)
		}
	} else {
		if err := client.Alert(types.ImportantNotice, "No Guests Allowed : send authentication token to continue"); err != nil {
			logger.Error(err)
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

		if ban := ParseMessage(client, rawmsg); ban > 0 {
			// TODO handle ban-score
			break
		}
	}

	// We broke out of the loop so disconnect the client.
	client.Conn.Close()
	if client.User != nil {
		if client.User.Type == "guest" {
			if err := db.DeleteUser(client.User); err != nil {
				logger.Error(err)
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
