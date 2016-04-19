package types

// MasterObj for unmarshalling all client objects on the server side.
type MasterObj struct {
	// Found in all objects
	Action string `json:"action"`
	Time   int64  `json:"time"`
	// Only found in message objects
	To   string `json:"to"`
	From string `json:"from"`
	Body string `json:"message"`
	// Found in join/part messages
	Room string `json:"room"`
	// Found only in authentication messages
	User AuthToken `json:"user"`
}
