package tiberious_test

import (
	"context"
	"database/sql"
	"flag"
	"io"
	"testing"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/stdlib"
	. "github.com/justanotherorganization/tiberious"
	pb "github.com/justanotherorganization/tiberious/proto/v1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestTiberious(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tiberious tests")
}

var (
	pgHost     = flag.String("pg_host", "localhost", "postgres host")
	pgPort     = flag.Int("pg_port", 5432, "postgres port")
	pgDatabase = flag.String("pg_database", "tiberious__test", "postgres database")
	pgUser     = flag.String("pg_user", "postgres", "postgres user")
	pgPassword = flag.String("pg_password", "", "postgres password")

	grpcAddr = flag.String("grpc_address", "localhost:4004", "the grpc address")
)

type fixture struct {
	tiberious Tiberious
	conn      *grpc.ClientConn
	client    pb.TiberiousClient
	stream    pb.Tiberious_StartStreamClient
	responses []*pb.StreamMessage
	count     int
	db        *sql.DB
}

var f fixture

var _ = BeforeSuite(func() {
	flag.Parse()

	driverConfig := stdlib.DriverConfig{
		ConnConfig: pgx.ConnConfig{
			Host:     *pgHost,
			Port:     uint16(*pgPort),
			Database: *pgDatabase,
			User:     *pgUser,
			Password: *pgPassword,
		},
	}
	stdlib.RegisterDriverConfig(&driverConfig)

	var err error
	f.db, err = sql.Open("pgx", driverConfig.ConnectionString(""))
	Expect(err).To(BeNil())

	f.tiberious, err = New(nil, f.db, nil)
	Expect(err).To(BeNil())

	go func() {
		Expect(f.tiberious.Start(*grpcAddr)).To(BeNil())
	}()

	f.conn, err = grpc.Dial(*grpcAddr, grpc.WithInsecure())
	Expect(err).To(BeNil())

	f.client = pb.NewTiberiousClient(f.conn)
	f.stream, err = f.client.StartStream(context.Background())
	Expect(err).To(BeNil())

	f.responses = []*pb.StreamMessage{}
	go func() {
		for {
			in, err := f.stream.Recv()
			if err != nil {
				if err == io.EOF {
					break
				}

				// An error is thrown during shutdown of the tests.
				stat, ok := status.FromError(err)
				Expect(ok).To(BeTrue())
				Expect(stat.Code()).To(Equal(codes.Unavailable))
			}
			if in != nil {
				f.responses = append(f.responses, in)
			}
		}
	}()
})

var _ = AfterSuite(func() {
	Expect(f.stream.CloseSend()).To(BeNil())
	Expect(f.conn.Close()).To(BeNil())
	Expect(f.db.Close()).To(BeNil())

	Expect(len(f.responses)).To(Equal(f.count))
})
