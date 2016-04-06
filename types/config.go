package types

// Config will hold all the standard configuration data for Tiberious
type Config struct {
	/* Port used to set what port to run the server on, this should be
	 * formatted like the following `:4002` (JIM standard ports are 4002
	 * through 4006) */
	Port string `yaml:"port"`
	// TODO benchmark and tune buffer-size defaults
	// ReadBufferSize used to set the websocket ReadBufferSize (default 1024)
	ReadBufferSize int `yaml:"readbuffer"`
	// WriteBufferSize used to set the websocket WriteBufferSize (default 1024)
	WriteBufferSize int `yaml:"writebuffer"`
	/* MessageStore defines whether or not to store messages for retrieval at
	 * a later time instead of throwing them out after passing them on. */
	MessageStore bool `yaml:"messagestore"`
	/* MessageExpire can be used to define whether to expire stored messages
	 * (if MessageStore is enabled)  after a given time period in days (0 means
	 * they do not expire). */
	MessageExpire int `yaml:"messageexpire"`
	/* MessageOverflow can be used to set a max number of stored messages (if
	 * MessageStore is enabled) where it will start deleting old messages after
	 * the Overflow size is reached (0 means they do not overflow) */
	MessageOverflow int `yaml:"messageoverflow"`
	/* UserDatabase can be used to select the database type (1 == Redis).
	 * If set to 1 RedisHost, RedisPass and RedisUser must be set. */
	UserDatabase int `yaml:"userdatabase"`
	/* RedisHost will need to be set if using built in redis handling:
	 * example: localhost:6379
	 */
	RedisHost string `yaml:"redishost"`
	// RedisPass will need to be set if using build in redis handling.
	RedisPass string `yaml:"redispass"`
	/* RedisUser should be set to the database number to use in redis.
	 * Unless using redis for something else 0 should suffice. */
	RedisUser int64 `yaml:"redisuser"`
	// ErrorLog allows setting a file location to log errors to.
	ErrorLog string `yaml:"errorlog"`
	// DebugLog allows setting a file location to log standard information.
	DebugLog string `yaml:"debuglog"`
}
