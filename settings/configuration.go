package settings

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"tiberious/types"

	"gopkg.in/yaml.v2"
)

var config types.Config

func init() {
	// If no config file is found set defaults.
	configfile, err := filepath.Abs("./config.yml")
	if err != nil {
		log.Println(err)
		setDefaults()
		return
	}
	// If unable to read the file set defaults.
	configyaml, err := ioutil.ReadFile(configfile)
	if err != nil {
		log.Println(err)
		setDefaults()
		return
	}
	// If unable to parse the yaml set defaults.
	if err := yaml.Unmarshal([]byte(configyaml), &config); err != nil {
		log.Println(err)
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
}

// GetConfig returns the current configuration file.
func GetConfig() types.Config {
	return config
}
