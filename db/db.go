package db

import (
	"tiberious/settings"
	"tiberious/types"

	"github.com/pborman/uuid"
)

var config settings.Config

func init() {
	config = settings.GetConfig()
}

/*NewUser generates a new UUID for a newly connected user and if using a
 * database initializes a hset with the UUID as the key and a username of
 * "guest". */
func NewUser(client *types.Client) error {
	id := uuid.NewRandom()
	client.ID = id
	switch {
	// userdatabase method 1, redis
	case config.UserDatabase == 1:
		redis := GetRedis()
		if err := redis.HMSet(client.ID.String(), "username", "guest", "registered", "false").Err(); err != nil {
			return err
		}
		break
	default:
		break
	}

	return nil
}
