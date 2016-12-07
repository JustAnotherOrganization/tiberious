package group

import (
	"tiberious/types"

	"github.com/pkg/errors"
)

// GetRoom check if a room exists (requires group) and if so return it
func (h *handler) GetRoom(gname, rname string) (*types.Room, error) {
	group, err := h.GetGroup(gname)
	if err != nil {
		return nil, errors.Wrap(err, "GetGroup")
	}
	if group == nil {
		return nil, nil
	}

	rexists, err := h.dbClient.RoomExists(gname, rname)
	if err != nil {
		return nil, errors.Wrap(err, "dbClient.RoomExists")
	}

	if !rexists {
		return nil, nil
	}

	room, err := h.dbClient.GetRoomData(gname, rname)
	if err != nil {
		return nil, errors.Wrap(err, "dbClient.GetRoomData")
	}

	return room, nil
}

// GetNewRoom should only be used if the room doesn't already exist in the
// provided group.
func (h *handler) GetNewRoom(gname, rname string) (*types.Room, error) {
	group, err := h.GetGroup(gname)
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

	if err := h.WriteGroupData(group); err != nil {
		return nil, errors.Wrap(err, "WriteGroupData")
	}

	return room, nil
}

// WriteRoomData writes the given room object to the current database.
func (h *handler) WriteRoomData(room *types.Room) error {
	return h.dbClient.WriteRoomData(room)
}
