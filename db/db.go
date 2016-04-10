package db

import (
	"tiberious/settings"
	"tiberious/types"
)

var config settings.Config

func init() {
	config = settings.GetConfig()
}

// WriteUserData writes a given user object to the current database.
func WriteUserData(user *types.User) error {
	switch {
	case config.UserDatabase == 1:
		if err := rdis.HMSet(
			"user-"+user.Type+"-"+user.ID.String(),
			"id", user.ID.String(),
			"type", user.Type,
			"username", user.Username,
			"loginname", user.LoginName,
			"email", user.Email,
			"password", user.Password,
			"salt", user.Salt,
			"connected", strbool(user.Connected),
			"authorized", strbool(user.Authorized),
		).Err(); err != nil {
			return err
		}

		for _, r := range user.Rooms {
			if err := rdis.SAdd("user-"+user.Type+"-"+user.ID.String()+"-rooms", r).Err(); err != nil {
				return err
			}
		}

		for _, g := range user.Groups {
			if err := rdis.SAdd("user-"+user.Type+"-"+user.ID.String()+"-groups", g).Err(); err != nil {
				return err
			}
		}
		break
	default:
		break
	}

	return nil
}

// WriteRoomData writes a given room object to the current database.
func WriteRoomData(room *types.Room) error {
	switch {
	case config.UserDatabase == 1:
		var private = "false"
		if room.Private {
			private = "true"
		}

		if err := rdis.HMSet("room-"+room.Group+"-"+room.Title+"-info", "private", private).Err(); err != nil {
			return err
		}

		for _, u := range room.Users {
			if err := rdis.SAdd("room-"+room.Group+"-"+room.Title+"-list", u.ID.String()).Err(); err != nil {
				return err
			}
		}
		break
	default:
		break
	}

	return nil
}

// WriteGroupData writes a given group object to the current database.
func WriteGroupData(group *types.Group) error {
	switch {
	case config.UserDatabase == 1:
		for _, r := range group.Rooms {
			if err := rdis.SAdd("group-"+group.Title+"-rooms", r.Title).Err(); err != nil {
				return err
			}
		}

		for _, u := range group.Users {
			if err := rdis.SAdd("group-"+group.Title+"-users", u.ID.String()).Err(); err != nil {
				return err
			}
		}
		break
	default:
		break
	}

	return nil
}

// UserExists returns whether a user exists in the database.
func UserExists(id string) (bool, error) {
	switch {
	case config.UserDatabase == 1:
		res, err := rdis.Keys("user-*-" + id).Result()
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
