package types

import (
	"encoding/json"
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
	Auth     bool
	BanScore int
}

const (
	// GeneralNotice for general informational alerts
	GeneralNotice = 100
	// ImportantNotice for importand informational alerts
	ImportantNotice = 101
	// OK response, to confirm an action completed
	OK = 200
	// Created response code
	Created = 201
	// Accepted response code
	Accepted = 202
	// BadRequestOrObject response code
	BadRequestOrObject = 400
	// NotAuthorized response code
	NotAuthorized = 401
	// IncorrectCredentials response code
	IncorrectCredentials = 402
	// Forbidden response code
	Forbidden = 403
	// NotFound response code
	NotFound = 404
	// Conflict response code
	Conflict = 409
	// Gone response code
	Gone = 410
	// ServerError response code
	ServerError = 500
)

// NewClient returns a Client
func NewClient() (client *Client) {
	return &Client{}
}

// Alert handles both full and minimal alerts.
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

	c.Conn.WriteMessage(websocket.BinaryMessage, ret)
	return nil
}

// Error handles both full and minimal errors.
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

	c.Conn.WriteMessage(websocket.BinaryMessage, ret)
	return nil
}
