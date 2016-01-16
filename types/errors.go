package types

// ErrorMin JIM minimum error response (message is only optional for errors)
type ErrorMin struct {
	Response int `json:"response"`
}

// ErrorFull JIM standard error response with optional message included
type ErrorFull struct {
	Response int    `json:"response"`
	Error    string `json:"error"`
}
