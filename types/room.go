package types

// Room struct contains the RoomList and RoomFlags for a given room.
type Room struct {
	Title   string
	Group   string
	Private bool
	Users   map[string]*User
}
