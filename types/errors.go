package types

// ErrorMin JIM minimum error response (message is only optional for errors)
type ErrorMin struct {
	Response int   `json:"response"`
	Time     int64 `json:"time"`
}

// ErrorFull JIM standard error response with optional message included
type ErrorFull struct {
	Response int    `json:"response"`
	Time     int64  `json:"time"`
	Error    string `json:"error"`
}
