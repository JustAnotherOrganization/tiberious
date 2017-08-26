// Tiberious.
// Originally the idea was an open JSON based messaging protocol (JIM), the
// spec for which was never fully completed.
// JIM has since been more or less abandoned for an unnamed protocol buffer
// based messaging protocol (written specifically for Tiberious).
//
// About the name:
// The first existing JIM server (for testing only) was written in Node and
// was originally called NodeJIM. Do to the complete boring nature of this
// name it was renamed KirkNode in honor of the Star Trek character
// James T. Kirk (one of the more famous Jims that came to mind at the time).
// KirkNode was abandoned for the purpose of writing a more complete JIM
// server in Go; in keeping with the naming scheme it was decided that we
// would call it Tiberious for the T in James T. Kirk.
package tiberious

import (
	"database/sql"
	"encoding/json"
	"io"
	"time"

	"github.com/justanotherorganization/tiberious/netstuffs"
	pb "github.com/justanotherorganization/tiberious/proto/v1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	// Tiberious provides access to the underlying server.
	Tiberious interface {
		pb.TiberiousServer

		StartGRPC(string) error
	}

	tiberious struct {
		log    *logrus.Entry
		db     *sql.DB
		config *Config
		cm     *clientManager

		// Tiberious can run with only some connections enabled, because of
		// that some of these can and will be nil in certain situations.
		grpcServer netstuffs.GRPCServer
	}
)

// New creates a new Tiberious instance.
func New(log *logrus.Entry, db *sql.DB, config *Config) (Tiberious, error) {
	if log == nil {
		logger := logrus.New()
		log = logrus.NewEntry(logger)
	}

	if db == nil {
		return nil, errors.New("db required")
	}

	// If no config is provided load one with defaults?
	if config == nil {
		config = NewConfig()
	}

	return &tiberious{
		log:    log,
		db:     db,
		config: config,
		cm:     newClientManager(log, db),
	}, nil
}

// StartGRPC starts a GRPC server.
func (t *tiberious) StartGRPC(address string) error {
	t.log.Infof("starting grpc server on %s", address)
	t.grpcServer = netstuffs.New(address)
	pb.RegisterTiberiousServer(t.grpcServer.Server(), t)
	return t.grpcServer.Serve()
}

func (t *tiberious) StartStream(stream pb.Tiberious_StartStreamServer) error {
	//var _client *client
	// TODO:
	// If guests are enabled create a guest ID and register a client with the
	// cm. Otherwise require authentication before registering a client with
	// the cm.
	needsAuth := true
	if t.config.EnableGuests {
		cid, err := t.cm.newGuestID()
		if err != nil {
			return status.Error(codes.Internal, errors.Wrap(err, "cm.newGuestID").Error())
		}

		// FIXME: this should set client...
		_ = t.cm.registerClient(cid, stream)
		needsAuth = false
	}

	for {
		in, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				t.log.Info("stream terminated by client")
				return nil
			}

			return status.Error(codes.Internal, errors.Wrap(err, "stream.Recv").Error())
		}

		if in.GetServerMessage() != nil {
			return status.Error(codes.PermissionDenied, errors.New("ServerMessage received from client, disconnecting").Error())
		}

		if msg := in.GetClientMessage(); msg != nil {
			if needsAuth {
				// TODO: if guest access is disabled non-authenticated
				// client connections should be kicked off if they send
				// any message other than an authentication message.
			}

			// Either guest access is enabled or the user is authenticated
			// so we pass the message onto the message manager.
			t.log.Info(msg)
		}
	}
}

func (t *tiberious) NewConversation(ctx context.Context, request *pb.NewConversationRequest) (response *pb.Conversation, err error) {
	if err = request.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(err, "request.Validate").Error())
	}

	ts := pb.NewTimestamp(time.Now())
	response = &pb.Conversation{
		Created:    ts,
		TimeOfLast: ts,
	}

	byt, err := json.Marshal(response)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	tx, err := t.db.Begin()
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrap(err, "db.Begin").Error())
	}

	defer func() {
		if err != nil {
			if e := tx.Rollback(); e != nil {
				t.log.Error(errors.Wrap(e, "tx.Rollback"))
			}
		}
	}()

	sql := `INSERT INTO "conversations" (
		id,
		conversation
	) VALUES (
		nextval('conversation_id_seq'),
		$1
	) RETURNING id;`
	if err := tx.QueryRow(sql, byt).Scan(&response.Id); err != nil {
		return nil, status.Error(codes.Internal, errors.Wrapf(err, "tx.QueryRow.Scan : sql : %s", sql).Error())
	}

	if err := tx.Commit(); err != nil {
		return nil, status.Error(codes.Internal, errors.Wrap(err, "tx.Commit").Error())
	}

	return response, nil
}
