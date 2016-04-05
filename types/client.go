package types

import (
	"encoding/json"

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
	Auth     bool
	BanScore int
}

// NewClient returns a Client
func NewClient() (client *Client) {
	return &Client{}
}

// Alert sends an alert with the current timestamp
func (c Client) Alert(code int, message string) error {
	ret, err := json.Marshal(NewAlert(code, message))

	if err != nil {
		return err
	}

	c.Conn.WriteMessage(websocket.BinaryMessage, ret)
	return nil
}

// Error sends an error with the current timestamp
func (c Client) Error(code int, message string) error {
	ret, err := json.Marshal(NewError(code, message))

	if err != nil {
		return err
	}

	c.Conn.WriteMessage(websocket.BinaryMessage, ret)
	return nil
}
