package types

// Message (used for channel and direct messages)
type Message struct {
	Action string `json:"action"`
	Time   int64  `json:"time"`
	To     string `json:"to"`
	From   string `json:"from"`
	Body   string `json:"message"`
}
