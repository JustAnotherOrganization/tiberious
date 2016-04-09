package redis

/* TODO make the presence of this optional and not loaded if redis is not
 * enabled for the application. If possible... */

import (
	"log"

	"tiberious/settings"

	"gopkg.in/redis.v3"
)

var (
	config settings.Config
	rdis   *redis.Client
)

func init() {
	config = settings.GetConfig()

	if config.UserDatabase == 1 {
		if config.RedisHost == "" {
			log.Fatalln("Missing redishost in config file")
		}

		if config.RedisPass == "" {
			log.Println("Insecure redis database is not recommended")
		}

		rdis = redis.NewClient(&redis.Options{
			Addr:     config.RedisHost,
			Password: config.RedisPass,
			DB:       config.RedisUser,
		})

		// Confirm we can communicate with the redis instance.
		_, err := rdis.Ping().Result()
		if err != nil {
			log.Fatalln("Unable to connect to redis database:", err)
		}

		log.Println("User database started on redis db", config.RedisUser)
	}
}

// GetRedis returns the current redis object.
func GetRedis() *redis.Client {
	return rdis
}

// Set a value for a given key.
func Set(key string, value string, client *redis.Client) error {
	if client == nil {
		client = rdis
	}

	if err := client.Set(key, value, 0).Err(); err != nil {
		return err
	}

	return nil
}

// Get a redis value for a given key.
func Get(key string, client *redis.Client) (string, error) {
	if client == nil {
		client = rdis
	}

	value, err := client.Get(key).Result()
	if err != nil {
		return "", err
	}

	return value, nil
}
