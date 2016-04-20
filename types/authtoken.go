package types

// AuthToken used inside of an authentication message.
type AuthToken struct {
	AccountName string `json:"account_name"`
	Password    string `json:"password"`
}
