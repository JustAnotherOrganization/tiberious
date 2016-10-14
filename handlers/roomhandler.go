package handlers

import (
	"log"
	"strings"
	"tiberious/types"

	"github.com/pkg/errors"
)

var (
	defgroup *types.Group
)

func init() {
	defgroup = GetNewGroup("#default")
	/* We can't get a room from the default group without first writing the
	 * group data (on first start). */
	if err := WriteGroupData(defgroup); err != nil {
		log.Fatalln(err)
	}

	genroom, err := GetNewRoom("#default", "#general")
	if err != nil {
		log.Fatalln(err)
	}

	if err := WriteRoomData(genroom); err != nil {
		log.Fatalln(err)
	}
}

// GetGroup check if a group exists and if so return it
func GetGroup(gname string) (*types.Group, error) {
	gexists, err := dbClient.GroupExists(gname)
	if err != nil {
		return nil, errors.Wrap(err, "dbClient.GroupExists")
	}
	if !gexists {
		return nil, nil
	}

	group, err := dbClient.GetGroupData(gname)
	if err != nil {
		return nil, errors.Wrap(err, "dbClient.GetGroupData")
	}

	return group, nil
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
	return dbClient.WriteGroupData(group)
}

// GetRoom check if a room exists (requires group) and if so return it
func GetRoom(gname, rname string) (*types.Room, error) {
	group, err := GetGroup(gname)
	if err != nil {
		return nil, errors.Wrap(err, "GetGroup")
	}
	if group == nil {
		return nil, nil
	}

	rexists, err := dbClient.RoomExists(gname, rname)
	if err != nil {
		return nil, errors.Wrap(err, "dbClient.RoomExists")
	}

	if !rexists {
		return nil, nil
	}

	room, err := dbClient.GetRoomData(gname, rname)
	if err != nil {
		return nil, errors.Wrap(err, "dbClient.GetRoomData")
	}

	return room, nil
}

/*GetNewRoom should only be used if the room doesn't already exist in the
 * provided group. */
func GetNewRoom(gname, rname string) (*types.Room, error) {
	group, err := GetGroup(gname)
	if err != nil {
		return nil, errors.Wrap(err, "GetGroup")
	}
	if group == nil {
		return nil, nil
	}

	room := new(types.Room)
	room.Users = make(map[string]*types.User)
	room.Title = rname
	room.Group = gname
	room.Private = false

	group.Rooms[rname] = room

	if err := WriteGroupData(group); err != nil {
		return nil, errors.Wrap(err, "WriteGroupData")
	}

	return room, nil
}

// WriteRoomData writes the given room object to the current database.
func WriteRoomData(room *types.Room) error {
	return dbClient.WriteRoomData(room)
}

// IsRoomName returns whether a string starts with "#"
func IsRoomName(str string) bool {
	return strings.HasPrefix(str, "#")
}
