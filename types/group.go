package types

type Group struct {
	Title string
	// List all the rooms in a group
	Rooms map[string]*Room
	// List all the users in a group
	Users map[string]*User
}
