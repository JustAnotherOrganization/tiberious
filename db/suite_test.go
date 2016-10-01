package db

import (
	"testing"
	"tiberious/settings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

func TestDB(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Redis DB tests")
}

var f fixture

var _ = BeforeSuite(func() {
	f = fixture{}
	if err := f.StartDB(); err != nil {
		panic(err)
	}
})

/*
var _ = AfterSuite(func() {
    f.StopDB()
})
*/

type fixture struct {
	client Client
}

func (f *fixture) StartDB() error {
	var err error
	f.client, err = NewDB(settings.GetConfig())
	if err != nil {
		return errors.Wrap(err, "NewDB")
	}

	return nil
}

/*
func (f *fixture) StopDB() {
    // Nothing to be done for redis
}
*/
