package tiberious_test

import (
	"context"
	"time"

	pb "github.com/justanotherorganization/tiberious/proto/v1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tiberious grpc tests", func() {
	var skipCount bool
	var _ = AfterEach(func() {
		if !skipCount {
			f.count++
		}

		skipCount = false
	})

	Describe("sending a message over the stream", func() {
		Context("with a client message", func() {
			It("works correctly", func() {
				Expect(f.stream.Send(&pb.StreamMessage{
					StreamMessage: &pb.StreamMessage_ClientMessage{
						ClientMessage: &pb.ClientMessage{
							Meta: &pb.Meta{
								Time: pb.NewTimestamp(time.Now()),
							},
							ConversationId: 42,
							Data:           []byte("A simple client message test"),
						},
					},
				})).To(BeNil())
			})
		})

		/* FIXME: figure out why the expected behavior doesn't occur...
		Context("with a server message", func() {
			It("disconnects the stream", func() {
				Expect(f.stream.Send(&pb.StreamMessage{
					StreamMessage: &pb.StreamMessage_ServerMessage{
						ServerMessage: &pb.ServerMessage{
							Meta: &pb.Meta{
								Time: pb.NewTimestamp(time.Now()),
							},
							Data: "A simple server message test!",
						},
					},
				})).ToNot(BeNil())
			})
		})
		*/
	})

	Describe("calling NewConversation", func() {
		newConversationRequest := func() *pb.NewConversationRequest {
			return &pb.NewConversationRequest{
				Meta: &pb.Meta{
					Time: pb.NewTimestamp(time.Now()),
					// TODO: fill in signatures...
					// TODO: fill in participants...
				},
			}
		}

		It("works correctly, even though it's missing a lot", func() {
			req := newConversationRequest()
			_, err := f.client.NewConversation(context.Background(), req)
			Expect(err).To(BeNil())
		})

		skipCount = true
	})
})
