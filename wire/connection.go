package wire

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kelleyk/go-cachecash/common"
	"github.com/sirupsen/logrus"
)

const (
	// GracefulCloseTimeout is how long the client will wait between telling a server that it'd like to end the
	// connection and forcefully closing it.
	GracefulCloseTimeout = 250 * time.Millisecond
)

// TODO: Move this elsewhere.
type connection struct {
	l *logrus.Logger

	// c is the underlying connection to the server.
	c *websocket.Conn
	// resign is called to notify the parent Server that this connection is over.
	resign func()

	// inboundCh is closed when the connection is closed (at the same time that finishedCh is).
	inboundCh  chan []byte
	outboundCh chan []byte

	// XXX: I am not sure that this should be necessary.
	outboundCloseOnce sync.Once

	finishedCh chan struct{}
}

var _ Connection = (*connection)(nil)

func NewConnection(ctx context.Context, l *logrus.Logger, c *websocket.Conn, resign func()) (*connection, error) {
	co := &connection{
		l:      l,
		c:      c,
		resign: resign,

		finishedCh: make(chan struct{}),
		inboundCh:  make(chan []byte),
		outboundCh: make(chan []byte),
	}

	go co.runTx(ctx)
	go co.runRx(ctx)
	return co, nil
}

// XXX: Not actually used.
var ErrNotConnected = errors.New("Peer is not connected")

// SendMessage transmits the message to the connected peer.  If the peer is not connected, returns ErrNotConnected.
// (XXX: Can we guarantee this?)
func (co *connection) SendMessage(data []byte) error {
	// return co.send(websocket.BinaryMessage, data)
	co.outboundCh <- data
	return nil
}

func (co *connection) MessageCh() chan []byte {
	return co.inboundCh
}

// Close begins a graceful shutdown of the connection.  If ctx expires, the connection is forcefully closed.
// XXX: Do we want to use Close() or the ctx passed to newConnection()?  I'm thinking the former might be easier.
func (co *connection) Close(ctx context.Context) error {
	// XXX: Need to implement timeout.

	co.l.Infof("%p connection.Close() - enter", co)
	co.outboundCloseOnce.Do(func() {
		close(co.outboundCh)
	})

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-co.finishedCh:
		return nil
	}
}

func (co *connection) runTx(ctx context.Context) {
	co.l.Info("(TMP) entering runTx")

	defer co.cleanupTx()

	for {
		select {
		// XXX: Why was this case missing?  What is the expected shutdown flow here?
		case <-co.finishedCh:
			return
		case <-ctx.Done():
			return
			// TODO: Need a channel for messages outbound on this connection.
		case msg, ok := <-co.outboundCh:
			if !ok {
				fmt.Printf("sending graceful close request\n")
				// outboundCh has been closed; let's send a graceful close message.
				// N.B.: status 1000 is normal.
				if err := co.send(websocket.CloseMessage, websocket.FormatCloseMessage(1000, "Be seeing you")); err != nil {
					fmt.Printf("failed to send graceful close request: %v\n", err)
				}
				return
			}
			if err := co.send(websocket.BinaryMessage, msg); err != nil {
				fmt.Printf("failed to send: %v\n", err)
			}
		}
	}
}

func (co *connection) cleanupTx() {
	co.l.Info("(TMP) entering cleanupTx")

	// TODO: We probably don't always want to try to gracefully close the connection---for example, if we are exiting
	// because the connection has already been closed.

	// TODO: Do we need to send a websocket.CloseMessage here?

	select {
	case <-co.finishedCh:
	case <-time.After(GracefulCloseTimeout):
		if err := co.c.Close(); err != nil && !common.IsNetErrClosing(err) {
			co.l.WithError(err).Error("error trying to close connection after timeout")
		}
	}
}

func (co *connection) runRx(ctx context.Context) {
	co.l.Info("(TMP) entering runRx")

	defer co.cleanupRx()

	for {
		msgType, data, err := co.c.ReadMessage()
		if err != nil {
			switch {
			case common.IsWebSocketClose(err, websocket.CloseNormalClosure, websocket.CloseGoingAway):
				co.l.Info("connection closed gracefully")
			case common.IsNetErrClosing(err):
				// If we got this particular error, another goroutine must have closed our connection for us.
				co.l.Warn("connection appears to have been uncleanly closed from our side")
			default:
				// The other side has uncleanly closed the connection; we should exit without trying to send a graceful
				// close message.
				co.l.WithError(err).Error("failed to read WebSocket message")
			}

			return
		}

		switch msgType {
		case websocket.BinaryMessage:
			co.inboundCh <- data
		case websocket.CloseMessage:
			// TODO: handle me; close gracefully
			co.l.Error("recv websock close message, but graceful close is not implemented")
			return
		default:
			co.l.Error("received a WebSocket message, but it is not a binary message")
		}
	}
}

func (co *connection) cleanupRx() {
	co.l.Info("(TMP) entering cleanupRx")

	close(co.finishedCh)
	close(co.inboundCh)
	co.resign()
}

// send actually transmits a WebSocket message.  msgType is one of the websocket.*Message constants.
func (co *connection) send(msgType int, data []byte) error {
	if co.c != nil {
		if err := co.c.WriteMessage(msgType, data); err != nil {
			return fmt.Errorf("write error: %v", err)
		}
		co.l.Debug("sent serialized message to connected server")
	} else {
		// XXX: Is this branch ever taken?
		co.l.Debug("skipping message send because underlying connection object is nil")
	}
	return nil
}
