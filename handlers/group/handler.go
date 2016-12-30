package group

import (
	"strings"
	"tiberious/db"
	"tiberious/settings"
	"tiberious/types"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

type (
	// Handler provides access to our group and room handler.
	Handler interface {
		GetGroup(gname string) (*types.Group, error)
		GetNewGroup(gname string) *types.Group
		WriteGroupData(group *types.Group) error
		GetRoom(gname, rname string) (*types.Room, error)
		GetNewRoom(gname, rname string) (*types.Room, error)
		WriteRoomData(room *types.Room) error
		IsRoomName(str string) bool
	}

	handler struct {
		config   *settings.Config
		log      *logrus.Logger
		dbClient db.Client

		// TODO implement multiple groups
		defgroup *types.Group
	}
)

// NewHandler returns a new Handler using the provided config, dbClient and
// default group and room names.
func NewHandler(config *settings.Config, dbClient db.Client, log *logrus.Logger, defGroupName, genRoomName string) (Handler, error) {
	h := &handler{
		config:   config,
		log:      log,
		dbClient: dbClient,
	}

	defgroup := h.GetNewGroup(defGroupName)

	if err := h.WriteGroupData(defgroup); err != nil {
		return nil, errors.Wrap(err, "WriteGroupData")
	}

	h.defgroup = defgroup

	genroom, err := h.GetNewRoom(defGroupName, genRoomName)
	if err != nil {
		return nil, errors.Wrap(err, "GetNewRoom")
	}

	if err := h.WriteRoomData(genroom); err != nil {
		return nil, errors.Wrap(err, "WriteRoomData")
	}

	return h, nil
}

// IsRoomName returns whether a string starts with "#"
func (*handler) IsRoomName(str string) bool {
	return strings.HasPrefix(str, "#")
}
