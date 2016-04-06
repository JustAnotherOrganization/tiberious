package handlers

import (
	"tiberious/types"
)

var (
	rooms = make(map[string]*types.Room)
)

// GetRoom check if a room exists and if so return it
func GetRoom(rname string) (bool, *types.Room) {
	var rexists = false
	var room *types.Room
	for k, r := range rooms {
		if rname == k {
			room = r
			rexists = true
			break
		}
	}

	return rexists, room
}

// GetNewRoom should only be used if the room doesn't already exist
func GetNewRoom(rname string) *types.Room {
	room := new(types.Room)
	room.List = make(types.RoomList)
	room.Title = rname
	room.Private = false
	rooms[rname] = room
	return room
}
