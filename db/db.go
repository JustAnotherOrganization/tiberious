package db

import (
	"strings"
	"tiberious/settings"
	"tiberious/types"
	"time"

	"gopkg.in/redis.v5"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

var (
	// ErrInvalidConfig is returned when the database is improperly setup and
	// no data can be written. This should never occur.
	ErrInvalidConfig = errors.New("Invalid config: no data written")

	// TestMode sets the db package to perform a few things differently then
	// it would otherwise.
	TestMode bool
)

type (
	// Client provides access to the database
	Client interface {
		Shutdown()
		GetKeySet(search string) ([]string, error)
		WriteUserData(user *types.User) error
		WriteRoomData(room *types.Room) error
		WriteGroupData(group *types.Group) error
		UserExists(id string) (bool, error)
		RoomExists(gname, rname string) (bool, error)
		GroupExists(gname string) (bool, error)
		GetUserData(id string) (*types.User, error)
		GetRoomData(gname, rname string) (*types.Room, error)
		GetGroupData(gname string) (*types.Group, error)
		DeleteUser(user *types.User) error

		RedisClient() *redis.Client
	}

	dbClient struct {
		config *settings.Config

		rdis rdisClient
		log  *logrus.Logger
	}
)

// NewDB returns a new database client
func NewDB(config *settings.Config, log *logrus.Logger) (Client, error) {
	client := &dbClient{
		config: config,
		log:    log,
	}

	if client.config.UserDatabase == 0 {
		// Load Redis DB
		var err error
		client.rdis, err = client.newRedisClient(log)
		if err != nil {
			return client, errors.Wrap(err, "newRedisClient")
		}

		client.log.Infof("User database started on redis db %d", client.config.DatabaseUser)
	}

	if !TestMode {
		if err := client.rdis.Client().SetNX("created", time.Now().String(), 0).Err(); err != nil {
			return client, errors.Wrap(err, "cliend.rdis.Client().SetNX")
		}
	}

	return client, nil
}

// Shutdown saves and quits database.
func (db *dbClient) Shutdown() {
	switch {
	case db.config.UserDatabase == 0:
		db.rdis.shutdown()
	default:
		break
	}
}

// GetKeySet returns all the keys that match a given search pattern.
func (db *dbClient) GetKeySet(search string) ([]string, error) {
	switch {
	case db.config.UserDatabase == 0:
		return db.rdis.getKeySet(search)
	default:
		break
	}

	return nil, types.NotInDB
}

// WriteUserData writes a given user object to the current database.
func (db *dbClient) WriteUserData(user *types.User) error {
	switch {
	case db.config.UserDatabase == 0:
		return db.rdis.writeUserData(user)
	default:
		return ErrInvalidConfig
	}
}

// WriteRoomData writes a given room object to the current database.
func (db *dbClient) WriteRoomData(room *types.Room) error {
	switch {
	case db.config.UserDatabase == 0:
		return db.rdis.writeRoomData(room)
	default:
		return ErrInvalidConfig
	}
}

// WriteGroupData writes a given group object to the current database.
func (db *dbClient) WriteGroupData(group *types.Group) error {
	switch {
	case db.config.UserDatabase == 0:
		return db.rdis.writeGroupData(group)
	default:
		return ErrInvalidConfig
	}
}

// UserExists returns whether a user exists in the database.
func (db *dbClient) UserExists(id string) (bool, error) {
	res, err := db.GetKeySet("user-*-*-" + id)
	if err != nil {
		return false, errors.Wrap(err, "GetKeySet")
	}

	if len(res) == 0 {
		return false, nil
	}

	return true, nil
}

// RoomExists returns whether a room exists in the database.
func (db *dbClient) RoomExists(gname, rname string) (bool, error) {
	res, err := db.GetKeySet("room-" + gname + "-" + rname + "*")
	if err != nil {
		return false, errors.Wrap(err, "GetKeySet")
	}

	if len(res) == 0 {
		return false, nil
	}

	return true, nil
}

// GroupExists returns whether a group exists in the database.
func (db *dbClient) GroupExists(gname string) (bool, error) {
	res, err := db.GetKeySet("group-" + gname + "-*")
	if err != nil {
		return false, errors.Wrap(err, "GetKeySet")
	}

	if len(res) == 0 {
		return false, nil
	}

	return true, nil
}

// GetUserData gets all the data for a given user ID from the database
func (db *dbClient) GetUserData(id string) (*types.User, error) {
	switch {
	case db.config.UserDatabase == 0:
		return db.rdis.getUserData(id)
	default:
		break
	}

	return nil, types.NotInDB
}

// GetRoomData gets all the data for a given room (group required) from the database
func (db *dbClient) GetRoomData(gname, rname string) (*types.Room, error) {
	switch {
	case db.config.UserDatabase == 0:
		return db.rdis.getRoomData(gname, rname)
	default:
		break
	}

	return nil, types.NotInDB
}

// GetGroupData gets all the data for a given group from the database
func (db *dbClient) GetGroupData(gname string) (*types.Group, error) {
	switch {
	case db.config.UserDatabase == 0:
		return db.rdis.getGroupData(gname)
	default:
		break
	}

	return nil, types.NotInDB
}

// DeleteUser removes a user from all rooms and groups and deletes it from the
// database (use sparingly). */
func (db *dbClient) DeleteUser(user *types.User) error {
	for _, gname := range user.Groups {
		group, err := db.GetGroupData(gname)
		if err != nil {
			return errors.Wrap(err, "GetGroupData")
		}

		delete(group.Users, user.ID.String())
		if err := db.WriteGroupData(group); err != nil {
			return errors.Wrap(err, "WriteGroupData")
		}
	}

	for _, rname := range user.Rooms {
		slice := strings.Split(rname, "/")
		room, err := db.GetRoomData(slice[0], slice[1])
		if err != nil {
			return errors.Wrap(err, "GetRoomData")
		}

		delete(room.Users, user.ID.String())
		if err := db.WriteRoomData(room); err != nil {
			return errors.Wrap(err, "WriteRoomData")
		}
	}

	switch {
	case db.config.UserDatabase == 0:
		return db.rdis.deleteUser(user)
	default:
		return ErrInvalidConfig
	}
}

// RedisClient returns the underlying *redis.Client for testing.
func (db *dbClient) RedisClient() *redis.Client {
	return db.rdis.Client()
}
