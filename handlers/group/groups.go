package group

import (
	"tiberious/types"

	"github.com/pkg/errors"
)

// GetGroup check if a group exists and if so return it.
func (h *handler) GetGroup(gname string) (*types.Group, error) {
	gexists, err := h.dbClient.GroupExists(gname)
	if err != nil {
		return nil, errors.Wrap(err, "dbClient.GroupExists")
	}

	if !gexists {
		return nil, nil
	}

	group, err := h.dbClient.GetGroupData(gname)
	if err != nil {
		return nil, errors.Wrap(err, "dbClient.GetGroupData")
	}

	return group, nil
}

// GetNewGroup should ony be used if the group doesn't already exist.
func (*handler) GetNewGroup(gname string) *types.Group {
	return &types.Group{
		Title: gname,
		Rooms: make(map[string]*types.Room),
		Users: make(map[string]*types.User),
	}
}

// WriteGroupData writes the given group object to the current database.
func (h *handler) WriteGroupData(group *types.Group) error {
	return h.dbClient.WriteGroupData(group)
}
