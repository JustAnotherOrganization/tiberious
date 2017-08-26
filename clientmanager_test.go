package tiberious

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

func TestClientManager(t *testing.T) {
	RegisterTestingT(t)

	log := logrus.NewEntry(logrus.New())
	cm := newClientManager(log, f.db)

	// TODO: utilize the cid from this.
	_, err := cm.newGuestID()
	Expect(err).To(BeNil())

	// FIXME: figure out how to make this work for internal tests separate
	// from the external gRPC tests.
	/*
		_client := cm.registerClient(cid, nil)
		Expect(_client).ToNot(BeNil())

		_client = cm.getClient(cid)
		Expect(_client).ToNot(BeNil())
	*/
}
