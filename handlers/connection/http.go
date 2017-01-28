package connection

// TODO either remove/replace go-json-rest or see about implementing logrus
// in a middleware layer to unify logging...
//
// Add perms!!!

import (
	"net/http"
	"tiberious/types"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/pkg/errors"
)

func (h *handler) getClients(w rest.ResponseWriter, req *rest.Request) {
	type clients struct {
		Clients map[string]*types.Client
	}

	c := clients{Clients: h.clientHandler.GetClients()}
	if err := w.WriteJson(&c); err != nil {
		h.log.Error(errors.Wrap(err, "w.WriteJson"))
	}
}

// TODO getRooms and getGroups (requires modifications to db)
//
// getRoom and getGroup (requires either json formatted requests or replacing
// go-json-rest with a less strict API layer).

func (h *handler) startAPIRoute() error {
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Get("/clients", h.getClients),
	)
	if err != nil {
		return errors.Wrap(err, "rest.MakeRouter")
	}

	api.SetApp(router)
	return http.ListenAndServe(":8080", api.MakeHandler())
}
