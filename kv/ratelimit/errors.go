package ratelimit

import "errors"

// ErrTooMuchData indicates when too much data has been sent.
var ErrTooMuchData = errors.New("too much data sent; please retry later")
