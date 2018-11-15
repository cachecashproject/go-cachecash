package wire

import (
	"context"
)

// // An Endpoint manages connections to one or more peers; messages received from any of those connections are
// // merged and delivered to the application, and messages sent are sent to all of those connections.
// type Endpoint interface {
// 	// Start causes the Client or Server to begin opening connections and delivering received messages.  It must be
// 	// called exactly once; subsequent calls are programming errors.  The Client or Server will stop when ctx is done.
// 	Start(ctx context.Context) error
// 	// Wait blocks until the Client or Server is done cleaning up all of its connections or until ctx is done.  You
// 	// should make sure to cancel the Context passed to Start first to make the client begin cleaning up.  Returns ctx's
// 	// error, if any, or nil if the Client or Server finished cleaning up.
// 	Wait(ctx context.Context) error
// }

// // MessageReceiver should be implemented by the type that will handle callbacks as messages are delivered.
// // Implementations of MessageReceiver must be safe for concurrent use by multiple goroutines.
// type MessageReceiver interface {
// 	HandleMessage(peerID uuid.UUID, data []byte)
// }

// EventReceiver receives callbacks when the Client or Server connects to or disconnects from a peer.
type EventReceiver interface {
	// OnConnectDisconnect is called when the Client or Server connects to or disconnects from a peer.  'delta' will be
	// 1 if a new peer has just connected and -1 if a peer has disconnected.
	//
	// It must be safe for concurrent use by multiple goroutines.
	OnConnect(co Connection)
}

type Connection interface {
	SendMessage(msg []byte) error
	MessageCh() chan []byte
	Close(ctx context.Context) error
}

// // XXX: Formerly in package discovery
// type ServiceInstanceKey int
