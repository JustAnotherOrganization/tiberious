package main

import (
	"net/http"
	"os"

	"tiberious/handlers/client"
	"tiberious/settings"
	"tiberious/types"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
)

var (
	config *settings.Config
	log    *logrus.Logger

	clientHandler client.Handler
)

// TODO move this into a separate handler of some form.
func newConnection(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  config.ReadBufferSize,
		WriteBufferSize: config.WriteBufferSize,
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); ok {
			w.WriteHeader(404)
			w.Write([]byte("Invalid websocket handshake"))
			return
		}
		logrus.Error(err)
		return
	}

	go clientHandler.HandleConnection(conn)
}

func init() {
	log = logrus.New()

	// Errors or warnings encountered when initializing settings will always
	// print to stdout because we don't know if there's a log file set in the
	// config yet.
	note, err := settings.Init()
	if err != nil {
		log.Println(err)
	}
	log.Println(note)

	config = settings.GetConfig()

	// At this point we have our config so set the log location in logrus (if
	// found in config).
	if config.Log != "" {
		writer, err := os.OpenFile(config.Log, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal(err)
		}

		log.Out = writer
	}
}

func main() {
	config = settings.GetConfig()
	var err error
	clientHandler, err = client.NewHandler(config, make(map[string]*types.Client), log)
	if err != nil {
		log.Error(err)
	}

	http.HandleFunc("/", http.NotFound)
	http.HandleFunc("/ws", newConnection)
	log.Infof("Starting Tiberious on %s", config.Port)
	if err := http.ListenAndServe(config.Port, nil); err != nil {
		log.Error(err)
	}
}
