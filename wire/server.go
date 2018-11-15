package wire

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/kelleyk/go-cachecash/common"
	"github.com/sirupsen/logrus"
	"github.com/zenazn/goji/graceful"
	graceful_listener "github.com/zenazn/goji/graceful/listener"
)

type server struct {
	l *logrus.Logger

	// laddr is the local address on which the server will listen.
	laddr net.Addr

	listener *graceful_listener.T

	// resignCh is used by the conncetion goroutines to inform the server goroutine when they have exited.
	resignCh  chan int
	newConnCh chan *websocket.Conn

	// shutdownCh is closed to signal to internal goroutines that they should begin gracefully shutting down.
	shutdownCh chan struct{}
	// finishedCh is closed to signal that shutdown is complete.
	finishedCh chan struct{}

	// TODO: Do we need a channel for passing messages to/from the actual application logic?

	// This state may only be accessed by the goroutine that executes run() and cleanup().
	nextConnID  int
	connections map[int]Connection

	receiver EventReceiver
}

var (
	// TODO: Do we need to tweak any of these options?
	upgrader = websocket.Upgrader{}
)

// A Server manages WebSocket connections from one or more clients.
type Server interface {
	Start() error
	Shutdown(context.Context) error
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

var _ Server = (*server)(nil)
var _ common.StarterShutdowner = (*server)(nil)

func NewServer(l *logrus.Logger, receiver EventReceiver) (Server, error) {
	return &server{
		l:        l,
		receiver: receiver,

		resignCh:  make(chan int),
		newConnCh: make(chan *websocket.Conn),

		shutdownCh: make(chan struct{}),
		finishedCh: make(chan struct{}),
	}, nil
}

func (s *server) Start() error {
	// TODO: Prevent multiple starts?

	/*
		// XXX: Should listening be done here, or by parent code?  Doing it here prevents us from attaching this to a parent
		// mux.
		l, err := net.Listen(s.laddr.Network(), s.laddr.String())
		if err != nil {
			return fmt.Errorf("failed to bind socket: %v", err)
		}
		// XXX: Is this still the right way to do this, or are there nicer ways built into newer versions of Go?
		s.listener = graceful_listener.Wrap(l, graceful_listener.Deadline)
	*/

	ctx := context.Background() // TEMP: refactoring dust
	go s.run(ctx)
	/*
	 go s.serve()
	*/
	return nil
}

func (s *server) Shutdown(ctx context.Context) error {
	close(s.shutdownCh)

	select {
	case <-s.finishedCh:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// TODO: Log error.
		return
	}

	// Causes the server goroutine to call addNewConn().
	s.newConnCh <- c
}

func (s *server) serve() {
	mux := http.NewServeMux()
	mux.Handle("/ws", s) // XXX: What should this path be?
	httpServer := &graceful.Server{Handler: mux}

	if err := httpServer.Serve(s.listener); err != nil {
		// TODO: We get net.ErrClosing when we Close() our listener; should not treat that as an error.
		fmt.Printf("listener error: %v", err)
	}
}

// N.B.: Only the goroutine that executes run() and cleanup() is allowed to access internal state such as the
// connections map.  Other goroutines must send it messages.
func (s *server) run(ctx context.Context) {
	defer s.cleanup()

	for {
		select {
		case <-s.shutdownCh:
			return
		case c := <-s.newConnCh:
			s.addNewConn(c)
		case id := <-s.resignCh:
			delete(s.connections, id)
		}
	}
}

func (s *server) cleanup() {
	// TODO: Ask all listeners to exit; wait for their resignations.
	close(s.finishedCh)
}

func (s *server) addNewConn(c *websocket.Conn) {
	id := s.nextConnID
	s.nextConnID++

	co, err := NewConnection(context.Background(), s.l, c, func() { s.resignCh <- id })
	if err != nil {
		panic(err)
	}
	// TODO: ...

	s.receiver.OnConnect(co)

}
