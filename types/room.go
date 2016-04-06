package types

// RoomList maps contain all the clients for a given room.
type RoomList map[string]*Client

// Room struct contains the RoomList and RoomFlags for a given room.
type Room struct {
	Title   string
	Private bool
	List    RoomList
}
