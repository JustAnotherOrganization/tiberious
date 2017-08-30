package tiberious

// Tiberious Client Manager.
// Currently this is mostly ephemeral (with the exception of the guest_id_seq),
// this is OK because if Tiberious crashes all the clients will be lost anyway.

import (
	"database/sql"
	"fmt"
	"sync"

	pb "github.com/justanotherorganization/tiberious/proto/v1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type (
	clientManager struct {
		log *logrus.Entry
		db  *sql.DB

		mu      *sync.RWMutex
		clients map[string]*client
	}

	client struct {
		stream pb.Tiberious_StartStreamServer
	}
)

func newClientManager(log *logrus.Entry, db *sql.DB) *clientManager {
	return &clientManager{
		log:     log,
		db:      db,
		mu:      &sync.RWMutex{},
		clients: make(map[string]*client),
	}
}

func (cm *clientManager) newGuestID() (string, error) {
	tx, err := cm.db.Begin()
	if err != nil {
		return "", errors.Wrap(err, "db.Begin")
	}

	defer func() {
		if err != nil {
			if e := tx.Rollback(); e != nil {
				cm.log.Error(errors.Wrap(e, "tx.Rollback"))
			}
		}
	}()

	sql := `SELECT nextval('guest_id_seq')`
	var id int64
	if err := tx.QueryRow(sql).Scan(&id); err != nil {
		return "", errors.Wrapf(err, "tx.QueryRow.Scan : sql : %s", sql)
	}

	if err := tx.Commit(); err != nil {
		return "", errors.Wrap(err, "tx.Commit")
	}

	return fmt.Sprintf("guest%d", id), nil
}

func (cm *clientManager) registerClient(clientID string, stream pb.Tiberious_StartStreamServer) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// TODO: refuse to register a client without a stream...
	// requires figuring out how to pass a stream from the internal tests...

	cm.clients[clientID] = &client{
		stream: stream,
	}
	return
}

func (cm *clientManager) getClient(clientID string) *client {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	client, ok := cm.clients[clientID]
	if !ok {
		return nil
	}

	return client
}

func (cm *clientManager) removeClient(clientID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	delete(cm.clients, clientID)
}
