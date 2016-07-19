package types

import (
	"encoding/json"

	"github.com/JustAnotherOrganization/jgordon"
	"github.com/gorilla/websocket"
)

// Client struct
type Client struct {
	// Store the connection interface for each client.
	Conn *websocket.Conn
	// Store a user object for each client.
	User       *User
	Authorized bool
	banScore   int
}

// NewClient returns a Client
func NewClient() (client *Client) {
	return &Client{}
}

// Alert sends an alert with the current timestamp
func (c Client) Alert(code int32, message string) error {
	ret, err := json.Marshal(jgordon.NewAlert(code, message))

	if err != nil {
		return err
	}

	c.Conn.WriteMessage(websocket.BinaryMessage, ret)
	return nil
}

// Error sends an error with the current timestamp
func (c Client) Error(code int32, message string) error {
	ret, err := json.Marshal(jgordon.NewError(code, message))

	if err != nil {
		return err
	}

	c.Conn.WriteMessage(websocket.BinaryMessage, ret)
	return nil
}

// RaiseBan raises the clients banScore by i.
func (c Client) RaiseBan(i int) {
	c.banScore = c.banScore + i
}

// BanScore returns the clients current banScore
func (c Client) BanScore() int {
	return c.banScore
}
