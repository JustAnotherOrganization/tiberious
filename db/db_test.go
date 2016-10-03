package db

import (
	"tiberious/types"

	"github.com/pborman/uuid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	un         = "dbtest"
	email      = "dbtest@tiberious.nowhere"
	verySecret = "ThisIsNotASecret"
)

var (
	id = uuid.NewRandom()

	user = &types.User{
		ID:        id,
		Type:      "test",
		Username:  un,
		LoginName: un,
		Email:     email,
		Password:  verySecret,
		Salt:      verySecret,
		Connected: false,
		Rooms:     []string{"#testing/#testing"},
		Groups:    []string{"#testing"},
	}
)

var _ = Describe("db", func() {
	Describe("Calling WriteUserData", func() {
		It("works correctly", func() {
			err := f.client.WriteUserData(user)
			Expect(err).To(BeNil())
		})
	})

	Describe("Calling UserExists", func() {
		It("works correctly", func() {
			b, err := f.client.UserExists(id.String())
			Expect(err).To(BeNil())
			Expect(b).To(BeTrue())
		})
		It("user doesn't exist", func() {
			b, err := f.client.UserExists(uuid.NewRandom().String())
			Expect(err).To(BeNil())
			Expect(b).To(BeFalse())
		})
	})

	Describe("Calling GetUserData", func() {
		It("works correctly", func() {
			u, err := f.client.GetUserData(id.String())
			Expect(err).To(BeNil())
			Expect(u.Username).To(Equal(un))
			Expect(u.Password).To(Equal(verySecret))
		})
	})

	Describe("Calling WriteRoomData", func() {
		It("works correctly", func() {
			room := &types.Room{
				Title:   "#testing",
				Group:   "#testing",
				Private: false,
			}
			room.Users = make(map[string]*types.User)
			room.Users[user.ID.String()] = user

			err := f.client.WriteRoomData(room)
			Expect(err).To(BeNil())
		})
	})

	Describe("Calling RoomExists", func() {
		It("works correctly", func() {
			b, err := f.client.RoomExists("#testing", "#testing")
			Expect(err).To(BeNil())
			Expect(b).To(BeTrue())
		})
		It("room doesn't exist", func() {
			b, err := f.client.RoomExists("#default", "#foo")
			Expect(err).To(BeNil())
			Expect(b).To(BeFalse())
		})
	})

	Describe("Calling GetRoomData", func() {
		It("works correctly", func() {
			room, err := f.client.GetRoomData("#testing", "#testing")
			Expect(err).To(BeNil())
			Expect(room.Private).To(BeFalse())
			Expect(room.Users[id.String()].Username).To(Equal(un))
		})
	})

	Describe("Calling WriteGroupData", func() {
		It("works correctly", func() {
			users := make(map[string]*types.User)
			users[user.ID.String()] = user

			group := &types.Group{
				Title: "#testing",
			}
			group.Users = users

			room := &types.Room{
				Title:   "#testing",
				Group:   "#testing",
				Private: false,
			}
			group.Rooms = make(map[string]*types.Room)
			group.Rooms["#testing"] = room

			err := f.client.WriteGroupData(group)
			Expect(err).To(BeNil())
		})
	})

	Describe("Calling GroupExists", func() {
		It("works correctly", func() {
			b, err := f.client.GroupExists("#testing")
			Expect(err).To(BeNil())
			Expect(b).To(BeTrue())
		})
		It("group doesn't exist", func() {
			b, err := f.client.GroupExists("#foo")
			Expect(err).To(BeNil())
			Expect(b).To(BeFalse())
		})
	})

	Describe("Calling GetGroupData", func() {
		It("works correctly", func() {
			group, err := f.client.GetGroupData("#testing")
			Expect(err).To(BeNil())
			Expect(group.Rooms["#testing"].Title).To(Equal("#testing"))
			Expect(group.Users[id.String()].Username).To(Equal(un))
		})
	})
})
