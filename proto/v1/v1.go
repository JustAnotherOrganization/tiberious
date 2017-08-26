// The Tiberious Protocol.
// Designed from scratch for the Tiberious messaging server this protocol
// is meant to be an open protocol alternative to XMPP and IRC.
// It is meant to allow for encrypted as well as plain text communications
// within the same environment while both securing the senders privacy and
// proving their identity to the recipient.
//
// This is a work in progress and is highly experimental. To date
// encryption has not been tested over the Tiberious Protocol, it is merely
// designed with it in mind; until this is tested consider the protocol
// and Tiberious to be unstable.
package v1

import (
	"time"
)

// NewTimestamp creates a new timestamp.
func NewTimestamp(t time.Time) *Timestamp {
	return &Timestamp{
		Seconds: t.Unix(),
		Nanos:   int32(t.Nanosecond()),
	}
}

// Validate a NewConversationRequest.
func (r *NewConversationRequest) Validate() error {
	// TODO: confirm the signature and timestamp data match up.
	return nil
}
