package types

import (
	"errors"
	"time"
)

const (
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

// Generic internal Go errors
var (
	NotInDB = errors.New("not found in db")
)

// Error JIM standard error response, error message may be optional.
type Error struct {
	Response int    `json:"response"`
	Time     int64  `json:"time"`
	Error    string `json:"error"`
}

// NewError returns a new error with the current timestamp.
func NewError(response int, message string) *Error {
	ret := new(Error)
	ret.Response = response
	ret.Time = time.Now().Unix()
	ret.Error = message
	return ret
}
