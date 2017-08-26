package tiberious_test

import (
	"context"
	"database/sql"
	"flag"
	"testing"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/stdlib"
	. "github.com/justanotherorganization/tiberious"
	pb "github.com/justanotherorganization/tiberious/proto/v1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
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
		if err := f.tiberious.StartGRPC(*grpcAddr); err != nil {
			panic(err)
		}
	}()

	f.conn, err = grpc.Dial(*grpcAddr, grpc.WithInsecure())
	Expect(err).To(BeNil())

	f.client = pb.NewTiberiousClient(f.conn)
	f.stream, err = f.client.StartStream(context.Background())
	Expect(err).To(BeNil())
})

var _ = AfterSuite(func() {
	Expect(f.stream.CloseSend()).To(BeNil())
	Expect(f.conn.Close()).To(BeNil())
	Expect(f.db.Close()).To(BeNil())
})
