package types

import "time"

// Message (used for channel and direct messages)
type Message struct {
	Action string `json:"action"`
	Time   int64  `json:"time"`
	To     string `json:"to"`
	From   string `json:"from"`
	Body   string `json:"message"`
}

// NewMessage returns a standard pre-constructed "msg"
func NewMessage(to, from, body string) *Message {
	ret := new(Message)
	ret.Action = "msg"
	ret.Time = time.Now().Unix()
	ret.To = to
	ret.From = from
	ret.Body = body
	return ret
}
