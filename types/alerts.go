package types

// AlertMin JIM minimum alert server response (message is only optional for 2xx alerts)
type AlertMin struct {
	Response int `json:"response"`
}

// AlertFull JIM standard alert server response with message included
type AlertFull struct {
	Response int    `json:"response"`
	Alert    string `json:"alert"`
}
