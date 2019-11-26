package ranger

import "errors"

// ErrMarshalLength is returned when the marshal length does not match the generated content.
var ErrMarshalLength = errors.New("marshal length does not match generated content")

// ErrShortRead indicates to users when an unmarshaling error has occurred because we ran out of bytes.
var ErrShortRead = errors.New("short read during unmarshal")

// ErrShortWrite indicates to users when a marshal error has occurred because we ran out of bytes.
var ErrShortWrite = errors.New("short write during marshal")

// ErrTooMany is for when a length constraint was exceeded.
var ErrTooMany = errors.New("too many items during (un)marshal")

// ErrBadInterface is for when interface types (enums) are unaccounted for.
var ErrBadInterface = errors.New("bad interface type")

// ErrLeftOverBytes is from when unmarhsals leave dangling content
var ErrLeftOverBytes = errors.New("leftover bytes during unmarshal")

// ErrTooLarge is for when values exceeed their input boundaries
var ErrTooLarge = errors.New("value is too large for type")

// ErrLengthMismatch is when two lengths should match but do not.
var ErrLengthMismatch = errors.New("lengths do not match")
