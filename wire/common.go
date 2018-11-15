// Package wire contains a WebSocket-based server and client.
//
// It handles connection negotiation and management.  It is agnostic to the content of the messages passed back and
// forth; they are byte slices.  It provides connection, disconnection, and message-received events.
//
// Its public interfaces are safe for concurrent use by multiple goroutines.
package wire

import uuid "github.com/satori/go.uuid"

// An outboundItem consists of a message to be sent and information about the peer(s) that it be sent to.
type outboundItem struct {
	data   []byte
	peerID *uuid.UUID // XXX: Are we still using peer UUIDs?
}
