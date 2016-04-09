package settings

import (
	"io/ioutil"
	"log"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

var config Config

/* Since logger relies on settings for the file location of logs init errors
 * here are passed directly to the standard logger. */
func init() {
	// If no config file is found set defaults.
	configfile, err := filepath.Abs("./config.yml")
	if err != nil {
		log.Println(err)
		log.Println("Using default settings")
		setDefaults()
		return
	}
	// If unable to read the file set defaults.
	configyaml, err := ioutil.ReadFile(configfile)
	if err != nil {
		log.Println(err)
		log.Println("Using default settings")
		setDefaults()
		return
	}
	// If unable to parse the yaml set defaults.
	if err := yaml.Unmarshal([]byte(configyaml), &config); err != nil {
		log.Println(err)
		log.Println("Using default settings")
		setDefaults()
	}
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
	config.ErrorLog = ""
	config.DebugLog = ""
}

// GetConfig returns the current configuration file.
func GetConfig() Config {
	return config
}
