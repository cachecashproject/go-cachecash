package common

import (
	"net"

	"github.com/gorilla/websocket"
)

// IsNetErrClosing returns true iff err is net.errClosing: "use of closed network connection".  This helper exists
// because that error is not exported.
func IsNetErrClosing(err error) bool {
	switch err := err.(type) {
	case *net.OpError:
		if err.Err.Error() == "use of closed network connection" {
			return true
		}
	default:
	}
	return false
}

// IsErrConnRefused returns true iff the error is a network error corresponding to ECONNREFUSED being returned from a
// syscall like listen(2).
func IsErrConnRefused(err error) bool {
	if oerr, ok := err.(*net.OpError); ok {
		if oerr.Err.Error() == "connect: connection refused" {
			return true
		}
	}
	return false
}

// IsWebSocketClose returns true iff err is indicates that a WebSocket connection has been closed with one of the
// indicated status codes.
func IsWebSocketClose(err error, statusCodes ...int) bool {
	if err, ok := err.(*websocket.CloseError); ok {
		for _, code := range statusCodes {
			if err.Code == code {
				return true
			}
		}
	}
	return false
}
