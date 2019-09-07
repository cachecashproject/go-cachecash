package log

import "errors"

// ErrBundleTooLarge indicates when the bundle has exceeded the limits the logpipe server has placed.
var ErrBundleTooLarge = errors.New("log bundle is too large")

// ErrSpoolFull is sent when the spool dir is full and no more logs can be accepted.
var ErrSpoolFull = errors.New("the spool is full, please retry later")

// ErrTooManyConnections is thrown when the server is overloaded
var ErrTooManyConnections = errors.New("too many connections, please retry later")

// ErrTooMuchData indicates when too much data has been sent.
var ErrTooMuchData = errors.New("too much data sent; please retry later")
