package types

type User struct {
	Username  string
	LoginName string
	Email     strign
	Password  string //TODO Hash this
	Salt      string //TODO Implement salting.
}
