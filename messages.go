package tiberious

import pb "github.com/justanotherorganization/tiberious/proto/v1"

type (
	ioMessage struct {
		clintID string
		message *pb.ClientMessage
	}
)

func (t *tiberious) manageMessagesRoutine() {
	defer t.wg.Done()

	for in := range t.incomingMessages {
		// Simply echo the message back for now
		t.outgoingMessages <- in

		// TODO: what should really happen here:
		// Validate the conversation_id on the message.
		// Relay the message to all currently connected users participating in
		// the conversation.
		// Write the message to the conversation archive (if archiving is enabled)
		// to allow offline users to fetch it later.
	}
}
