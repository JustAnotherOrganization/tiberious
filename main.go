package main

import (
	"io"
	"os"

	"tiberious/handlers/connection"
	"tiberious/settings"

	"github.com/Sirupsen/logrus"
)

func main() {
	log := logrus.New()

	// Errors or warnings encountered when initializing settings will always
	// print to stdout because we don't know if there's a log file set in the
	// config yet.
	note, err := settings.Init(false)
	if err != nil {
		log.Println(err)
	}
	log.Println(note)

	config := settings.GetConfig()

	// At this point we have our config so set the log location in logrus (if
	// found in config).
	if config.Log != "" {
		var writer io.Writer
		writer, err = os.OpenFile(config.Log, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal(err)
		}

		log.Out = writer
	}

	log.Level = logrus.DebugLevel

	connectionHandler, err := connection.NewHandler(config, log)
	if err != nil {
		log.Fatal(err)
	}

	// TODO add shutdown/stop logic
	connectionHandler.ListenAndServe()
	for {
		select {}
	}
}
