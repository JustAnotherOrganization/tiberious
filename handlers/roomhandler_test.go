package handlers

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"tiberious/types"
)

var _ = Describe("group handlers", func() {
	var g *types.Group

	Describe("Calling GetGroup on #test", func() {
		It("No database is enabled, should return nil", func() {
			g = GetGroup("#test")
			Expect(g).To(BeNil())
		})
	})

	Describe("Calling GetGroup on #default", func() {
		It("#default exists with or with out a database", func() {
			g = GetGroup("#default")
			Expect(g).NotTo(BeNil())
			Expect(g.Title).To(Equal("#default"))
		})
	})
})

var _ = Describe("new group handlers", func() {
	var g *types.Group

	Describe("Calling GetNewGroup on #test", func() {
		It("", func() {
			g = GetNewGroup("#anything")
			Expect(g).NotTo(BeNil())
			Expect(g.Title).To(Equal("#anything"))
		})
	})

	Describe("Calling GetNewGroup on #test", func() {
		It("should get group", func() {
			g = GetNewGroup("#test")
			Expect(g).NotTo(BeNil())
			Expect(g.Title).To(Equal("#test"))
		})
	})
})

var _ = Describe("group room", func() {
	var r *types.Room

	Describe("Calling GetRoom on #anything, #anywhere", func() {
		It("No database is enabled so only the #default and #test groups exist", func() {
			r = GetRoom("#anything", "#anywhere")
			Expect(r).To(BeNil())
		})
	})

	Describe("Calling GetRoom on #default, #anywhere", func() {
		It("#anywhere should not exist", func() {
			r = GetRoom("#default", "#anywhere")
			Expect(r).To(BeNil())
		})
	})

	Describe("Calling GetRoom on #default, #general", func() {
		It("#default/#general exists by default", func() {
			r = GetRoom("#default", "#general")
			Expect(r).NotTo(BeNil())
			Expect(r.Title).To(Equal("#general"))
		})
	})
})

var _ = Describe("new room", func() {
	var r *types.Room

	Describe("Calling GetNewRoom on #anything, #anywhere", func() {
		It("#anything group does not exist", func() {
			r = GetNewRoom("#anything", "#anywhere")
			Expect(r).To(BeNil())
		})
	})

	Describe("Calling GetRoom on #default, #anywhere", func() {
		It("#default exists by default", func() {
			r = GetNewRoom("#default", "#anywhere")
			Expect(r).NotTo(BeNil())
			Expect(r.Title).To(Equal("#anywhere"))
		})
	})
})
