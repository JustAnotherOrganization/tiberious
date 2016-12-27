package connection

import (
	"net/http"
	"tiberious/handlers/client"
	"tiberious/settings"
	"tiberious/types"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

type (
	// Handler provides access to handler.
	Handler interface {
		ListenAndServe() error
	}

	handler struct {
		config        *settings.Config
		log           *logrus.Logger
		clientHandler client.Handler
	}
)

// NewHandler returns a new Handler
func NewHandler(config *settings.Config, log *logrus.Logger) (Handler, error) {
	clientHandler, err := client.NewHandler(config, make(map[string]*types.Client), log)
	if err != nil {
		return nil, errors.Wrap(err, "client.NewHandler")
	}

	return &handler{
		config:        config,
		log:           log,
		clientHandler: clientHandler,
	}, nil
}

func (h *handler) newClientConnection(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  h.config.ReadBufferSize,
		WriteBufferSize: h.config.WriteBufferSize,
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

	go h.clientHandler.HandleConnection(conn)
}

// ListenAndServe is a wrapper around http.ListenAndServe
func (h *handler) ListenAndServe() error {
	http.HandleFunc("/", http.NotFound)
	http.HandleFunc("/ws", h.newClientConnection)
	h.log.Infof("Starting Tiberious on %s", h.config.Port)
	return http.ListenAndServe(h.config.Port, nil)
}
