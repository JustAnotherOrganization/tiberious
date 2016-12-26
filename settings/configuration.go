package settings

import (
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

var config *Config

const usingDefaults = "Using default settings"

// Init loads are configuration data.
func Init() (string, error) {
	/* Set default values then overwrite them with ones in the yml, this way
	 * if something is missing from the yml the defaults apply properly. */
	config = &Config{}
	setDefaults()
	// If no config file is found set defaults.
	configfile, err := filepath.Abs("./config.yml")
	if err != nil {
		return usingDefaults, err
	}
	// If unable to read the file set defaults.
	configyaml, err := ioutil.ReadFile(configfile)
	if err != nil {
		return usingDefaults, err
	}
	// If unable to parse the yaml set defaults.
	if err := yaml.Unmarshal([]byte(configyaml), &config); err != nil {
		return usingDefaults, err
	}

	return "Settings loaded from config.yml", nil
}

func setDefaults() {
	config.Port = ":4002"
	// TODO benchmark and tune default buffer-sizes
	config.ReadBufferSize = 1024
	config.WriteBufferSize = 1024
	config.MessageStore = false
	config.MessageExpire = 0
	config.MessageOverflow = 0
	config.UserDatabase = 0
	config.Log = ""
	config.AllowGuests = true
	config.DatabaseAddress = "localhost:6379"
	config.DatabasePass = ""
	config.DatabaseUser = 0
}

// GetConfig returns the current configuration file.
func GetConfig() *Config {
	return config
}
