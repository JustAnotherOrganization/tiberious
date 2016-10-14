package handlers

import (
	"log"
	"strconv"
	"strings"
	"tiberious/db"
	"tiberious/logger"
	"tiberious/settings"
	"tiberious/types"

	"github.com/gorilla/websocket"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
)

var (
	config   settings.Config
	dbClient db.Client

	clients = make(map[string]*types.Client)
)

func init() {
	config = settings.GetConfig()

	var err error
	dbClient, err = db.NewDB(config)
	if err != nil {
		log.Fatalln(err)
	}
}

func authenticate(client *types.Client, token types.AuthToken) (banScore int, err error) {
	banScore = 0
	keys, err := dbClient.GetKeySet("user-*-" + token.AccountName + "-*")
	if err != nil {
		err = errors.Wrap(err, "dbClient.GetKeySet")
		return
	}

	if len(keys) == 0 {
		banScore = 1
		if err = client.Error(types.IncorrectCredentials, ""); err != nil {
			err = errors.Wrap(err, "client.Error")
		}
		return
	}

	slice := strings.Split(keys[0], "-")
	user, err := dbClient.GetUserData(strings.Join(slice[3:], "-"))
	if err != nil {
		if err == types.NotInDB {
			if err = client.Error(types.IncorrectCredentials, ""); err != nil {
				err = errors.Wrap(err, "client.Error")
			}
		} else {
			err = errors.Wrap(err, "dbClient.GetUserData")
		}
		return
	}

	if token.Password != user.Password {
		banScore = 1
		if err = client.Error(types.IncorrectCredentials, ""); err != nil {
			err = errors.Wrap(err, "client.Error")
		}
		return
	}

	if client.User.Type == "guest" {
		if err = dbClient.DeleteUser(client.User); err != nil {
			err = errors.Wrap(err, "dbClient.DeleteUser")
		}
	}

	delete(clients, client.User.ID.String())

	/* TODO add any rooms that exist in the current user to the new user
	 * before discarding it. */
	client.User = user
	client.Authorized = true
	client.User.Connected = true
	if err = dbClient.WriteUserData(client.User); err != nil {
		err = errors.Wrap(err, "dbClient.WriteUserData")
	}

	clients[client.User.ID.String()] = client

	if err = client.Alert(types.OK, ""); err != nil {
		err = errors.Wrap(err, "client.Alert")
	}

	return
}

/* Always make sure a new ID is unique...
 * the probability of a UUID collision is somewhere around 1% in 100 million
 * UUIDs but we'll be overly cautious and check anyway. */
func getUniqueID() (uuid.UUID, error) {
	var id uuid.UUID
	for {
		id = uuid.NewRandom()
		exists, err := dbClient.UserExists(id.String())
		if err != nil {
			return nil, errors.Wrap(err, "dbClient.UserExists")
		}
		if exists {
			continue
		}
		break
	}

	return id, nil
}

// ClientHandler handles all client interactions
func ClientHandler(conn *websocket.Conn) {
	var err error

	client := types.NewClient()
	client.Conn = conn
	client.User = new(types.User)
	// Set the UUID and initialize a username of "guest"
	client.User.ID, err = getUniqueID()
	if err != nil {
		logger.Error(err)
	}

	guests, err := dbClient.GetKeySet("user-guest-*-*")
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
		defgroup, err := GetGroup("#default")
		if err != nil {
			logger.Error(err)
		}
		defgroup.Users[client.User.ID.String()] = client.User
		client.User.Groups = append(client.User.Groups, "#default")
		room, err := GetRoom("#default", "#general")
		if err != nil {
			logger.Error(err)
		}
		client.User.Rooms = append(client.User.Rooms, "#default/#general")
		room.Users[client.User.ID.String()] = client.User

		dbClient.WriteUserData(client.User)
		dbClient.WriteGroupData(defgroup)
		dbClient.WriteRoomData(room)

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

		ban, err := ParseMessage(client, rawmsg)
		if err != nil {
			logger.Info(err)
		}
		if ban > 0 {
			// TODO handle ban-score
			break
		}
	}

	// We broke out of the loop so disconnect the client.
	client.Conn.Close()
	if client.User != nil {
		if client.User.Type == "guest" {
			if err := dbClient.DeleteUser(client.User); err != nil {
				logger.Error(err)
			}
		} else {
			client.User.Connected = false
			dbClient.WriteUserData(client.User)
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
