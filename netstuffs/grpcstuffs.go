package netstuffs

import (
	"net"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type (
	// GRPCServer provides access to the underlying grpcServer.
	GRPCServer interface {
		Server() *grpc.Server
		Serve() error
	}

	grpcServer struct {
		server  *grpc.Server
		address string
	}
)

// New returns a new GRPCServer.
func New(address string) GRPCServer {
	return &grpcServer{
		server:  grpc.NewServer(),
		address: address,
	}
}

// Server returns the underlying grpc server.
func (s *grpcServer) Server() *grpc.Server {
	return s.server
}

// Serve a GRPC connection.
func (s *grpcServer) Serve() error {
	lis, err := net.Listen("tcp", s.address)
	if err != nil {
		return errors.Wrap(err, "tcp for grpc failed to listen")
	}

	return s.server.Serve(lis)
}
