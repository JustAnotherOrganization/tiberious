package db

import (
	"strings"
	"tiberious/settings"
	"tiberious/types"

	"github.com/pborman/uuid"
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

var config settings.Config

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

	return nil, types.NotInDB
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
			return err
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
		if err := rdis.HMSet("room-"+room.Group+"-"+room.Title+"-info", "title", room.Title, "group", room.Group, "private", strbool(room.Private)).Err(); err != nil {
			return err
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
		if err := rdis.HSet("group-"+group.Title+"-info", "title", group.Title).Err(); err != nil {
			return err
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
		res, err := GetKeySet("user-*-*-" + id)
		if err != nil {
			return false, nil
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
		res, err := GetKeySet("room-" + gname + "-" + rname + "*")
		if err != nil {
			return false, err
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
		res, err := GetKeySet("group-" + gname + "-*")
		if err != nil {
			return false, err
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

		keys, err := GetKeySet("user-*-*" + id)
		if err != nil {
			return nil, err
		}

		info, err := rdis.HGetAllMap(keys[0]).Result()
		if err != nil {
			return nil, err
		}

		user.ID = uuid.Parse(info["id"])
		user.Type = info["type"]
		user.Username = info["username"]
		user.LoginName = info["loginname"]
		user.Email = info["email"]
		user.Password = info["password"]
		user.Salt = info["salt"]
		user.Connected = boolstr(info["connected"])

		rooms, err := rdis.SMembers("user-" + user.Type + "-" + user.ID.String() + "-rooms").Result()
		if err != nil {
			return nil, err
		}

		if len(rooms) > 0 {
			user.Rooms = rooms
		}

		groups, err := rdis.SMembers("user-" + user.Type + "-" + user.ID.String() + "-groups").Result()
		if err != nil {
			return nil, err
		}

		if len(groups) > 0 {
			user.Groups = groups
		}

		return user, nil
	default:
		break
	}

	return nil, types.NotInDB
}

// GetRoomData gets all the data for a given room (group required) from the database
func GetRoomData(gname, rname string) (*types.Room, error) {
	switch {
	case config.UserDatabase == 0:
		room := new(types.Room)
		room.Users = make(map[string]*types.User)

		info, err := rdis.HGetAllMap("room-" + gname + "-" + rname + "-info").Result()
		if err != nil {
			return nil, err
		}

		room.Title = info["title"]
		room.Group = info["group"]
		room.Private = boolstr(info["private"])

		users, err := rdis.SMembers("room-" + gname + "-" + rname + "-list").Result()
		if err != nil {
			return nil, err
		}

		if len(users) > 0 {
			for _, v := range users {
				u, err := GetUserData(v)
				if err != nil {
					return nil, err
				}
				room.Users[u.ID.String()] = u
			}
		}

		return room, nil
	default:
		break
	}

	return nil, types.NotInDB
}

// GetGroupData gets all the data for a given group from the database
func GetGroupData(gname string) (*types.Group, error) {
	switch {
	case config.UserDatabase == 0:
		group := new(types.Group)

		group.Title = gname
		group.Rooms = make(map[string]*types.Room)
		group.Users = make(map[string]*types.User)

		users, err := rdis.SMembers("group-" + gname + "-users").Result()
		if err != nil {
			return nil, err
		}

		if len(users) > 0 {
			for _, v := range users {
				/* For some reason the length of this keeps coming up as 1 above
				 * the actual number of entries so confirm it's not nil before
				 * attempting to run GetUserData with the given string. */
				if v != "" {
					u, stat := GetUserData(v)
					if stat != nil {
						return nil, stat
					}
					group.Users[u.ID.String()] = u
				}
			}
		}

		rooms, err := rdis.SMembers("group-" + gname + "-rooms").Result()
		if err != nil {
			return nil, err
		}

		if len(rooms) > 0 {
			for _, v := range rooms {
				if v != "" {
					r, err := GetRoomData(gname, v)
					if err != nil {
						return nil, err
					}
					group.Rooms[r.Title] = r
				}
			}
		}

		return group, nil
	default:
		break
	}

	return nil, types.NotInDB
}

/*DeleteUser removes a user from all rooms and groups and deletes it from the
 * database (user sparingly). */
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
		if err := rdis.Del("user-" + user.Type + "-" + user.ID.String() + "-groups").Err(); err != nil {
			return err
		}
		if err := rdis.Del("user-" + user.Type + "-" + user.ID.String() + "-rooms").Err(); err != nil {
			return err
		}
		return rdis.Del("user-" + user.Type + "-" + user.LoginName + "-" + user.ID.String()).Err()
	default:
		break
	}

	return nil
}
