package db

import (
	"strings"
	"tiberious/settings"
	"tiberious/types"

	"github.com/pborman/uuid"
	"github.com/pkg/errors"
)

/*
Data map:
	Redis:
		Users:
			Main Key: "user-"+<user-type>+"-"+<loginname>+<uuid> (hash)
			Joined Rooms: "user-"+<user-type+"-"+<uuid>+"rooms" (set)
			Joined Groups: "user-"+<user-type+"-"+<uuid>+"groups" (set)
		Rooms:
			Info: "room-"+<group name>+<room name>+"-info" (hash)
			User List: "room-"+<group name>+<room name>+"-list" (set)
		Groups:
			Info: "group-"+<group name>+"info" (hash)
			User List: "group-"+<group name>+"-users" (set)
			Room List: "group-"+<group name>+"-rooms" (set)
*/

var (
	config settings.Config
	// ErrNotInDB ...
	ErrNotInDB = errors.New("not found in db")
)

func init() {
	config = settings.GetConfig()
}

// GetKeySet returns all the keys that match a given search pattern.
func GetKeySet(search string) ([]string, error) {
	switch {
	case config.UserDatabase == 0:
		return rdis.Keys(search).Result()
	default:
		break
	}

	return nil, ErrNotInDB
}

// WriteUserData writes a given user object to the current database.
func WriteUserData(user *types.User) error {
	switch {
	case config.UserDatabase == 0:
		if err := rdis.HMSet(
			"user-"+user.Type+"-"+user.LoginName+"-"+user.ID.String(),
			"id", user.ID.String(),
			"type", user.Type,
			"username", user.Username,
			"loginname", user.LoginName,
			"email", user.Email,
			"password", user.Password,
			"salt", user.Salt,
			"connected", strbool(user.Connected),
		).Err(); err != nil {
			return errors.Wrap(err, "rdis.HMSet")
		}

		go updateSet("user-"+user.Type+"-"+user.ID.String()+"-rooms", user.Rooms)
		go updateSet("user-"+user.Type+"-"+user.ID.String()+"-groups", user.Groups)
		break
	default:
		break
	}

	return nil
}

// WriteRoomData writes a given room object to the current database.
func WriteRoomData(room *types.Room) error {
	switch {
	case config.UserDatabase == 0:
		str := "room-" + room.Group + "-" + room.Title + "-info"
		if err := rdis.HMSet(str, "title", room.Title, "group", room.Group, "private", strbool(room.Private)).Err(); err != nil {
			return errors.Wrapf(err, "rdis.HMSet %s", str)
		}

		var slice []string
		for _, u := range room.Users {
			slice = append(slice, u.ID.String())
		}

		go updateSet("room-"+room.Group+"-"+room.Title+"-list", slice)
		break
	default:
		break
	}

	return nil
}

// WriteGroupData writes a given group object to the current database.
func WriteGroupData(group *types.Group) error {
	switch {
	case config.UserDatabase == 0:
		str := "group-" + group.Title + "-info"
		if err := rdis.HSet(str, "title", group.Title).Err(); err != nil {
			return errors.Wrapf(err, "rdis.HSet %s", str)
		}

		var slice []string
		for _, r := range group.Rooms {
			slice = append(slice, r.Title)
		}
		go updateSet("group-"+group.Title+"-rooms", slice)

		slice = nil
		for _, u := range group.Users {
			slice = append(slice, u.ID.String())
		}

		go updateSet("group-"+group.Title+"-users", slice)

		break
	default:
		break
	}

	return nil
}

// UserExists returns whether a user exists in the database.
func UserExists(id string) (bool, error) {
	switch {
	case config.UserDatabase == 0:
		str := "user-*-*-" + id
		res, err := GetKeySet(str)
		if err != nil {
			return false, errors.Wrapf(err, "GetKeySet %s", str)
		}

		if len(res) == 0 {
			return false, nil
		}

		return true, nil
	default:
		break
	}

	return false, nil
}

// RoomExists returns whether a room exists in the database.
func RoomExists(gname, rname string) (bool, error) {
	switch {
	case config.UserDatabase == 0:
		str := "room-" + gname + "-" + rname + "*"
		res, err := GetKeySet(str)
		if err != nil {
			return false, errors.Wrapf(err, "GetKeySet %s", str)
		}

		if len(res) == 0 {
			return false, nil
		}

		return true, nil
	default:
		break
	}

	return false, nil
}

// GroupExists returns whether a group exists in the database.
func GroupExists(gname string) (bool, error) {
	switch {
	case config.UserDatabase == 0:
		str := "group-" + gname + "-*"
		res, err := GetKeySet(str)
		if err != nil {
			return false, errors.Wrapf(err, "GetKeySet %s", str)
		}

		if len(res) == 0 {
			return false, nil
		}

		return true, nil
	default:
		break
	}

	return false, nil
}

