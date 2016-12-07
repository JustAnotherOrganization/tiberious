package group_test

import (
	"fmt"
	"testing"
	"tiberious/db"
	"tiberious/handlers/group"
	"tiberious/settings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGroupHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Group Handler Suite")
}

type fixture struct {
	groupHandler group.Handler
}

var f fixture

var _ = BeforeSuite(func() {
	fmt.Println("Before Suite")
	defer fmt.Println("Done Before Suite")

	f = fixture{}

	config := settings.GetConfig()
	dbClient, err := db.NewDB(config)
	if err != nil {
		panic(err)
	}

	f.groupHandler, err = group.NewHandler(config, dbClient, "#test", "#general")
	if err != nil {
		panic(err)
	}
})
