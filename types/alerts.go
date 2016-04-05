package types

import "time"

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
)

// Alert JIM standard alert server response, alert message may be optional
type Alert struct {
	Response int    `json:"response"`
	Time     int64  `json:"time"`
	Alert    string `json:"alert"`
}

// NewAlert returns a new alert with the current timestamp.
func NewAlert(response int, message string) *Alert {
	ret := new(Alert)
	ret.Response = response
	ret.Time = time.Now().Unix()
	ret.Alert = message
	return ret
}
