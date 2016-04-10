package handlers

import (
	"tiberious/db"
	"tiberious/settings"
	"tiberious/types"
)

var (
	groups = make(map[string]*types.Group)
	config settings.Config
)

func init() {
	config = settings.GetConfig()

	defgroup := GetNewGroup("#default", true)
	genroom := GetNewRoom("#default", "#general")
	if config.UserDatabase != 0 {
		WriteGroupData(defgroup)
		WriteRoomData(genroom)
	}
}

// GetGroup check if a group exists and if so return it
func GetGroup(gname string) *types.Group {
	// Only the default group exists with UserDatabase disabled.
	if config.UserDatabase == 0 && gname != "default" {
		return nil
	}

	var gexists = false
	var group *types.Group
	for k, g := range groups {
		if gname == k {
			group = g
			gexists = true
			break
		}
	}

	if !gexists {
		return nil
	}

	return group
}

/*GetNewGroup should ony be used if the group doesn't already exist
 * and should not be called if the UserDatabase is disabled. */
func GetNewGroup(gname string, init bool) *types.Group {
	if config.UserDatabase == 0 {
		return nil
	}

	group := new(types.Group)
	group.Title = gname
	group.Rooms = make(map[string]*types.Room)
	group.Users = make(map[string]*types.User)
	groups[gname] = group
	return group
}

// WriteGroupData writes the given group object to the current database.
func WriteGroupData(group *types.Group) error {
	return db.WriteGroupData(group)
}

// GetRoom check if a room exists (requires group) and if so return it
func GetRoom(gname, rname string) *types.Room {
	group := GetGroup(gname)
	if group == nil {
		return nil
	}

	var rexists = false
	var room *types.Room

	for _, r := range group.Rooms {
		if rname == r.Title {
			room = r
			rexists = true
		}
	}

	if !rexists {
		return nil
	}

	return room
}

/*GetNewRoom should only be used if the room doesn't already exist in the
 * provided group. */
func GetNewRoom(gname, rname string) *types.Room {
	group := GetGroup(gname)
	if group == nil {
		return nil
	}

	room := new(types.Room)
	room.Users = make(map[string]*types.User)
	room.Title = rname
	room.Group = gname
	room.Private = false

	group.Rooms[rname] = room
	return room
}

// WriteRoomData writes the given room object to the current database.
func WriteRoomData(room *types.Room) error {
	return db.WriteRoomData(room)
}
