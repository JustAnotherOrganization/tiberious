package db

import (
	"log"
	"tiberious/logger"

	"gopkg.in/redis.v3"
)

var rdis *redis.Client

func init() {
	if config.UserDatabase == 0 {
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

// Stupid helper function because redis only handles strings.
func strbool(b bool) string {
	if b {
		return "true"
	}

	return "false"
}

func boolstr(s string) bool {
	if s == "true" {
		return true
	}

	return false
}

/* This seems extremely cumbersome but it's the best way I can think to handle
 * this without deleting the entire set and recreating it. */
func updateSet(key string, new []string) {
	old, err := rdis.SMembers(key).Result()
	if err != nil {
		logger.Error(err)
	}

	for _, o := range old {
		var rem = true
		for _, n := range new {
			if o == n {
				rem = false
			}
		}

		if rem {
			if err := rdis.SRem(key, o).Err(); err != nil {
				logger.Error(err)
			}
		}
	}

	for _, n := range new {
		var add = true
		for _, o := range old {
			if n == o {
				add = false
			}
		}

		if add {
			if err := rdis.SAdd(key, n).Err(); err != nil {
				logger.Error(err)
			}
		}
	}
}
