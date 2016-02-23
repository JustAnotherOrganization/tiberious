package types

import (
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
