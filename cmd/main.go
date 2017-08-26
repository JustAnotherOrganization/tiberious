package main

import (
	"database/sql"
	"flag"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/stdlib"
	"github.com/justanotherorganization/tiberious"
	"github.com/sirupsen/logrus"
)

var (
	pgHost     = flag.String("pg_host", "localhost", "postgres host")
	pgPort     = flag.Int("pg_port", 5432, "postgres port")
	pgDatabase = flag.String("pg_database", "tiberious", "postgres database")
	pgUser     = flag.String("pg_user", "postgres", "postgres user")
	pgPassword = flag.String("pg_password", "", "postgres password")

	grpcAddr = flag.String("grpc_address", "localhost:4004", "the grpc address")
)

func main() {
	flag.Parse()

	logger := logrus.New()
	log := logrus.NewEntry(logger)

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

	db, err := sql.Open("pgx", driverConfig.ConnectionString(""))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	t, err := tiberious.New(log, db)
	if err != nil {
		log.Error(err)
		return
	}

	log.Error(t.StartGRPC(*grpcAddr))
}
