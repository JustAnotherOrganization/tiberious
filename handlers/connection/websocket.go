package connection

import (
	"net/http"

	"github.com/gorilla/websocket"
)

func (h *handler) newWebsocketConnection(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  h.config.ReadBufferSize,
		WriteBufferSize: h.config.WriteBufferSize,
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); ok {
			w.WriteHeader(404)
			if _, er2 := w.Write([]byte("Invalid websocket handshake")); er2 != nil {
				h.log.Error(er2)
			}
			return
		}
		h.log.Error(err)
		return
	}

	go h.clientHandler.HandleConnection(conn)
}

func (h *handler) startWebsocketRoute() error {
	http.HandleFunc("/", http.NotFound)
	http.HandleFunc("/ws", h.newWebsocketConnection)
	h.log.Infof("Starting Tiberious on %s", h.config.Port)
	return http.ListenAndServe(h.config.Port, nil)
}
