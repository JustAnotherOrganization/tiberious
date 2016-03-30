package types

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pborman/uuid"
)

// Client struct
type Client struct {
	// Store the connection interface for each client.
	Conn *websocket.Conn
	// Unique ID for all users.
	ID uuid.UUID
	// Store a client name.
	Name string
	/* Is a client authorized (logged in w/ a password), by default this is
	 * false. */
	Auth bool
}

// NewClient returns a Client
func NewClient() (client *Client) {
	return &Client{}
}

func (c Client) Alert(code int, message string) error {
	var ret []byte
	var err error
	if message != "" {
		ret, err = json.Marshal(AlertFull{Response: code, Time: time.Now().Unix(), Alert: message})
	} else {
		ret, err = json.Marshal(AlertMin{Response: code, Time: time.Now().Unix()})
	}

	if err != nil {
		return err
	}

	log.Println("returned", string(ret), "to client", c.ID.String())
	c.Conn.WriteMessage(websocket.BinaryMessage, ret)
	return nil
}

func (c Client) Error(code int, message string) error {
	var ret []byte
	var err error
	if message != "" {
		ret, err = json.Marshal(ErrorFull{Response: code, Time: time.Now().Unix(), Error: message})
	} else {
		ret, err = json.Marshal(ErrorMin{Response: code, Time: time.Now().Unix()})
	}

	if err != nil {
		return err
	}

	log.Println("returned", string(ret), "to client", c.ID.String())
	c.Conn.WriteMessage(websocket.BinaryMessage, ret)
	return nil
}
