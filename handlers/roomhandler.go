package handlers

import (
	"log"
	"strings"
	"tiberious/db"
	"tiberious/logger"
	"tiberious/settings"
	"tiberious/types"
)

var (
	defgroup *types.Group
	config   settings.Config
)

func init() {
	config = settings.GetConfig()
	defgroup = GetNewGroup("#default")
	/* We can't get a room from the default group without first writing the
	 * group data (on first start). */
	if err := WriteGroupData(defgroup); err != nil {
		log.Fatalln(err)
	}

	genroom := GetNewRoom("#default", "#general")
	if err := WriteRoomData(genroom); err != nil {
		log.Fatalln(err)
	}
}

// GetGroup check if a group exists and if so return it
func GetGroup(gname string) *types.Group {
	gexists, err := db.GroupExists(gname)
	if err != nil {
		logger.Error(err)
	}
	if !gexists {
		return nil
	}

	group, err := db.GetGroupData(gname)
	if err != nil {
		logger.Error(err)
	}

	return group
}

// GetNewGroup should ony be used if the group doesn't already exist
func GetNewGroup(gname string) *types.Group {
	group := new(types.Group)
	group.Title = gname
	group.Rooms = make(map[string]*types.Room)
	group.Users = make(map[string]*types.User)
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

	rexists, err := db.RoomExists(gname, rname)
	if err != nil {
		logger.Error(err)
	}

	if !rexists {
		return nil
	}

	room, err := db.GetRoomData(gname, rname)
	if err != nil {
		logger.Error(err)
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

	WriteGroupData(group)

	return room
}

// WriteRoomData writes the given room object to the current database.
func WriteRoomData(room *types.Room) error {
	return db.WriteRoomData(room)
}

// IsRoomName returns whether a string starts with "#"
func IsRoomName(str string) bool {
	return strings.HasPrefix(str, "#")
}
