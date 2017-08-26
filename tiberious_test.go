package tiberious_test

import (
	"time"

	pb "github.com/justanotherorganization/tiberious/proto/v1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tiberious grpc tests", func() {
	Describe("calling StartStream", func() {
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

			/*
				Expect(stream.Send(&pb.StreamMessage{
					StreamMessage: &pb.StreamMessage_ServerMessage{
						ServerMessage: &pb.ServerMessage{
							Meta: &pb.Meta{
								Time: pb.NewTimestamp(time.Now()),
							},
							Data: "A simple server message test!",
						},
					},
				})).To(BeNil())
			*/
		})
	})
	/*
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
		})
	*/
})
