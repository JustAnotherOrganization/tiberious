package tiberious

import (
	"database/sql"
	"flag"
	"os"
	"testing"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/stdlib"
	"github.com/sirupsen/logrus"
)

var (
	// TODO: this needs to be in a config since used by our suite_test.go file
	// as well.
	// FIXME: figure out why this is clashing with the flags defined in suite_test
	// when using the same env names.
	pgHost     = flag.String("pg_host_wtf", "localhost", "postgres host")
	pgPort     = flag.Int("pg_port_wtf", 5432, "postgres port")
	pgDatabase = flag.String("pg_database_wtf", "tiberious__test", "postgres database")
	pgUser     = flag.String("pg_user_wtf", "postgres", "postgres user")
	pgPassword = flag.String("pg_password_wtf", "", "postgres password")
)

type fixture struct {
	db *sql.DB
}

var f fixture

// TestMain sets up a DB for running internal tests separate from the external
// GRPC API tests.
func TestMain(m *testing.M) {
	logger := logrus.New()
	log := logrus.NewEntry(logger)

	log.Infoln("Starting internal tests")
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
	if err != nil {
		panic(err)
	}

	x := m.Run()
	f.db.Close()

	os.Exit(x)
}
