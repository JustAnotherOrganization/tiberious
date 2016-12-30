package db_test

import (
	"fmt"
	"testing"
	"tiberious/db"
	"tiberious/settings"

	"gopkg.in/redis.v5"

	"github.com/Sirupsen/logrus"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDB(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Redis DB tests")
}

type fixture struct {
	client db.Client
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
	f.client, err = db.NewDB(config, log)
	if err != nil {
		panic(err)
	}

	dbi := config.DatabaseUser
	for {
		var (
			sCmd *redis.StatusCmd
			iCmd *redis.IntCmd
		)

		if _, err := f.client.RedisClient().Pipelined(func(pipe *redis.Pipeline) error {
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
		f.client, err = db.NewDB(config, log)
		if err != nil {
			panic(err)
		}
	}
	fmt.Println("using database", dbi)
})

var _ = AfterSuite(func() {
	if err := f.client.RedisClient().FlushDb().Err(); err != nil {
		panic(err)
	}
})
