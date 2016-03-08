package types

// AlertMin JIM minimum alert server response (message is only optional for 2xx alerts)
type AlertMin struct {
	Response int   `json:"response"`
	Time     int64 `json:"time"`
}

// AlertFull JIM standard alert server response with message included
type AlertFull struct {
	Response int    `json:"response"`
	Time     int64  `json:"time"`
	Alert    string `json:"alert"`
}
