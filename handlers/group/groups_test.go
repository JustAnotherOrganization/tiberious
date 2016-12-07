package group_test

import (
	"tiberious/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("group handler", func() {
	Describe("calling GetGroup", func() {
		It("Group doesn't exist and there is no error", func() {
			g, err := f.groupHandler.GetGroup("#foo")
			Expect(err).To(BeNil())
			Expect(g).To(BeNil())
		})
		It("Should return a group", func() {
			g, err := f.groupHandler.GetGroup("#test")
			Expect(err).To(BeNil())
			Expect(g).ToNot(BeNil())
			Expect(g.Title).To(Equal("#test"))
		})
	})
	var g *types.Group
	Describe("calling GetNewGroup", func() {
		It("GetNewGroup should return a group", func() {
			g = f.groupHandler.GetNewGroup("#foo")
			Expect(g).ToNot(BeNil())
			Expect(g.Title).To(Equal("#foo"))
		})
	})
	Describe("calling WriteGroupData", func() {
		It("Should return nil", func() {
			Expect(f.groupHandler.WriteGroupData(g))
		})
	})
})
