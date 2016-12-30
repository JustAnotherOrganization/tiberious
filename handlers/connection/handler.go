package connection

import (
	"tiberious/db"
	"tiberious/handlers/client"
	"tiberious/handlers/group"
	"tiberious/settings"
	"tiberious/types"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

type (
	// Handler provides access to handler.
	Handler interface {
		ListenAndServe()
	}

	handler struct {
		config        *settings.Config
		log           *logrus.Logger
		clientHandler client.Handler
		groupHandler  group.Handler
	}
)

// NewHandler returns a new Handler.
func NewHandler(config *settings.Config, log *logrus.Logger) (Handler, error) {
	dbClient, err := db.NewDB(config, log)
	if err != nil {
		return nil, errors.Wrap(err, "db.NewDB")
	}

	groupHandler, err := group.NewHandler(config, dbClient, log, "#default", "#general")
	if err != nil {
		return nil, errors.Wrap(err, "group.NewHandler")
	}

	clientHandler, err := client.NewHandler(config, dbClient, groupHandler, make(map[string]*types.Client), log)
	if err != nil {
		return nil, errors.Wrap(err, "client.NewHandler")
	}

	return &handler{
		config:        config,
		log:           log,
		clientHandler: clientHandler,
		groupHandler:  groupHandler,
	}, nil
}

// ListenAndServe starts our inbound routes.
func (h *handler) ListenAndServe() {
	go func() {
		if err := h.startWebsocketRoute(); err != nil {
			h.log.Fatal(err)
		}
	}()

	go func() {
		if err := h.startAPIRoute(); err != nil {
			h.log.Fatal(err)
		}
	}()
}
