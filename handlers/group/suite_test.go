package group_test

import (
	"fmt"
	"testing"

	"gopkg.in/redis.v5"

	"tiberious/db"
	"tiberious/handlers/group"
	"tiberious/settings"

	"github.com/Sirupsen/logrus"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGroupHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Group Handler Suite")
}

type fixture struct {
	dbClient     db.Client
	groupHandler group.Handler
}

var f fixture

var _ = BeforeSuite(func() {
	log := logrus.New()

	fmt.Println("Before Suite")
	defer fmt.Println("Done Before Suite")

	db.TestMode = true

	f = fixture{}

	_, err := settings.Init(true)
	if err != nil {
		panic(err)
	}
	config := settings.GetConfig()

	reconnect := false
	f.dbClient, err = db.NewDB(config, log)
	if err != nil {
		panic(err)
	}

	dbi := config.DatabaseUser
	for {
		var (
			sCmd *redis.StatusCmd
			iCmd *redis.IntCmd
		)

		if _, err := f.dbClient.RedisClient().Pipelined(func(pipe *redis.Pipeline) error {
			sCmd = pipe.Select(dbi)
			iCmd = pipe.DbSize()
			return nil
		}); err != nil {
			panic(err)
		}
		size, err := iCmd.Result()
		if err != nil {
			panic(err)
		}

		if size > 0 {
			dbi++
			reconnect = true
			continue
		}

		break
	}

	if reconnect {
		// Quit panics with "not implemented" so for now just reassign a new client
		config.DatabaseUser = dbi
		f.dbClient, err = db.NewDB(config, log)
		if err != nil {
			panic(err)
		}
	}
	fmt.Println("using database", dbi)

	f.groupHandler, err = group.NewHandler(config, f.dbClient, log, "#test", "#general")
	if err != nil {
		panic(err)
	}
})

var _ = AfterSuite(func() {
	if err := f.dbClient.RedisClient().FlushDb().Err(); err != nil {
		panic(err)
	}
})
