package types

import "github.com/gorilla/websocket"

// Client struct
type Client struct {
	// Store the connection interface for each client.
	Conn *websocket.Conn
	// Unique ID for all users (go truncates this down (removing the 0s).
	ID int64
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
