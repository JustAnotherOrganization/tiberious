package db

/* TODO make the presence of this optional and not loaded if redis is not
 * enabled for the application. If possible... */

import (
	"log"

	"gopkg.in/redis.v3"
)

var (
	rdis *redis.Client
)

func init() {
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
