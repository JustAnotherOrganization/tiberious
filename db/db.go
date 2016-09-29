package db

import (
	"log"
	"strings"
	"tiberious/settings"
	"tiberious/types"
)

var (
	config settings.Config
	rdis   rdisClient
)

func init() {
	config = settings.GetConfig()

	if config.UserDatabase == 0 {
		// Load Redis DB
		var err error
		rdis, err = newRedisClient()
		if err != nil {
			log.Fatalln("Unable to connect to redis database:", err)
		}
		log.Println("User database started on redis db", config.RedisUser)
	}
}

// GetKeySet returns all the keys that match a given search pattern.
func GetKeySet(search string) ([]string, error) {
	switch {
	case config.UserDatabase == 0:
		return rdis.getKeySet(search)
	default:
		break
	}

	return nil, types.NotInDB
}

// WriteUserData writes a given user object to the current database.
func WriteUserData(user *types.User) error {
	switch {
	case config.UserDatabase == 0:
		return rdis.writeUserData(user)
	default:
		break
	}

	// TODO log this, it should not occur
	return nil
}

// WriteRoomData writes a given room object to the current database.
func WriteRoomData(room *types.Room) error {
	switch {
	case config.UserDatabase == 0:
		return rdis.writeRoomData(room)
	default:
		break
	}

	// TODO log this, it should not occur
	return nil
}

// WriteGroupData writes a given group object to the current database.
func WriteGroupData(group *types.Group) error {
	switch {
	case config.UserDatabase == 0:
		return rdis.writeGroupData(group)
	default:
		break
	}

	return nil
}

// UserExists returns whether a user exists in the database.
func UserExists(id string) (bool, error) {
	res, err := GetKeySet("user-*-*-" + id)
	if err != nil {
		return false, nil
	}

	if len(res) == 0 {
		return false, nil
	}

	return true, nil
}

// RoomExists returns whether a room exists in the database.
func RoomExists(gname, rname string) (bool, error) {
	res, err := GetKeySet("room-" + gname + "-" + rname + "*")
	if err != nil {
		return false, err
	}

	if len(res) == 0 {
		return false, nil
	}

	return true, nil
}

// GroupExists returns whether a group exists in the database.
func GroupExists(gname string) (bool, error) {
	res, err := GetKeySet("group-" + gname + "-*")
	if err != nil {
		return false, err
	}

	if len(res) == 0 {
		return false, nil
	}

	return true, nil
}

// GetUserData gets all the data for a given user ID from the database
func GetUserData(id string) (*types.User, error) {
	switch {
	case config.UserDatabase == 0:
		return rdis.getUserData(id)
	default:
		break
	}

	return nil, types.NotInDB
}

// GetRoomData gets all the data for a given room (group required) from the database
func GetRoomData(gname, rname string) (*types.Room, error) {
	switch {
	case config.UserDatabase == 0:
		return rdis.getRoomData(gname, rname)
	default:
		break
	}

	return nil, types.NotInDB
}

// GetGroupData gets all the data for a given group from the database
func GetGroupData(gname string) (*types.Group, error) {
	switch {
	case config.UserDatabase == 0:
		return rdis.getGroupData(gname)
	default:
		break
	}

	return nil, types.NotInDB
}

// DeleteUser removes a user from all rooms and groups and deletes it from the
// database (use sparingly). */
func DeleteUser(user *types.User) error {
	for _, gname := range user.Groups {
		group, err := GetGroupData(gname)
		if err != nil {
			return err
		}

		delete(group.Users, user.ID.String())
		if err := WriteGroupData(group); err != nil {
			return err
		}
	}

	for _, rname := range user.Rooms {
		slice := strings.Split(rname, "/")
		room, err := GetRoomData(slice[0], slice[1])
		if err != nil {
			return err
		}

		delete(room.Users, user.ID.String())
		if err := WriteRoomData(room); err != nil {
			return err
		}
	}

	switch {
	case config.UserDatabase == 0:
		return rdis.deleteUser(user)
	default:
		break
	}

	return nil
}
