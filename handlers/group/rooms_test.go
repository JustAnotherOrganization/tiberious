package group_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("room handler", func() {
	Describe("Calling GetRoom", func() {
		It("Group doesn't exist so neither will room", func() {
			r, err := f.groupHandler.GetRoom("#bar", "#general")
			Expect(err).To(BeNil())
			Expect(r).To(BeNil())
		})
		It("Group exists but room doesn't", func() {
			r, err := f.groupHandler.GetRoom("#test", "#foo")
			Expect(err).To(BeNil())
			Expect(r).To(BeNil())
		})
		It("Group and Room exist", func() {
			r, err := f.groupHandler.GetRoom("#test", "#general")
			Expect(err).To(BeNil())
			Expect(r).ToNot(BeNil())
			Expect(r.Title).To(Equal("#general"))
		})
	})
	Describe("Calling GetNewRoom", func() {
		It("Group doesn't exist so we can't create a room", func() {
			r, err := f.groupHandler.GetNewRoom("#bar", "#foo")
			Expect(err).To(BeNil())
			Expect(r).To(BeNil())
		})
		It("Room already exists", func() {
			r, err := f.groupHandler.GetNewRoom("#test", "#general")
			Expect(err).To(BeNil())
			Expect(r).ToNot(BeNil())
			Expect(r.Title).To(Equal("#general"))
		})
		It("Should return a new room", func() {
			r, err := f.groupHandler.GetNewRoom("#test", "#foo")
			Expect(err).To(BeNil())
			Expect(r).ToNot(BeNil())
			Expect(r.Title).To(Equal("#foo"))
		})
	})
})
