package handlers

import (
	"tiberious/logger"
	"tiberious/types"

	"github.com/gorilla/websocket"
	"github.com/pborman/uuid"
)

var (
	clients = make(map[string]*types.Client)
)

// ClientHandler handles all client interactions
func ClientHandler(conn *websocket.Conn) {
	client := types.NewClient()
	client.Conn = conn
	// Set the UUID and initialize a username of "guest"
	client.ID = uuid.NewRandom()

	clients[client.ID.String()] = client
	logger.Info("client", client.ID, "connected")

	if err := client.Alert(types.OK, ""); err != nil {
		logger.Error(err)
	}

	/* TODO we may want to remove this later it's just for easy testing.
	 * to allow a client to get their UUID back from the server after
	 * connecting. */
	if err := client.Alert(types.GeneralNotice, string("Connected with ID "+client.ID.String())); err != nil {
		logger.Error(err)
	}

	// TODO handle authentication for servers with user databases.

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
