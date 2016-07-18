package handlers

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RoomHandler", func() {
	Describe("Testing groups", func() {
		Describe("calling GetGroup", func() {
			It("default is always present", func() {
				str := "#default"
				g := GetGroup(str)
				Expect(g).NotTo(BeNil())
				Expect(g.Title).To(Equal(str))
			})
			It("group doesn't exist prior to creation", func() {
				Expect(GetGroup("#test")).To(BeNil())
			})
		})
		str := "#test"
		Describe("calling GetNewGroup", func() {
			It("GetNewGroup works properly", func() {
				g := GetNewGroup(str)
				Expect(g).NotTo(BeNil())
				Expect(g.Title).To(Equal(str))
			})
			It("GetNewGroup works properly if group already exists", func() {
				g := GetNewGroup(str)
				Expect(g).NotTo(BeNil())
				Expect(g.Title).To(Equal(str))
				Expect(WriteGroupData(g)).To(BeNil())
			})
		})
		Describe("calling GetGroup again after creation", func() {
			It("GetGroup works properly after creation", func() {
				g := GetGroup(str)
				Expect(g).NotTo(BeNil())
				Expect(g.Title).To(Equal(str))
			})
		})
	})
	Describe("Testing rooms", func() {
		Describe("calling GetRoom", func() {
			It("group doesn't exist", func() {
				r, err := GetRoom("#test2", "#anywhere")
				Expect(r).To(BeNil())
				Expect(err).ToNot(BeNil())
			})
			It("room doesn't exist", func() {
				r, err := GetRoom("#default", "#anywhere")
				Expect(r).To(BeNil())
				Expect(err).ToNot(BeNil())
			})
			It("#default/#general always exists", func() {
				str := "#general"
				r, err := GetRoom("#default", str)
				Expect(r).ToNot(BeNil())
				Expect(err).To(BeNil())
				Expect(r.Title).To(Equal(str))
				Expect(r.Private).To(BeFalse())
			})
		})
		Describe("calling GetNewRoom", func() {
			It("group doesn't exist", func() {
				r, err := GetNewRoom("#test2", "#anywhere")
				Expect(r).To(BeNil())
				Expect(err).ToNot(BeNil())
			})
			It("room already exists", func() {
				str := "#general"
				r, err := GetNewRoom("#default", str)
				Expect(r).ToNot(BeNil())
				Expect(err).To(BeNil())
				Expect(r.Title).To(Equal(str))
				Expect(r.Private).To(BeFalse())
			})
			It("works correctly", func() {
				str := "#test"
				r, err := GetNewRoom("#default", str)
				Expect(r).ToNot(BeNil())
				Expect(err).To(BeNil())
				Expect(r.Title).To(Equal(str))
				Expect(r.Private).To(BeFalse())
				Expect(WriteRoomData(r)).To(BeNil())
			})
		})
		Describe("calling GetRoom again after creation", func() {
			It("works correctly", func() {
				str := "#test"
				r, err := GetRoom("#default", str)
				Expect(r).ToNot(BeNil())
				Expect(err).To(BeNil())
				Expect(r.Title).To(Equal(str))
				Expect(r.Private).To(BeFalse())
			})
		})
	})
})
