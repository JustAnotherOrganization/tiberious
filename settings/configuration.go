package settings

import (
	"io/ioutil"
	"path/filepath"
	"tiberious/logger"
	"tiberious/types"

	"gopkg.in/yaml.v2"
)

var config types.Config

func init() {
	// If no config file is found set defaults.
	configfile, err := filepath.Abs("./config.yml")
	if err != nil {
		logger.Info(err)
		setDefaults()
		return
	}
	// If unable to read the file set defaults.
	configyaml, err := ioutil.ReadFile(configfile)
	if err != nil {
		logger.Info(err)
		setDefaults()
		return
	}
	// If unable to parse the yaml set defaults.
	if err := yaml.Unmarshal([]byte(configyaml), &config); err != nil {
		logger.Info(err)
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
}

// GetConfig returns the current configuration file.
func GetConfig() types.Config {
	return config
}
