package handlers

import (
	"tiberious/db"
	"tiberious/logger"
	"tiberious/settings"
	"tiberious/types"

	"github.com/gorilla/websocket"
	"github.com/pborman/uuid"
)

var (
	clients = make(map[string]*types.Client)
)

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

	client.User.Type = "default"

	clients[client.User.ID.String()] = client
	logger.Info("client", client.User.ID, "connected")

	if err := client.Alert(types.OK, ""); err != nil {
		logger.Error(err)
	}

	/* TODO we may want to remove this later it's just for easy testing.
	 * to allow a client to get their UUID back from the server after
	 * connecting. */
	if err := client.Alert(types.GeneralNotice, string("Connected with ID "+client.User.ID.String())); err != nil {
		logger.Error(err)
	}

	// TODO handle authentication for servers with user databases.

	if settings.GetConfig().UserDatabase != 0 {
		db.WriteUserData(client.User)
	}

	/* Never return from this loop!
	 * Never break from this loop unless intending to disconnect the client. */
	for {
		_, rawmsg, err := client.Conn.ReadMessage()
		if err != nil {
			logger.Info(err)
			break
		}

		if ban := ParseMessage(client, rawmsg); ban > 0 {
			// TODO handle ban-score
			break
		}
	}

	// We broke out of the loop so disconnect the client.
	client.Conn.Close()
}
