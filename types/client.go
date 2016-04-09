package types

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

// Client struct
type Client struct {
	// Store the connection interface for each client.
	Conn *websocket.Conn
	// Store a user object for each client.
	User     *User
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