// GetUserData gets all the data for a given user ID from the database
func GetUserData(id string) (*types.User, error) {
	switch {
	case config.UserDatabase == 0:
		user := new(types.User)

		str := "user-*-*-" + id
		keys, err := GetKeySet(str)
		if err != nil {
			return nil, errors.Wrapf(err, "GetKeySet %s", str)
		}

		if len(keys) == 0 {
			return nil, nil
		}

		info, err := rdis.HGetAllMap(keys[0]).Result()
		if err != nil {
			return nil, errors.Wrapf(err, "rdis.HGetAllMap %s", keys[0])
		}

		user.ID = uuid.Parse(info["id"])
		user.Type = info["type"]
		user.Username = info["username"]
		user.LoginName = info["loginname"]
		user.Email = info["email"]
		user.Password = info["password"]
		user.Salt = info["salt"]
		user.Connected = boolstr(info["connected"])

		str = "user-" + user.Type + "-" + user.ID.String() + "-rooms"
		rooms, err := rdis.SMembers(str).Result()
		if err != nil {
			return nil, errors.Wrapf(err, "rdis.SMembers %s", str)
		}

		if len(rooms) > 0 {
			user.Rooms = rooms
		}

		str = "user-" + user.Type + "-" + user.ID.String() + "-groups"
		groups, err := rdis.SMembers(str).Result()
		if err != nil {
			return nil, errors.Wrapf(err, "rdis.SMembers %s", str)
		}

		if len(groups) > 0 {
			user.Groups = groups
		}

		return user, nil
	default:
		break
	}

	return nil, ErrNotInDB
}

// GetRoomData gets all the data for a given room (group required) from the database
func GetRoomData(gname, rname string) (*types.Room, error) {
	switch {
	case config.UserDatabase == 0:
		room := new(types.Room)
		room.Users = make(map[string]*types.User)

		str := "room-" + gname + "-" + rname + "-info"
		info, err := rdis.HGetAllMap(str).Result()
		if err != nil {
			return nil, errors.Wrapf(err, "rdis.HGetAllMap %s", err)
		}

		room.Title = info["title"]
		room.Group = info["group"]
		room.Private = boolstr(info["private"])

		str = "room-" + gname + "-" + rname + "-list"
		users, err := rdis.SMembers(str).Result()
		if err != nil {
			return nil, errors.Wrapf(err, "rdis.SMembers %s", str)
		}

		room.Users = make(map[string]*types.User)
		if len(users) > 0 {
			for _, v := range users {
				u, err := GetUserData(v)
				if err != nil {
					return nil, errors.Wrapf(err, "GetUserData %s", v)
				}
				room.Users[u.ID.String()] = u
			}
		}

		return room, nil
	default:
		break
	}

	return nil, ErrNotInDB
}

// GetGroupData gets all the data for a given group from the database
func GetGroupData(gname string) (*types.Group, error) {
	switch {
	case config.UserDatabase == 0:
		group := new(types.Group)

		group.Title = gname
		group.Rooms = make(map[string]*types.Room)
		group.Users = make(map[string]*types.User)

		str := "group-" + gname + "-users"
		users, err := rdis.SMembers(str).Result()
		if err != nil {
			return nil, errors.Wrapf(err, "rdis.SMembers %s", str)
		}

		if len(users) > 0 {
			for _, v := range users {
				/* For some reason the length of this keeps coming up as 1 above
				 * the actual number of entries so confirm it's not nil before
				 * attempting to run GetUserData with the given string. */
				if v != "" {
					u, stat := GetUserData(v)
					if stat != nil {
						return nil, errors.Wrapf(stat, "GetUserData %s", v)
					}
					group.Users[u.ID.String()] = u
				}
			}
		}

		str = "group-" + gname + "-rooms"
		rooms, err := rdis.SMembers(str).Result()
		if err != nil {
			return nil, errors.Wrapf(err, "rdis.SMembers %s", str)
		}

		if len(rooms) > 0 {
			for _, v := range rooms {
				if v != "" {
					r, err := GetRoomData(gname, v)
					if err != nil {
						return nil, errors.Wrapf(err, "GetRoomData %s", v)
					}
					group.Rooms[r.Title] = r
				}
			}
		}

		return group, nil
	default:
		break
	}

	return nil, ErrNotInDB
}

/*DeleteUser removes a user from all rooms and groups and deletes it from the
 * database (user sparingly). */
func DeleteUser(user *types.User) error {
	for _, gname := range user.Groups {
		group, err := GetGroupData(gname)
		if err != nil {
			return errors.Wrapf(err, "GetGroupData %s", gname)
		}

		delete(group.Users, user.ID.String())
		if err := WriteGroupData(group); err != nil {
			return errors.Wrapf(err, "WriteGroupData %s", group)
		}
	}

	for _, rname := range user.Rooms {
		slice := strings.Split(rname, "/")
		room, err := GetRoomData(slice[0], slice[1])
		if err != nil {
			return errors.Wrapf(err, "GetRoomData %s/%s", slice[0], slice[1])
		}

		delete(room.Users, user.ID.String())
		if err := WriteRoomData(room); err != nil {
			return errors.Wrapf(err, "WriteRoomData %s", room)
		}
	}

	switch {
	case config.UserDatabase == 0:
		str := "user-" + user.Type + "-" + user.ID.String() + "-groups"
		if err := rdis.Del(str).Err(); err != nil {
			return errors.Wrapf(err, "rdis.Del", str)
		}
		str = "user-" + user.Type + "-" + user.ID.String() + "-rooms"
		if err := rdis.Del(str).Err(); err != nil {
			return errors.Wrapf(err, "rdis.Del", str)
		}
		str = "user-" + user.Type + "-" + user.LoginName + "-" + user.ID.String()
		if err := rdis.Del(str).Err(); err != nil {
			return errors.Wrapf(err, "rdis.Del", str)
		}
	default:
		break
	}

	return nil
}
